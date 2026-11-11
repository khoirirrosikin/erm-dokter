package pemeriksaan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Service interface {
	DaftarPemeriksaan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterDaftarPemeriksaan) ([]Pemeriksaan, shared.PaginationMeta, error)
	DaftarPemeriksaanByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterDaftarPemeriksaan) ([]Pemeriksaan, shared.PaginationMeta, error)
	DetailPemeriksaan(ctx context.Context, id IdPemeriksaan, statusLanjut shared.StatusLanjut) (*Pemeriksaan, error)
	DaftarKesadaran(ctx context.Context) []OpsiReferensi
	SimpanPemeriksaan(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPemeriksaanRequest) (*Pemeriksaan, error)
	UpdatePemeriksaan(ctx context.Context, kodeDokter string, id IdPemeriksaan, statusLanjut shared.StatusLanjut, req UpdatePemeriksaanRequest) (*Pemeriksaan, error)
	HapusPemeriksaan(ctx context.Context, kodeDokter string, id IdPemeriksaan, statusLanjut shared.StatusLanjut) error
}

type service struct {
	repo              Repository
	rawatJalanService rawatjalan.Service
	rawatInapService  rawatinap.Service
	maxEditJam        int
	log               *logger.Logger
}

func NewService(repo Repository, rawatJalanService rawatjalan.Service, rawatInapService rawatinap.Service, maxEditJam int, log *logger.Logger) Service {
	if maxEditJam <= 0 {
		maxEditJam = 48
	}
	return &service{
		repo:              repo,
		rawatJalanService: rawatJalanService,
		rawatInapService:  rawatInapService,
		maxEditJam:        maxEditJam,
		log:               log,
	}
}

func (s *service) DaftarPemeriksaan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterDaftarPemeriksaan) ([]Pemeriksaan, shared.PaginationMeta, error) {
	rawParts := strings.Split(noRawat, ",")
	var listNoRawat []string
	for _, p := range rawParts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			listNoRawat = append(listNoRawat, trimmed)
		}
	}

	daftarPemeriksaan, totalData, err := s.repo.DaftarPemeriksaan(ctx, listNoRawat, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal query daftar pemeriksaan %s: %v", noRawat, err)
		return nil, shared.PaginationMeta{}, err
	}

	return daftarPemeriksaan, shared.NewPaginationMeta(totalData, filter.Page, filter.Limit), nil
}

func (s *service) DaftarPemeriksaanByRM(ctx context.Context, noRekamMedis string, statusLanjut shared.StatusLanjut, filter FilterDaftarPemeriksaan) ([]Pemeriksaan, shared.PaginationMeta, error) {
	riwayatKunjungan, err := s.rawatJalanService.RiwayatKunjunganPasien(ctx, noRekamMedis)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat kunjungan untuk RM %s: %v", noRekamMedis, err)
		return nil, shared.PaginationMeta{}, err
	}

	if len(riwayatKunjungan) == 0 {
		return make([]Pemeriksaan, 0), shared.NewPaginationMeta(0, filter.Page, filter.Limit), nil
	}

	listNoRawat := make([]string, len(riwayatKunjungan))
	for i, k := range riwayatKunjungan {
		listNoRawat[i] = k.NoRawat
	}

	daftarPemeriksaan, totalData, err := s.repo.DaftarPemeriksaan(ctx, listNoRawat, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal query daftar pemeriksaan by RM %s: %v", noRekamMedis, err)
		return nil, shared.PaginationMeta{}, err
	}

	return daftarPemeriksaan, shared.NewPaginationMeta(totalData, filter.Page, filter.Limit), nil
}

func (s *service) DetailPemeriksaan(ctx context.Context, id IdPemeriksaan, statusLanjut shared.StatusLanjut) (*Pemeriksaan, error) {
	pemeriksaan, err := s.repo.DetailPemeriksaan(ctx, id, statusLanjut)
	if err != nil {
		s.log.Error("Gagal query detail pemeriksaan %+v (%s): %v", id, statusLanjut, err)
		return nil, err
	}
	if pemeriksaan == nil {
		return nil, apperror.NewNotFoundError(
			"Data pemeriksaan tidak ditemukan",
			fmt.Sprintf("Data pemeriksaan tidak ditemukan untuk no_rawat '%s' tanggal '%s' jam '%s' (%s)", id.NoRawat, id.TanggalPemeriksaan, id.JamPemeriksaan, statusLanjut),
		)
	}

	return pemeriksaan, nil
}

func (s *service) DaftarKesadaran(ctx context.Context) []OpsiReferensi {
	opsi := make([]OpsiReferensi, len(ListKesadaran))
	for i, k := range ListKesadaran {
		opsi[i] = OpsiReferensi{
			Value: string(k),
			Label: string(k),
		}
	}
	return opsi
}

func (s *service) SimpanPemeriksaan(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPemeriksaanRequest) (*Pemeriksaan, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, req.NoRawat, req.TanggalPemeriksaan, req.JamPemeriksaan, statusLanjut, "disimpan"); err != nil {
		return nil, err
	}

	if err := s.repo.SimpanPemeriksaan(ctx, kodeDokter, statusLanjut, req); err != nil {
		if isDuplicateEntry(err) {
			return nil, apperror.ValidationError{
				"jam_pemeriksaan": fmt.Sprintf("Pemeriksaan pada %s %s sudah ada", req.TanggalPemeriksaan, req.JamPemeriksaan),
			}
		}
		s.log.Error("Gagal menyimpan data pemeriksaan no_rawat '%s' (%s) oleh dokter '%s': %v", req.NoRawat, statusLanjut, kodeDokter, err)
		return nil, err
	}

	idPemeriksaan := IdPemeriksaan{
		NoRawat:            req.NoRawat,
		TanggalPemeriksaan: req.TanggalPemeriksaan,
		JamPemeriksaan:     req.JamPemeriksaan,
	}

	detail, err := s.DetailPemeriksaan(ctx, idPemeriksaan, statusLanjut)
	if err != nil {
		return nil, err
	}

	s.log.Info("Berhasil menyimpan data pemeriksaan no_rawat '%s' (%s %s, %s) oleh dokter '%s'", req.NoRawat, req.TanggalPemeriksaan, req.JamPemeriksaan, statusLanjut, kodeDokter)
	return detail, nil
}

func (s *service) UpdatePemeriksaan(ctx context.Context, kodeDokter string, id IdPemeriksaan, statusLanjut shared.StatusLanjut, req UpdatePemeriksaanRequest) (*Pemeriksaan, error) {
	pemeriksaan, err := s.DetailPemeriksaan(ctx, id, statusLanjut)
	if err != nil {
		return nil, err
	}

	if err := s.validasiKepemilikanDokter(pemeriksaan, kodeDokter, "mengubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, id.NoRawat, req.TanggalPemeriksaan, req.JamPemeriksaan, statusLanjut, "diubah"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePemeriksaan(ctx, id, statusLanjut, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError(
				"Data pemeriksaan tidak ditemukan",
				fmt.Sprintf("Data pemeriksaan tidak ditemukan untuk no_rawat '%s' tanggal '%s' jam '%s' (%s)", id.NoRawat, id.TanggalPemeriksaan, id.JamPemeriksaan, statusLanjut),
			)
		}
		if isDuplicateEntry(err) {
			return nil, apperror.ValidationError{
				"jam_pemeriksaan": fmt.Sprintf("Pemeriksaan pada %s %s sudah ada", req.TanggalPemeriksaan, req.JamPemeriksaan),
			}
		}
		s.log.Error("Gagal memperbarui data pemeriksaan no_rawat '%s' (%s) oleh dokter '%s': %v", id.NoRawat, statusLanjut, kodeDokter, err)
		return nil, err
	}

	updatedId := IdPemeriksaan{
		NoRawat:            id.NoRawat,
		TanggalPemeriksaan: req.TanggalPemeriksaan,
		JamPemeriksaan:     req.JamPemeriksaan,
	}

	detail, err := s.DetailPemeriksaan(ctx, updatedId, statusLanjut)
	if err != nil {
		return nil, err
	}

	s.log.Info("Berhasil memperbarui data pemeriksaan no_rawat '%s' (%s %s -> %s %s) oleh dokter '%s'", id.NoRawat, id.TanggalPemeriksaan, id.JamPemeriksaan, req.TanggalPemeriksaan, req.JamPemeriksaan, kodeDokter)
	return detail, nil
}

func (s *service) HapusPemeriksaan(ctx context.Context, kodeDokter string, id IdPemeriksaan, statusLanjut shared.StatusLanjut) error {
	pemeriksaan, err := s.DetailPemeriksaan(ctx, id, statusLanjut)
	if err != nil {
		return err
	}

	if err := s.validasiKepemilikanDokter(pemeriksaan, kodeDokter, "menghapus"); err != nil {
		return err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, id.NoRawat, "", "", statusLanjut, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPemeriksaan(ctx, id, statusLanjut); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError(
				"Data pemeriksaan tidak ditemukan",
				fmt.Sprintf("Data pemeriksaan tidak ditemukan untuk no_rawat '%s' tanggal '%s' jam '%s' (%s)", id.NoRawat, id.TanggalPemeriksaan, id.JamPemeriksaan, statusLanjut),
			)
		}
		s.log.Error("Gagal menghapus data pemeriksaan no_rawat '%s' (%s) oleh dokter '%s': %v", id.NoRawat, statusLanjut, kodeDokter, err)
		return err
	}

	s.log.Info("Berhasil menghapus data pemeriksaan no_rawat '%s' (%s %s) oleh dokter '%s'", id.NoRawat, id.TanggalPemeriksaan, id.JamPemeriksaan, kodeDokter)
	return nil
}

func (s *service) validasiKepemilikanDokter(pemeriksaan *Pemeriksaan, kodeDokter, aksi string) error {
	if pemeriksaan.KodeDokterPetugas != kodeDokter {
		return apperror.NewForbiddenError(
			fmt.Sprintf("Data pemeriksaan ini diinput oleh dokter/petugas lain (%s)", pemeriksaan.NamaDokterPetugas),
			fmt.Sprintf("Dokter '%s' mencoba %s data pemeriksaan no_rawat '%s' (%s %s) milik '%s' (%s)",
				kodeDokter, aksi, pemeriksaan.NoRawat, pemeriksaan.TanggalPemeriksaan, pemeriksaan.JamPemeriksaan,
				pemeriksaan.KodeDokterPetugas, pemeriksaan.NamaDokterPetugas),
		)
	}
	return nil
}

func (s *service) validasiRegistrasiDanStatus(ctx context.Context, noRawat, tanggalPemeriksaan, jamPemeriksaan string, statusLanjut shared.StatusLanjut, aksi string) error {
	tanggalRegistrasi, jamRegistrasi, exists, err := s.rawatJalanService.GetWaktuRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal mengambil data registrasi no_rawat '%s': %v", noRawat, err)
		return err
	}
	if !exists {
		return apperror.NewNotFoundError("Data registrasi kunjungan pasien tidak ditemukan")
	}

	waktuRegistrasi, err := shared.ParseWaktu(tanggalRegistrasi, jamRegistrasi)
	if err != nil {
		s.log.Error("Gagal parse waktu registrasi no_rawat '%s' (%s %s): %v", noRawat, tanggalRegistrasi, jamRegistrasi, err)
		return err
	}

	if tanggalPemeriksaan != "" && jamPemeriksaan != "" {
		waktuPemeriksaan, err := shared.ParseWaktu(tanggalPemeriksaan, jamPemeriksaan)
		if err != nil {
			return apperror.NewBusinessError(err.Error())
		}

		if waktuPemeriksaan.Before(waktuRegistrasi) {
			return apperror.ValidationError{
				"tanggal_pemeriksaan": fmt.Sprintf("Waktu pemeriksaan tidak boleh sebelum registrasi (%s %s)", tanggalRegistrasi, jamRegistrasi),
			}
		}
	}

	isAktifRanap, hasRecordKamar, err := s.rawatInapService.CekStatusKamarInap(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal cek status kamar inap untuk no_rawat '%s': %v", noRawat, err)
		return err
	}

	if hasRecordKamar {
		if !isAktifRanap {
			return apperror.NewBusinessError("Pasien sudah keluar dari kamar inap")
		}
		return nil
	}

	if strings.EqualFold(string(statusLanjut), string(shared.StatusLanjutRawatInap)) {
		return apperror.NewBusinessError("Pasien belum terdaftar di kamar inap, gunakan status 'ralan'")
	}

	batasWaktu := waktuRegistrasi.Add(time.Duration(s.maxEditJam) * time.Hour)
	if time.Now().After(batasWaktu) {
		return apperror.NewForbiddenError(
			fmt.Sprintf("Pemeriksaan tidak dapat %s, melebihi batas %d jam", aksi, s.maxEditJam),
			fmt.Sprintf("Pemeriksaan no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari registrasi (%s)", noRawat, aksi, s.maxEditJam, waktuRegistrasi.Format("2006-01-02 15:04:05")),
		)
	}

	return nil
}

func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	return (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) ||
		strings.Contains(err.Error(), "1062") ||
		strings.Contains(err.Error(), "Duplicate entry")
}
