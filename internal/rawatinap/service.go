package rawatinap

import (
	"context"
	"strings"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Service interface {
	CekStatusKamarInap(ctx context.Context, noRawat string) (isKamarAktif bool, hasRecordKamar bool, err error)
	GetKelasRawat(ctx context.Context, noRawat string) (string, error)
	DaftarPasienRawatInap(ctx context.Context, kodeDokterLogin string, filter FilterPasienRawatInap) ([]KunjunganRawatInap, shared.PaginationMeta, error)
	DaftarStatusPulang(ctx context.Context) []OpsiReferensi
	DetailPasienRawatInap(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*KunjunganRawatInap, error)
}

type service struct {
	repo Repository
	log  *logger.Logger
}

func NewService(repo Repository, log *logger.Logger) Service {
	return &service{
		repo: repo,
		log:  log,
	}
}

func (s *service) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return false, false, nil
	}

	isAktif, hasRecord, err := s.repo.CekStatusKamarInap(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal memeriksa status kamar inap pasien %s: %v", noRawat, err)
		return false, false, err
	}

	return isAktif, hasRecord, nil
}

func (s *service) GetKelasRawat(ctx context.Context, noRawat string) (string, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return "", nil
	}

	kelas, err := s.repo.GetKelasRawat(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal mengambil kelas rawat inap pasien %s: %v", noRawat, err)
		return "", err
	}

	return kelas, nil
}

func (s *service) DaftarPasienRawatInap(ctx context.Context, kodeDokterLogin string, filter FilterPasienRawatInap) ([]KunjunganRawatInap, shared.PaginationMeta, error) {
	kodeDokterLogin = strings.TrimSpace(kodeDokterLogin)
	list, total, err := s.repo.DaftarPasienRawatInap(ctx, kodeDokterLogin, filter)
	if err != nil {
		s.log.Error("Gagal mengambil daftar pasien rawat inap oleh dokter %s: %v", kodeDokterLogin, err)
		return nil, shared.PaginationMeta{}, err
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return list, meta, nil
}

func (s *service) DaftarStatusPulang(ctx context.Context) []OpsiReferensi {
	res := make([]OpsiReferensi, len(ListStatusPulang))
	for i, item := range ListStatusPulang {
		res[i] = OpsiReferensi{
			Value: string(item.Value),
			Label: item.Label,
		}
	}
	return res
}

func (s *service) DetailPasienRawatInap(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*KunjunganRawatInap, error) {
	noRawat = strings.TrimSpace(noRawat)
	tglMasuk = strings.TrimSpace(tglMasuk)
	jamMasuk = strings.TrimSpace(jamMasuk)

	if noRawat == "" {
		return nil, apperror.NewBusinessError("Nomor rawat wajib diisi")
	}
	if tglMasuk == "" {
		return nil, apperror.NewBusinessError("Tanggal masuk kamar inap wajib diisi")
	}
	if jamMasuk == "" {
		return nil, apperror.NewBusinessError("Jam masuk kamar inap wajib diisi")
	}

	item, err := s.repo.DetailPasienRawatInap(ctx, noRawat, tglMasuk, jamMasuk)
	if err != nil {
		s.log.Error("Gagal mengambil detail pasien rawat inap %s (%s %s): %v", noRawat, tglMasuk, jamMasuk, err)
		return nil, err
	}

	return item, nil
}
