package diagnosa_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	"erm-dokter/internal/diagnosa"
	"erm-dokter/internal/master"
	"erm-dokter/internal/pkg/eklaim"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	getDiagnosaByNoRawatFunc     func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error)
	getProsedurByNoRawatFunc     func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error)
	getRiwayatDiagnosaByListFunc func(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error)
	getRiwayatProsedurByListFunc func(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error)
	cekRiwayatPenyakitFunc       func(ctx context.Context, listPastNoRawat []string, kodePenyakit string, status shared.StatusLanjut) (bool, error)
	getMaxPrioritasDiagnosaFunc  func(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error)
	getMaxPrioritasProsedurFunc  func(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error)
	cekDuplikasiDiagnosaFunc     func(ctx context.Context, noRawat string, kodePenyakit string, status shared.StatusLanjut) (bool, error)
	cekDuplikasiProsedurFunc     func(ctx context.Context, noRawat string, kodeProsedur string, status shared.StatusLanjut) (bool, error)
	tambahDiagnosaFunc           func(ctx context.Context, d diagnosa.DiagnosaPasien) error
	tambahProsedurFunc           func(ctx context.Context, p diagnosa.ProsedurPasien) error
	updateDiagnosaFunc           func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas *int, statusPenyakit string) error
	updateProsedurFunc           func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas int) error
	hapusDiagnosaFunc            func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error
	hapusProsedurFunc            func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error
	reorderDiagnosaFunc          func(ctx context.Context, noRawat string, status shared.StatusLanjut, items []diagnosa.ReorderItemRequest) error
	reorderProsedurFunc          func(ctx context.Context, noRawat string, status shared.StatusLanjut, items []diagnosa.ReorderItemRequest) error
}

func (m *mockRepository) GetDiagnosaByNoRawat(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
	if m.getDiagnosaByNoRawatFunc != nil {
		return m.getDiagnosaByNoRawatFunc(ctx, noRawat, status)
	}
	return []diagnosa.DiagnosaPasien{}, nil
}

func (m *mockRepository) GetProsedurByNoRawat(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error) {
	if m.getProsedurByNoRawatFunc != nil {
		return m.getProsedurByNoRawatFunc(ctx, noRawat, status)
	}
	return []diagnosa.ProsedurPasien{}, nil
}

func (m *mockRepository) GetRiwayatDiagnosaByListNoRawat(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
	if m.getRiwayatDiagnosaByListFunc != nil {
		return m.getRiwayatDiagnosaByListFunc(ctx, listNoRawat, status)
	}
	return []diagnosa.DiagnosaPasien{}, nil
}

func (m *mockRepository) GetRiwayatProsedurByListNoRawat(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error) {
	if m.getRiwayatProsedurByListFunc != nil {
		return m.getRiwayatProsedurByListFunc(ctx, listNoRawat, status)
	}
	return []diagnosa.ProsedurPasien{}, nil
}

func (m *mockRepository) CekRiwayatPenyakit(ctx context.Context, listPastNoRawat []string, kodePenyakit string, status shared.StatusLanjut) (bool, error) {
	if m.cekRiwayatPenyakitFunc != nil {
		return m.cekRiwayatPenyakitFunc(ctx, listPastNoRawat, kodePenyakit, status)
	}
	return false, nil
}

func (m *mockRepository) GetMaxPrioritasDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
	if m.getMaxPrioritasDiagnosaFunc != nil {
		return m.getMaxPrioritasDiagnosaFunc(ctx, noRawat, status)
	}
	return 0, nil
}

func (m *mockRepository) GetMaxPrioritasProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
	if m.getMaxPrioritasProsedurFunc != nil {
		return m.getMaxPrioritasProsedurFunc(ctx, noRawat, status)
	}
	return 0, nil
}

func (m *mockRepository) CekDuplikasiDiagnosa(ctx context.Context, noRawat string, kodePenyakit string, status shared.StatusLanjut) (bool, error) {
	if m.cekDuplikasiDiagnosaFunc != nil {
		return m.cekDuplikasiDiagnosaFunc(ctx, noRawat, kodePenyakit, status)
	}
	return false, nil
}

func (m *mockRepository) CekDuplikasiProsedur(ctx context.Context, noRawat string, kodeProsedur string, status shared.StatusLanjut) (bool, error) {
	if m.cekDuplikasiProsedurFunc != nil {
		return m.cekDuplikasiProsedurFunc(ctx, noRawat, kodeProsedur, status)
	}
	return false, nil
}

func (m *mockRepository) TambahDiagnosa(ctx context.Context, d diagnosa.DiagnosaPasien) error {
	if m.tambahDiagnosaFunc != nil {
		return m.tambahDiagnosaFunc(ctx, d)
	}
	return nil
}

func (m *mockRepository) TambahProsedur(ctx context.Context, p diagnosa.ProsedurPasien) error {
	if m.tambahProsedurFunc != nil {
		return m.tambahProsedurFunc(ctx, p)
	}
	return nil
}

func (m *mockRepository) UpdateDiagnosa(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas *int, statusPenyakit string) error {
	if m.updateDiagnosaFunc != nil {
		return m.updateDiagnosaFunc(ctx, noRawat, kode, status, prioritas, statusPenyakit)
	}
	return nil
}

func (m *mockRepository) UpdateProsedur(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas int) error {
	if m.updateProsedurFunc != nil {
		return m.updateProsedurFunc(ctx, noRawat, kode, status, prioritas)
	}
	return nil
}

func (m *mockRepository) HapusDiagnosa(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
	if m.hapusDiagnosaFunc != nil {
		return m.hapusDiagnosaFunc(ctx, noRawat, kode, status)
	}
	return nil
}

func (m *mockRepository) HapusProsedur(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
	if m.hapusProsedurFunc != nil {
		return m.hapusProsedurFunc(ctx, noRawat, kode, status)
	}
	return nil
}

func (m *mockRepository) ReorderDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut, items []diagnosa.ReorderItemRequest) error {
	if m.reorderDiagnosaFunc != nil {
		return m.reorderDiagnosaFunc(ctx, noRawat, status, items)
	}
	return nil
}

func (m *mockRepository) ReorderProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut, items []diagnosa.ReorderItemRequest) error {
	if m.reorderProsedurFunc != nil {
		return m.reorderProsedurFunc(ctx, noRawat, status, items)
	}
	return nil
}

type mockRawatJalanService struct {
	rawatjalan.Service
	riwayatKunjunganPasienFunc func(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error)
	getWaktuRegistrasiFunc     func(ctx context.Context, noRawat string) (string, string, bool, error)
	detailKunjunganFunc        func(ctx context.Context, noRawat, kdDokter string) (*rawatjalan.KunjunganRawatJalan, error)
}

func (m *mockRawatJalanService) RiwayatKunjunganPasien(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
	if m.riwayatKunjunganPasienFunc != nil {
		return m.riwayatKunjunganPasienFunc(ctx, noRM)
	}
	return []rawatjalan.KunjunganRawatJalan{}, nil
}

func (m *mockRawatJalanService) GetWaktuRegistrasi(ctx context.Context, noRawat string) (string, string, bool, error) {
	if m.getWaktuRegistrasiFunc != nil {
		return m.getWaktuRegistrasiFunc(ctx, noRawat)
	}
	now := time.Now()
	return now.Format("2006-01-02"), now.Add(-1 * time.Hour).Format("15:04:05"), true, nil
}

func (m *mockRawatJalanService) DetailKunjungan(ctx context.Context, noRawat, kdDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
	if m.detailKunjunganFunc != nil {
		return m.detailKunjunganFunc(ctx, noRawat, kdDokter)
	}
	return &rawatjalan.KunjunganRawatJalan{
		NoRawat:      noRawat,
		NoRekamMedis: "123456",
		NamaPasien:   "Pasien Uji",
	}, nil
}

type mockRawatInapService struct {
	rawatinap.Service
	cekStatusKamarInapFunc     func(ctx context.Context, noRawat string) (bool, bool, error)
	getWaktuKeluarTerakhirFunc func(ctx context.Context, noRawat string) (string, string, error)
	getKelasRawatFunc          func(ctx context.Context, noRawat string) (string, error)
}

func (m *mockRawatInapService) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	if m.cekStatusKamarInapFunc != nil {
		return m.cekStatusKamarInapFunc(ctx, noRawat)
	}
	return false, false, nil
}

func (m *mockRawatInapService) GetWaktuKeluarTerakhir(ctx context.Context, noRawat string) (string, string, error) {
	if m.getWaktuKeluarTerakhirFunc != nil {
		return m.getWaktuKeluarTerakhirFunc(ctx, noRawat)
	}
	return "", "", nil
}

func (m *mockRawatInapService) GetKelasRawat(ctx context.Context, noRawat string) (string, error) {
	if m.getKelasRawatFunc != nil {
		return m.getKelasRawatFunc(ctx, noRawat)
	}
	return "Kelas 3", nil
}

type mockMasterService struct {
	master.Service
	cekKeberadaanICD10Func func(ctx context.Context, kodeList []string) (map[string]bool, error)
	cekKeberadaanICD9Func  func(ctx context.Context, kodeList []string) (map[string]bool, error)
}

func (m *mockMasterService) CekKeberadaanICD10(ctx context.Context, kodeList []string) (map[string]bool, error) {
	if m.cekKeberadaanICD10Func != nil {
		return m.cekKeberadaanICD10Func(ctx, kodeList)
	}
	res := make(map[string]bool)
	for _, k := range kodeList {
		res[k] = true
	}
	return res, nil
}

func (m *mockMasterService) CekKeberadaanICD9(ctx context.Context, kodeList []string) (map[string]bool, error) {
	if m.cekKeberadaanICD9Func != nil {
		return m.cekKeberadaanICD9Func(ctx, kodeList)
	}
	res := make(map[string]bool)
	for _, k := range kodeList {
		res[k] = true
	}
	return res, nil
}

type mockEklaimClient struct {
	simulasiGrouperFunc func(ctx context.Context, param eklaim.ParameterSimulasi) (*eklaim.HasilSimulasi, error)
}

func (m *mockEklaimClient) SimulasiGrouper(ctx context.Context, param eklaim.ParameterSimulasi) (*eklaim.HasilSimulasi, error) {
	if m.simulasiGrouperFunc != nil {
		return m.simulasiGrouperFunc(ctx, param)
	}
	return &eklaim.HasilSimulasi{
		KodeCBG:       "I-4-17-I",
		DeskripsiCBG:  "GAGAL JANTUNG & SYOK KARDIOGENIK RINGAN",
		Tarif:         4250000,
		BaseTarif:     4250000,
		Kelas:         "rawat_jalan",
		JenisRawat:    "Rawat Jalan",
		SeverityLevel: "I",
	}, nil
}

func setupTestService(
	repo diagnosa.Repository,
	rj rawatjalan.Service,
	ri rawatinap.Service,
	mst master.Service,
) diagnosa.Service {
	return setupTestServiceWithEklaim(repo, rj, ri, mst, &mockEklaimClient{})
}

func setupTestServiceWithEklaim(
	repo diagnosa.Repository,
	rj rawatjalan.Service,
	ri rawatinap.Service,
	mst master.Service,
	ek eklaim.Client,
) diagnosa.Service {
	if ek == nil {
		ek = &mockEklaimClient{}
	}
	return diagnosa.NewService(repo, rj, ri, mst, ek, 48, logger.New())
}

func TestDaftarDiagnosaProsedur(t *testing.T) {
	t.Run("Sukses mengambil daftar diagnosa dan prosedur", func(t *testing.T) {
		repo := &mockRepository{
			getDiagnosaByNoRawatFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
				return []diagnosa.DiagnosaPasien{
					{NoRawat: noRawat, Kode: "I63.9", Nama: "Cerebral infarction", Status: status, Prioritas: 1, StatusPenyakit: "Baru"},
				}, nil
			},
			getProsedurByNoRawatFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error) {
				return []diagnosa.ProsedurPasien{
					{NoRawat: noRawat, Kode: "87.03", Nama: "CT scan of head", Status: status, Prioritas: 1},
				}, nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		res, err := svc.DaftarDiagnosaProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan)
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if len(res.Diagnosa) != 1 || res.Diagnosa[0].Kode != "I63.9" {
			t.Errorf("Expected diagnosa I63.9, got %+v", res.Diagnosa)
		}
		if len(res.Prosedur) != 1 || res.Prosedur[0].Kode != "87.03" {
			t.Errorf("Expected prosedur 87.03, got %+v", res.Prosedur)
		}
	})

	t.Run("Gagal mengambil diagnosa", func(t *testing.T) {
		repo := &mockRepository{
			getDiagnosaByNoRawatFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
				return nil, errors.New("db error")
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		_, err := svc.DaftarDiagnosaProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}

func TestRiwayatPasien(t *testing.T) {
	t.Run("Sukses riwayat pasien dengan kunjungan", func(t *testing.T) {
		rj := &mockRawatJalanService{
			riwayatKunjunganPasienFunc: func(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
				return []rawatjalan.KunjunganRawatJalan{
					{NoRawat: "RAWAT-001"},
					{NoRawat: "RAWAT-002"},
				}, nil
			},
		}
		repo := &mockRepository{
			getRiwayatDiagnosaByListFunc: func(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
				return []diagnosa.DiagnosaPasien{
					{NoRawat: "RAWAT-001", Kode: "I10", Nama: "Essential (primary) hypertension", Status: status, Prioritas: 1, StatusPenyakit: "Baru"},
				}, nil
			},
			getRiwayatProsedurByListFunc: func(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error) {
				return []diagnosa.ProsedurPasien{
					{NoRawat: "RAWAT-001", Kode: "89.52", Nama: "Electrocardiogram", Status: status, Prioritas: 1},
				}, nil
			},
		}
		svc := setupTestService(repo, rj, &mockRawatInapService{}, &mockMasterService{})

		res, err := svc.RiwayatPasien(context.Background(), "123456", shared.StatusLanjutRawatJalan)
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if len(res.Diagnosa) != 1 || len(res.Prosedur) != 1 {
			t.Errorf("Unexpected riwayat result: %+v", res)
		}
	})

	t.Run("Sukses riwayat pasien dengan status Semua", func(t *testing.T) {
		rj := &mockRawatJalanService{
			riwayatKunjunganPasienFunc: func(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
				return []rawatjalan.KunjunganRawatJalan{
					{NoRawat: "RAWAT-001"},
					{NoRawat: "RAWAT-002"},
				}, nil
			},
		}
		repo := &mockRepository{
			getRiwayatDiagnosaByListFunc: func(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
				return []diagnosa.DiagnosaPasien{
					{NoRawat: "RAWAT-001", Kode: "I10", Nama: "Hipertensi", Status: shared.StatusLanjutRawatJalan, Prioritas: 1, StatusPenyakit: "Baru"},
					{NoRawat: "RAWAT-002", Kode: "I63.9", Nama: "Stroke", Status: shared.StatusLanjutRawatInap, Prioritas: 1, StatusPenyakit: "Baru"},
				}, nil
			},
			getRiwayatProsedurByListFunc: func(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error) {
				return []diagnosa.ProsedurPasien{
					{NoRawat: "RAWAT-001", Kode: "89.52", Nama: "EKG", Status: shared.StatusLanjutRawatJalan, Prioritas: 1},
					{NoRawat: "RAWAT-002", Kode: "87.03", Nama: "CT Scan", Status: shared.StatusLanjutRawatInap, Prioritas: 1},
				}, nil
			},
		}
		svc := setupTestService(repo, rj, &mockRawatInapService{}, &mockMasterService{})

		res, err := svc.RiwayatPasien(context.Background(), "123456", "Semua")
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if len(res.Diagnosa) != 2 || len(res.Prosedur) != 2 {
			t.Errorf("Expected 2 diagnosa and 2 prosedur, got %+v", res)
		}
	})

	t.Run("Pasien belum ada riwayat kunjungan", func(t *testing.T) {
		rj := &mockRawatJalanService{
			riwayatKunjunganPasienFunc: func(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
				return []rawatjalan.KunjunganRawatJalan{}, nil
			},
		}
		svc := setupTestService(&mockRepository{}, rj, &mockRawatInapService{}, &mockMasterService{})

		res, err := svc.RiwayatPasien(context.Background(), "123456", shared.StatusLanjutRawatJalan)
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if len(res.Diagnosa) != 0 || len(res.Prosedur) != 0 {
			t.Errorf("Expected empty lists, got %+v", res)
		}
	})
}

func TestTambahDiagnosa(t *testing.T) {
	t.Run("Sukses menambah diagnosa baru", func(t *testing.T) {
		repo := &mockRepository{
			getMaxPrioritasDiagnosaFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
				return 0, nil
			},
			tambahDiagnosaFunc: func(ctx context.Context, d diagnosa.DiagnosaPasien) error {
				if d.Prioritas != 1 {
					t.Errorf("Expected prioritas 1, got %d", d.Prioritas)
				}
				if d.StatusPenyakit != "Baru" {
					t.Errorf("Expected status penyakit 'Baru', got %s", d.StatusPenyakit)
				}
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.TambahDiagnosaRequest{
			Kode: "I63.9",
		}
		res, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if res.Kode != "I63.9" {
			t.Errorf("Expected kode I63.9, got %s", res.Kode)
		}
	})

	t.Run("Diagnosa status penyakit Lama jika pernah terdiagnosa sebelumnya", func(t *testing.T) {
		rj := &mockRawatJalanService{
			detailKunjunganFunc: func(ctx context.Context, noRawat, kdDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
				return &rawatjalan.KunjunganRawatJalan{
					NoRawat:      noRawat,
					NoRekamMedis: "123456",
				}, nil
			},
			riwayatKunjunganPasienFunc: func(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
				return []rawatjalan.KunjunganRawatJalan{
					{NoRawat: "PAST-001"},
					{NoRawat: "2026/09/17/000001"},
				}, nil
			},
		}
		repo := &mockRepository{
			cekRiwayatPenyakitFunc: func(ctx context.Context, listPastNoRawat []string, kodePenyakit string, status shared.StatusLanjut) (bool, error) {
				return true, nil
			},
			tambahDiagnosaFunc: func(ctx context.Context, d diagnosa.DiagnosaPasien) error {
				if d.StatusPenyakit != "Lama" {
					t.Errorf("Expected status penyakit 'Lama', got %s", d.StatusPenyakit)
				}
				return nil
			},
		}
		svc := setupTestService(repo, rj, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.TambahDiagnosaRequest{Kode: "I10"}
		res, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if res.StatusPenyakit != "Lama" {
			t.Errorf("Expected status penyakit 'Lama', got %s", res.StatusPenyakit)
		}
	})

	t.Run("Gagal karena master ICD-10 tidak ditemukan", func(t *testing.T) {
		mst := &mockMasterService{
			cekKeberadaanICD10Func: func(ctx context.Context, kodeList []string) (map[string]bool, error) {
				return map[string]bool{"INVALID": false}, nil
			},
		}
		svc := setupTestService(&mockRepository{}, &mockRawatJalanService{}, &mockRawatInapService{}, mst)

		req := diagnosa.TambahDiagnosaRequest{Kode: "INVALID"}
		_, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("Expected NotFoundError, got %v", err)
		}
	})

	t.Run("Gagal karena duplikasi diagnosa pada kunjungan", func(t *testing.T) {
		repo := &mockRepository{
			cekDuplikasiDiagnosaFunc: func(ctx context.Context, noRawat string, kodePenyakit string, status shared.StatusLanjut) (bool, error) {
				return true, nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.TambahDiagnosaRequest{Kode: "I63.9"}
		_, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("Expected BusinessError, got %v", err)
		}
	})

	t.Run("Gagal karena melebihi batas 48 jam rawat jalan", func(t *testing.T) {
		rj := &mockRawatJalanService{
			getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
				old := time.Now().Add(-50 * time.Hour)
				return old.Format("2006-01-02"), old.Format("15:04:05"), true, nil
			},
		}
		svc := setupTestService(&mockRepository{}, rj, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.TambahDiagnosaRequest{Kode: "I63.9"}
		_, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		var fErr *apperror.ForbiddenError
		if !errors.As(err, &fErr) {
			t.Fatalf("Expected ForbiddenError, got %v", err)
		}
	})

	t.Run("Gagal rawat inap belum terdaftar kamar inap", func(t *testing.T) {
		ri := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, false, nil
			},
		}
		svc := setupTestService(&mockRepository{}, &mockRawatJalanService{}, ri, &mockMasterService{})

		req := diagnosa.TambahDiagnosaRequest{Kode: "I63.9"}
		_, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatInap, req)
		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("Expected BusinessError, got %v", err)
		}
	})

	t.Run("Gagal rawat inap pasien sudah checkout", func(t *testing.T) {
		ri := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, true, nil
			},
		}
		svc := setupTestService(&mockRepository{}, &mockRawatJalanService{}, ri, &mockMasterService{})

		req := diagnosa.TambahDiagnosaRequest{Kode: "I63.9"}
		_, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatInap, req)
		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("Expected BusinessError, got %v", err)
		}
	})

	t.Run("Sukses rawat inap kamar aktif", func(t *testing.T) {
		ri := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return true, true, nil
			},
		}
		svc := setupTestService(&mockRepository{}, &mockRawatJalanService{}, ri, &mockMasterService{})

		req := diagnosa.TambahDiagnosaRequest{Kode: "I63.9"}
		res, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatInap, req)
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if res.Kode != "I63.9" {
			t.Errorf("Expected I63.9, got %s", res.Kode)
		}
	})
}

func TestUpdateDiagnosa(t *testing.T) {
	id := diagnosa.IdDiagnosa{
		NoRawat: "2026/09/17/000001",
		Kode:    "I63.9",
		Status:  shared.StatusLanjutRawatJalan,
	}

	t.Run("Sukses update", func(t *testing.T) {
		prio := 2
		repo := &mockRepository{
			updateDiagnosaFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas *int, statusPenyakit string) error {
				if *prioritas != 2 || statusPenyakit != "Lama" {
					t.Errorf("Unexpected args: %v, %s", prioritas, statusPenyakit)
				}
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		err := svc.UpdateDiagnosa(context.Background(), id, diagnosa.UpdateDiagnosaRequest{
			Prioritas:      &prio,
			StatusPenyakit: "Lama",
		})
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
	})

	t.Run("Item not found", func(t *testing.T) {
		repo := &mockRepository{
			updateDiagnosaFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas *int, statusPenyakit string) error {
				return sql.ErrNoRows
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		err := svc.UpdateDiagnosa(context.Background(), id, diagnosa.UpdateDiagnosaRequest{})
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("Expected NotFoundError, got %v", err)
		}
	})
}

func TestHapusDiagnosa(t *testing.T) {
	id := diagnosa.IdDiagnosa{
		NoRawat: "2026/09/17/000001",
		Kode:    "I63.9",
		Status:  shared.StatusLanjutRawatJalan,
	}

	t.Run("Sukses hapus", func(t *testing.T) {
		repo := &mockRepository{
			hapusDiagnosaFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		if err := svc.HapusDiagnosa(context.Background(), id); err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
	})

	t.Run("Item not found", func(t *testing.T) {
		repo := &mockRepository{
			hapusDiagnosaFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
				return sql.ErrNoRows
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		err := svc.HapusDiagnosa(context.Background(), id)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("Expected NotFoundError, got %v", err)
		}
	})
}

func TestReorderDiagnosa(t *testing.T) {
	t.Run("Sukses reorder", func(t *testing.T) {
		repo := &mockRepository{
			reorderDiagnosaFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut, items []diagnosa.ReorderItemRequest) error {
				if len(items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(items))
				}
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.ReorderRequest{
			Items: []diagnosa.ReorderItemRequest{
				{Kode: "I63.9", Prioritas: 1},
				{Kode: "I10", Prioritas: 2},
			},
		}
		if err := svc.ReorderDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req); err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
	})

	t.Run("Reorder item not found", func(t *testing.T) {
		repo := &mockRepository{
			reorderDiagnosaFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut, items []diagnosa.ReorderItemRequest) error {
				return sql.ErrNoRows
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.ReorderRequest{
			Items: []diagnosa.ReorderItemRequest{{Kode: "I63.9", Prioritas: 1}},
		}
		err := svc.ReorderDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("Expected NotFoundError, got %v", err)
		}
	})
}

func TestTambahProsedur(t *testing.T) {
	t.Run("Sukses menambah prosedur baru", func(t *testing.T) {
		repo := &mockRepository{
			getMaxPrioritasProsedurFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
				return 0, nil
			},
			tambahProsedurFunc: func(ctx context.Context, p diagnosa.ProsedurPasien) error {
				if p.Prioritas != 1 {
					t.Errorf("Expected prioritas 1, got %d", p.Prioritas)
				}
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.TambahProsedurRequest{Kode: "87.03"}
		res, err := svc.TambahProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
		if res.Kode != "87.03" {
			t.Errorf("Expected 87.03, got %s", res.Kode)
		}
	})

	t.Run("Gagal karena master ICD-9 tidak ditemukan", func(t *testing.T) {
		mst := &mockMasterService{
			cekKeberadaanICD9Func: func(ctx context.Context, kodeList []string) (map[string]bool, error) {
				return map[string]bool{"INVALID": false}, nil
			},
		}
		svc := setupTestService(&mockRepository{}, &mockRawatJalanService{}, &mockRawatInapService{}, mst)

		req := diagnosa.TambahProsedurRequest{Kode: "INVALID"}
		_, err := svc.TambahProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("Expected NotFoundError, got %v", err)
		}
	})

	t.Run("Gagal karena duplikasi prosedur", func(t *testing.T) {
		repo := &mockRepository{
			cekDuplikasiProsedurFunc: func(ctx context.Context, noRawat string, kodeProsedur string, status shared.StatusLanjut) (bool, error) {
				return true, nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.TambahProsedurRequest{Kode: "87.03"}
		_, err := svc.TambahProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req)
		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("Expected BusinessError, got %v", err)
		}
	})
}

func TestUpdateProsedur(t *testing.T) {
	id := diagnosa.IdProsedur{
		NoRawat: "2026/09/17/000001",
		Kode:    "87.03",
		Status:  shared.StatusLanjutRawatJalan,
	}

	t.Run("Sukses update", func(t *testing.T) {
		prio := 3
		repo := &mockRepository{
			updateProsedurFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas int) error {
				if prioritas != 3 {
					t.Errorf("Expected prioritas 3, got %d", prioritas)
				}
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		err := svc.UpdateProsedur(context.Background(), id, diagnosa.UpdateProsedurRequest{Prioritas: &prio})
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
	})

	t.Run("Item not found", func(t *testing.T) {
		prio := 1
		repo := &mockRepository{
			updateProsedurFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas int) error {
				return sql.ErrNoRows
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		err := svc.UpdateProsedur(context.Background(), id, diagnosa.UpdateProsedurRequest{Prioritas: &prio})
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("Expected NotFoundError, got %v", err)
		}
	})
}

func TestHapusProsedur(t *testing.T) {
	id := diagnosa.IdProsedur{
		NoRawat: "2026/09/17/000001",
		Kode:    "87.03",
		Status:  shared.StatusLanjutRawatJalan,
	}

	t.Run("Sukses hapus", func(t *testing.T) {
		repo := &mockRepository{
			hapusProsedurFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		if err := svc.HapusProsedur(context.Background(), id); err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
	})

	t.Run("Item not found", func(t *testing.T) {
		repo := &mockRepository{
			hapusProsedurFunc: func(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
				return sql.ErrNoRows
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		err := svc.HapusProsedur(context.Background(), id)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("Expected NotFoundError, got %v", err)
		}
	})
}

func TestReorderProsedur(t *testing.T) {
	t.Run("Sukses reorder", func(t *testing.T) {
		repo := &mockRepository{
			reorderProsedurFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut, items []diagnosa.ReorderItemRequest) error {
				if len(items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(items))
				}
				return nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		req := diagnosa.ReorderRequest{
			Items: []diagnosa.ReorderItemRequest{
				{Kode: "87.03", Prioritas: 1},
				{Kode: "89.52", Prioritas: 2},
			},
		}
		if err := svc.ReorderProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, req); err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
	})
}

func TestSimulasiEklaim(t *testing.T) {
	t.Run("Sukses simulasi dengan draft diagnosa dan prosedur", func(t *testing.T) {
		rj := &mockRawatJalanService{
			detailKunjunganFunc: func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
				return &rawatjalan.KunjunganRawatJalan{
					NoRawat:           noRawat,
					NoRekamMedis:      "123456",
					NamaPasien:        "Budi Santoso",
					JenisKelamin:      "L",
					TanggalLahir:      "1990-01-01",
					TanggalRegistrasi: "2026-09-18",
					NoPeserta:         "0001234567890",
					StatusLanjut:      shared.StatusLanjutRawatJalan,
					NamaDokterAsal:    "dr. Ahmad",
				}, nil
			},
		}

		mockEk := &mockEklaimClient{
			simulasiGrouperFunc: func(ctx context.Context, param eklaim.ParameterSimulasi) (*eklaim.HasilSimulasi, error) {
				if len(param.Diagnosa) != 2 || param.Diagnosa[0] != "I50.9" {
					t.Errorf("Unexpected param.Diagnosa: %v", param.Diagnosa)
				}
				if len(param.Prosedur) != 1 || param.Prosedur[0] != "88.72" {
					t.Errorf("Unexpected param.Prosedur: %v", param.Prosedur)
				}
				return &eklaim.HasilSimulasi{
					KodeCBG:       "I-4-17-I",
					DeskripsiCBG:  "GAGAL JANTUNG & SYOK KARDIOGENIK RINGAN",
					Tarif:         4250000,
					BaseTarif:     4250000,
					Kelas:         "rawat_jalan",
					JenisRawat:    "Rawat Jalan",
					SeverityLevel: "I",
				}, nil
			},
		}

		svc := setupTestServiceWithEklaim(&mockRepository{}, rj, &mockRawatInapService{}, &mockMasterService{}, mockEk)

		req := diagnosa.SimulasiEklaimRequest{
			Diagnosa: []string{"I50.9", "I10"},
			Prosedur: []string{"88.72"},
		}

		res, err := svc.SimulasiEklaim(context.Background(), "2026/09/18/000001", shared.StatusLanjutRawatJalan, req)
		if err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}

		if res.KodeCBG != "I-4-17-I" {
			t.Errorf("Expected KodeCBG 'I-4-17-I', got '%s'", res.KodeCBG)
		}
		if res.Tarif != 4250000 {
			t.Errorf("Expected Tarif 4250000, got %d", res.Tarif)
		}
	})

	t.Run("Sukses simulasi dengan diagnosa tersimpan", func(t *testing.T) {
		rj := &mockRawatJalanService{
			detailKunjunganFunc: func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
				return &rawatjalan.KunjunganRawatJalan{
					NoRawat:           noRawat,
					NoRekamMedis:      "123456",
					NamaPasien:        "Siti Aminah",
					JenisKelamin:      "P",
					TanggalLahir:      "1985-05-15",
					TanggalRegistrasi: "2026-09-18",
					NoPeserta:         "0009876543210",
					StatusLanjut:      shared.StatusLanjutRawatJalan,
				}, nil
			},
		}

		repo := &mockRepository{
			getDiagnosaByNoRawatFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
				return []diagnosa.DiagnosaPasien{
					{Kode: "E11.9", Status: status},
				}, nil
			},
			getProsedurByNoRawatFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.ProsedurPasien, error) {
				return []diagnosa.ProsedurPasien{}, nil
			},
		}

		mockEk := &mockEklaimClient{
			simulasiGrouperFunc: func(ctx context.Context, param eklaim.ParameterSimulasi) (*eklaim.HasilSimulasi, error) {
				if param.Gender != "2" {
					t.Errorf("Expected Gender '2' (Perempuan), got '%s'", param.Gender)
				}
				if len(param.Diagnosa) != 1 || param.Diagnosa[0] != "E11.9" {
					t.Errorf("Unexpected param.Diagnosa: %v", param.Diagnosa)
				}
				return &eklaim.HasilSimulasi{
					KodeCBG:       "E-4-10-I",
					DeskripsiCBG:  "DIABETES MELLITUS RINGAN",
					Tarif:         3200000,
					BaseTarif:     3200000,
					Kelas:         "rawat_jalan",
					JenisRawat:    "Rawat Jalan",
					SeverityLevel: "I",
				}, nil
			},
		}

		svc := setupTestServiceWithEklaim(repo, rj, &mockRawatInapService{}, &mockMasterService{}, mockEk)

		req := diagnosa.SimulasiEklaimRequest{}

		res, err := svc.SimulasiEklaim(context.Background(), "2026/09/18/000001", shared.StatusLanjutRawatJalan, req)
		if err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}

		if res.KodeCBG != "E-4-10-I" {
			t.Errorf("Expected KodeCBG 'E-4-10-I', got '%s'", res.KodeCBG)
		}
	})

	t.Run("Gagal simulasi jika tidak ada diagnosa sama sekali", func(t *testing.T) {
		rj := &mockRawatJalanService{
			detailKunjunganFunc: func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
				return &rawatjalan.KunjunganRawatJalan{
					NoRawat:      noRawat,
					StatusLanjut: shared.StatusLanjutRawatJalan,
				}, nil
			},
		}

		repo := &mockRepository{
			getDiagnosaByNoRawatFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]diagnosa.DiagnosaPasien, error) {
				return []diagnosa.DiagnosaPasien{}, nil
			},
		}

		svc := setupTestService(repo, rj, &mockRawatInapService{}, &mockMasterService{})

		_, err := svc.SimulasiEklaim(context.Background(), "2026/09/18/000001", shared.StatusLanjutRawatJalan, diagnosa.SimulasiEklaimRequest{})
		if err == nil {
			t.Fatal("Expected error for empty diagnoses, got nil")
		}

		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("Expected BusinessError, got %v", err)
		}
	})

	t.Run("Gagal simulasi jika E-Klaim client mengembalikan error", func(t *testing.T) {
		rj := &mockRawatJalanService{
			detailKunjunganFunc: func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
				return &rawatjalan.KunjunganRawatJalan{
					NoRawat:      noRawat,
					StatusLanjut: shared.StatusLanjutRawatJalan,
				}, nil
			},
		}

		mockEk := &mockEklaimClient{
			simulasiGrouperFunc: func(ctx context.Context, param eklaim.ParameterSimulasi) (*eklaim.HasilSimulasi, error) {
				return nil, errors.New("connection timed out")
			},
		}

		svc := setupTestServiceWithEklaim(&mockRepository{}, rj, &mockRawatInapService{}, &mockMasterService{}, mockEk)

		req := diagnosa.SimulasiEklaimRequest{
			Diagnosa: []string{"I50.9"},
		}

		_, err := svc.SimulasiEklaim(context.Background(), "2026/09/18/000001", shared.StatusLanjutRawatJalan, req)
		if err == nil {
			t.Fatal("Expected error when eklaim fails, got nil")
		}

		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("Expected BusinessError, got %v", err)
		}
	})

	t.Run("Gagal simulasi ranap jika pasien belum terdaftar di kamar inap", func(t *testing.T) {
		rj := &mockRawatJalanService{
			detailKunjunganFunc: func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
				return &rawatjalan.KunjunganRawatJalan{
					NoRawat:      noRawat,
					StatusLanjut: shared.StatusLanjutRawatJalan,
				}, nil
			},
		}
		ri := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, false, nil
			},
		}
		svc := setupTestService(&mockRepository{}, rj, ri, &mockMasterService{})
		_, err := svc.SimulasiEklaim(context.Background(), "2026/09/18/000001", shared.StatusLanjutRawatInap, diagnosa.SimulasiEklaimRequest{Diagnosa: []string{"I50.9"}})
		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("Expected BusinessError for ranap without kamar_inap, got %v", err)
		}
	})

	t.Run("Simulasi ranap memakai TanggalPulang hari ini dan meneruskan context pembatalan", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		regDate := time.Now().Add(-3 * 24 * time.Hour).Format("2006-01-02")

		rj := &mockRawatJalanService{
			detailKunjunganFunc: func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
				return &rawatjalan.KunjunganRawatJalan{
					NoRawat:           noRawat,
					TanggalRegistrasi: regDate,
					StatusLanjut:      shared.StatusLanjutRawatInap,
				}, nil
			},
		}
		ri := &mockRawatInapService{
			getKelasRawatFunc: func(ctx context.Context, noRawat string) (string, error) {
				return "Kelas 2", nil
			},
		}

		var gotTglMasuk, gotTglPulang string
		mockEk := &mockEklaimClient{
			simulasiGrouperFunc: func(ctx context.Context, param eklaim.ParameterSimulasi) (*eklaim.HasilSimulasi, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				gotTglMasuk = param.TanggalMasuk
				gotTglPulang = param.TanggalPulang
				return &eklaim.HasilSimulasi{KodeCBG: "I-1-10-I", Tarif: 5000000}, nil
			},
		}

		svc := setupTestServiceWithEklaim(&mockRepository{}, rj, ri, &mockMasterService{}, mockEk)
		if _, err := svc.SimulasiEklaim(context.Background(), "2026/09/18/000001", shared.StatusLanjutRawatInap, diagnosa.SimulasiEklaimRequest{Diagnosa: []string{"I50.9"}}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotTglMasuk != regDate || gotTglPulang != today {
			t.Errorf("expected TanggalMasuk=%s TanggalPulang=%s, got Masuk=%s Pulang=%s", regDate, today, gotTglMasuk, gotTglPulang)
		}

		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := svc.SimulasiEklaim(canceledCtx, "2026/09/18/000001", shared.StatusLanjutRawatInap, diagnosa.SimulasiEklaimRequest{Diagnosa: []string{"I50.9"}}); err == nil {
			t.Fatal("expected canceled context to fail SimulasiEklaim")
		}
	})
}

func TestTambahDiagnosa_PostCheckoutRanapGraceWindow(t *testing.T) {
	oldReg := time.Now().Add(-5 * 24 * time.Hour)
	rj := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return oldReg.Format("2006-01-02"), oldReg.Format("15:04:05"), true, nil
		},
	}

	t.Run("Diizinkan pada ranap dan ralan jika checkout <= 48 jam", func(t *testing.T) {
		checkout := time.Now().Add(-12 * time.Hour)
		ri := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, true, nil
			},
			getWaktuKeluarTerakhirFunc: func(ctx context.Context, noRawat string) (string, string, error) {
				return checkout.Format("2006-01-02"), checkout.Format("15:04:05"), nil
			},
		}
		svc := setupTestService(&mockRepository{}, rj, ri, &mockMasterService{})

		for _, st := range []shared.StatusLanjut{shared.StatusLanjutRawatInap, shared.StatusLanjutRawatJalan} {
			if _, err := svc.TambahDiagnosa(context.Background(), "2026/09/10/000001", st, diagnosa.TambahDiagnosaRequest{Kode: "I10"}); err != nil {
				t.Fatalf("expected allowed for %s within +48h post-checkout, got %v", st, err)
			}
		}
	})

	t.Run("Ditolak ForbiddenError jika checkout > 48 jam", func(t *testing.T) {
		checkout := time.Now().Add(-50 * time.Hour)
		ri := &mockRawatInapService{
			cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, true, nil
			},
			getWaktuKeluarTerakhirFunc: func(ctx context.Context, noRawat string) (string, string, error) {
				return checkout.Format("2006-01-02"), checkout.Format("15:04:05"), nil
			},
		}
		svc := setupTestService(&mockRepository{}, rj, ri, &mockMasterService{})

		for _, st := range []shared.StatusLanjut{shared.StatusLanjutRawatInap, shared.StatusLanjutRawatJalan} {
			_, err := svc.TambahDiagnosa(context.Background(), "2026/09/10/000001", st, diagnosa.TambahDiagnosaRequest{Kode: "I10"})
			var fErr *apperror.ForbiddenError
			if !errors.As(err, &fErr) {
				t.Fatalf("expected ForbiddenError for %s > 48h post-checkout, got %v", st, err)
			}
		}
	})
}

func TestTambahDiagnosaDanProsedur_DuplicateEntry1062AndMaxPrioritas(t *testing.T) {
	dupErr := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}

	t.Run("TambahDiagnosa race condition 1062 mengembalikan BusinessError", func(t *testing.T) {
		repo := &mockRepository{
			cekDuplikasiDiagnosaFunc: func(ctx context.Context, noRawat, kodePenyakit string, status shared.StatusLanjut) (bool, error) {
				return false, nil
			},
			tambahDiagnosaFunc: func(ctx context.Context, d diagnosa.DiagnosaPasien) error {
				return dupErr
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})
		_, err := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, diagnosa.TambahDiagnosaRequest{Kode: "I10"})
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Fatalf("expected *apperror.BusinessError on MySQL 1062, got %T (%v)", err, err)
		}
	})

	t.Run("TambahProsedur race condition 1062 mengembalikan BusinessError", func(t *testing.T) {
		repo := &mockRepository{
			cekDuplikasiProsedurFunc: func(ctx context.Context, noRawat, kodeProsedur string, status shared.StatusLanjut) (bool, error) {
				return false, nil
			},
			tambahProsedurFunc: func(ctx context.Context, p diagnosa.ProsedurPasien) error {
				return dupErr
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})
		_, err := svc.TambahProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, diagnosa.TambahProsedurRequest{Kode: "87.03"})
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Fatalf("expected *apperror.BusinessError on MySQL 1062, got %T (%v)", err, err)
		}
	})

	t.Run("Auto-increment prioritas menolak jika sudah mencapai 127", func(t *testing.T) {
		repo := &mockRepository{
			getMaxPrioritasDiagnosaFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
				return 127, nil
			},
			getMaxPrioritasProsedurFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
				return 127, nil
			},
		}
		svc := setupTestService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, &mockMasterService{})

		_, errDiag := svc.TambahDiagnosa(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, diagnosa.TambahDiagnosaRequest{Kode: "I10"})
		var bizDiag *apperror.BusinessError
		if !errors.As(errDiag, &bizDiag) {
			t.Fatalf("expected BusinessError when diagnosa max prioritas is 127, got %v", errDiag)
		}

		_, errPros := svc.TambahProsedur(context.Background(), "2026/09/17/000001", shared.StatusLanjutRawatJalan, diagnosa.TambahProsedurRequest{Kode: "87.03"})
		var bizPros *apperror.BusinessError
		if !errors.As(errPros, &bizPros) {
			t.Fatalf("expected BusinessError when prosedur max prioritas is 127, got %v", errPros)
		}
	})
}
