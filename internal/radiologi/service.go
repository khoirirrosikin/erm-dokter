package radiologi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
	"erm-dokter/internal/tindakan"
)

type Service interface {
	DaftarHasilRadiologi(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, shared.PaginationMeta, error)
	DaftarHasilRadiologiByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, shared.PaginationMeta, error)
	DetailHasilRadiologi(ctx context.Context, idHasil IdHasilRadiologi) (*HasilRadiologi, error)

	SimpanPermintaanRadiologi(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPermintaanRadiologiRequest) (*DetailPermintaanRadiologi, error)
	DaftarPermintaanRadiologi(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]DetailPermintaanRadiologi, error)
	DaftarPermintaanRadiologiByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatPermintaanRadiologi) ([]DetailPermintaanRadiologi, shared.PaginationMeta, error)
	DetailPermintaanRadiologi(ctx context.Context, noPermintaan string) (*DetailPermintaanRadiologi, error)
	UpdatePermintaanRadiologi(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req SimpanPermintaanRadiologiRequest) (*DetailPermintaanRadiologi, error)
	HapusPermintaanRadiologi(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut) error
}

type service struct {
	repo                   Repository
	rawatJalanService      rawatjalan.Service
	rawatInapService       rawatinap.Service
	tindakanService        tindakan.Service
	maxEditJam             int
	kodePenjaminKunciBayar []string
	log                    *logger.Logger
}

func NewService(
	repo Repository,
	rawatJalanService rawatjalan.Service,
	rawatInapService rawatinap.Service,
	tindakanService tindakan.Service,
	maxEditJam int,
	log *logger.Logger,
	kodePenjaminKunciBayar ...string,
) Service {
	if maxEditJam <= 0 {
		maxEditJam = 48
	}
	return &service{
		repo:                   repo,
		rawatJalanService:      rawatJalanService,
		rawatInapService:       rawatInapService,
		tindakanService:        tindakanService,
		maxEditJam:             maxEditJam,
		kodePenjaminKunciBayar: shared.NormalizeKodePenjaminKunciBayar(kodePenjaminKunciBayar),
		log:                    log,
	}
}

func (s *service) DaftarHasilRadiologi(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, shared.PaginationMeta, error) {
	list, total, err := s.repo.DaftarHasilRadiologiKunjungan(ctx, noRawat, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat radiologi kunjungan %s: %v", noRawat, err)
		return nil, shared.PaginationMeta{}, err
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return list, meta, nil
}

func (s *service) DaftarHasilRadiologiByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, shared.PaginationMeta, error) {
	list, total, err := s.repo.DaftarHasilRadiologiPasien(ctx, noRM, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat radiologi pasien %s: %v", noRM, err)
		return nil, shared.PaginationMeta{}, err
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return list, meta, nil
}

func (s *service) DetailHasilRadiologi(ctx context.Context, idHasil IdHasilRadiologi) (*HasilRadiologi, error) {
	item, err := s.repo.DetailHasilRadiologi(ctx, idHasil)
	if err != nil {
		s.log.Error("Gagal mengambil detail hasil radiologi no_rawat %s kode_tindakan %s: %v", idHasil.NoRawat, idHasil.KodeTindakan, err)
		return nil, err
	}
	if item == nil {
		return nil, apperror.NewNotFoundError("Data hasil pemeriksaan radiologi tidak ditemukan")
	}

	return item, nil
}

func (s *service) SimpanPermintaanRadiologi(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPermintaanRadiologiRequest) (*DetailPermintaanRadiologi, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, req.NoRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, "disimpan"); err != nil {
		return nil, err
	}

	kodeTindakanList, err := s.validasiTindakanRadiologi(ctx, req)
	if err != nil {
		return nil, err
	}

	noPermintaan, err := s.repo.SimpanPermintaanRadiologi(ctx, req.NoRawat, kodeDokter, statusLanjut, req, kodeTindakanList)
	if err != nil {
		s.log.Error("Gagal menyimpan permintaan radiologi untuk no_rawat %s: %v", req.NoRawat, err)
		return nil, err
	}

	detail, err := s.repo.DetailPermintaanRadiologi(ctx, noPermintaan)
	if err != nil {
		s.log.Error("Gagal mengambil detail setelah simpan permintaan radiologi no_permintaan %s: %v", noPermintaan, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan permintaan radiologi no_permintaan '%s' untuk no_rawat '%s' (%s %s, %s) oleh dokter '%s'", noPermintaan, req.NoRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, kodeDokter)
	return detail, nil
}

func (s *service) DaftarPermintaanRadiologi(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]DetailPermintaanRadiologi, error) {
	list, err := s.repo.DaftarPermintaanRadiologiKunjungan(ctx, noRawat, statusLanjut)
	if err != nil {
		s.log.Error("Gagal mengambil daftar permintaan radiologi no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	return list, nil
}

func (s *service) DaftarPermintaanRadiologiByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatPermintaanRadiologi) ([]DetailPermintaanRadiologi, shared.PaginationMeta, error) {
	list, total, err := s.repo.DaftarPermintaanRadiologiPasien(ctx, noRM, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat permintaan radiologi pasien no_rm %s: %v", noRM, err)
		return nil, shared.PaginationMeta{}, err
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return list, meta, nil
}

func (s *service) DetailPermintaanRadiologi(ctx context.Context, noPermintaan string) (*DetailPermintaanRadiologi, error) {
	detail, err := s.repo.DetailPermintaanRadiologi(ctx, noPermintaan)
	if err != nil {
		s.log.Error("Gagal mengambil detail permintaan radiologi no_permintaan %s: %v", noPermintaan, err)
		return nil, err
	}
	if detail == nil {
		return nil, apperror.NewNotFoundError(
			"Data permintaan radiologi tidak ditemukan",
			fmt.Sprintf("Data permintaan radiologi no_permintaan '%s' tidak ditemukan di database", noPermintaan),
		)
	}

	return detail, nil
}

func (s *service) UpdatePermintaanRadiologi(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req SimpanPermintaanRadiologiRequest) (*DetailPermintaanRadiologi, error) {
	detail, err := s.DetailPermintaanRadiologi(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	if err := s.validasiAksesDanStatusPermintaan(detail, noRawat, kodeDokter, statusLanjut, "diubah"); err != nil {
		return nil, err
	}

	orderStatus := statusLanjut
	if orderStatus == "" {
		if strings.EqualFold(detail.Status, string(shared.StatusLanjutRawatInap)) {
			orderStatus = shared.StatusLanjutRawatInap
		} else {
			orderStatus = shared.StatusLanjutRawatJalan
		}
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPermintaan, req.JamPermintaan, orderStatus, "diubah"); err != nil {
		return nil, err
	}

	kodeTindakanList, err := s.validasiTindakanRadiologi(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePermintaanRadiologi(ctx, noPermintaan, req, kodeTindakanList); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError(
				"Data permintaan radiologi tidak ditemukan",
				fmt.Sprintf("Data permintaan radiologi no_permintaan '%s' tidak ditemukan di database", noPermintaan),
			)
		}
		s.log.Error("Gagal memperbarui permintaan radiologi no_permintaan '%s' untuk no_rawat '%s' (%s) oleh dokter '%s': %v", noPermintaan, noRawat, orderStatus, kodeDokter, err)
		return nil, err
	}

	updatedDetail, err := s.DetailPermintaanRadiologi(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	s.log.Info("Berhasil memperbarui permintaan radiologi no_permintaan '%s' untuk no_rawat '%s' (%s %s, %s) oleh dokter '%s'", noPermintaan, noRawat, req.TanggalPermintaan, req.JamPermintaan, orderStatus, kodeDokter)
	return updatedDetail, nil
}

func (s *service) HapusPermintaanRadiologi(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut) error {
	detail, err := s.DetailPermintaanRadiologi(ctx, noPermintaan)
	if err != nil {
		return err
	}

	if err := s.validasiAksesDanStatusPermintaan(detail, noRawat, kodeDokter, statusLanjut, "dihapus"); err != nil {
		return err
	}

	orderStatus := statusLanjut
	if orderStatus == "" {
		if strings.EqualFold(detail.Status, string(shared.StatusLanjutRawatInap)) {
			orderStatus = shared.StatusLanjutRawatInap
		} else {
			orderStatus = shared.StatusLanjutRawatJalan
		}
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", "", orderStatus, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPermintaanRadiologi(ctx, noPermintaan); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError(
				"Data permintaan radiologi tidak ditemukan",
				fmt.Sprintf("Data permintaan radiologi no_permintaan '%s' tidak ditemukan di database", noPermintaan),
			)
		}
		s.log.Error("Gagal menghapus permintaan radiologi no_permintaan '%s' untuk no_rawat '%s' oleh dokter '%s': %v", noPermintaan, noRawat, kodeDokter, err)
		return err
	}

	s.log.Info("Berhasil menghapus permintaan radiologi no_permintaan '%s' untuk no_rawat '%s' (%s) oleh dokter '%s'", noPermintaan, noRawat, orderStatus, kodeDokter)
	return nil
}

func (s *service) validasiAksesDanStatusPermintaan(detail *DetailPermintaanRadiologi, noRawat, kodeDokter string, statusLanjut shared.StatusLanjut, action string) error {
	if detail.NoRawat != noRawat {
		return apperror.NewNotFoundError(
			"Permintaan radiologi tidak ditemukan",
			fmt.Sprintf("Data permintaan radiologi no_permintaan '%s' tidak ditemukan untuk no_rawat '%s'", detail.NoPermintaan, noRawat),
		)
	}

	if statusLanjut != "" && statusLanjut != shared.StatusLanjutSemua && !strings.EqualFold(strings.TrimSpace(detail.Status), string(statusLanjut)) {
		return apperror.NewNotFoundError(
			"Permintaan radiologi tidak ditemukan",
			fmt.Sprintf("Data permintaan radiologi no_permintaan '%s' berstatus '%s', tidak sesuai dengan status '%s'", detail.NoPermintaan, detail.Status, statusLanjut),
		)
	}

	if detail.DokterPerujuk.KodeDokter != kodeDokter {
		s.log.Warn("Percobaan %s permintaan radiologi no_order %s oleh dokter %s ditolak: dibuat oleh %s (%s)", action, detail.NoPermintaan, kodeDokter, detail.DokterPerujuk.KodeDokter, detail.DokterPerujuk.NamaDokter)
		return apperror.NewForbiddenError(fmt.Sprintf("Permintaan radiologi dokter lain tidak dapat %s", action))
	}

	isSampelDiambil := shared.IsNonZeroDate(detail.TanggalSampel) || shared.IsNonZeroTime(detail.JamSampel)
	isHasilKeluar := shared.IsNonZeroDate(detail.TanggalHasil) || shared.IsNonZeroTime(detail.JamHasil)
	if isSampelDiambil || isHasilKeluar {
		s.log.Warn("Percobaan %s permintaan radiologi no_order %s ditolak: sudah diproses petugas (sampel: %s %s, hasil: %s %s)", action, detail.NoPermintaan, detail.TanggalSampel, detail.JamSampel, detail.TanggalHasil, detail.JamHasil)
		return apperror.NewForbiddenError(fmt.Sprintf("Permintaan radiologi sudah diproses oleh petugas, tidak dapat %s", action))
	}

	for _, p := range detail.Pemeriksaan {
		if strings.EqualFold(p.StatusBayar, "Sudah") {
			s.log.Warn("Percobaan %s permintaan radiologi no_order %s ditolak: pemeriksaan '%s' telah dibayar", action, detail.NoPermintaan, p.NamaTindakan)
			return apperror.NewForbiddenError(fmt.Sprintf("Pemeriksaan radiologi sudah dibayar, tidak dapat %s", action))
		}
	}

	return nil
}

func (s *service) validasiTindakanRadiologi(ctx context.Context, req SimpanPermintaanRadiologiRequest) ([]string, error) {
	dupErrs := make(apperror.ValidationError)
	seenTindakan := make(map[string]bool, len(req.Pemeriksaan))
	var kodeTindakanList []string
	for i, item := range req.Pemeriksaan {
		if item.KodeTindakan != "" {
			if seenTindakan[item.KodeTindakan] {
				dupErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: Tindakan radiologi tidak boleh duplikat", i+1)
				continue
			}
			seenTindakan[item.KodeTindakan] = true
		}
		kodeTindakanList = append(kodeTindakanList, item.KodeTindakan)
	}
	if len(dupErrs) > 0 {
		return nil, dupErrs
	}

	if len(kodeTindakanList) == 0 {
		return nil, nil
	}

	adaMap, err := s.tindakanService.CekKeberadaanTindakanRadiologi(ctx, kodeTindakanList)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan tindakan radiologi: %v", err)
		return nil, err
	}

	valErrs := make(apperror.ValidationError)
	for i, item := range req.Pemeriksaan {
		if !adaMap[item.KodeTindakan] {
			valErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = "Tindakan radiologi tidak ditemukan"
		}
	}

	if len(valErrs) > 0 {
		s.log.Warn("Validasi keberadaan tindakan radiologi gagal untuk no_rawat %s: %+v", req.NoRawat, valErrs)
		return nil, valErrs
	}

	return kodeTindakanList, nil
}

func (s *service) validasiRegistrasiDanStatus(ctx context.Context, noRawat, tanggalPermintaan, jamPermintaan string, statusLanjut shared.StatusLanjut, aksi string) error {
	infoRegistrasi, err := s.rawatJalanService.GetInfoRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal mengambil data registrasi no_rawat '%s': %v", noRawat, err)
		return err
	}

	if shared.IsPenjaminTerkunciSudahBayar(infoRegistrasi.StatusBayar, infoRegistrasi.KodePenjamin, s.kodePenjaminKunciBayar) {
		return apperror.NewBusinessError(shared.PesanTerkunciSudahBayar(infoRegistrasi.KodePenjamin, "permintaan radiologi", aksi))
	}

	tanggalRegistrasi := infoRegistrasi.TanggalRegistrasi
	jamRegistrasi := infoRegistrasi.JamRegistrasi

	waktuRegistrasi, err := shared.ParseWaktu(tanggalRegistrasi, jamRegistrasi)
	if err != nil {
		s.log.Error("Gagal parse waktu registrasi no_rawat '%s' (%s %s): %v", noRawat, tanggalRegistrasi, jamRegistrasi, err)
		return err
	}

	var waktuPermintaan time.Time
	hasWaktuInput := tanggalPermintaan != "" && jamPermintaan != ""
	if hasWaktuInput {
		waktuPermintaan, err = shared.ParseWaktu(tanggalPermintaan, jamPermintaan)
		if err != nil {
			return apperror.NewBusinessError(err.Error())
		}

		if waktuPermintaan.Before(waktuRegistrasi) {
			return apperror.ValidationError{
				"tanggal_permintaan": fmt.Sprintf("Waktu permintaan radiologi (%s %s) tidak boleh sebelum waktu registrasi (%s %s)", tanggalPermintaan, jamPermintaan, tanggalRegistrasi, jamRegistrasi),
			}
		}
	}

	if strings.EqualFold(string(statusLanjut), string(shared.StatusLanjutRawatInap)) {
		isAktifRanap, hasRecordKamar, err := s.rawatInapService.CekStatusKamarInap(ctx, noRawat)
		if err != nil {
			s.log.Error("Gagal cek status kamar inap untuk no_rawat '%s': %v", noRawat, err)
			return err
		}
		if !hasRecordKamar {
			return apperror.NewBusinessError("Pasien belum terdaftar di kamar inap, gunakan status 'ralan'")
		}
		if isAktifRanap {
			return nil
		}

		tglKeluar, jamKeluar, err := s.rawatInapService.GetWaktuKeluarTerakhir(ctx, noRawat)
		if err != nil {
			s.log.Error("Gagal mengambil waktu keluar kamar inap untuk no_rawat '%s': %v", noRawat, err)
			return err
		}
		if tglKeluar == "" {
			return apperror.NewBusinessError("Pasien sudah keluar dari kamar inap")
		}

		jamKeluarNorm := strings.TrimSpace(jamKeluar)
		if jamKeluarNorm == "" {
			jamKeluarNorm = "23:59:59"
		}
		waktuKeluar, err := shared.ParseWaktu(tglKeluar, jamKeluarNorm)
		if err != nil {
			s.log.Error("Gagal parse waktu keluar kamar inap no_rawat '%s' (%s %s): %v", noRawat, tglKeluar, jamKeluarNorm, err)
			return err
		}

		batasWaktu := waktuKeluar.Add(time.Duration(s.maxEditJam) * time.Hour)
		if time.Now().After(batasWaktu) {
			return apperror.NewForbiddenError(
				fmt.Sprintf("Permintaan radiologi tidak dapat %s, melewati batas %d jam setelah pasien keluar kamar inap", aksi, s.maxEditJam),
				fmt.Sprintf("Permintaan radiologi no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari waktu keluar kamar inap (%s)", noRawat, aksi, s.maxEditJam, waktuKeluar.Format("2006-01-02 15:04:05")),
			)
		}

		if hasWaktuInput && waktuPermintaan.After(waktuKeluar) {
			return apperror.ValidationError{
				"tanggal_permintaan": fmt.Sprintf("Waktu permintaan radiologi (%s %s) tidak boleh melebihi waktu pulang kamar inap (%s %s)", tanggalPermintaan, jamPermintaan, tglKeluar, jamKeluarNorm),
			}
		}

		return nil
	}

	batasWaktu := waktuRegistrasi.Add(time.Duration(s.maxEditJam) * time.Hour)
	if time.Now().After(batasWaktu) {
		return apperror.NewForbiddenError(
			fmt.Sprintf("Permintaan radiologi tidak dapat %s, melewati batas %d jam", aksi, s.maxEditJam),
			fmt.Sprintf("Permintaan radiologi no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari registrasi (%s)", noRawat, aksi, s.maxEditJam, waktuRegistrasi.Format("2006-01-02 15:04:05")),
		)
	}

	return nil
}
