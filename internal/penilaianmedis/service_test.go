package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	detailData      *PenilaianMedisRalan
	detailErr       error
	riwayatData     []PenilaianMedisRalan
	riwayatErr      error
	adaResult       bool
	adaErr          error
	simpanErr       error
	updateErr       error
	hapusErr        error
	simpanCalledReq SimpanPenilaianMedisRalanRequest
	updateCalledReq UpdatePenilaianMedisRalanRequest
	hapusCalledNo   string

	detailIGDData      *PenilaianMedisIGD
	detailIGDErr       error
	riwayatIGDData     []PenilaianMedisIGD
	riwayatIGDErr      error
	adaIGDResult       bool
	adaIGDErr          error
	simpanIGDErr       error
	updateIGDErr       error
	hapusIGDErr        error
	simpanIGDCalledReq SimpanPenilaianMedisIGDRequest
	updateIGDCalledReq UpdatePenilaianMedisIGDRequest
	hapusIGDCalledNo   string

	detailRanapData      *PenilaianMedisRanap
	detailRanapErr       error
	riwayatRanapData     []PenilaianMedisRanap
	riwayatRanapErr      error
	adaRanapResult       bool
	adaRanapErr          error
	simpanRanapErr       error
	updateRanapErr       error
	hapusRanapErr        error
	simpanRanapCalledReq SimpanPenilaianMedisRanapRequest
	updateRanapCalledReq UpdatePenilaianMedisRanapRequest
	hapusRanapCalledNo   string

	detailRalanKandunganData      *PenilaianMedisRalanKandungan
	detailRalanKandunganErr       error
	riwayatRalanKandunganData     []PenilaianMedisRalanKandungan
	riwayatRalanKandunganErr      error
	adaRalanKandunganResult       bool
	adaRalanKandunganErr          error
	simpanRalanKandunganErr       error
	updateRalanKandunganErr       error
	hapusRalanKandunganErr        error
	simpanRalanKandunganCalledReq SimpanPenilaianMedisRalanKandunganRequest
	updateRalanKandunganCalledReq UpdatePenilaianMedisRalanKandunganRequest
	hapusRalanKandunganCalledNo   string

	detailRanapKandunganData      *PenilaianMedisRanapKandungan
	detailRanapKandunganErr       error
	riwayatRanapKandunganData     []PenilaianMedisRanapKandungan
	riwayatRanapKandunganErr      error
	adaRanapKandunganResult       bool
	adaRanapKandunganErr          error
	simpanRanapKandunganErr       error
	updateRanapKandunganErr       error
	hapusRanapKandunganErr        error
	simpanRanapKandunganCalledReq SimpanPenilaianMedisRanapKandunganRequest
	updateRanapKandunganCalledReq UpdatePenilaianMedisRanapKandunganRequest
	hapusRanapKandunganCalledNo   string
}

func (m *mockRepository) DetailPenilaianMedisRalan(ctx context.Context, noRawat string) (*PenilaianMedisRalan, error) {
	return m.detailData, m.detailErr
}

func (m *mockRepository) RiwayatPenilaianMedisRalanByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalan, error) {
	return m.riwayatData, m.riwayatErr
}

func (m *mockRepository) CekPenilaianMedisRalanAda(ctx context.Context, noRawat string) (bool, error) {
	return m.adaResult, m.adaErr
}

func (m *mockRepository) SimpanPenilaianMedisRalan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRalanRequest) error {
	m.simpanCalledReq = req
	return m.simpanErr
}

func (m *mockRepository) UpdatePenilaianMedisRalan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRalanRequest) error {
	m.updateCalledReq = req
	return m.updateErr
}

func (m *mockRepository) HapusPenilaianMedisRalan(ctx context.Context, noRawat string) error {
	m.hapusCalledNo = noRawat
	return m.hapusErr
}

func (m *mockRepository) DetailPenilaianMedisIGD(ctx context.Context, noRawat string) (*PenilaianMedisIGD, error) {
	return m.detailIGDData, m.detailIGDErr
}

func (m *mockRepository) RiwayatPenilaianMedisIGDByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisIGD, error) {
	return m.riwayatIGDData, m.riwayatIGDErr
}

func (m *mockRepository) CekPenilaianMedisIGDAda(ctx context.Context, noRawat string) (bool, error) {
	return m.adaIGDResult, m.adaIGDErr
}

func (m *mockRepository) SimpanPenilaianMedisIGD(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisIGDRequest) error {
	m.simpanIGDCalledReq = req
	return m.simpanIGDErr
}

func (m *mockRepository) UpdatePenilaianMedisIGD(ctx context.Context, noRawat string, req UpdatePenilaianMedisIGDRequest) error {
	m.updateIGDCalledReq = req
	return m.updateIGDErr
}

func (m *mockRepository) HapusPenilaianMedisIGD(ctx context.Context, noRawat string) error {
	m.hapusIGDCalledNo = noRawat
	return m.hapusIGDErr
}

func (m *mockRepository) DetailPenilaianMedisRanap(ctx context.Context, noRawat string) (*PenilaianMedisRanap, error) {
	return m.detailRanapData, m.detailRanapErr
}

func (m *mockRepository) RiwayatPenilaianMedisRanapByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanap, error) {
	return m.riwayatRanapData, m.riwayatRanapErr
}

func (m *mockRepository) CekPenilaianMedisRanapAda(ctx context.Context, noRawat string) (bool, error) {
	return m.adaRanapResult, m.adaRanapErr
}

func (m *mockRepository) SimpanPenilaianMedisRanap(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRanapRequest) error {
	m.simpanRanapCalledReq = req
	return m.simpanRanapErr
}

func (m *mockRepository) UpdatePenilaianMedisRanap(ctx context.Context, noRawat string, req UpdatePenilaianMedisRanapRequest) error {
	m.updateRanapCalledReq = req
	return m.updateRanapErr
}

func (m *mockRepository) HapusPenilaianMedisRanap(ctx context.Context, noRawat string) error {
	m.hapusRanapCalledNo = noRawat
	return m.hapusRanapErr
}

func (m *mockRepository) DetailPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRalanKandungan, error) {
	return m.detailRalanKandunganData, m.detailRalanKandunganErr
}

func (m *mockRepository) RiwayatPenilaianMedisRalanKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalanKandungan, error) {
	return m.riwayatRalanKandunganData, m.riwayatRalanKandunganErr
}

func (m *mockRepository) CekPenilaianMedisRalanKandunganAda(ctx context.Context, noRawat string) (bool, error) {
	return m.adaRalanKandunganResult, m.adaRalanKandunganErr
}

func (m *mockRepository) SimpanPenilaianMedisRalanKandungan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRalanKandunganRequest) error {
	m.simpanRalanKandunganCalledReq = req
	return m.simpanRalanKandunganErr
}

func (m *mockRepository) UpdatePenilaianMedisRalanKandungan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRalanKandunganRequest) error {
	m.updateRalanKandunganCalledReq = req
	return m.updateRalanKandunganErr
}

func (m *mockRepository) HapusPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) error {
	m.hapusRalanKandunganCalledNo = noRawat
	return m.hapusRalanKandunganErr
}

func (m *mockRepository) DetailPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRanapKandungan, error) {
	return m.detailRanapKandunganData, m.detailRanapKandunganErr
}

func (m *mockRepository) RiwayatPenilaianMedisRanapKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanapKandungan, error) {
	return m.riwayatRanapKandunganData, m.riwayatRanapKandunganErr
}

func (m *mockRepository) CekPenilaianMedisRanapKandunganAda(ctx context.Context, noRawat string) (bool, error) {
	return m.adaRanapKandunganResult, m.adaRanapKandunganErr
}

func (m *mockRepository) SimpanPenilaianMedisRanapKandungan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRanapKandunganRequest) error {
	m.simpanRanapKandunganCalledReq = req
	return m.simpanRanapKandunganErr
}

func (m *mockRepository) UpdatePenilaianMedisRanapKandungan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRanapKandunganRequest) error {
	m.updateRanapKandunganCalledReq = req
	return m.updateRanapKandunganErr
}

func (m *mockRepository) HapusPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) error {
	m.hapusRanapKandunganCalledNo = noRawat
	return m.hapusRanapKandunganErr
}

type mockRawatJalanService struct {
	rawatjalan.Service
	tglReg, jamReg string
	exists         bool
	err            error
}

func (m *mockRawatJalanService) DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter rawatjalan.FilterAntreanDokter) ([]rawatjalan.KunjunganRawatJalan, shared.PaginationMeta, error) {
	return nil, shared.PaginationMeta{}, nil
}
func (m *mockRawatJalanService) DetailKunjungan(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
	return nil, nil
}
func (m *mockRawatJalanService) RiwayatKunjunganPasien(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
	return nil, nil
}
func (m *mockRawatJalanService) GetWaktuRegistrasi(ctx context.Context, noRawat string) (string, string, bool, error) {
	return m.tglReg, m.jamReg, m.exists, m.err
}
func (m *mockRawatJalanService) GetInfoRegistrasi(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
	return nil, nil
}
func (m *mockRawatJalanService) DaftarStatusPemeriksaan(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}
func (m *mockRawatJalanService) DaftarStatusLanjut(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}
func (m *mockRawatJalanService) DaftarStatusBayar(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}
func (m *mockRawatJalanService) DaftarJenisAntrean(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}

type mockRawatInapService struct {
	rawatinap.Service
	cekStatusKamarInapFunc func(ctx context.Context, noRawat string) (bool, bool, error)
}

func (m *mockRawatInapService) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	if m.cekStatusKamarInapFunc != nil {
		return m.cekStatusKamarInapFunc(ctx, noRawat)
	}
	return false, false, nil
}

func TestService_Referensi(t *testing.T) {
	log := logger.New()
	svc := NewService(&mockRepository{}, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)
	ref := svc.Referensi(context.Background())

	if len(ref.Anamnesis) == 0 || len(ref.Keadaan) == 0 || len(ref.Kesadaran) == 0 || len(ref.StatusFisik) == 0 {
		t.Errorf("expected all reference options to be non-empty, got: %+v", ref)
	}
}

func TestService_DetailPenilaianMedisRalan_Success(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
			DataPenilaianMedisRalan: DataPenilaianMedisRalan{
				KeluhanUtama: "Batuk",
			},
		},
	}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	result, err := svc.DetailPenilaianMedisRalan(context.Background(), "2026/04/22/000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NoRawat != "2026/04/22/000001" {
		t.Errorf("expected no_rawat 2026/04/22/000001, got %s", result.NoRawat)
	}
}

func TestService_DetailPenilaianMedisRalan_NotFound(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		detailErr: sql.ErrNoRows,
	}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	_, err := svc.DetailPenilaianMedisRalan(context.Background(), "2026/04/22/999999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *apperror.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestService_RiwayatPenilaianMedisRalanByNoRM_Success(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		riwayatData: []PenilaianMedisRalan{
			{
				NoRawat: "2026/04/22/000001",
			},
			{
				NoRawat: "2026/04/21/000002",
			},
		},
	}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	list, err := svc.RiwayatPenilaianMedisRalanByNoRM(context.Background(), "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}

func TestService_SimpanPenilaianMedisRalan_Expired48Hours(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: "2020-01-01", jamReg: "10:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := SimpanPenilaianMedisRalanRequest{
		NoRawat: "2026/04/22/000001",
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: "2026-04-22 09:30:00",
			KeluhanUtama:     "Demam",
			Diagnosis:        "Febris",
			TataLaksana:      "Paracetamol",
		},
	}

	_, err := svc.SimpanPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err == nil {
		t.Fatal("expected 48-hour expiration error, got nil")
	}
	var forbErr *apperror.ForbiddenError
	if !errors.As(err, &forbErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisRalan_EarlierThanRegistration(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "10:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := SimpanPenilaianMedisRalanRequest{
		NoRawat: "2026/04/22/000001",
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: today + " 08:00:00",
			KeluhanUtama:     "Demam",
			Diagnosis:        "Febris",
			TataLaksana:      "Paracetamol",
		},
	}

	_, err := svc.SimpanPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err == nil {
		t.Fatal("expected validation error when waktu penilaian is before waktu_registrasi")
	}
	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if valErr["tanggal_penilaian"] == "" {
		t.Errorf("expected error on tanggal_penilaian, got: %+v", valErr)
	}
}

func TestService_SimpanPenilaianMedisRalan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		adaResult: false,
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := SimpanPenilaianMedisRalanRequest{
		NoRawat: "2026/04/22/000001",
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Demam",
			Diagnosis:        "Febris",
			TataLaksana:      "Paracetamol",
		},
	}

	result, err := svc.SimpanPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NoRawat != "2026/04/22/000001" {
		t.Errorf("expected no_rawat 2026/04/22/000001, got %s", result.NoRawat)
	}
}

func TestService_SimpanPenilaianMedisRalan_Duplicate(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		adaResult: true,
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := SimpanPenilaianMedisRalanRequest{
		NoRawat: "2026/04/22/000001",
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Demam",
			Diagnosis:        "Febris",
			TataLaksana:      "Paracetamol",
		},
	}

	_, err := svc.SimpanPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err == nil {
		t.Fatal("expected duplicate error, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_UpdatePenilaianMedisRalan_Forbidden(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := UpdatePenilaianMedisRalanRequest{
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Demam berkurang",
			Diagnosis:        "Febris H2",
			TataLaksana:      "Lanjut Paracetamol",
		},
	}

	_, err := svc.UpdatePenilaianMedisRalan(context.Background(), "DR002", "2026/04/22/000001", req)
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
	var forbErr *apperror.ForbiddenError
	if !errors.As(err, &forbErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_UpdatePenilaianMedisRalan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := UpdatePenilaianMedisRalanRequest{
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Demam berkurang",
			Diagnosis:        "Febris H2",
			TataLaksana:      "Lanjut Paracetamol",
		},
	}

	result, err := svc.UpdatePenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NoRawat != "2026/04/22/000001" {
		t.Errorf("expected no_rawat 2026/04/22/000001, got %s", result.NoRawat)
	}
}

func TestService_HapusPenilaianMedisRalan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	err := svc.HapusPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.hapusCalledNo != "2026/04/22/000001" {
		t.Errorf("expected hapusCalledNo to be 2026/04/22/000001, got %s", repo.hapusCalledNo)
	}
}

func TestService_HapusPenilaianMedisRalan_Forbidden(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	err := svc.HapusPenilaianMedisRalan(context.Background(), "DR002", "2026/04/22/000001")
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
	var forbErr *apperror.ForbiddenError
	if !errors.As(err, &forbErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_DetailPenilaianMedisIGD_Success(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
			DataPenilaianMedisIGD: DataPenilaianMedisIGD{
				KeluhanUtama: "Sesak Napas Akut",
				EKG:          "Sinus Takikardia",
			},
		},
	}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	result, err := svc.DetailPenilaianMedisIGD(context.Background(), "2026/04/22/000002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NoRawat != "2026/04/22/000002" {
		t.Errorf("expected no_rawat 2026/04/22/000002, got %s", result.NoRawat)
	}
}

func TestService_DetailPenilaianMedisIGD_NotFound(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		detailIGDErr: sql.ErrNoRows,
	}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	_, err := svc.DetailPenilaianMedisIGD(context.Background(), "2026/04/22/999999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *apperror.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestService_RiwayatPenilaianMedisIGDByNoRM_Success(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		riwayatIGDData: []PenilaianMedisIGD{
			{NoRawat: "2026/04/22/000002"},
		},
	}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	list, err := svc.RiwayatPenilaianMedisIGDByNoRM(context.Background(), "654321")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
}

func TestService_SimpanPenilaianMedisIGD_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		adaIGDResult: false,
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := SimpanPenilaianMedisIGDRequest{
		NoRawat: "2026/04/22/000002",
		DataPenilaianMedisIGD: DataPenilaianMedisIGD{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Nyeri dada hebat",
			Diagnosis:        "STEMI",
			TataLaksana:      "Oksigen + Aspilet",
			EKG:              "ST Elevasi V1-V4",
		},
	}

	result, err := svc.SimpanPenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NoRawat != "2026/04/22/000002" {
		t.Errorf("expected no_rawat 2026/04/22/000002, got %s", result.NoRawat)
	}
}

func TestService_SimpanPenilaianMedisIGD_Duplicate(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		adaIGDResult: true,
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := SimpanPenilaianMedisIGDRequest{
		NoRawat: "2026/04/22/000002",
		DataPenilaianMedisIGD: DataPenilaianMedisIGD{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Nyeri dada hebat",
			Diagnosis:        "STEMI",
			TataLaksana:      "Oksigen",
		},
	}

	_, err := svc.SimpanPenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002", req)
	if err == nil {
		t.Fatal("expected duplicate error for IGD, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_UpdatePenilaianMedisIGD_Forbidden(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := UpdatePenilaianMedisIGDRequest{
		DataPenilaianMedisIGD: DataPenilaianMedisIGD{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Nyeri dada membaik",
			Diagnosis:        "STEMI Post PCI",
			TataLaksana:      "Lanjut terapi",
		},
	}

	_, err := svc.UpdatePenilaianMedisIGD(context.Background(), "DR002", "2026/04/22/000002", req)
	if err == nil {
		t.Fatal("expected forbidden error for IGD, got nil")
	}
	var forbErr *apperror.ForbiddenError
	if !errors.As(err, &forbErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_UpdatePenilaianMedisIGD_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	req := UpdatePenilaianMedisIGDRequest{
		DataPenilaianMedisIGD: DataPenilaianMedisIGD{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Nyeri dada membaik",
			Diagnosis:        "STEMI Post PCI",
			TataLaksana:      "Lanjut terapi",
		},
	}

	result, err := svc.UpdatePenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NoRawat != "2026/04/22/000002" {
		t.Errorf("expected no_rawat 2026/04/22/000002, got %s", result.NoRawat)
	}
}

func TestService_HapusPenilaianMedisIGD_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	err := svc.HapusPenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.hapusIGDCalledNo != "2026/04/22/000002" {
		t.Errorf("expected hapusIGDCalledNo to be 2026/04/22/000002, got %s", repo.hapusIGDCalledNo)
	}
}

func TestService_HapusPenilaianMedisIGD_Forbidden(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

	err := svc.HapusPenilaianMedisIGD(context.Background(), "DR002", "2026/04/22/000002")
	if err == nil {
		t.Fatal("expected forbidden error for IGD, got nil")
	}
	var forbErr *apperror.ForbiddenError
	if !errors.As(err, &forbErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisRalan_RanapCheckout_Ditolak(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRalanRequest{
		NoRawat: "2026/04/22/000001",
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Demam",
			Diagnosis:        "Febris",
			TataLaksana:      "Paracetamol",
		},
	}

	_, err := svc.SimpanPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err == nil {
		t.Fatal("expected BusinessError when patient checked out of rawat inap, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
	if busErr.Message != "Pasien sudah keluar dari kamar inap" {
		t.Errorf("unexpected error message: %s", busErr.Message)
	}
}

func TestService_SimpanPenilaianMedisRalan_RanapAktif_Diizinkan(t *testing.T) {
	log := logger.New()

	repo := &mockRepository{
		adaResult: false,
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: "2026-04-15", jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRalanRequest{
		NoRawat: "2026/04/22/000001",
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: "2026-04-16 09:00:00",
			KeluhanUtama:     "Demam",
			Diagnosis:        "Febris",
			TataLaksana:      "Paracetamol",
		},
	}

	res, err := svc.SimpanPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err != nil {
		t.Fatalf("expected success for active ranap patient, got %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestService_UpdatePenilaianMedisRalan_RanapCheckout_Ditolak(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisRalanRequest{
		DataPenilaianMedisRalan: DataPenilaianMedisRalan{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Demam berkurang",
			Diagnosis:        "Febris H2",
			TataLaksana:      "Lanjut Paracetamol",
		},
	}

	_, err := svc.UpdatePenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001", req)
	if err == nil {
		t.Fatal("expected BusinessError when updating checkout patient, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_HapusPenilaianMedisRalan_RanapCheckout_Ditolak(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailData: &PenilaianMedisRalan{
			NoRawat:    "2026/04/22/000001",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisRalan(context.Background(), "DR001", "2026/04/22/000001")
	if err == nil {
		t.Fatal("expected BusinessError when deleting checkout patient, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisIGD_RanapCheckout_Ditolak(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisIGDRequest{
		NoRawat: "2026/04/22/000002",
		DataPenilaianMedisIGD: DataPenilaianMedisIGD{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Nyeri dada hebat",
			Diagnosis:        "STEMI",
			TataLaksana:      "Oksigen + Aspilet",
			EKG:              "ST Elevasi V1-V4",
		},
	}

	_, err := svc.SimpanPenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002", req)
	if err == nil {
		t.Fatal("expected BusinessError when saving checked out patient IGD assessment, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisIGD_RanapAktif_Diizinkan(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		adaIGDResult: false,
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: "2026-04-15", jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisIGDRequest{
		NoRawat: "2026/04/22/000002",
		DataPenilaianMedisIGD: DataPenilaianMedisIGD{
			TanggalPenilaian: "2026-04-16 09:00:00",
			KeluhanUtama:     "Nyeri dada hebat",
			Diagnosis:        "STEMI",
			TataLaksana:      "Oksigen + Aspilet",
			EKG:              "ST Elevasi V1-V4",
		},
	}

	res, err := svc.SimpanPenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002", req)
	if err != nil {
		t.Fatalf("expected success for active ranap patient IGD assessment, got %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestService_UpdatePenilaianMedisIGD_RanapCheckout_Ditolak(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisIGDRequest{
		DataPenilaianMedisIGD: DataPenilaianMedisIGD{
			TanggalPenilaian: today + " 09:00:00",
			KeluhanUtama:     "Nyeri dada membaik",
			Diagnosis:        "STEMI Post PCI",
			TataLaksana:      "Lanjut terapi",
		},
	}

	_, err := svc.UpdatePenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002", req)
	if err == nil {
		t.Fatal("expected BusinessError when updating checkout patient IGD assessment, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_HapusPenilaianMedisIGD_RanapCheckout_Ditolak(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		detailIGDData: &PenilaianMedisIGD{
			NoRawat:    "2026/04/22/000002",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisIGD(context.Background(), "DR001", "2026/04/22/000002")
	if err == nil {
		t.Fatal("expected BusinessError when deleting checkout patient IGD assessment, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_DetailPenilaianMedisRanap_Success(t *testing.T) {
	log := logger.New()
	expected := &PenilaianMedisRanap{
		NoRawat:    "2026/04/22/000003",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			KeluhanUtama: "Sesak nafas dan batuk",
			Diagnosis:    "Pneumonia",
			TataLaksana:  "O2 nasal canul 3 lpm",
		},
	}
	repo := &mockRepository{detailRanapData: expected}
	rjSvc := &mockRawatJalanService{tglReg: "2026-04-22", jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	res, err := svc.DetailPenilaianMedisRanap(context.Background(), "2026/04/22/000003")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != expected.NoRawat || res.Diagnosis != expected.Diagnosis {
		t.Errorf("expected %+v, got %+v", expected, res)
	}
}

func TestService_DetailPenilaianMedisRanap_NotFound(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{detailRanapErr: sql.ErrNoRows}
	rjSvc := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	_, err := svc.DetailPenilaianMedisRanap(context.Background(), "2026/04/22/000099")
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	var notFound *apperror.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestService_RiwayatPenilaianMedisRanapByNoRM_Success(t *testing.T) {
	log := logger.New()
	expectedList := []PenilaianMedisRanap{
		{
			NoRawat:    "2026/04/22/000003",
			KodeDokter: "DR001",
			NamaDokter: "dr. Budi",
		},
	}
	repo := &mockRepository{riwayatRanapData: expectedList}
	rjSvc := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	list, err := svc.RiwayatPenilaianMedisRanapByNoRM(context.Background(), "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
}

func TestService_SimpanPenilaianMedisRanap_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	expected := &PenilaianMedisRanap{
		NoRawat:    "2026/04/22/000003",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Nyeri dada hebat",
			Diagnosis:        "STEMI",
			TataLaksana:      "ISDN sublingual, rujuk ICCU",
		},
	}
	repo := &mockRepository{
		adaRanapResult:  false,
		detailRanapData: expected,
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapRequest{
		NoRawat: "2026/04/22/000003",
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Nyeri dada hebat",
			Diagnosis:        "STEMI",
			TataLaksana:      "ISDN sublingual, rujuk ICCU",
		},
	}

	res, err := svc.SimpanPenilaianMedisRanap(context.Background(), "DR001", "2026/04/22/000003", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000003" {
		t.Errorf("expected no_rawat 2026/04/22/000003, got %s", res.NoRawat)
	}
}

func TestService_SimpanPenilaianMedisRanap_BelumTerdaftarKamarInap(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapRequest{
		NoRawat: "2026/04/22/000003",
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Nyeri dada",
			Diagnosis:        "STEMI",
			TataLaksana:      "ISDN",
		},
	}

	_, err := svc.SimpanPenilaianMedisRanap(context.Background(), "DR001", "2026/04/22/000003", req)
	if err == nil {
		t.Fatal("expected error when patient not admitted to inpatient, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisRanap_CheckoutKamarInap(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapRequest{
		NoRawat: "2026/04/22/000003",
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Nyeri dada",
			Diagnosis:        "STEMI",
			TataLaksana:      "ISDN",
		},
	}

	_, err := svc.SimpanPenilaianMedisRanap(context.Background(), "DR001", "2026/04/22/000003", req)
	if err == nil {
		t.Fatal("expected error when patient already checked out from inpatient, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisRanap_Duplikasi(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{adaRanapResult: true}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapRequest{
		NoRawat: "2026/04/22/000003",
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Nyeri dada",
			Diagnosis:        "STEMI",
			TataLaksana:      "ISDN",
		},
	}

	_, err := svc.SimpanPenilaianMedisRanap(context.Background(), "DR001", "2026/04/22/000003", req)
	if err == nil {
		t.Fatal("expected BusinessError on duplicate assessment, got nil")
	}
	var busErr *apperror.BusinessError
	if !errors.As(err, &busErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_UpdatePenilaianMedisRanap_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanap{
		NoRawat:    "2026/04/22/000003",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisRanapRequest{
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			TanggalPenilaian: fmt.Sprintf("%s 10:00:00", today),
			KeluhanUtama:     "Keluhan membaik",
			Diagnosis:        "STEMI perbaikan",
			TataLaksana:      "Lanjut terapi oral",
		},
	}

	res, err := svc.UpdatePenilaianMedisRanap(context.Background(), "DR001", "2026/04/22/000003", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000003" {
		t.Errorf("expected no_rawat 2026/04/22/000003, got %s", res.NoRawat)
	}
}

func TestService_UpdatePenilaianMedisRanap_DokterLain(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanap{
		NoRawat:    "2026/04/22/000003",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisRanapRequest{
		DataPenilaianMedisRanap: DataPenilaianMedisRanap{
			TanggalPenilaian: fmt.Sprintf("%s 10:00:00", today),
			KeluhanUtama:     "Keluhan membaik",
			Diagnosis:        "STEMI perbaikan",
			TataLaksana:      "Lanjut terapi oral",
		},
	}

	_, err := svc.UpdatePenilaianMedisRanap(context.Background(), "DR002", "2026/04/22/000003", req)
	if err == nil {
		t.Fatal("expected ForbiddenError when updated by different doctor, got nil")
	}
	var fErr *apperror.ForbiddenError
	if !errors.As(err, &fErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_HapusPenilaianMedisRanap_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanap{
		NoRawat:    "2026/04/22/000003",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisRanap(context.Background(), "DR001", "2026/04/22/000003")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.hapusRanapCalledNo != "2026/04/22/000003" {
		t.Errorf("expected repo.Hapus called with 2026/04/22/000003, got %s", repo.hapusRanapCalledNo)
	}
}

func TestService_HapusPenilaianMedisRanap_DokterLain(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanap{
		NoRawat:    "2026/04/22/000003",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisRanap(context.Background(), "DR002", "2026/04/22/000003")
	if err == nil {
		t.Fatal("expected ForbiddenError when deleted by different doctor, got nil")
	}
	var fErr *apperror.ForbiddenError
	if !errors.As(err, &fErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_DetailPenilaianMedisRalanKandungan_Success(t *testing.T) {
	log := logger.New()
	expected := &PenilaianMedisRalanKandungan{
		NoRawat:    "2026/04/22/000004",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
		DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
			KeluhanUtama: "Mules-mules",
			Diagnosis:    "G1P0A0",
			TataLaksana:  "Observasi",
			Kontraksi:    KontraksiAda,
		},
	}
	repo := &mockRepository{detailRalanKandunganData: expected}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	res, err := svc.DetailPenilaianMedisRalanKandungan(context.Background(), "2026/04/22/000004")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000004" || res.Kontraksi != KontraksiAda {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestService_DetailPenilaianMedisRalanKandungan_NotFound(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{detailRalanKandunganErr: sql.ErrNoRows}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	_, err := svc.DetailPenilaianMedisRalanKandungan(context.Background(), "2026/04/22/999999")
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	var nfErr *apperror.NotFoundError
	if !errors.As(err, &nfErr) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestService_RiwayatPenilaianMedisRalanKandunganByNoRM_Success(t *testing.T) {
	log := logger.New()
	expected := []PenilaianMedisRalanKandungan{
		{
			NoRawat: "2026/04/22/000004",
			DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
				KeluhanUtama: "Kontrol kehamilan",
				Kontraksi:    KontraksiTidak,
			},
		},
	}
	repo := &mockRepository{riwayatRalanKandunganData: expected}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	res, err := svc.RiwayatPenilaianMedisRalanKandunganByNoRM(context.Background(), "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0].Kontraksi != KontraksiTidak {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestService_SimpanPenilaianMedisRalanKandungan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		adaRalanKandunganResult: false,
		detailRalanKandunganData: &PenilaianMedisRalanKandungan{
			NoRawat:    "2026/04/22/000004",
			KodeDokter: "DR001",
			DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
				KeluhanUtama: "Perut kencang",
				Kontraksi:    KontraksiAda,
			},
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRalanKandunganRequest{
		NoRawat: "2026/04/22/000004",
		DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Perut kencang",
			Diagnosis:        "G1P0A0",
			TataLaksana:      "Observasi",
			Kontraksi:        KontraksiAda,
		},
	}

	res, err := svc.SimpanPenilaianMedisRalanKandungan(context.Background(), "DR001", "2026/04/22/000004", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000004" {
		t.Errorf("expected no_rawat 2026/04/22/000004, got %s", res.NoRawat)
	}
	if repo.simpanRalanKandunganCalledReq.KeluhanUtama != "Perut kencang" {
		t.Errorf("expected repo called with KeluhanUtama, got %s", repo.simpanRalanKandunganCalledReq.KeluhanUtama)
	}
}

func TestService_SimpanPenilaianMedisRalanKandungan_Duplikasi(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{adaRalanKandunganResult: true}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRalanKandunganRequest{
		NoRawat: "2026/04/22/000004",
		DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Perut kencang",
			Diagnosis:        "G1P0A0",
			TataLaksana:      "Observasi",
		},
	}

	_, err := svc.SimpanPenilaianMedisRalanKandungan(context.Background(), "DR001", "2026/04/22/000004", req)
	if err == nil {
		t.Fatal("expected BusinessError on duplicate, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisRalanKandungan_Lewat48Jam(t *testing.T) {
	log := logger.New()
	threeDaysAgo := time.Now().Add(-72 * time.Hour).Format("2006-01-02")
	repo := &mockRepository{adaRalanKandunganResult: false}
	rjSvc := &mockRawatJalanService{tglReg: threeDaysAgo, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRalanKandunganRequest{
		NoRawat: "2026/04/22/000004",
		DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", threeDaysAgo),
			KeluhanUtama:     "Perut kencang",
			Diagnosis:        "G1P0A0",
			TataLaksana:      "Observasi",
		},
	}

	_, err := svc.SimpanPenilaianMedisRalanKandungan(context.Background(), "DR001", "2026/04/22/000004", req)
	if err == nil {
		t.Fatal("expected ForbiddenError when > 48h, got nil")
	}
	var fErr *apperror.ForbiddenError
	if !errors.As(err, &fErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_UpdatePenilaianMedisRalanKandungan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRalanKandungan{
		NoRawat:    "2026/04/22/000004",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRalanKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisRalanKandunganRequest{
		DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 10:00:00", today),
			KeluhanUtama:     "Mules-mules berkurang",
			Diagnosis:        "G1P0A0 belum inpartu",
			TataLaksana:      "Rawat jalan",
			Kontraksi:        KontraksiTidak,
		},
	}

	res, err := svc.UpdatePenilaianMedisRalanKandungan(context.Background(), "DR001", "2026/04/22/000004", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000004" {
		t.Errorf("expected no_rawat 2026/04/22/000004, got %s", res.NoRawat)
	}
}

func TestService_UpdatePenilaianMedisRalanKandungan_DokterLain(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRalanKandungan{
		NoRawat:    "2026/04/22/000004",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRalanKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisRalanKandunganRequest{
		DataPenilaianMedisRalanKandungan: DataPenilaianMedisRalanKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 10:00:00", today),
			KeluhanUtama:     "Mules-mules berkurang",
			Diagnosis:        "G1P0A0 belum inpartu",
			TataLaksana:      "Rawat jalan",
		},
	}

	_, err := svc.UpdatePenilaianMedisRalanKandungan(context.Background(), "DR002", "2026/04/22/000004", req)
	if err == nil {
		t.Fatal("expected ForbiddenError when updated by different doctor, got nil")
	}
	var fErr *apperror.ForbiddenError
	if !errors.As(err, &fErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_HapusPenilaianMedisRalanKandungan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRalanKandungan{
		NoRawat:    "2026/04/22/000004",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRalanKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisRalanKandungan(context.Background(), "DR001", "2026/04/22/000004")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.hapusRalanKandunganCalledNo != "2026/04/22/000004" {
		t.Errorf("expected repo.Hapus called with 2026/04/22/000004, got %s", repo.hapusRalanKandunganCalledNo)
	}
}

func TestService_HapusPenilaianMedisRalanKandungan_DokterLain(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRalanKandungan{
		NoRawat:    "2026/04/22/000004",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRalanKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisRalanKandungan(context.Background(), "DR002", "2026/04/22/000004")
	if err == nil {
		t.Fatal("expected ForbiddenError when deleted by different doctor, got nil")
	}
	var fErr *apperror.ForbiddenError
	if !errors.As(err, &fErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_DetailPenilaianMedisRanapKandungan_Success(t *testing.T) {
	log := logger.New()
	expected := &PenilaianMedisRanapKandungan{
		NoRawat:    "2026/04/22/000005",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
		DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
			KeluhanUtama: "Mules-mules teratur",
			Diagnosis:    "G2P1A0 inpartu",
			TataLaksana:  "Observasi persalinan",
			Kontraksi:    KontraksiAda,
			Edukasi:      "Edukasi proses melahirkan",
		},
	}
	repo := &mockRepository{detailRanapKandunganData: expected}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	res, err := svc.DetailPenilaianMedisRanapKandungan(context.Background(), "2026/04/22/000005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000005" || res.Kontraksi != KontraksiAda || res.Edukasi != "Edukasi proses melahirkan" {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestService_DetailPenilaianMedisRanapKandungan_NotFound(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{detailRanapKandunganErr: sql.ErrNoRows}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	_, err := svc.DetailPenilaianMedisRanapKandungan(context.Background(), "2026/04/22/999999")
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	var nfErr *apperror.NotFoundError
	if !errors.As(err, &nfErr) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestService_RiwayatPenilaianMedisRanapKandunganByNoRM_Success(t *testing.T) {
	log := logger.New()
	expected := []PenilaianMedisRanapKandungan{
		{
			NoRawat: "2026/04/22/000005",
			DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
				KeluhanUtama: "Inpartu kala I",
				Kontraksi:    KontraksiAda,
			},
		},
	}
	repo := &mockRepository{riwayatRanapKandunganData: expected}
	svc := NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, log)

	res, err := svc.RiwayatPenilaianMedisRanapKandunganByNoRM(context.Background(), "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0].Kontraksi != KontraksiAda {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestService_SimpanPenilaianMedisRanapKandungan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		adaRanapKandunganResult: false,
		detailRanapKandunganData: &PenilaianMedisRanapKandungan{
			NoRawat:    "2026/04/22/000005",
			KodeDokter: "DR001",
			DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
				KeluhanUtama: "Mules teratur",
				Kontraksi:    KontraksiAda,
			},
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapKandunganRequest{
		NoRawat: "2026/04/22/000005",
		DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Mules teratur",
			Diagnosis:        "G2P1A0 inpartu",
			TataLaksana:      "Observasi ketat",
			Kontraksi:        KontraksiAda,
			Edukasi:          "Edukasi proses melahirkan",
		},
	}

	res, err := svc.SimpanPenilaianMedisRanapKandungan(context.Background(), "DR001", "2026/04/22/000005", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000005" {
		t.Errorf("expected no_rawat 2026/04/22/000005, got %s", res.NoRawat)
	}
	if repo.simpanRanapKandunganCalledReq.KeluhanUtama != "Mules teratur" {
		t.Errorf("expected repo called with KeluhanUtama, got %s", repo.simpanRanapKandunganCalledReq.KeluhanUtama)
	}
}

func TestService_SimpanPenilaianMedisRanapKandungan_BelumTerdaftarKamarInap(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{adaRanapKandunganResult: false}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapKandunganRequest{
		NoRawat: "2026/04/22/000005",
		DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Mules teratur",
			Diagnosis:        "G2P1A0",
			TataLaksana:      "Observasi",
		},
	}

	_, err := svc.SimpanPenilaianMedisRanapKandungan(context.Background(), "DR001", "2026/04/22/000005", req)
	if err == nil {
		t.Fatal("expected BusinessError when not registered in kamar inap, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisRanapKandungan_CheckoutKamarInap(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{adaRanapKandunganResult: false}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapKandunganRequest{
		NoRawat: "2026/04/22/000005",
		DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Mules teratur",
			Diagnosis:        "G2P1A0",
			TataLaksana:      "Observasi",
		},
	}

	_, err := svc.SimpanPenilaianMedisRanapKandungan(context.Background(), "DR001", "2026/04/22/000005", req)
	if err == nil {
		t.Fatal("expected BusinessError when already checked out, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_SimpanPenilaianMedisRanapKandungan_Duplikasi(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{adaRanapKandunganResult: true}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := SimpanPenilaianMedisRanapKandunganRequest{
		NoRawat: "2026/04/22/000005",
		DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 09:00:00", today),
			KeluhanUtama:     "Mules teratur",
			Diagnosis:        "G2P1A0",
			TataLaksana:      "Observasi",
		},
	}

	_, err := svc.SimpanPenilaianMedisRanapKandungan(context.Background(), "DR001", "2026/04/22/000005", req)
	if err == nil {
		t.Fatal("expected BusinessError on duplicate, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
}

func TestService_UpdatePenilaianMedisRanapKandungan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanapKandungan{
		NoRawat:    "2026/04/22/000005",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisRanapKandunganRequest{
		DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 10:00:00", today),
			KeluhanUtama:     "Mules-mules bertambah kuat",
			Diagnosis:        "G2P1A0 inpartu kala I fase aktif",
			TataLaksana:      "Observasi ketat",
			Kontraksi:        KontraksiAda,
			Edukasi:          "Edukasi pendamping persalinan",
		},
	}

	res, err := svc.UpdatePenilaianMedisRanapKandungan(context.Background(), "DR001", "2026/04/22/000005", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoRawat != "2026/04/22/000005" {
		t.Errorf("expected no_rawat 2026/04/22/000005, got %s", res.NoRawat)
	}
}

func TestService_UpdatePenilaianMedisRanapKandungan_DokterLain(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanapKandungan{
		NoRawat:    "2026/04/22/000005",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	req := UpdatePenilaianMedisRanapKandunganRequest{
		DataPenilaianMedisRanapKandungan: DataPenilaianMedisRanapKandungan{
			TanggalPenilaian: fmt.Sprintf("%s 10:00:00", today),
			KeluhanUtama:     "Mules-mules bertambah kuat",
			Diagnosis:        "G2P1A0 inpartu",
			TataLaksana:      "Observasi",
		},
	}

	_, err := svc.UpdatePenilaianMedisRanapKandungan(context.Background(), "DR002", "2026/04/22/000005", req)
	if err == nil {
		t.Fatal("expected ForbiddenError when updated by different doctor, got nil")
	}
	var fErr *apperror.ForbiddenError
	if !errors.As(err, &fErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_HapusPenilaianMedisRanapKandungan_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanapKandungan{
		NoRawat:    "2026/04/22/000005",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisRanapKandungan(context.Background(), "DR001", "2026/04/22/000005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.hapusRanapKandunganCalledNo != "2026/04/22/000005" {
		t.Errorf("expected repo.Hapus called with 2026/04/22/000005, got %s", repo.hapusRanapKandunganCalledNo)
	}
}

func TestService_HapusPenilaianMedisRanapKandungan_DokterLain(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	existing := &PenilaianMedisRanapKandungan{
		NoRawat:    "2026/04/22/000005",
		KodeDokter: "DR001",
		NamaDokter: "dr. Budi",
	}
	repo := &mockRepository{detailRanapKandunganData: existing}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := NewService(repo, rjSvc, mockRI, 48, log)

	err := svc.HapusPenilaianMedisRanapKandungan(context.Background(), "DR002", "2026/04/22/000005")
	if err == nil {
		t.Fatal("expected ForbiddenError when deleted by different doctor, got nil")
	}
	var fErr *apperror.ForbiddenError
	if !errors.As(err, &fErr) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
}

func TestService_ValidasiKepemilikanDokter_StandardizedMessage(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")

	t.Run("Update Ralan by different doctor", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &PenilaianMedisRalan{
				NoRawat:    "2026/04/22/000001",
				KodeDokter: "DR001",
				NamaDokter: "dr. Budi",
			},
		}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
		svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

		_, err := svc.UpdatePenilaianMedisRalan(context.Background(), "DR002", "2026/04/22/000001", UpdatePenilaianMedisRalanRequest{})
		if err == nil {
			t.Fatal("expected forbidden error, got nil")
		}
		var fErr *apperror.ForbiddenError
		if !errors.As(err, &fErr) {
			t.Fatalf("expected ForbiddenError, got %v", err)
		}
		if fErr.Message != "Penilaian medis dokter lain tidak dapat diubah" {
			t.Errorf("expected 'Penilaian medis dokter lain tidak dapat diubah', got: %s", fErr.Message)
		}
	})

	t.Run("Hapus Ralan by different doctor", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &PenilaianMedisRalan{
				NoRawat:    "2026/04/22/000001",
				KodeDokter: "DR001",
				NamaDokter: "dr. Budi",
			},
		}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
		svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

		err := svc.HapusPenilaianMedisRalan(context.Background(), "DR002", "2026/04/22/000001")
		if err == nil {
			t.Fatal("expected forbidden error, got nil")
		}
		var fErr *apperror.ForbiddenError
		if !errors.As(err, &fErr) {
			t.Fatalf("expected ForbiddenError, got %v", err)
		}
		if fErr.Message != "Penilaian medis dokter lain tidak dapat dihapus" {
			t.Errorf("expected 'Penilaian medis dokter lain tidak dapat dihapus', got: %s", fErr.Message)
		}
	})

	t.Run("Update IGD by different doctor", func(t *testing.T) {
		repo := &mockRepository{
			detailIGDData: &PenilaianMedisIGD{
				NoRawat:    "2026/04/22/000002",
				KodeDokter: "DR001",
				NamaDokter: "dr. Budi",
			},
		}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
		svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

		_, err := svc.UpdatePenilaianMedisIGD(context.Background(), "DR002", "2026/04/22/000002", UpdatePenilaianMedisIGDRequest{})
		if err == nil {
			t.Fatal("expected forbidden error, got nil")
		}
		var fErr *apperror.ForbiddenError
		if !errors.As(err, &fErr) {
			t.Fatalf("expected ForbiddenError, got %v", err)
		}
		if fErr.Message != "Penilaian medis dokter lain tidak dapat diubah" {
			t.Errorf("expected 'Penilaian medis dokter lain tidak dapat diubah', got: %s", fErr.Message)
		}
	})

	t.Run("Hapus IGD by different doctor", func(t *testing.T) {
		repo := &mockRepository{
			detailIGDData: &PenilaianMedisIGD{
				NoRawat:    "2026/04/22/000002",
				KodeDokter: "DR001",
				NamaDokter: "dr. Budi",
			},
		}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
		svc := NewService(repo, rjSvc, &mockRawatInapService{}, 48, log)

		err := svc.HapusPenilaianMedisIGD(context.Background(), "DR002", "2026/04/22/000002")
		if err == nil {
			t.Fatal("expected forbidden error, got nil")
		}
		var fErr *apperror.ForbiddenError
		if !errors.As(err, &fErr) {
			t.Fatalf("expected ForbiddenError, got %v", err)
		}
		if fErr.Message != "Penilaian medis dokter lain tidak dapat dihapus" {
			t.Errorf("expected 'Penilaian medis dokter lain tidak dapat dihapus', got: %s", fErr.Message)
		}
	})
}
