package penilaianmedis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Service interface {
	Referensi(ctx context.Context) ReferensiPenilaianMedis

	DetailPenilaianMedisRalan(ctx context.Context, noRawat string) (*PenilaianMedisRalan, error)
	RiwayatPenilaianMedisRalanByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalan, error)
	SimpanPenilaianMedisRalan(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRalanRequest) (*PenilaianMedisRalan, error)
	UpdatePenilaianMedisRalan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRalanRequest) (*PenilaianMedisRalan, error)
	HapusPenilaianMedisRalan(ctx context.Context, kodeDokterLogin, noRawat string) error

	DetailPenilaianMedisIGD(ctx context.Context, noRawat string) (*PenilaianMedisIGD, error)
	RiwayatPenilaianMedisIGDByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisIGD, error)
	SimpanPenilaianMedisIGD(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisIGDRequest) (*PenilaianMedisIGD, error)
	UpdatePenilaianMedisIGD(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisIGDRequest) (*PenilaianMedisIGD, error)
	HapusPenilaianMedisIGD(ctx context.Context, kodeDokterLogin, noRawat string) error

	DetailPenilaianMedisRanap(ctx context.Context, noRawat string) (*PenilaianMedisRanap, error)
	RiwayatPenilaianMedisRanapByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanap, error)
	SimpanPenilaianMedisRanap(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRanapRequest) (*PenilaianMedisRanap, error)
	UpdatePenilaianMedisRanap(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRanapRequest) (*PenilaianMedisRanap, error)
	HapusPenilaianMedisRanap(ctx context.Context, kodeDokterLogin, noRawat string) error

	DetailPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRalanKandungan, error)
	RiwayatPenilaianMedisRalanKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalanKandungan, error)
	SimpanPenilaianMedisRalanKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRalanKandunganRequest) (*PenilaianMedisRalanKandungan, error)
	UpdatePenilaianMedisRalanKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRalanKandunganRequest) (*PenilaianMedisRalanKandungan, error)
	HapusPenilaianMedisRalanKandungan(ctx context.Context, kodeDokterLogin, noRawat string) error

	DetailPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRanapKandungan, error)
	RiwayatPenilaianMedisRanapKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanapKandungan, error)
	SimpanPenilaianMedisRanapKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRanapKandunganRequest) (*PenilaianMedisRanapKandungan, error)
	UpdatePenilaianMedisRanapKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRanapKandunganRequest) (*PenilaianMedisRanapKandungan, error)
	HapusPenilaianMedisRanapKandungan(ctx context.Context, kodeDokterLogin, noRawat string) error
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

func (s *service) Referensi(ctx context.Context) ReferensiPenilaianMedis {
	return GetReferensiPenilaianMedis()
}

func (s *service) validasiKepemilikanDokter(kodeDokterPembuat, namaDokterPembuat, kodeDokterLogin, noRawat, aksi string) error {
	if kodeDokterPembuat != kodeDokterLogin {
		s.log.Warn("Percobaan %s penilaian medis no_rawat %s ditolak: dibuat oleh dokter %s (%s), dicoba oleh %s",
			aksi, noRawat, kodeDokterPembuat, namaDokterPembuat, kodeDokterLogin)
		return apperror.NewForbiddenError(fmt.Sprintf("Penilaian medis dokter lain tidak dapat %s", aksi))
	}
	return nil
}

func (s *service) validasiRegistrasiDanStatus(ctx context.Context, noRawat, tglPenilaian string, statusLanjut shared.StatusLanjut, aksi string) error {
	tanggalRegistrasi, jamRegistrasi, exists, err := s.rawatJalanService.GetWaktuRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal mengambil data registrasi no_rawat '%s': %v", noRawat, err)
		return err
	}
	if !exists {
		return apperror.NewNotFoundError("Data kunjungan pasien tidak ditemukan")
	}

	waktuRegistrasi, err := shared.ParseWaktu(tanggalRegistrasi, jamRegistrasi)
	if err != nil {
		s.log.Error("Gagal parse waktu registrasi no_rawat '%s' (%s %s): %v", noRawat, tanggalRegistrasi, jamRegistrasi, err)
		return err
	}

	var waktuPenilaian time.Time
	var hasWaktuInput bool
	if tglPenilaian != "" {
		parsedWaktu, errPer := time.ParseInLocation("2006-01-02 15:04:05", tglPenilaian, time.Local)
		if errPer == nil {
			waktuPenilaian = parsedWaktu
			hasWaktuInput = true
			if waktuPenilaian.Before(waktuRegistrasi) {
				errs := apperror.ValidationError{
					"tanggal_penilaian": fmt.Sprintf("Waktu penilaian medis (%s) tidak boleh sebelum registrasi (%s %s)", tglPenilaian, tanggalRegistrasi, jamRegistrasi),
				}
				s.log.Warn("Validasi waktu penilaian medis gagal untuk no_rawat %s: %+v", noRawat, errs)
				return errs
			}
		}
	}

	isAktifRanap, hasRecordKamar, err := s.rawatInapService.CekStatusKamarInap(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal cek status kamar inap untuk no_rawat '%s': %v", noRawat, err)
		return err
	}

	if hasRecordKamar {
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
				fmt.Sprintf("Penilaian medis tidak dapat %s, melewati batas %d jam setelah pasien keluar kamar inap", aksi, s.maxEditJam),
				fmt.Sprintf("Penilaian medis no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari waktu keluar kamar inap (%s)", noRawat, aksi, s.maxEditJam, waktuKeluar.Format("2006-01-02 15:04:05")),
			)
		}

		if hasWaktuInput && waktuPenilaian.After(waktuKeluar) {
			errs := apperror.ValidationError{
				"tanggal_penilaian": fmt.Sprintf("Waktu penilaian medis (%s) tidak boleh melebihi waktu pulang kamar inap (%s %s)", tglPenilaian, tglKeluar, jamKeluarNorm),
			}
			s.log.Warn("Validasi waktu penilaian medis pasca-pulang gagal untuk no_rawat %s: %+v", noRawat, errs)
			return errs
		}

		return nil
	}

	if strings.EqualFold(string(statusLanjut), string(shared.StatusLanjutRawatInap)) {
		return apperror.NewBusinessError("Pasien belum terdaftar di kamar inap, gunakan status 'ralan'")
	}

	batasWaktu := waktuRegistrasi.Add(time.Duration(s.maxEditJam) * time.Hour)
	if time.Now().After(batasWaktu) {
		return apperror.NewForbiddenError(
			fmt.Sprintf("Penilaian medis tidak dapat %s, melewati batas %d jam", aksi, s.maxEditJam),
			fmt.Sprintf("Penilaian medis no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari registrasi (%s)", noRawat, aksi, s.maxEditJam, waktuRegistrasi.Format("2006-01-02 15:04:05")),
		)
	}

	return nil
}
