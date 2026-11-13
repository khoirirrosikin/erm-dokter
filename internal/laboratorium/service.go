package laboratorium

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
	"erm-dokter/internal/tindakan"
)

type Service interface {
	GetRiwayatLabKunjungan(ctx context.Context, kategori shared.KategoriLab, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) (*HasilLaboratoriumKunjungan, shared.PaginationMeta, error)
	GetRiwayatLabPasien(ctx context.Context, kategori shared.KategoriLab, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, shared.PaginationMeta, error)
	GetDetailHasilLab(ctx context.Context, kategori shared.KategoriLab, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa string) (*HasilLaboratorium, error)

	SimpanPermintaanLabPK(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPKRequest) (*DetailPermintaanLabPK, error)
	GetDaftarPermintaanLabPK(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]PermintaanLabPK, error)
	GetRiwayatPermintaanLabPKByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]PermintaanLabPK, shared.PaginationMeta, error)
	GetDetailPermintaanLabPK(ctx context.Context, noPermintaan string) (*DetailPermintaanLabPK, error)
	UpdatePermintaanLabPK(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPKRequest) (*DetailPermintaanLabPK, error)
	HapusPermintaanLabPK(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error

	SimpanPermintaanLabPA(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPARequest) (*DetailPermintaanLabPA, error)
	GetDaftarPermintaanLabPA(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]PermintaanLabPA, error)
	GetRiwayatPermintaanLabPAByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]PermintaanLabPA, shared.PaginationMeta, error)
	GetDetailPermintaanLabPA(ctx context.Context, noPermintaan string) (*DetailPermintaanLabPA, error)
	UpdatePermintaanLabPA(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPARequest) (*DetailPermintaanLabPA, error)
	HapusPermintaanLabPA(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error

	SimpanPermintaanLabMB(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabMBRequest) (*DetailPermintaanLabMB, error)
	GetDaftarPermintaanLabMB(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]PermintaanLabMB, error)
	GetRiwayatPermintaanLabMBByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]PermintaanLabMB, shared.PaginationMeta, error)
	GetDetailPermintaanLabMB(ctx context.Context, noPermintaan string) (*DetailPermintaanLabMB, error)
	UpdatePermintaanLabMB(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabMBRequest) (*DetailPermintaanLabMB, error)
	HapusPermintaanLabMB(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error
}

type service struct {
	repo                   Repository
	berkasSvc              berkasdigital.Service
	rawatJalanService      rawatjalan.Service
	rawatInapService       rawatinap.Service
	tindakanService        tindakan.Service
	maxEditJam             int
	kodePenjaminKunciBayar []string
	urlBerkasDigital       string
	kodeBerkasPK           []string
	kodeBerkasPA           []string
	kodeBerkasMB           []string
	log                    *logger.Logger
}

func NewService(
	repo Repository,
	berkasSvc berkasdigital.Service,
	rawatJalanService rawatjalan.Service,
	rawatInapService rawatinap.Service,
	tindakanService tindakan.Service,
	maxEditJam int,
	urlBerkasDigital string,
	kodeBerkasPK, kodeBerkasPA, kodeBerkasMB []string,
	log *logger.Logger,
	kodePenjaminKunciBayar ...string,
) Service {
	if maxEditJam <= 0 {
		maxEditJam = 48
	}
	baseURL := ""
	if urlBerkasDigital != "" {
		baseURL = strings.TrimRight(urlBerkasDigital, "/") + "/"
	}
	return &service{
		repo:                   repo,
		berkasSvc:              berkasSvc,
		rawatJalanService:      rawatJalanService,
		rawatInapService:       rawatInapService,
		tindakanService:        tindakanService,
		maxEditJam:             maxEditJam,
		kodePenjaminKunciBayar: shared.NormalizeKodePenjaminKunciBayar(kodePenjaminKunciBayar),
		urlBerkasDigital:       baseURL,
		kodeBerkasPK:           kodeBerkasPK,
		kodeBerkasPA:           kodeBerkasPA,
		kodeBerkasMB:           kodeBerkasMB,
		log:                    log,
	}
}

func (s *service) GetRiwayatLabKunjungan(ctx context.Context, kategori shared.KategoriLab, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) (*HasilLaboratoriumKunjungan, shared.PaginationMeta, error) {
	var items []HasilLaboratorium
	var total int
	var kodeBerkas []string
	var err error

	switch kategori {
	case shared.KategoriLabPK:
		items, total, err = s.repo.DaftarHasilLabPK(ctx, noRawat, statusLanjut, filter)
		kodeBerkas = s.kodeBerkasPK
	case shared.KategoriLabPA:
		items, total, err = s.repo.DaftarHasilLabPA(ctx, noRawat, statusLanjut, filter)
		kodeBerkas = s.kodeBerkasPA
	case shared.KategoriLabMB:
		items, total, err = s.repo.DaftarHasilLabMB(ctx, noRawat, statusLanjut, filter)
		kodeBerkas = s.kodeBerkasMB
	}

	if err != nil {
		s.log.Error("gagal mengambil riwayat lab kunjungan %s kategori %s: %v", noRawat, kategori, err)
		return nil, shared.PaginationMeta{}, err
	}

	berkasList := s.fetchBerkasDigitalKunjungan(ctx, noRawat, kodeBerkas)

	kunjunganData := &HasilLaboratoriumKunjungan{
		HasilPemeriksaan: items,
		BerkasDigital:    berkasList,
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return kunjunganData, meta, nil
}

func (s *service) GetRiwayatLabPasien(ctx context.Context, kategori shared.KategoriLab, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, shared.PaginationMeta, error) {
	var items []HasilLaboratorium
	var total int
	var err error

	switch kategori {
	case shared.KategoriLabPK:
		items, total, err = s.repo.DaftarHasilLabPKByRM(ctx, noRM, statusLanjut, filter)
	case shared.KategoriLabPA:
		items, total, err = s.repo.DaftarHasilLabPAByRM(ctx, noRM, statusLanjut, filter)
	case shared.KategoriLabMB:
		items, total, err = s.repo.DaftarHasilLabMBByRM(ctx, noRM, statusLanjut, filter)
	}

	if err != nil {
		s.log.Error("gagal mengambil riwayat lab pasien %s kategori %s: %v", noRM, kategori, err)
		return nil, shared.PaginationMeta{}, err
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return items, meta, nil
}

func (s *service) GetDetailHasilLab(ctx context.Context, kategori shared.KategoriLab, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa string) (*HasilLaboratorium, error) {
	var item *HasilLaboratorium
	var err error

	switch kategori {
	case shared.KategoriLabPK:
		item, err = s.repo.DetailHasilLabPK(ctx, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa)
	case shared.KategoriLabPA:
		item, err = s.repo.DetailHasilLabPA(ctx, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa)
	case shared.KategoriLabMB:
		item, err = s.repo.DetailHasilLabMB(ctx, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data hasil laboratorium tidak ditemukan")
		}
		s.log.Error("gagal mengambil detail hasil lab %s %s: %v", noRawat, kodeTindakan, err)
		return nil, err
	}

	return item, nil
}

func (s *service) fetchBerkasDigitalKunjungan(ctx context.Context, noRawat string, kodeBerkas []string) []berkasdigital.BerkasDigital {
	if len(kodeBerkas) == 0 {
		return make([]berkasdigital.BerkasDigital, 0)
	}

	berkasList, err := s.berkasSvc.GetBerkasByNoRawat(ctx, noRawat, kodeBerkas)
	if err != nil || len(berkasList) == 0 {
		return make([]berkasdigital.BerkasDigital, 0)
	}

	berkasDigital := make([]berkasdigital.BerkasDigital, 0, len(berkasList))
	for _, b := range berkasList {
		if b.LokasiFile == "" {
			continue
		}
		fullURL := s.berkasSvc.BuildFullURL(b.LokasiFile)
		if fullURL == "" {
			continue
		}
		berkasDigital = append(berkasDigital, berkasdigital.BerkasDigital{
			Kode:       b.Kode,
			NamaBerkas: b.NamaBerkas,
			IdBerkas:   fullURL,
		})
	}
	return berkasDigital
}
func (s *service) validasiAksesDanStatusPermintaanLab(header PermintaanLabHeader, noRawat, kodeDokter string, statusLanjut shared.StatusLanjut, aksi string) error {
	if header.NoRawat != noRawat {
		return apperror.NewNotFoundError(
			"Permintaan laboratorium tidak ditemukan",
			fmt.Sprintf("Data permintaan laboratorium no_permintaan '%s' tidak ditemukan untuk no_rawat '%s'", header.NoPermintaan, noRawat),
		)
	}

	if statusLanjut != "Semua" && statusLanjut != "" && !strings.EqualFold(header.Status, string(statusLanjut)) {
		return apperror.NewNotFoundError("Data permintaan laboratorium tidak ditemukan")
	}

	if header.KodeDokterPerujuk != kodeDokter {
		s.log.Warn("Percobaan %s permintaan laboratorium %s oleh dokter %s ditolak: dibuat oleh %s (%s)", aksi, header.NoPermintaan, kodeDokter, header.KodeDokterPerujuk, header.NamaDokterPerujuk)
		return apperror.NewForbiddenError(fmt.Sprintf("Permintaan laboratorium dokter lain tidak dapat %s", aksi))
	}

	isSampelDiambil := shared.IsNonZeroDate(header.TanggalSampel) || shared.IsNonZeroTime(header.JamSampel)
	isHasilKeluar := shared.IsNonZeroDate(header.TanggalHasil) || shared.IsNonZeroTime(header.JamHasil)

	if isSampelDiambil || isHasilKeluar {
		s.log.Warn("Percobaan %s permintaan laboratorium %s ditolak: sudah diproses petugas (sampel: %s %s, hasil: %s %s)", aksi, header.NoPermintaan, header.TanggalSampel, header.JamSampel, header.TanggalHasil, header.JamHasil)
		return apperror.NewForbiddenError(fmt.Sprintf("Permintaan laboratorium sudah diproses oleh petugas, tidak dapat %s", aksi))
	}

	return nil
}

func (s *service) cekStatusBayarItemLab(noPermintaan, aksi, namaItem, statusBayar string) error {
	if strings.EqualFold(strings.TrimSpace(statusBayar), "Sudah") {
		s.log.Warn("Percobaan %s permintaan laboratorium %s ditolak: pemeriksaan '%s' telah dibayar", aksi, noPermintaan, namaItem)
		return apperror.NewForbiddenError(fmt.Sprintf("Pemeriksaan laboratorium sudah dibayar, tidak dapat %s", aksi))
	}
	return nil
}

func (s *service) validasiStatusBayarPK(noPermintaan, aksi string, items []PemeriksaanLabPKItem) error {
	for _, p := range items {
		if err := s.cekStatusBayarItemLab(noPermintaan, aksi, p.NamaTindakan, p.StatusBayar); err != nil {
			return err
		}
		for _, t := range p.DetailTemplate {
			if err := s.cekStatusBayarItemLab(noPermintaan, aksi, t.NamaPemeriksaan, t.StatusBayar); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) validasiStatusBayarPA(noPermintaan, aksi string, items []PemeriksaanLabPAItem) error {
	for _, p := range items {
		if err := s.cekStatusBayarItemLab(noPermintaan, aksi, p.NamaTindakan, p.StatusBayar); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) validasiStatusBayarMB(noPermintaan, aksi string, items []PemeriksaanLabMBItem) error {
	for _, p := range items {
		if err := s.cekStatusBayarItemLab(noPermintaan, aksi, p.NamaTindakan, p.StatusBayar); err != nil {
			return err
		}
		for _, t := range p.DetailTemplate {
			if err := s.cekStatusBayarItemLab(noPermintaan, aksi, t.NamaPemeriksaan, t.StatusBayar); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) validasiRegistrasiDanStatus(ctx context.Context, noRawat, tanggalPermintaan, jamPermintaan string, statusLanjut shared.StatusLanjut, aksi string) error {
	infoRegistrasi, err := s.rawatJalanService.GetInfoRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal mengambil data registrasi no_rawat '%s': %v", noRawat, err)
		return err
	}

	if shared.IsPenjaminTerkunciSudahBayar(infoRegistrasi.StatusBayar, infoRegistrasi.KodePenjamin, s.kodePenjaminKunciBayar) {
		return apperror.NewBusinessError(shared.PesanTerkunciSudahBayar(infoRegistrasi.KodePenjamin, "permintaan laboratorium", aksi))
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
				"tanggal_permintaan": fmt.Sprintf("Waktu permintaan laboratorium (%s %s) tidak boleh sebelum waktu registrasi (%s %s)", tanggalPermintaan, jamPermintaan, tanggalRegistrasi, jamRegistrasi),
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
				fmt.Sprintf("Permintaan laboratorium tidak dapat %s, melewati batas %d jam setelah pasien keluar kamar inap", aksi, s.maxEditJam),
				fmt.Sprintf("Permintaan laboratorium no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari waktu keluar kamar inap (%s)", noRawat, aksi, s.maxEditJam, waktuKeluar.Format("2006-01-02 15:04:05")),
			)
		}

		if hasWaktuInput && waktuPermintaan.After(waktuKeluar) {
			return apperror.ValidationError{
				"tanggal_permintaan": fmt.Sprintf("Waktu permintaan laboratorium (%s %s) tidak boleh melebihi waktu pulang kamar inap (%s %s)", tanggalPermintaan, jamPermintaan, tglKeluar, jamKeluarNorm),
			}
		}

		return nil
	}

	batasWaktu := waktuRegistrasi.Add(time.Duration(s.maxEditJam) * time.Hour)
	if time.Now().After(batasWaktu) {
		return apperror.NewForbiddenError(
			fmt.Sprintf("Permintaan laboratorium tidak dapat %s, melewati batas %d jam", aksi, s.maxEditJam),
			fmt.Sprintf("Permintaan laboratorium no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari registrasi (%s)", noRawat, aksi, s.maxEditJam, waktuRegistrasi.Format("2006-01-02 15:04:05")),
		)
	}

	return nil
}
