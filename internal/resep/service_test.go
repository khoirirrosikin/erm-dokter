package resep_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"erm-dokter/internal/obat"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/resep"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	daftarResepFunc              func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error)
	daftarResepByRMFunc          func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error)
	detailResepFunc              func(ctx context.Context, noResep string) (*resep.Resep, error)
	simpanResepFunc              func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error)
	daftarAturanPakaiFunc        func(ctx context.Context, keyword string) ([]resep.AturanPakai, error)
	daftarMetodeRacikFunc        func(ctx context.Context) ([]resep.MetodeRacik, error)
	cekKeberadaanMetodeRacikFunc func(ctx context.Context, listKodeRacik []string) (map[string]bool, error)
	hapusResepFunc               func(ctx context.Context, noResep string) error
	updateResepFunc              func(ctx context.Context, noResep string, req resep.SimpanResepRequest) (*resep.Resep, error)
}

func (m *mockRepository) DaftarResep(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error) {
	if m.daftarResepFunc != nil {
		return m.daftarResepFunc(ctx, noRawat, statusLanjut, filter)
	}
	return []resep.Resep{}, 0, nil
}

func (m *mockRepository) DaftarResepByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error) {
	if m.daftarResepByRMFunc != nil {
		return m.daftarResepByRMFunc(ctx, noRM, statusLanjut, filter)
	}
	return []resep.Resep{}, 0, nil
}

func (m *mockRepository) DetailResep(ctx context.Context, noResep string) (*resep.Resep, error) {
	if m.detailResepFunc != nil {
		return m.detailResepFunc(ctx, noResep)
	}
	return nil, nil
}

func (m *mockRepository) SimpanResep(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error) {
	if m.simpanResepFunc != nil {
		return m.simpanResepFunc(ctx, kodeDokter, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockRepository) DaftarAturanPakai(ctx context.Context, keyword string) ([]resep.AturanPakai, error) {
	if m.daftarAturanPakaiFunc != nil {
		return m.daftarAturanPakaiFunc(ctx, keyword)
	}
	return []resep.AturanPakai{}, nil
}

func (m *mockRepository) DaftarMetodeRacik(ctx context.Context) ([]resep.MetodeRacik, error) {
	if m.daftarMetodeRacikFunc != nil {
		return m.daftarMetodeRacikFunc(ctx)
	}
	return []resep.MetodeRacik{}, nil
}

func (m *mockRepository) CekKeberadaanMetodeRacik(ctx context.Context, listKodeRacik []string) (map[string]bool, error) {
	if m.cekKeberadaanMetodeRacikFunc != nil {
		return m.cekKeberadaanMetodeRacikFunc(ctx, listKodeRacik)
	}
	res := make(map[string]bool)
	for _, k := range listKodeRacik {
		res[k] = true
	}
	return res, nil
}

func (m *mockRepository) HapusResep(ctx context.Context, noResep string) error {
	if m.hapusResepFunc != nil {
		return m.hapusResepFunc(ctx, noResep)
	}
	return nil
}

func (m *mockRepository) UpdateResep(ctx context.Context, noResep string, req resep.SimpanResepRequest) (*resep.Resep, error) {
	if m.updateResepFunc != nil {
		return m.updateResepFunc(ctx, noResep, req)
	}
	return nil, nil
}

type mockRawatJalanService struct {
	rawatjalan.Service
	getWaktuRegistrasiFunc func(ctx context.Context, noRawat string) (string, string, bool, error)
	getInfoRegistrasiFunc  func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error)
}

func (m *mockRawatJalanService) GetWaktuRegistrasi(ctx context.Context, noRawat string) (string, string, bool, error) {
	if m.getWaktuRegistrasiFunc != nil {
		return m.getWaktuRegistrasiFunc(ctx, noRawat)
	}
	return time.Now().Format("2006-01-02"), "08:00:00", true, nil
}

func (m *mockRawatJalanService) GetInfoRegistrasi(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
	if m.getInfoRegistrasiFunc != nil {
		return m.getInfoRegistrasiFunc(ctx, noRawat)
	}
	tgl := time.Now().Format("2006-01-02")
	jam := "08:00:00"
	if m.getWaktuRegistrasiFunc != nil {
		t, j, exists, err := m.getWaktuRegistrasiFunc(ctx, noRawat)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, apperror.NewNotFoundError("Data registrasi kunjungan pasien tidak ditemukan")
		}
		tgl = t
		jam = j
	}
	return &rawatjalan.InfoRegistrasiPasien{
		TanggalRegistrasi: tgl,
		JamRegistrasi:     jam,
		KodePenjamin:      "UMU",
		StatusBayar:       "Belum Bayar",
	}, nil
}

type mockObatService struct {
	obat.Service
	cekKeberadaanObatFunc func(ctx context.Context, listKodeObat []string) (map[string]bool, error)
}

func (m *mockObatService) CekKeberadaanObat(ctx context.Context, listKodeObat []string) (map[string]bool, error) {
	if m.cekKeberadaanObatFunc != nil {
		return m.cekKeberadaanObatFunc(ctx, listKodeObat)
	}
	res := make(map[string]bool)
	for _, k := range listKodeObat {
		res[k] = true
	}
	return res, nil
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

func TestFilterDaftarResep_Validate(t *testing.T) {
	t.Run("Valid tanggal range", func(t *testing.T) {
		f := resep.FilterDaftarResep{
			Tanggal: "2026-04-01,2026-04-30",
			Page:    1,
			Limit:   10,
		}
		if errs := f.Validate(); errs != nil {
			t.Fatalf("expected nil validation errors, got %v", errs)
		}
	})

	t.Run("Invalid tanggal format", func(t *testing.T) {
		f := resep.FilterDaftarResep{
			Tanggal: "invalid-date",
		}
		if errs := f.Validate(); errs == nil {
			t.Fatal("expected validation error for invalid date format, got nil")
		}
	})

	t.Run("Invalid tanggal reversed range", func(t *testing.T) {
		f := resep.FilterDaftarResep{
			Tanggal: "2026-04-30,2026-04-01",
		}
		if errs := f.Validate(); errs == nil {
			t.Fatal("expected validation error for reversed date range, got nil")
		}
	})

	t.Run("Sanitize limit and page", func(t *testing.T) {
		f := resep.FilterDaftarResep{
			Page:  0,
			Limit: 200,
		}
		f.Sanitize()
		if f.Page != 1 {
			t.Errorf("expected page 1, got %d", f.Page)
		}
		if f.Limit != 100 {
			t.Errorf("expected limit 100, got %d", f.Limit)
		}
	})
}

func TestDaftarResep_Success(t *testing.T) {
	mockData := []resep.Resep{
		{
			NoResep:          "202604230001",
			NoRawat:          "2026/04/23/000001",
			TanggalPeresepan: "2026-04-23",
			JamPeresepan:     "10:00:00",
			Status:           "ralan",
			KodeDokter:       "DK001",
			NamaDokter:       "dr. Budi",
			ResepDokter: []resep.ResepDokter{
				{
					ItemObatResep: resep.ItemObatResep{
						KodeObat: "B001",
						NamaObat: "Paracetamol 500mg",
						Jumlah:   10,
						Satuan:   "TAB",
					},
					AturanPakai: "3x1",
				},
			},
			ResepDokterRacikan: []resep.ResepDokterRacikan{},
		},
	}

	mockRepo := &mockRepository{
		daftarResepFunc: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error) {
			return mockData, 1, nil
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	t.Run("Status Ralan", func(t *testing.T) {
		data, meta, err := svc.DaftarResep(ctx, "2026/04/23/000001", shared.StatusLanjutRawatJalan, resep.FilterDaftarResep{Page: 1, Limit: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("expected 1 resep, got %d", len(data))
		}
		if meta.TotalRecords != 1 {
			t.Fatalf("expected total records 1, got %d", meta.TotalRecords)
		}
	})

	t.Run("Status Semua", func(t *testing.T) {
		data, meta, err := svc.DaftarResep(ctx, "2026/04/23/000001", "Semua", resep.FilterDaftarResep{Page: 1, Limit: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("expected 1 resep, got %d", len(data))
		}
		if meta.TotalPages != 1 {
			t.Fatalf("expected total pages 1, got %d", meta.TotalPages)
		}
	})
}

func TestDaftarResep_RepoError(t *testing.T) {
	mockRepo := &mockRepository{
		daftarResepFunc: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error) {
			return nil, 0, errors.New("database error")
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	_, _, err := svc.DaftarResep(ctx, "2026/04/23/000001", shared.StatusLanjutRawatJalan, resep.FilterDaftarResep{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDaftarResepByRM_Success(t *testing.T) {
	mockData := []resep.Resep{
		{
			NoResep:          "202604230002",
			NoRawat:          "2026/04/23/000002",
			TanggalPeresepan: "2026-04-23",
			JamPeresepan:     "11:00:00",
			Status:           "ranap",
			KodeDokter:       "DK002",
			NamaDokter:       "dr. Siti",
			ResepDokter:      []resep.ResepDokter{},
			ResepDokterRacikan: []resep.ResepDokterRacikan{
				{
					NoRacik:         "1",
					NamaRacik:       "Puyer Flu",
					KodeMetodeRacik: "PULV",
					NamaMetodeRacik: "Puyer",
					JumlahRacikan:   10,
					AturanPakai:     "3x1 bungkus",
					Keterangan:      "Sesudah makan",
					DetailRacikan: []resep.ResepDokterRacikanDetail{
						{
							ItemObatResep: resep.ItemObatResep{
								KodeObat: "B002",
								NamaObat: "CTM",
								Jumlah:   2.5,
								Satuan:   "TAB",
							},
							Kandungan: "4mg",
						},
					},
				},
			},
		},
	}

	mockRepo := &mockRepository{
		daftarResepByRMFunc: func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error) {
			return mockData, 1, nil
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	t.Run("Status Ranap", func(t *testing.T) {
		data, meta, err := svc.DaftarResepByRM(ctx, "123456", shared.StatusLanjutRawatInap, resep.FilterDaftarResep{Page: 1, Limit: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("expected 1 resep, got %d", len(data))
		}
		if len(data[0].ResepDokterRacikan) != 1 {
			t.Fatalf("expected 1 racikan, got %d", len(data[0].ResepDokterRacikan))
		}
		if len(data[0].ResepDokterRacikan[0].DetailRacikan) != 1 {
			t.Fatalf("expected 1 detail racikan, got %d", len(data[0].ResepDokterRacikan[0].DetailRacikan))
		}
		if meta.TotalRecords != 1 {
			t.Fatalf("expected total records 1, got %d", meta.TotalRecords)
		}
	})

	t.Run("Status Semua", func(t *testing.T) {
		data, meta, err := svc.DaftarResepByRM(ctx, "123456", "Semua", resep.FilterDaftarResep{Page: 1, Limit: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("expected 1 resep, got %d", len(data))
		}
		if meta.TotalRecords != 1 {
			t.Fatalf("expected total records 1, got %d", meta.TotalRecords)
		}
	})
}

func TestDaftarResepByRM_RepoError(t *testing.T) {
	mockRepo := &mockRepository{
		daftarResepByRMFunc: func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, int, error) {
			return nil, 0, errors.New("database error")
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	_, _, err := svc.DaftarResepByRM(ctx, "123456", shared.StatusLanjutRawatJalan, resep.FilterDaftarResep{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDetailResep_Success(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep: noResep,
				NoRawat: "2026/08/28/000001",
			}, nil
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	res, err := svc.DetailResep(context.Background(), "202608280001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoResep != "202608280001" {
		t.Errorf("expected 202608280001, got %s", res.NoResep)
	}
}

func TestDetailResep_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return nil, nil
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	_, err := svc.DetailResep(context.Background(), "202608280001")
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestSimpanResep_Success(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{
		simpanResepFunc: func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:          "202608280001",
				NoRawat:          req.NoRawat,
				TanggalPeresepan: req.TanggalPeresepan,
				JamPeresepan:     req.JamPeresepan,
				KodeDokter:       kodeDokter,
				Status:           string(statusLanjut),
			}, nil
		},
	}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "07:00:00", true, nil
		},
	}

	log := logger.New()
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "08:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					KodeObat: "OBAT001",
					Jumlah:   10,
				},
				AturanPakai: "3x1",
			},
		},
	}

	res, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoResep != "202608280001" {
		t.Errorf("expected 202608280001, got %s", res.NoResep)
	}
}

func TestSimpanResep_RegistrationNotFound(t *testing.T) {
	mockRepo := &mockRepository{}
	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return "", "", false, nil
		},
	}

	log := logger.New()
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: time.Now().Format("2006-01-02"),
		JamPeresepan:     "08:00:00",
	}

	_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error registration not found, got nil")
	}
}

func TestSimpanResep_WaktuSebelumRegistrasi(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{}
	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "10:00:00", true, nil
		},
	}

	log := logger.New()
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "08:00:00",
	}

	_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error peresepan before registration, got nil")
	}
	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestSimpanResep_Lewat48Jam(t *testing.T) {
	threeDaysAgo := time.Now().Add(-72 * time.Hour).Format("2006-01-02")
	mockRepo := &mockRepository{}
	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return threeDaysAgo, "08:00:00", true, nil
		},
	}

	log := logger.New()
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: time.Now().Format("2006-01-02"),
		JamPeresepan:     "08:00:00",
	}

	_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error 48 hours exceeded, got nil")
	}
}

func TestSimpanResep_PasienBPJSSudahBayar(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRJ := &mockRawatJalanService{
		getInfoRegistrasiFunc: func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
			return &rawatjalan.InfoRegistrasiPasien{
				TanggalRegistrasi: today,
				JamRegistrasi:     "08:00:00",
				KodePenjamin:      "BPJ",
				StatusBayar:       "Sudah Bayar",
			}, nil
		},
	}
	mockRepo := &mockRepository{}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	mockObat := &mockObatService{}
	log := logger.New()
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "09:00:00",
	}
	_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error for paid BPJS patient, got nil")
	}
	if err.Error() != "Pasien BPJS sudah bayar, resep tidak dapat disimpan" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSimpanResep_StatusKamarInap(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}
	log := logger.New()

	t.Run("Pasien aktif ranap tetap bisa buat resep ralan", func(t *testing.T) {
		mockRepo := &mockRepository{
			simpanResepFunc: func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error) {
				return &resep.Resep{NoResep: "202608280099"}, nil
			},
		}
		mockRI := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return true, true, nil
			},
		}
		mockObat := &mockObatService{}
		svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
		req := resep.SimpanResepRequest{
			NoRawat:          "2026/08/28/000001",
			TanggalPeresepan: today,
			JamPeresepan:     "09:00:00",
		}
		res, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
		if err != nil {
			t.Fatalf("expected success for ralan prescription when patient is active in ranap, got: %v", err)
		}
		if res.NoResep != "202608280099" {
			t.Errorf("expected no_resep 202608280099, got %s", res.NoResep)
		}
	})

	t.Run("Pasien checkout ranap tidak bisa buat resep ralan", func(t *testing.T) {
		mockRepo := &mockRepository{}
		mockRI := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, true, nil
			},
		}
		mockObat := &mockObatService{}
		svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
		req := resep.SimpanResepRequest{
			NoRawat:          "2026/08/28/000001",
			TanggalPeresepan: today,
			JamPeresepan:     "09:00:00",
		}
		_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
		if err == nil {
			t.Fatal("expected error patient checked out ranap for ralan prescription, got nil")
		}
		if !strings.Contains(err.Error(), "sudah keluar dari kamar inap") {
			t.Errorf("expected error to mention checkout, got: %v", err)
		}
	})

	t.Run("Pasien checkout ranap tidak bisa buat resep ranap", func(t *testing.T) {
		mockRepo := &mockRepository{}
		mockRI := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, true, nil
			},
		}
		mockObat := &mockObatService{}
		svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
		req := resep.SimpanResepRequest{
			NoRawat:          "2026/08/28/000001",
			TanggalPeresepan: today,
			JamPeresepan:     "09:00:00",
		}
		_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatInap, req)
		if err == nil {
			t.Fatal("expected error patient checked out ranap, got nil")
		}
	})

	t.Run("Pasien tidak pernah ranap mencoba resep ranap", func(t *testing.T) {
		mockRepo := &mockRepository{}
		mockRI := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, false, nil
			},
		}
		mockObat := &mockObatService{}
		svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
		req := resep.SimpanResepRequest{
			NoRawat:          "2026/08/28/000001",
			TanggalPeresepan: today,
			JamPeresepan:     "09:00:00",
		}
		_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatInap, req)
		if err == nil {
			t.Fatal("expected error patient never in ranap, got nil")
		}
	})
}

func TestSimpanResep_ObatNotFound(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}
	mockRepo := &mockRepository{}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	mockObat := &mockObatService{
		cekKeberadaanObatFunc: func(ctx context.Context, listKodeObat []string) (map[string]bool, error) {
			return map[string]bool{
				"OBAT_ADA": true,
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					KodeObat: "OBAT_TIDAK_ADA",
					Jumlah:   10,
				},
				AturanPakai: "3x1",
			},
		},
		ResepRacikan: []resep.ResepRacikanInput{
			{
				NamaRacik:     "Puyer Flu",
				KodeRacik:     "PULV",
				JumlahRacikan: 10,
				AturanPakai:   "3x1",
				Detail: []resep.ResepRacikanDetailInput{
					{
						ItemObatInput: resep.ItemObatInput{
							KodeObat: "OBAT_TIDAK_ADA_2",
							Jumlah:   5,
						},
					},
				},
			},
		},
	}

	_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected ValidationError for non-existent medicines, got nil")
	}

	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if _, exists := valErr["resep_dokter[0].id_obat"]; !exists {
		t.Errorf("expected error for resep_dokter[0].id_obat, got %v", valErr)
	}
	if _, exists := valErr["resep_racikan[0].detail[0].id_obat"]; !exists {
		t.Errorf("expected error for resep_racikan[0].detail[0].id_obat, got %v", valErr)
	}
}

func TestSimpanResep_MetodeRacikNotFound(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}
	mockRepo := &mockRepository{
		cekKeberadaanMetodeRacikFunc: func(ctx context.Context, listKodeRacik []string) (map[string]bool, error) {
			return map[string]bool{
				"PULV": true,
			}, nil
		},
	}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	mockObat := &mockObatService{
		cekKeberadaanObatFunc: func(ctx context.Context, listKodeObat []string) (map[string]bool, error) {
			return map[string]bool{
				"OBAT001": true,
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "09:00:00",
		ResepRacikan: []resep.ResepRacikanInput{
			{
				NamaRacik:     "Puyer Flu",
				KodeRacik:     "METODE_TIDAK_ADA",
				JumlahRacikan: 10,
				AturanPakai:   "3x1",
				Detail: []resep.ResepRacikanDetailInput{
					{
						ItemObatInput: resep.ItemObatInput{
							KodeObat: "OBAT001",
							Jumlah:   5,
						},
					},
				},
			},
		},
	}

	_, err := svc.SimpanResep(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected ValidationError for non-existent metode racik, got nil")
	}

	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if _, exists := valErr["resep_racikan[0].kode_racik"]; !exists {
		t.Errorf("expected error for resep_racikan[0].kode_racik, got %v", valErr)
	}
}

func TestDaftarAturanPakai_Success(t *testing.T) {
	mockData := []resep.AturanPakai{
		{AturanPakai: "3 x 1 tablet sesudah makan"},
		{AturanPakai: "2 x 1 tablet sesudah makan"},
	}

	mockRepo := &mockRepository{
		daftarAturanPakaiFunc: func(ctx context.Context, keyword string) ([]resep.AturanPakai, error) {
			if keyword == "3 x 1" {
				return []resep.AturanPakai{mockData[0]}, nil
			}
			return mockData, nil
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	t.Run("Tanpa keyword", func(t *testing.T) {
		data, err := svc.DaftarAturanPakai(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) != 2 {
			t.Fatalf("expected 2 aturan pakai, got %d", len(data))
		}
	})

	t.Run("Dengan keyword", func(t *testing.T) {
		data, err := svc.DaftarAturanPakai(ctx, "3 x 1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("expected 1 aturan pakai, got %d", len(data))
		}
		if data[0].AturanPakai != "3 x 1 tablet sesudah makan" {
			t.Errorf("expected '3 x 1 tablet sesudah makan', got '%s'", data[0].AturanPakai)
		}
	})
}

func TestDaftarAturanPakai_RepoError(t *testing.T) {
	mockRepo := &mockRepository{
		daftarAturanPakaiFunc: func(ctx context.Context, keyword string) ([]resep.AturanPakai, error) {
			return nil, errors.New("database connection failed")
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	_, err := svc.DaftarAturanPakai(ctx, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDaftarMetodeRacik_Success(t *testing.T) {
	mockData := []resep.MetodeRacik{
		{Kode: "PULV", Nama: "Puyer / Serbuk"},
		{Kode: "CAPS", Nama: "Kapsul"},
	}

	mockRepo := &mockRepository{
		daftarMetodeRacikFunc: func(ctx context.Context) ([]resep.MetodeRacik, error) {
			return mockData, nil
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	data, err := svc.DaftarMetodeRacik(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 metode racik, got %d", len(data))
	}
	if data[0].Kode != "PULV" || data[0].Nama != "Puyer / Serbuk" {
		t.Errorf("expected PULV, got %+v", data[0])
	}
}

func TestDaftarMetodeRacik_RepoError(t *testing.T) {
	mockRepo := &mockRepository{
		daftarMetodeRacikFunc: func(ctx context.Context) ([]resep.MetodeRacik, error) {
			return nil, errors.New("database error")
		},
	}

	log := logger.New()
	mockRJ := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	mockObat := &mockObatService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)
	ctx := context.Background()

	_, err := svc.DaftarMetodeRacik(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHapusResep_Success(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				NamaDokter:        "dr. Budi",
				TanggalPerawatan:  "0000-00-00",
				JamPerawatan:      "00:00:00",
				TanggalPenyerahan: "0000-00-00",
				JamPenyerahan:     "00:00:00",
			}, nil
		},
		hapusResepFunc: func(ctx context.Context, noResep string) error {
			return nil
		},
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, mockRJ, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err != nil {
		t.Fatalf("expected nil error on success, got: %v", err)
	}
}

func TestHapusResep_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return nil, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error for not found resep, got nil")
	}

	var notFoundErr *apperror.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestHapusResep_NoRawatMismatch(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:    "202608280001",
				NoRawat:    "2026/08/28/999999",
				KodeDokter: "DK001",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error for no_rawat mismatch, got nil")
	}

	var notFoundErr *apperror.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestHapusResep_ForbiddenOtherDoctor(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:    "202608280001",
				NoRawat:    "2026/08/28/000001",
				KodeDokter: "DK002",
				NamaDokter: "dr. Lain",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error when deleting other doctor's prescription, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestHapusResep_ForbiddenAlreadyValidatedFarmasi(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:          "202608280001",
				NoRawat:          "2026/08/28/000001",
				KodeDokter:       "DK001",
				TanggalPerawatan: "2026-08-28",
				JamPerawatan:     "10:00:00",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error when prescription is already validated by pharmacy, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestHapusResep_ForbiddenAlreadyHandedOverFarmasi(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				TanggalPenyerahan: "2026-08-28",
				JamPenyerahan:     "11:00:00",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error when prescription is already handed over by pharmacy, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestHapusResep_ForbiddenOver48HoursRalan(t *testing.T) {
	oldDate := time.Now().Add(-50 * time.Hour).Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:    "202608280001",
				NoRawat:    "2026/08/28/000001",
				KodeDokter: "DK001",
			}, nil
		},
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return oldDate, "08:00:00", true, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, mockRJ, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error when deleting prescription over 48 hours for Ralan, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestHapusResep_ForbiddenInactiveRanap(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				Status:            "ranap",
				TanggalPerawatan:  "0000-00-00",
				JamPerawatan:      "00:00:00",
				TanggalPenyerahan: "0000-00-00",
				JamPenyerahan:     "00:00:00",
			}, nil
		},
	}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {

			return false, true, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, mockRI, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatInap)
	if err == nil {
		t.Fatal("expected error when deleting prescription for checked-out Ranap patient, got nil")
	}

	var bizErr *apperror.BusinessError
	if !errors.As(err, &bizErr) {
		t.Fatalf("expected BusinessError, got %T: %v", err, err)
	}
}

func TestHapusResep_ForbiddenBPJSSudahBayar(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				Status:            "ralan",
				TanggalPerawatan:  "0000-00-00",
				JamPerawatan:      "00:00:00",
				TanggalPenyerahan: "0000-00-00",
				JamPenyerahan:     "00:00:00",
			}, nil
		},
	}
	mockRJ := &mockRawatJalanService{
		getInfoRegistrasiFunc: func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
			return &rawatjalan.InfoRegistrasiPasien{
				TanggalRegistrasi: today,
				JamRegistrasi:     "08:00:00",
				KodePenjamin:      "BPJ",
				StatusBayar:       "Sudah Bayar",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, mockRJ, &mockRawatInapService{}, &mockObatService{}, 48, log)

	err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error when deleting prescription for paid BPJS patient, got nil")
	}

	var bizErr *apperror.BusinessError
	if !errors.As(err, &bizErr) {
		t.Fatalf("expected BusinessError, got %T: %v", err, err)
	}
}

func TestUpdateResep_Success(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				NamaDokter:        "dr. Budi",
				TanggalPerawatan:  "0000-00-00",
				JamPerawatan:      "00:00:00",
				TanggalPenyerahan: "0000-00-00",
				JamPenyerahan:     "00:00:00",
			}, nil
		},
		updateResepFunc: func(ctx context.Context, noResep string, req resep.SimpanResepRequest) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:          noResep,
				NoRawat:          req.NoRawat,
				TanggalPeresepan: req.TanggalPeresepan,
				JamPeresepan:     req.JamPeresepan,
				KodeDokter:       "DK001",
			}, nil
		},
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}

	mockObat := &mockObatService{
		cekKeberadaanObatFunc: func(ctx context.Context, listKode []string) (map[string]bool, error) {
			return map[string]bool{"OBAT01": true}, nil
		},
	}

	log := logger.New()
	mockRI := &mockRawatInapService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	res, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err != nil {
		t.Fatalf("expected nil error on success, got: %v", err)
	}
	if res.NoResep != "202608280001" {
		t.Errorf("expected no_resep 202608280001, got %s", res.NoResep)
	}
}

func TestUpdateResep_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return nil, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: "2026-08-28",
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error for not found resep, got nil")
	}

	var notFoundErr *apperror.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUpdateResep_ForbiddenOtherDoctor(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:    "202608280001",
				NoRawat:    "2026/08/28/000001",
				KodeDokter: "DK002",
				NamaDokter: "dr. Lain",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: "2026-08-28",
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error when updating other doctor's prescription, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestUpdateResep_ForbiddenAlreadyValidatedFarmasi(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:          "202608280001",
				NoRawat:          "2026/08/28/000001",
				KodeDokter:       "DK001",
				TanggalPerawatan: "2026-08-28",
				JamPerawatan:     "10:00:00",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: "2026-08-28",
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error when updating validated prescription, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestUpdateResep_ForbiddenAlreadyHandedOverFarmasi(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				TanggalPenyerahan: "2026-08-28",
				JamPenyerahan:     "11:00:00",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: "2026-08-28",
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error when updating handed-over prescription, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestUpdateResep_NoRawatMismatch(t *testing.T) {
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:    "202608280001",
				NoRawat:    "2026/08/28/999999",
				KodeDokter: "DK001",
			}, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockObatService{}, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: "2026-08-28",
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error for no_rawat mismatch, got nil")
	}

	var notFoundErr *apperror.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestUpdateResep_ForbiddenOver48HoursRalan(t *testing.T) {
	oldDate := time.Now().Add(-50 * time.Hour).Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:    "202608280001",
				NoRawat:    "2026/08/28/000001",
				KodeDokter: "DK001",
			}, nil
		},
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return oldDate, "08:00:00", true, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, mockRJ, &mockRawatInapService{}, &mockObatService{}, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: oldDate,
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error when updating prescription over 48 hours for Ralan, got nil")
	}

	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestUpdateResep_ForbiddenInactiveRanap(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				Status:            "ranap",
				TanggalPerawatan:  "0000-00-00",
				JamPerawatan:      "00:00:00",
				TanggalPenyerahan: "0000-00-00",
				JamPenyerahan:     "00:00:00",
			}, nil
		},
	}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {

			return false, true, nil
		},
	}

	log := logger.New()
	svc := resep.NewService(mockRepo, &mockRawatJalanService{}, mockRI, &mockObatService{}, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatInap, req)
	if err == nil {
		t.Fatal("expected error when updating prescription for checked-out Ranap patient, got nil")
	}

	var bizErr *apperror.BusinessError
	if !errors.As(err, &bizErr) {
		t.Fatalf("expected BusinessError, got %T: %v", err, err)
	}
}

func TestUpdateResep_ObatNotFound(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				TanggalPerawatan:  "0000-00-00",
				JamPerawatan:      "00:00:00",
				TanggalPenyerahan: "0000-00-00",
				JamPenyerahan:     "00:00:00",
			}, nil
		},
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}

	mockObat := &mockObatService{
		cekKeberadaanObatFunc: func(ctx context.Context, listKode []string) (map[string]bool, error) {
			return map[string]bool{"OBAT01": false}, nil
		},
	}

	log := logger.New()
	mockRI := &mockRawatInapService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "09:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 10}, AturanPakai: "3 x 1"},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error when obat is not found, got nil")
	}

	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUpdateResep_MetodeRacikNotFound(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	mockRepo := &mockRepository{
		detailResepFunc: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:           "202608280001",
				NoRawat:           "2026/08/28/000001",
				KodeDokter:        "DK001",
				TanggalPerawatan:  "0000-00-00",
				JamPerawatan:      "00:00:00",
				TanggalPenyerahan: "0000-00-00",
				JamPenyerahan:     "00:00:00",
			}, nil
		},
		cekKeberadaanMetodeRacikFunc: func(ctx context.Context, listKodeRacik []string) (map[string]bool, error) {
			return map[string]bool{"NONEXISTENT": false}, nil
		},
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}

	mockObat := &mockObatService{
		cekKeberadaanObatFunc: func(ctx context.Context, listKode []string) (map[string]bool, error) {
			return map[string]bool{"OBAT01": true}, nil
		},
	}

	log := logger.New()
	mockRI := &mockRawatInapService{}
	svc := resep.NewService(mockRepo, mockRJ, mockRI, mockObat, 48, log)

	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: today,
		JamPeresepan:     "09:00:00",
		ResepRacikan: []resep.ResepRacikanInput{
			{
				NamaRacik:     "Puyer A",
				KodeRacik:     "NONEXISTENT",
				JumlahRacikan: 10,
				AturanPakai:   "3x1",
				Detail: []resep.ResepRacikanDetailInput{
					{ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT01", Jumlah: 5}},
				},
			},
		},
	}

	_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error when metode racik is not found, got nil")
	}

	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUpdateAndHapusResep_ErrNoRowsMapsToNotFound(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	existingDetail := func(ctx context.Context, noResep string) (*resep.Resep, error) {
		return &resep.Resep{
			NoResep:           noResep,
			NoRawat:           "2026/08/28/000001",
			KodeDokter:        "DK001",
			Status:            "ralan",
			TanggalPerawatan:  "0000-00-00",
			JamPerawatan:      "00:00:00",
			TanggalPenyerahan: "0000-00-00",
			JamPenyerahan:     "00:00:00",
		}, nil
	}

	mockRJ := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return today, "08:00:00", true, nil
		},
	}

	t.Run("UpdateResep maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		mockRepo := &mockRepository{
			detailResepFunc: existingDetail,
			updateResepFunc: func(ctx context.Context, noResep string, req resep.SimpanResepRequest) (*resep.Resep, error) {
				return nil, sql.ErrNoRows
			},
		}
		svc := resep.NewService(mockRepo, mockRJ, &mockRawatInapService{}, &mockObatService{}, 48, logger.New())

		req := resep.SimpanResepRequest{
			NoRawat:          "2026/08/28/000001",
			TanggalPeresepan: today,
			JamPeresepan:     "09:00:00",
			ResepDokter: []resep.ResepDokterInput{
				{
					ItemObatInput: resep.ItemObatInput{KodeObat: "OBAT001", Jumlah: 10},
					AturanPakai:   "3x1",
				},
			},
		}

		_, err := svc.UpdateResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan, req)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})

	t.Run("HapusResep maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		mockRepo := &mockRepository{
			detailResepFunc: existingDetail,
			hapusResepFunc: func(ctx context.Context, noResep string) error {
				return sql.ErrNoRows
			},
		}
		svc := resep.NewService(mockRepo, mockRJ, &mockRawatInapService{}, &mockObatService{}, 48, logger.New())

		err := svc.HapusResep(context.Background(), "DK001", "2026/08/28/000001", "202608280001", shared.StatusLanjutRawatJalan)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})
}
