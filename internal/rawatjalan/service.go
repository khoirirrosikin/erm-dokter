package rawatjalan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Service interface {
	DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter FilterAntreanDokter) ([]KunjunganRawatJalan, shared.PaginationMeta, error)
	DetailKunjungan(ctx context.Context, noRawat string, kodeDokter string) (*KunjunganRawatJalan, error)
	RiwayatKunjunganPasien(ctx context.Context, noRM string) ([]KunjunganRawatJalan, error)
	GetWaktuRegistrasi(ctx context.Context, noRawat string) (tanggal string, jam string, exists bool, err error)
	GetInfoRegistrasi(ctx context.Context, noRawat string) (*InfoRegistrasiPasien, error)
	PanggilAntrean(ctx context.Context, noRawat string, kdPoli string, kodeDokter string) (*HasilPanggilAntrean, error)
	UpdateStatusPasien(ctx context.Context, noRawat string, kodeDokter string, req UpdateStatusPasienRequest) (*HasilUpdateStatusPasien, error)

	DaftarStatusPemeriksaan(ctx context.Context) []OpsiReferensi
	DaftarStatusLanjut(ctx context.Context) []OpsiReferensi
	DaftarStatusBayar(ctx context.Context) []OpsiReferensi
	DaftarJenisAntrean(ctx context.Context) []OpsiReferensi
}

type service struct {
	repo           Repository
	log            *logger.Logger
	urlAntreanPoli string
	httpClient     *http.Client
}

func NewService(repo Repository, log *logger.Logger, urlAntreanPoli ...string) Service {
	var antreanURL string
	if len(urlAntreanPoli) > 0 {
		antreanURL = strings.TrimSpace(urlAntreanPoli[0])
	}
	return &service{
		repo:           repo,
		log:            log,
		urlAntreanPoli: antreanURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *service) DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter FilterAntreanDokter) ([]KunjunganRawatJalan, shared.PaginationMeta, error) {
	daftarAntrean, totalData, err := s.repo.DaftarAntreanDokter(ctx, kodeDokter, filter)
	if err != nil {
		s.log.Error("Gagal query antrean dokter %s: %v", kodeDokter, err)
		return nil, shared.PaginationMeta{}, err
	}

	return daftarAntrean, shared.NewPaginationMeta(totalData, filter.Page, filter.Limit), nil
}

func (s *service) DetailKunjungan(ctx context.Context, noRawat string, kodeDokter string) (*KunjunganRawatJalan, error) {
	if noRawat == "" {
		return nil, apperror.NewBusinessError("Nomor rawat wajib diisi")
	}

	kunjungan, err := s.repo.DetailKunjungan(ctx, noRawat, kodeDokter)
	if err != nil {
		s.log.Error("Gagal query detail kunjungan %s oleh dokter %s: %v", noRawat, kodeDokter, err)
		return nil, err
	}

	return kunjungan, nil
}

func (s *service) RiwayatKunjunganPasien(ctx context.Context, noRM string) ([]KunjunganRawatJalan, error) {
	if noRM == "" {
		return nil, apperror.NewBusinessError("Nomor rekam medis wajib diisi")
	}
	listKunjungan, err := s.repo.RiwayatKunjunganPasien(ctx, noRM)
	if err != nil {
		s.log.Error("Gagal query riwayat kunjungan pasien %s: %v", noRM, err)
		return nil, err
	}
	return listKunjungan, nil
}

func (s *service) GetWaktuRegistrasi(ctx context.Context, noRawat string) (string, string, bool, error) {
	if noRawat == "" {
		return "", "", false, apperror.NewBusinessError("Nomor rawat wajib diisi")
	}
	tanggalRegistrasi, jamRegistrasi, exists, err := s.repo.GetWaktuRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal query waktu registrasi %s: %v", noRawat, err)
		return "", "", false, err
	}
	return tanggalRegistrasi, jamRegistrasi, exists, nil
}

func (s *service) GetInfoRegistrasi(ctx context.Context, noRawat string) (*InfoRegistrasiPasien, error) {
	if noRawat == "" {
		return nil, apperror.NewBusinessError("Nomor rawat wajib diisi")
	}
	infoReg, err := s.repo.GetInfoRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal query info registrasi %s: %v", noRawat, err)
		return nil, err
	}
	if infoReg == nil {
		return nil, apperror.NewNotFoundError("Data kunjungan pasien tidak ditemukan")
	}
	return infoReg, nil
}

func (s *service) DaftarStatusPemeriksaan(ctx context.Context) []OpsiReferensi {
	opsi := make([]OpsiReferensi, len(ListStatusPemeriksaan))
	for i, item := range ListStatusPemeriksaan {
		opsi[i] = OpsiReferensi{
			Value: string(item.Value),
			Label: item.Label,
		}
	}
	return opsi
}

func (s *service) DaftarStatusLanjut(ctx context.Context) []OpsiReferensi {
	return []OpsiReferensi{
		{Value: string(shared.StatusLanjutRawatJalan), Label: "Rawat Jalan"},
		{Value: string(shared.StatusLanjutRawatInap), Label: "Rawat Inap"},
	}
}

func (s *service) DaftarStatusBayar(ctx context.Context) []OpsiReferensi {
	opsi := make([]OpsiReferensi, len(ListStatusBayar))
	for i, item := range ListStatusBayar {
		opsi[i] = OpsiReferensi{
			Value: string(item.Value),
			Label: item.Label,
		}
	}
	return opsi
}

func (s *service) DaftarJenisAntrean(ctx context.Context) []OpsiReferensi {
	opsi := make([]OpsiReferensi, len(ListJenisAntrean))
	for i, item := range ListJenisAntrean {
		opsi[i] = OpsiReferensi{
			Value: string(item.Value),
			Label: item.Label,
		}
	}
	return opsi
}

func (s *service) PanggilAntrean(ctx context.Context, noRawat string, kdPoli string, kodeDokter string) (*HasilPanggilAntrean, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return nil, apperror.NewBusinessError("Nomor rawat wajib diisi")
	}

	kdPoli = strings.TrimSpace(kdPoli)
	if kdPoli == "" {
		var err error
		kdPoli, err = s.repo.GetKodePoliByNoRawat(ctx, noRawat, kodeDokter)
		if err != nil {
			s.log.Error("Gagal mengambil kode poli untuk no_rawat %s: %v", noRawat, err)
			return nil, err
		}
	}

	if kdPoli == "" {
		return nil, apperror.NewNotFoundError("Kode poli tidak ditemukan untuk nomor rawat ini")
	}

	targetURL := strings.TrimSpace(s.urlAntreanPoli)
	if targetURL == "" {
		s.log.Error("URL display antrean belum disetel di environment variable URL_ANTREAN_POLI")
		return nil, apperror.NewBusinessError("URL display antrean poli belum dikonfigurasi pada sistem")
	}

	formData := url.Values{}
	formData.Set("no_rawat", noRawat)
	formData.Set("kd_poli", kdPoli)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, strings.NewReader(formData.Encode()))
	if err != nil {
		s.log.Error("Gagal membuat HTTP request antrean: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Error("Gagal mengirim request panggil antrean ke display (%s): %v", targetURL, err)
		return nil, apperror.NewBusinessError("Gagal menghubungi server display antrean", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		s.log.Error("Display antrean merespon error status %d", resp.StatusCode)
		return nil, apperror.NewBusinessError(fmt.Sprintf("Server display antrean merespon status error %d", resp.StatusCode))
	}

	s.log.Info("Berhasil panggil antrean display untuk no_rawat: %s, kd_poli: %s", noRawat, kdPoli)

	return &HasilPanggilAntrean{
		NoRawat:  noRawat,
		KodePoli: kdPoli,
	}, nil
}

func (s *service) UpdateStatusPasien(ctx context.Context, noRawat string, kodeDokter string, req UpdateStatusPasienRequest) (*HasilUpdateStatusPasien, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return nil, apperror.NewBusinessError("Nomor rawat wajib diisi")
	}

	kodeDokter = strings.TrimSpace(kodeDokter)
	exists, authorized, err := s.repo.CekAksesKunjungan(ctx, noRawat, kodeDokter)
	if err != nil {
		s.log.Error("Gagal memeriksa akses kunjungan %s: %v", noRawat, err)
		return nil, err
	}

	if !exists {
		return nil, apperror.NewNotFoundError("Data kunjungan pasien tidak ditemukan")
	}

	if !authorized {
		return nil, apperror.NewForbiddenError("Anda tidak memiliki akses untuk mengubah status pemeriksaan pasien ini")
	}

	if err := s.repo.UpdateStatusPemeriksaan(ctx, noRawat, req.StatusPemeriksaan); err != nil {
		s.log.Error("Gagal update status pemeriksaan kunjungan %s ke %s: %v", noRawat, req.StatusPemeriksaan, err)
		return nil, err
	}

	s.log.Info("Berhasil mengubah status pemeriksaan no_rawat: %s menjadi %s oleh dokter %s", noRawat, req.StatusPemeriksaan, kodeDokter)

	return &HasilUpdateStatusPasien{
		NoRawat:           noRawat,
		StatusPemeriksaan: req.StatusPemeriksaan,
	}, nil
}

