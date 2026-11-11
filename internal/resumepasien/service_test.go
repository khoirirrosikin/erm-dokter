package resumepasien_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/resumepasien"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	detailData      *resumepasien.ResumePasienRalan
	detailErr       error
	riwayatData     []resumepasien.ResumePasienRalan
	riwayatErr      error
	adaResult       bool
	adaErr          error
	simpanErr       error
	updateErr       error
	hapusErr        error
	simpanCalledReq resumepasien.SimpanResumePasienRalanRequest
	updateCalledReq resumepasien.UpdateResumePasienRalanRequest

	detailRanapData      *resumepasien.ResumePasienRanap
	detailRanapErr       error
	riwayatRanapData     []resumepasien.ResumePasienRanap
	riwayatRanapErr      error
	adaRanapResult       bool
	adaRanapErr          error
	simpanRanapErr       error
	updateRanapErr       error
	hapusRanapErr        error
	simpanRanapCalledReq resumepasien.SimpanResumePasienRanapRequest
	updateRanapCalledReq resumepasien.UpdateResumePasienRanapRequest
}

func (m *mockRepository) DetailResumePasienRalan(ctx context.Context, noRawat string) (*resumepasien.ResumePasienRalan, error) {
	if m.detailErr != nil {
		return nil, m.detailErr
	}
	if m.detailData != nil {
		return m.detailData, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockRepository) RiwayatResumePasienRalanByNoRM(ctx context.Context, noRM string) ([]resumepasien.ResumePasienRalan, error) {
	if m.riwayatErr != nil {
		return nil, m.riwayatErr
	}
	return m.riwayatData, nil
}

func (m *mockRepository) CekResumePasienRalanAda(ctx context.Context, noRawat string) (bool, error) {
	if m.adaErr != nil {
		return false, m.adaErr
	}
	return m.adaResult, nil
}

func (m *mockRepository) SimpanResumePasienRalan(ctx context.Context, noRawat, kodeDokter string, req resumepasien.SimpanResumePasienRalanRequest) (*resumepasien.ResumePasienRalan, error) {
	if m.simpanErr != nil {
		return nil, m.simpanErr
	}
	m.simpanCalledReq = req
	return &resumepasien.ResumePasienRalan{
		NoRawat:               noRawat,
		KodeDokter:            kodeDokter,
		DataResumePasienRalan: req.DataResumePasienRalan,
	}, nil
}

func (m *mockRepository) UpdateResumePasienRalan(ctx context.Context, noRawat string, req resumepasien.UpdateResumePasienRalanRequest) (*resumepasien.ResumePasienRalan, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.updateCalledReq = req
	return &resumepasien.ResumePasienRalan{
		NoRawat:               noRawat,
		DataResumePasienRalan: req.DataResumePasienRalan,
	}, nil
}

func (m *mockRepository) HapusResumePasienRalan(ctx context.Context, noRawat string) error {
	return m.hapusErr
}

func (m *mockRepository) DetailResumePasienRanap(ctx context.Context, noRawat string) (*resumepasien.ResumePasienRanap, error) {
	if m.detailRanapErr != nil {
		return nil, m.detailRanapErr
	}
	if m.detailRanapData != nil {
		return m.detailRanapData, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockRepository) RiwayatResumePasienRanapByNoRM(ctx context.Context, noRM string) ([]resumepasien.ResumePasienRanap, error) {
	if m.riwayatRanapErr != nil {
		return nil, m.riwayatRanapErr
	}
	return m.riwayatRanapData, nil
}

func (m *mockRepository) CekResumePasienRanapAda(ctx context.Context, noRawat string) (bool, error) {
	if m.adaRanapErr != nil {
		return false, m.adaRanapErr
	}
	return m.adaRanapResult, nil
}

func (m *mockRepository) SimpanResumePasienRanap(ctx context.Context, noRawat, kodeDokter string, req resumepasien.SimpanResumePasienRanapRequest) (*resumepasien.ResumePasienRanap, error) {
	if m.simpanRanapErr != nil {
		return nil, m.simpanRanapErr
	}
	m.simpanRanapCalledReq = req
	return &resumepasien.ResumePasienRanap{
		NoRawat:               noRawat,
		KodeDokter:            kodeDokter,
		DataResumePasienRanap: req.DataResumePasienRanap,
	}, nil
}

func (m *mockRepository) UpdateResumePasienRanap(ctx context.Context, noRawat string, req resumepasien.UpdateResumePasienRanapRequest) (*resumepasien.ResumePasienRanap, error) {
	if m.updateRanapErr != nil {
		return nil, m.updateRanapErr
	}
	m.updateRanapCalledReq = req
	return &resumepasien.ResumePasienRanap{
		NoRawat:               noRawat,
		DataResumePasienRanap: req.DataResumePasienRanap,
	}, nil
}

func (m *mockRepository) HapusResumePasienRanap(ctx context.Context, noRawat string) error {
	return m.hapusRanapErr
}

type mockRawatJalanService struct {
	rawatjalan.Service
	tglReg string
	jamReg string
	exists bool
	err    error
}

func (m *mockRawatJalanService) GetWaktuRegistrasi(ctx context.Context, noRawat string) (string, string, bool, error) {
	if m.err != nil {
		return "", "", false, m.err
	}
	return m.tglReg, m.jamReg, m.exists, nil
}

type mockRawatInapService struct {
	rawatinap.Service
	isAktifRanap   bool
	hasRecordKamar bool
	err            error
}

func (m *mockRawatInapService) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	if m.err != nil {
		return false, false, m.err
	}
	return m.isAktifRanap, m.hasRecordKamar, nil
}

func setupTestService(repo *mockRepository, ralanSvc *mockRawatJalanService, ranapSvc *mockRawatInapService) resumepasien.Service {
	log := logger.New()
	return resumepasien.NewService(repo, ralanSvc, ranapSvc, 48, log)
}

func TestService_DetailResumePasienRalan(t *testing.T) {
	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &resumepasien.ResumePasienRalan{
				NoRawat:    "2026/09/07/000001",
				KodeDokter: "DR01",
				NamaDokter: "dr. Budi",
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		res, err := svc.DetailResumePasienRalan(context.Background(), "2026/09/07/000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.NoRawat != "2026/09/07/000001" {
			t.Errorf("expected NoRawat '2026/09/07/000001', got '%s'", res.NoRawat)
		}
	})

	t.Run("Tidak Ditemukan (404)", func(t *testing.T) {
		repo := &mockRepository{detailData: nil}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.DetailResumePasienRalan(context.Background(), "2026/09/07/999999")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Errorf("expected NotFoundError, got %T: %v", err, err)
		}
	})
}

func TestService_RiwayatResumePasienRalanByNoRM(t *testing.T) {
	repo := &mockRepository{
		riwayatData: []resumepasien.ResumePasienRalan{
			{NoRawat: "RAWAT1"},
			{NoRawat: "RAWAT2"},
		},
	}
	svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

	list, err := svc.RiwayatResumePasienRalanByNoRM(context.Background(), "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 records, got %d", len(list))
	}
}

func TestService_SimpanResumePasienRalan(t *testing.T) {
	now := time.Now()
	tglHariIni := now.Format("2006-01-02")
	jamSekarang := now.Format("15:04:05")

	validReq := resumepasien.SimpanResumePasienRalanRequest{
		NoRawat: "2026/09/07/000001",
		DataResumePasienRalan: resumepasien.DataResumePasienRalan{
			KeluhanUtama:  "Batuk",
			DiagnosaUtama: "ISPA",
			KeadaanPulang: resumepasien.KeadaanPulangHidup,
		},
	}

	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{adaResult: false}
		ralanSvc := &mockRawatJalanService{
			tglReg: tglHariIni,
			jamReg: jamSekarang,
			exists: true,
		}
		ranapSvc := &mockRawatInapService{hasRecordKamar: false}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		res, err := svc.SimpanResumePasienRalan(context.Background(), "2026/09/07/000001", "DR01", validReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.NoRawat != "2026/09/07/000001" {
			t.Errorf("expected NoRawat '2026/09/07/000001', got '%s'", res.NoRawat)
		}
	})

	t.Run("Kunjungan Tidak Ditemukan", func(t *testing.T) {
		repo := &mockRepository{}
		ralanSvc := &mockRawatJalanService{exists: false}
		svc := setupTestService(repo, ralanSvc, &mockRawatInapService{})

		_, err := svc.SimpanResumePasienRalan(context.Background(), "2026/09/07/000001", "DR01", validReq)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Errorf("expected NotFoundError, got: %v", err)
		}
	})

	t.Run("Duplikasi Data Kunjungan", func(t *testing.T) {
		repo := &mockRepository{adaResult: true}
		ralanSvc := &mockRawatJalanService{
			tglReg: tglHariIni,
			jamReg: jamSekarang,
			exists: true,
		}
		svc := setupTestService(repo, ralanSvc, &mockRawatInapService{})

		_, err := svc.SimpanResumePasienRalan(context.Background(), "2026/09/07/000001", "DR01", validReq)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Errorf("expected BusinessError, got: %v", err)
		}
	})

	t.Run("Lewat Batas 48 Jam Rawat Jalan", func(t *testing.T) {
		repo := &mockRepository{adaResult: false}
		ralanSvc := &mockRawatJalanService{
			tglReg: "2020-01-01",
			jamReg: "08:00:00",
			exists: true,
		}
		svc := setupTestService(repo, ralanSvc, &mockRawatInapService{hasRecordKamar: false})

		_, err := svc.SimpanResumePasienRalan(context.Background(), "2026/09/07/000001", "DR01", validReq)
		if err == nil {
			t.Fatal("expected error for expired rawat jalan, got nil")
		}
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("expected ForbiddenError, got: %v", err)
		}
	})

	t.Run("Pasien Sudah Terdaftar di Rawat Inap (Ditolak)", func(t *testing.T) {
		repo := &mockRepository{adaResult: false}
		ralanSvc := &mockRawatJalanService{
			tglReg: tglHariIni,
			jamReg: jamSekarang,
			exists: true,
		}
		ranapSvc := &mockRawatInapService{
			hasRecordKamar: true,
			isAktifRanap:   true,
		}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		_, err := svc.SimpanResumePasienRalan(context.Background(), "2026/09/07/000001", "DR01", validReq)
		if err == nil {
			t.Fatal("expected error for inpatient, got nil")
		}
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Errorf("expected BusinessError, got: %v", err)
		}
	})
}

func TestService_UpdateResumePasienRalan(t *testing.T) {
	now := time.Now()
	tglHariIni := now.Format("2006-01-02")
	jamSekarang := now.Format("15:04:05")

	validUpdateReq := resumepasien.UpdateResumePasienRalanRequest{
		DataResumePasienRalan: resumepasien.DataResumePasienRalan{
			KeluhanUtama:  "Update keluhan",
			DiagnosaUtama: "Update diagnosa",
			KeadaanPulang: resumepasien.KeadaanPulangHidup,
		},
	}

	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &resumepasien.ResumePasienRalan{
				NoRawat:    "2026/09/07/000001",
				KodeDokter: "DR01",
			},
		}
		ralanSvc := &mockRawatJalanService{
			tglReg: tglHariIni,
			jamReg: jamSekarang,
			exists: true,
		}
		svc := setupTestService(repo, ralanSvc, &mockRawatInapService{})

		res, err := svc.UpdateResumePasienRalan(context.Background(), "DR01", "2026/09/07/000001", validUpdateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.NoRawat != "2026/09/07/000001" {
			t.Errorf("expected NoRawat '2026/09/07/000001', got '%s'", res.NoRawat)
		}
	})

	t.Run("Proteksi Dokter Lain (403 Forbidden)", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &resumepasien.ResumePasienRalan{
				NoRawat:    "2026/09/07/000001",
				KodeDokter: "DR01",
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.UpdateResumePasienRalan(context.Background(), "DR99", "2026/09/07/000001", validUpdateReq)
		if err == nil {
			t.Fatal("expected ForbiddenError for different doctor, got nil")
		}
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("expected ForbiddenError, got: %v", err)
		}
	})

	t.Run("Data Tidak Ditemukan", func(t *testing.T) {
		repo := &mockRepository{detailData: nil}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.UpdateResumePasienRalan(context.Background(), "DR01", "2026/09/07/000001", validUpdateReq)
		if err == nil {
			t.Fatal("expected NotFoundError, got nil")
		}
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Errorf("expected NotFoundError, got: %v", err)
		}
	})
}

func TestService_HapusResumePasienRalan(t *testing.T) {
	now := time.Now()
	tglHariIni := now.Format("2006-01-02")
	jamSekarang := now.Format("15:04:05")

	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &resumepasien.ResumePasienRalan{
				NoRawat:    "2026/09/07/000001",
				KodeDokter: "DR01",
			},
		}
		ralanSvc := &mockRawatJalanService{
			tglReg: tglHariIni,
			jamReg: jamSekarang,
			exists: true,
		}
		svc := setupTestService(repo, ralanSvc, &mockRawatInapService{})

		err := svc.HapusResumePasienRalan(context.Background(), "DR01", "2026/09/07/000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Proteksi Dokter Lain (403 Forbidden)", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &resumepasien.ResumePasienRalan{
				NoRawat:    "2026/09/07/000001",
				KodeDokter: "DR01",
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		err := svc.HapusResumePasienRalan(context.Background(), "DR99", "2026/09/07/000001")
		if err == nil {
			t.Fatal("expected ForbiddenError for different doctor, got nil")
		}
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("expected ForbiddenError, got: %v", err)
		}
	})
}

func TestService_DetailResumePasienRanap(t *testing.T) {
	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			detailRanapData: &resumepasien.ResumePasienRanap{
				NoRawat:    "2026/09/07/000002",
				KodeDokter: "DR01",
				NamaDokter: "dr. Budi",
				DataResumePasienRanap: resumepasien.DataResumePasienRanap{
					KeluhanUtama:  "Nyeri dada",
					DiagnosaUtama: "STEMI",
					CaraKeluar:    resumepasien.CaraKeluarAtasIzinDokter,
					KeadaanPulang: resumepasien.KeadaanPulangMembaik,
					Dilanjutkan:   resumepasien.DilanjutkanKembaliKeRS,
				},
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		res, err := svc.DetailResumePasienRanap(context.Background(), "2026/09/07/000002")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.NoRawat != "2026/09/07/000002" {
			t.Errorf("expected NoRawat '2026/09/07/000002', got '%s'", res.NoRawat)
		}
		if res.DiagnosaUtama != "STEMI" {
			t.Errorf("expected DiagnosaUtama 'STEMI', got '%s'", res.DiagnosaUtama)
		}
	})

	t.Run("Data Tidak Ditemukan (404)", func(t *testing.T) {
		repo := &mockRepository{detailRanapData: nil}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.DetailResumePasienRanap(context.Background(), "2026/09/07/000002")
		if err == nil {
			t.Fatal("expected NotFoundError, got nil")
		}
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Errorf("expected NotFoundError, got: %v", err)
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		repo := &mockRepository{detailRanapErr: errors.New("db error")}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.DetailResumePasienRanap(context.Background(), "2026/09/07/000002")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_RiwayatResumePasienRanapByNoRM(t *testing.T) {
	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			riwayatRanapData: []resumepasien.ResumePasienRanap{
				{
					NoRawat: "2026/09/07/000002",
					DataResumePasienRanap: resumepasien.DataResumePasienRanap{
						DiagnosaUtama: "STEMI",
					},
				},
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		list, err := svc.RiwayatResumePasienRanapByNoRM(context.Background(), "123456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 item, got %d", len(list))
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		repo := &mockRepository{riwayatRanapErr: errors.New("db error")}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.RiwayatResumePasienRanapByNoRM(context.Background(), "123456")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_SimpanResumePasienRanap(t *testing.T) {
	validReq := resumepasien.SimpanResumePasienRanapRequest{
		NoRawat: "2026/09/07/000002",
		DataResumePasienRanap: resumepasien.DataResumePasienRanap{
			DiagnosaAwal:  "Chest Pain",
			AlasanRawat:   "Evaluasi nyeri dada",
			KeluhanUtama:  "Nyeri dada kiri menjalar",
			DiagnosaUtama: "STEMI Anterior",
			CaraKeluar:    resumepasien.CaraKeluarAtasIzinDokter,
			KeadaanPulang: resumepasien.KeadaanPulangMembaik,
			Dilanjutkan:   resumepasien.DilanjutkanKembaliKeRS,
			WaktuKontrol:  "2026-09-12 09:00:00",
		},
	}

	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			adaRanapResult: false,
		}
		ralanSvc := &mockRawatJalanService{exists: true}
		ranapSvc := &mockRawatInapService{hasRecordKamar: true}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		res, err := svc.SimpanResumePasienRanap(context.Background(), "2026/09/07/000002", "DR01", validReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.NoRawat != "2026/09/07/000002" {
			t.Errorf("expected NoRawat '2026/09/07/000002', got '%s'", res.NoRawat)
		}
		if res.KodeDokter != "DR01" {
			t.Errorf("expected KodeDokter 'DR01', got '%s'", res.KodeDokter)
		}
	})

	t.Run("Kunjungan Pasien Tidak Ditemukan", func(t *testing.T) {
		repo := &mockRepository{}
		ralanSvc := &mockRawatJalanService{exists: false}
		ranapSvc := &mockRawatInapService{}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		_, err := svc.SimpanResumePasienRanap(context.Background(), "2026/09/07/000002", "DR01", validReq)
		if err == nil {
			t.Fatal("expected NotFoundError, got nil")
		}
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Errorf("expected NotFoundError, got: %v", err)
		}
	})

	t.Run("Pasien Belum Masuk Rawat Inap", func(t *testing.T) {
		repo := &mockRepository{}
		ralanSvc := &mockRawatJalanService{exists: true}
		ranapSvc := &mockRawatInapService{hasRecordKamar: false}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		_, err := svc.SimpanResumePasienRanap(context.Background(), "2026/09/07/000002", "DR01", validReq)
		if err == nil {
			t.Fatal("expected BusinessError, got nil")
		}
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Errorf("expected BusinessError, got: %v", err)
		}
	})

	t.Run("Resume Sudah Ada (Duplikasi 400)", func(t *testing.T) {
		repo := &mockRepository{
			adaRanapResult: true,
		}
		ralanSvc := &mockRawatJalanService{exists: true}
		ranapSvc := &mockRawatInapService{hasRecordKamar: true}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		_, err := svc.SimpanResumePasienRanap(context.Background(), "2026/09/07/000002", "DR01", validReq)
		if err == nil {
			t.Fatal("expected BusinessError for duplicate, got nil")
		}
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Errorf("expected BusinessError, got: %v", err)
		}
	})
}

func TestService_UpdateResumePasienRanap(t *testing.T) {
	validUpdateReq := resumepasien.UpdateResumePasienRanapRequest{
		DataResumePasienRanap: resumepasien.DataResumePasienRanap{
			DiagnosaAwal:  "Chest Pain",
			AlasanRawat:   "Evaluasi nyeri dada",
			KeluhanUtama:  "Nyeri dada berkurang",
			DiagnosaUtama: "STEMI Anterior Resolving",
			CaraKeluar:    resumepasien.CaraKeluarAtasIzinDokter,
			KeadaanPulang: resumepasien.KeadaanPulangSembuh,
			Dilanjutkan:   resumepasien.DilanjutkanKembaliKeRS,
			WaktuKontrol:  "2026-09-12 09:00:00",
		},
	}

	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			detailRanapData: &resumepasien.ResumePasienRanap{
				NoRawat:    "2026/09/07/000002",
				KodeDokter: "DR01",
			},
		}
		ralanSvc := &mockRawatJalanService{exists: true}
		ranapSvc := &mockRawatInapService{hasRecordKamar: true}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		res, err := svc.UpdateResumePasienRanap(context.Background(), "DR01", "2026/09/07/000002", validUpdateReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.NoRawat != "2026/09/07/000002" {
			t.Errorf("expected NoRawat '2026/09/07/000002', got '%s'", res.NoRawat)
		}
	})

	t.Run("Proteksi Dokter Lain (403 Forbidden)", func(t *testing.T) {
		repo := &mockRepository{
			detailRanapData: &resumepasien.ResumePasienRanap{
				NoRawat:    "2026/09/07/000002",
				KodeDokter: "DR01",
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.UpdateResumePasienRanap(context.Background(), "DR99", "2026/09/07/000002", validUpdateReq)
		if err == nil {
			t.Fatal("expected ForbiddenError for different doctor, got nil")
		}
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("expected ForbiddenError, got: %v", err)
		}
	})

	t.Run("Data Tidak Ditemukan", func(t *testing.T) {
		repo := &mockRepository{detailRanapData: nil}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		_, err := svc.UpdateResumePasienRanap(context.Background(), "DR01", "2026/09/07/000002", validUpdateReq)
		if err == nil {
			t.Fatal("expected NotFoundError, got nil")
		}
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Errorf("expected NotFoundError, got: %v", err)
		}
	})
}

func TestService_HapusResumePasienRanap(t *testing.T) {
	t.Run("Sukses", func(t *testing.T) {
		repo := &mockRepository{
			detailRanapData: &resumepasien.ResumePasienRanap{
				NoRawat:    "2026/09/07/000002",
				KodeDokter: "DR01",
			},
		}
		ralanSvc := &mockRawatJalanService{exists: true}
		ranapSvc := &mockRawatInapService{hasRecordKamar: true}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		err := svc.HapusResumePasienRanap(context.Background(), "DR01", "2026/09/07/000002")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Proteksi Dokter Lain (403 Forbidden)", func(t *testing.T) {
		repo := &mockRepository{
			detailRanapData: &resumepasien.ResumePasienRanap{
				NoRawat:    "2026/09/07/000002",
				KodeDokter: "DR01",
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{})

		err := svc.HapusResumePasienRanap(context.Background(), "DR99", "2026/09/07/000002")
		if err == nil {
			t.Fatal("expected ForbiddenError for different doctor, got nil")
		}
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("expected ForbiddenError, got: %v", err)
		}
	})
}

func TestService_Referensi(t *testing.T) {
	svc := setupTestService(&mockRepository{}, &mockRawatJalanService{}, &mockRawatInapService{})

	t.Run("ReferensiRalan", func(t *testing.T) {
		ref := svc.ReferensiRalan(context.Background())
		if len(ref.KeadaanPulang) == 0 {
			t.Error("expected non-empty KeadaanPulang")
		}
	})

	t.Run("ReferensiRanap", func(t *testing.T) {
		ref := svc.ReferensiRanap(context.Background())
		if len(ref.CaraKeluar) == 0 {
			t.Error("expected non-empty CaraKeluar")
		}
		if len(ref.KeadaanPulang) == 0 {
			t.Error("expected non-empty KeadaanPulang")
		}
		if len(ref.Dilanjutkan) == 0 {
			t.Error("expected non-empty Dilanjutkan")
		}
	})
}

func TestService_UpdateAndHapusResumePasien_ErrNoRowsMapsToNotFound(t *testing.T) {
	now := time.Now()
	tglHariIni := now.Format("2006-01-02")
	jamSekarang := now.Format("15:04:05")

	t.Run("UpdateResumePasienRalan maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &resumepasien.ResumePasienRalan{NoRawat: "2026/09/07/000001", KodeDokter: "DR01"},
			updateErr:  sql.ErrNoRows,
		}
		ralanSvc := &mockRawatJalanService{tglReg: tglHariIni, jamReg: jamSekarang, exists: true}
		svc := setupTestService(repo, ralanSvc, &mockRawatInapService{})

		_, err := svc.UpdateResumePasienRalan(context.Background(), "DR01", "2026/09/07/000001", resumepasien.UpdateResumePasienRalanRequest{
			DataResumePasienRalan: resumepasien.DataResumePasienRalan{
				KeluhanUtama:  "Keluhan",
				DiagnosaUtama: "Diagnosa",
				KeadaanPulang: resumepasien.KeadaanPulangHidup,
			},
		})
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})

	t.Run("HapusResumePasienRalan maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		repo := &mockRepository{
			detailData: &resumepasien.ResumePasienRalan{NoRawat: "2026/09/07/000001", KodeDokter: "DR01"},
			hapusErr:   sql.ErrNoRows,
		}
		ralanSvc := &mockRawatJalanService{tglReg: tglHariIni, jamReg: jamSekarang, exists: true}
		svc := setupTestService(repo, ralanSvc, &mockRawatInapService{})

		err := svc.HapusResumePasienRalan(context.Background(), "DR01", "2026/09/07/000001")
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})

	t.Run("UpdateResumePasienRanap maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		repo := &mockRepository{
			detailRanapData: &resumepasien.ResumePasienRanap{NoRawat: "2026/09/07/000002", KodeDokter: "DR01"},
			updateRanapErr:  sql.ErrNoRows,
		}
		ralanSvc := &mockRawatJalanService{exists: true}
		ranapSvc := &mockRawatInapService{hasRecordKamar: true}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		_, err := svc.UpdateResumePasienRanap(context.Background(), "DR01", "2026/09/07/000002", resumepasien.UpdateResumePasienRanapRequest{
			DataResumePasienRanap: resumepasien.DataResumePasienRanap{
				DiagnosaAwal:  "Chest Pain",
				AlasanRawat:   "Evaluasi nyeri dada",
				KeluhanUtama:  "Nyeri dada berkurang",
				DiagnosaUtama: "STEMI Anterior Resolving",
				CaraKeluar:    resumepasien.CaraKeluarAtasIzinDokter,
				KeadaanPulang: resumepasien.KeadaanPulangSembuh,
				Dilanjutkan:   resumepasien.DilanjutkanKembaliKeRS,
				WaktuKontrol:  "2026-09-12 09:00:00",
			},
		})
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})

	t.Run("HapusResumePasienRanap maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		repo := &mockRepository{
			detailRanapData: &resumepasien.ResumePasienRanap{NoRawat: "2026/09/07/000002", KodeDokter: "DR01"},
			hapusRanapErr:   sql.ErrNoRows,
		}
		ralanSvc := &mockRawatJalanService{exists: true}
		ranapSvc := &mockRawatInapService{hasRecordKamar: true}
		svc := setupTestService(repo, ralanSvc, ranapSvc)

		err := svc.HapusResumePasienRanap(context.Background(), "DR01", "2026/09/07/000002")
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})
}
