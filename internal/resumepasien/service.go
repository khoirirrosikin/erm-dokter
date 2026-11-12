package resumepasien

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
	DetailResumePasienRalan(ctx context.Context, noRawat string) (*ResumePasienRalan, error)
	RiwayatResumePasienRalanByNoRM(ctx context.Context, noRM string) ([]ResumePasienRalan, error)
	SimpanResumePasienRalan(ctx context.Context, noRawat, kodeDokter string, req SimpanResumePasienRalanRequest) (*ResumePasienRalan, error)
	UpdateResumePasienRalan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdateResumePasienRalanRequest) (*ResumePasienRalan, error)
	HapusResumePasienRalan(ctx context.Context, kodeDokterLogin, noRawat string) error

	DetailResumePasienRanap(ctx context.Context, noRawat string) (*ResumePasienRanap, error)
	RiwayatResumePasienRanapByNoRM(ctx context.Context, noRM string) ([]ResumePasienRanap, error)
	SimpanResumePasienRanap(ctx context.Context, noRawat, kodeDokter string, req SimpanResumePasienRanapRequest) (*ResumePasienRanap, error)
	UpdateResumePasienRanap(ctx context.Context, kodeDokterLogin, noRawat string, req UpdateResumePasienRanapRequest) (*ResumePasienRanap, error)
	HapusResumePasienRanap(ctx context.Context, kodeDokterLogin, noRawat string) error

	ReferensiRanap(ctx context.Context) ReferensiResumeRanap
	ReferensiRalan(ctx context.Context) ReferensiResumeRalan
}

type service struct {
	repo              Repository
	rawatJalanService rawatjalan.Service
	rawatInapService  rawatinap.Service
	maxEditJam        int
	log               *logger.Logger
}

func NewService(
	repo Repository,
	rawatJalanService rawatjalan.Service,
	rawatInapService rawatinap.Service,
	maxEditJam int,
	log *logger.Logger,
) Service {
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

func (s *service) validasiKepemilikanDokter(kodeDokterPembuat, namaDokterPembuat, kodeDokterLogin, noRawat, aksi string) error {
	if kodeDokterPembuat != kodeDokterLogin {
		s.log.Warn("Percobaan %s resume pasien no_rawat %s ditolak: dibuat oleh dokter %s (%s), dicoba oleh %s",
			aksi, noRawat, kodeDokterPembuat, namaDokterPembuat, kodeDokterLogin)
		return apperror.NewForbiddenError(fmt.Sprintf("Resume pasien dokter lain tidak dapat %s", aksi))
	}
	return nil
}

func (s *service) validasiRegistrasiDanStatus(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, aksi string) error {
	tanggalRegistrasiStr, jamRegistrasiStr, exists, err := s.rawatJalanService.GetWaktuRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal mengambil waktu registrasi no_rawat %s: %v", noRawat, err)
		return err
	}
	if !exists {
		return apperror.NewNotFoundError("Data kunjungan pasien tidak ditemukan")
	}

	isAktifRanap, hasRecordKamar, err := s.rawatInapService.CekStatusKamarInap(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal cek status kamar inap untuk no_rawat %s: %v", noRawat, err)
		return err
	}

	if strings.EqualFold(string(statusLanjut), string(shared.StatusLanjutRawatJalan)) {
		if hasRecordKamar {
			return apperror.NewBusinessError("Pasien sudah terdaftar di rawat inap. Resume medis rawat jalan hanya untuk kunjungan rawat jalan.")
		}

		waktuRegistrasi, err := shared.ParseWaktu(tanggalRegistrasiStr, jamRegistrasiStr)
		if err != nil {
			s.log.Error("Gagal parse waktu registrasi no_rawat %s (%s %s): %v", noRawat, tanggalRegistrasiStr, jamRegistrasiStr, err)
			return err
		}

		batasWaktu := waktuRegistrasi.Add(time.Duration(s.maxEditJam) * time.Hour)
		if time.Now().After(batasWaktu) {
			return apperror.NewForbiddenError(
				fmt.Sprintf("Resume pasien rawat jalan tidak dapat %s, melewati batas %d jam", aksi, s.maxEditJam),
				fmt.Sprintf("Resume pasien no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari registrasi (%s)", noRawat, aksi, s.maxEditJam, waktuRegistrasi.Format("2006-01-02 15:04:05")),
			)
		}
		return nil
	}

	if strings.EqualFold(string(statusLanjut), string(shared.StatusLanjutRawatInap)) {
		if !hasRecordKamar {
			return apperror.NewBusinessError("Pasien belum/tidak terdaftar di rawat inap. Resume medis rawat inap hanya untuk pasien rawat inap.")
		}
		if isAktifRanap {
			return nil
		}

		infoKamar, err := s.repo.GetInfoKamarInap(ctx, noRawat)
		if err != nil {
			s.log.Error("Gagal mengambil info kamar inap untuk no_rawat %s: %v", noRawat, err)
			return err
		}

		if infoKamar != nil && infoKamar.TanggalKeluar != "" {
			jamKeluar := strings.TrimSpace(infoKamar.JamKeluar)
			if jamKeluar == "" {
				jamKeluar = "00:00:00"
			}

			waktuKeluar, err := shared.ParseWaktu(infoKamar.TanggalKeluar, jamKeluar)
			if err != nil {
				s.log.Error("Gagal parse waktu keluar kamar inap no_rawat %s (%s %s): %v", noRawat, infoKamar.TanggalKeluar, jamKeluar, err)
				return err
			}

			batasWaktu := waktuKeluar.Add(time.Duration(s.maxEditJam) * time.Hour)
			if time.Now().After(batasWaktu) {
				return apperror.NewForbiddenError(
					fmt.Sprintf("Resume pasien rawat inap tidak dapat %s, melewati batas %d jam setelah pasien pulang", aksi, s.maxEditJam),
					fmt.Sprintf("Resume pasien rawat inap no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari waktu pulang (%s)", noRawat, aksi, s.maxEditJam, waktuKeluar.Format("2006-01-02 15:04:05")),
				)
			}
		}
		return nil
	}

	return nil
}
