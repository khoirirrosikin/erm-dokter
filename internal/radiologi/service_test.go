package radiologi_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/radiologi"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
	"erm-dokter/internal/tindakan"
)

type mockRepository struct {
	daftarHasilRadiologiKunjunganFn func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, int, error)
	daftarHasilRadiologiPasienFn    func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, int, error)
	detailHasilRadiologiFn          func(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error)

	simpanPermintaanRadiologiFn          func(ctx context.Context, noRawat string, kodeDokter string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest, kodeTindakanList []string) (string, error)
	daftarPermintaanRadiologiKunjunganFn func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]radiologi.DetailPermintaanRadiologi, error)
	daftarPermintaanRadiologiPasienFn    func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatPermintaanRadiologi) ([]radiologi.DetailPermintaanRadiologi, int, error)
	detailPermintaanRadiologiFn          func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error)
	updatePermintaanRadiologiFn          func(ctx context.Context, noPermintaan string, req radiologi.SimpanPermintaanRadiologiRequest, kodeTindakanList []string) error
	hapusPermintaanRadiologiFn           func(ctx context.Context, noPermintaan string) error
}

func (m *mockRepository) DaftarHasilRadiologiKunjungan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, int, error) {
	if m.daftarHasilRadiologiKunjunganFn != nil {
		return m.daftarHasilRadiologiKunjunganFn(ctx, noRawat, statusLanjut, filter)
	}
	return nil, 0, nil
}

func (m *mockRepository) DaftarHasilRadiologiPasien(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, int, error) {
	if m.daftarHasilRadiologiPasienFn != nil {
		return m.daftarHasilRadiologiPasienFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, 0, nil
}

func (m *mockRepository) DetailHasilRadiologi(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error) {
	if m.detailHasilRadiologiFn != nil {
		return m.detailHasilRadiologiFn(ctx, idHasil)
	}
	return nil, nil
}

func (m *mockRepository) SimpanPermintaanRadiologi(ctx context.Context, noRawat string, kodeDokter string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest, kodeTindakanList []string) (string, error) {
	if m.simpanPermintaanRadiologiFn != nil {
		return m.simpanPermintaanRadiologiFn(ctx, noRawat, kodeDokter, statusLanjut, req, kodeTindakanList)
	}
	return "RAD202609050001", nil
}

func (m *mockRepository) DaftarPermintaanRadiologiKunjungan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]radiologi.DetailPermintaanRadiologi, error) {
	if m.daftarPermintaanRadiologiKunjunganFn != nil {
		return m.daftarPermintaanRadiologiKunjunganFn(ctx, noRawat, statusLanjut)
	}
	return nil, nil
}

func (m *mockRepository) DaftarPermintaanRadiologiPasien(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatPermintaanRadiologi) ([]radiologi.DetailPermintaanRadiologi, int, error) {
	if m.daftarPermintaanRadiologiPasienFn != nil {
		return m.daftarPermintaanRadiologiPasienFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, 0, nil
}

func (m *mockRepository) DetailPermintaanRadiologi(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
	if m.detailPermintaanRadiologiFn != nil {
		return m.detailPermintaanRadiologiFn(ctx, noPermintaan)
	}
	return nil, nil
}

func (m *mockRepository) UpdatePermintaanRadiologi(ctx context.Context, noPermintaan string, req radiologi.SimpanPermintaanRadiologiRequest, kodeTindakanList []string) error {
	if m.updatePermintaanRadiologiFn != nil {
		return m.updatePermintaanRadiologiFn(ctx, noPermintaan, req, kodeTindakanList)
	}
	return nil
}

func (m *mockRepository) HapusPermintaanRadiologi(ctx context.Context, noPermintaan string) error {
	if m.hapusPermintaanRadiologiFn != nil {
		return m.hapusPermintaanRadiologiFn(ctx, noPermintaan)
	}
	return nil
}

type mockRawatJalanService struct {
	rawatjalan.Service
	getInfoRegistrasiFn func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error)
}

func (m *mockRawatJalanService) GetInfoRegistrasi(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
	if m.getInfoRegistrasiFn != nil {
		return m.getInfoRegistrasiFn(ctx, noRawat)
	}
	now := time.Now().Add(-2 * time.Hour)
	return &rawatjalan.InfoRegistrasiPasien{
		TanggalRegistrasi: now.Format("2006-01-02"),
		JamRegistrasi:     now.Format("15:04:05"),
		KodePenjamin:      "UMU",
		StatusBayar:       "Belum Bayar",
	}, nil
}

type mockRawatInapService struct {
	rawatinap.Service
	cekStatusKamarInapFn     func(ctx context.Context, noRawat string) (bool, bool, error)
	getWaktuKeluarTerakhirFn func(ctx context.Context, noRawat string) (string, string, error)
}

func (m *mockRawatInapService) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	if m.cekStatusKamarInapFn != nil {
		return m.cekStatusKamarInapFn(ctx, noRawat)
	}
	return false, false, nil
}

func (m *mockRawatInapService) GetWaktuKeluarTerakhir(ctx context.Context, noRawat string) (string, string, error) {
	if m.getWaktuKeluarTerakhirFn != nil {
		return m.getWaktuKeluarTerakhirFn(ctx, noRawat)
	}
	return "", "", nil
}

type mockTindakanService struct {
	tindakan.Service
	cekKeberadaanTindakanRadiologiFn func(ctx context.Context, listKodeTindakan []string) (map[string]bool, error)
}

func (m *mockTindakanService) CekKeberadaanTindakanRadiologi(ctx context.Context, listKodeTindakan []string) (map[string]bool, error) {
	if m.cekKeberadaanTindakanRadiologiFn != nil {
		return m.cekKeberadaanTindakanRadiologiFn(ctx, listKodeTindakan)
	}
	res := make(map[string]bool)
	for _, k := range listKodeTindakan {
		res[k] = true
	}
	return res, nil
}

func createTestService(repo radiologi.Repository, rj *mockRawatJalanService, ranap *mockRawatInapService, tnd *mockTindakanService) radiologi.Service {
	log := logger.New()
	if rj == nil {
		rj = &mockRawatJalanService{}
	}
	if ranap == nil {
		ranap = &mockRawatInapService{}
	}
	if tnd == nil {
		tnd = &mockTindakanService{}
	}
	return radiologi.NewService(repo, rj, ranap, tnd, 48, log)
}

func TestService_GetRiwayatRadiologiKunjungan_Success(t *testing.T) {
	mockRepo := &mockRepository{
		daftarHasilRadiologiKunjunganFn: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, int, error) {
			if noRawat != "2026/04/22/000001" {
				t.Errorf("Expected noRawat 2026/04/22/000001, got %s", noRawat)
			}
			return []radiologi.HasilRadiologi{
				{
					NoRawat:        noRawat,
					KodeTindakan:   "RAD001",
					NamaTindakan:   "Rontgen Thorax AP/PA",
					Status:         "Ralan",
					TanggalPeriksa: "2026-04-22",
					JamPeriksa:     "10:00:00",
					Hasil:          "Cor dan Pulmo dalam batas normal",
					GambarPACS:     []string{"http://pacs.example.com/viewer?token=xyz"},
				},
			}, 1, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	list, meta, err := svc.DaftarHasilRadiologi(context.Background(), "2026/04/22/000001", shared.StatusLanjutRawatJalan, radiologi.FilterRiwayatRadiologi{Page: 1, Limit: 5})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(list))
	}
	if list[0].KodeTindakan != "RAD001" {
		t.Errorf("Expected RAD001, got %s", list[0].KodeTindakan)
	}
	if len(list[0].GambarPACS) != 1 {
		t.Errorf("Expected 1 PACS image, got %d", len(list[0].GambarPACS))
	}
	if meta.TotalRecords != 1 {
		t.Errorf("Expected 1 total record, got %d", meta.TotalRecords)
	}
}

func TestService_GetRiwayatRadiologiKunjungan_Error(t *testing.T) {
	mockRepo := &mockRepository{
		daftarHasilRadiologiKunjunganFn: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, int, error) {
			return nil, 0, errors.New("database connection failed")
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	_, _, err := svc.DaftarHasilRadiologi(context.Background(), "2026/04/22/000001", shared.StatusLanjutRawatJalan, radiologi.FilterRiwayatRadiologi{})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestService_GetRiwayatRadiologiPasien_Success(t *testing.T) {
	mockRepo := &mockRepository{
		daftarHasilRadiologiPasienFn: func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, int, error) {
			return []radiologi.HasilRadiologi{
				{
					NoRawat:        "2026/04/22/000001",
					KodeTindakan:   "RAD002",
					NamaTindakan:   "USG Abdomen",
					Status:         "Ranap",
					TanggalPeriksa: "2026-04-22",
					JamPeriksa:     "14:00:00",
					Hasil:          "Hepar dan Lien homogen",
				},
			}, 1, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	list, meta, err := svc.DaftarHasilRadiologiByRM(context.Background(), "123456", shared.StatusLanjutRawatInap, radiologi.FilterRiwayatRadiologi{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(list))
	}
	if meta.TotalPages != 1 {
		t.Errorf("Expected 1 total page, got %d", meta.TotalPages)
	}
}

func TestService_GetDetailHasilRadiologi_Success(t *testing.T) {
	mockRepo := &mockRepository{
		detailHasilRadiologiFn: func(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error) {
			return &radiologi.HasilRadiologi{
				NoRawat:        idHasil.NoRawat,
				KodeTindakan:   idHasil.KodeTindakan,
				NamaTindakan:   "Rontgen Thorax",
				Status:         "Ralan",
				TanggalPeriksa: idHasil.TanggalPeriksa,
				JamPeriksa:     idHasil.JamPeriksa,
				Hasil:          "Cor dan Pulmo normal",
				GambarPACS:     []string{"http://pacs.example.com/viewer?token=abc"},
			}, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	idHasil := radiologi.IdHasilRadiologi{
		NoRawat:        "2026/04/22/000001",
		KodeTindakan:   "RAD001",
		TanggalPeriksa: "2026-04-22",
		JamPeriksa:     "10:00:00",
	}
	res, err := svc.DetailHasilRadiologi(context.Background(), idHasil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Hasil != "Cor dan Pulmo normal" {
		t.Errorf("Expected hasil bacaan, got %s", res.Hasil)
	}
}

func TestService_GetDetailHasilRadiologi_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		detailHasilRadiologiFn: func(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error) {
			return nil, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	idHasil := radiologi.IdHasilRadiologi{
		NoRawat:        "2026/04/22/000001",
		KodeTindakan:   "RAD999",
		TanggalPeriksa: "2026-04-22",
		JamPeriksa:     "10:00:00",
	}
	_, err := svc.DetailHasilRadiologi(context.Background(), idHasil)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	var notFoundErr *apperror.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected NotFoundError, got %T: %v", err, err)
	}
}

func TestService_SimpanPermintaanRadiologi_Success(t *testing.T) {
	now := time.Now().Add(-1 * time.Hour)
	tgl := now.Format("2006-01-02")
	jam := now.Format("15:04:05")

	mockRepo := &mockRepository{
		simpanPermintaanRadiologiFn: func(ctx context.Context, noRawat, kodeDokter string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest, kodeTindakanList []string) (string, error) {
			return "RAD202609050001", nil
		},
		detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
			return &radiologi.DetailPermintaanRadiologi{
				PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
					NoPermintaan:      noPermintaan,
					NoRawat:           "2026/09/05/000001",
					TanggalPermintaan: tgl,
					JamPermintaan:     jam,
					DokterPerujuk:     radiologi.DokterInfo{KodeDokter: "DR01", NamaDokter: "dr. SpRad"},
					Status:            "Ralan",
					InformasiTambahan: "Thorax AP",
					DiagnosaKlinis:    "Pneumonia",
				},
				Pemeriksaan: []radiologi.PemeriksaanRadiologiItem{
					{KodeTindakan: "RAD001", NamaTindakan: "Thorax", StatusBayar: "Belum"},
				},
			}, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	req := radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat:           "2026/09/05/000001",
		TanggalPermintaan: tgl,
		JamPermintaan:     jam,
		InformasiTambahan: "Thorax AP",
		DiagnosaKlinis:    "Pneumonia",
		Pemeriksaan: []radiologi.ItemPemeriksaanRadiologiRequest{
			{KodeTindakan: "RAD001"},
		},
	}

	detail, err := svc.SimpanPermintaanRadiologi(context.Background(), "DR01", shared.StatusLanjutRawatJalan, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if detail.NoPermintaan != "RAD202609050001" {
		t.Errorf("Expected NoPermintaan RAD202609050001, got %s", detail.NoPermintaan)
	}
	if len(detail.Pemeriksaan) != 1 {
		t.Errorf("Expected 1 pemeriksaan, got %d", len(detail.Pemeriksaan))
	}
}

func TestService_SimpanPermintaanRadiologi_TindakanNotFound(t *testing.T) {
	now := time.Now().Add(-1 * time.Hour)
	mockRepo := &mockRepository{}
	mockTnd := &mockTindakanService{
		cekKeberadaanTindakanRadiologiFn: func(ctx context.Context, listKodeTindakan []string) (map[string]bool, error) {
			return map[string]bool{"INVALID": false}, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, mockTnd)
	req := radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat:           "2026/09/05/000001",
		TanggalPermintaan: now.Format("2006-01-02"),
		JamPermintaan:     now.Format("15:04:05"),
		Pemeriksaan: []radiologi.ItemPemeriksaanRadiologiRequest{
			{KodeTindakan: "INVALID"},
		},
	}

	_, err := svc.SimpanPermintaanRadiologi(context.Background(), "DR01", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatalf("Expected error for missing tindakan, got nil")
	}
}

func TestService_UpdatePermintaanRadiologi_ForbiddenOtherDoctor(t *testing.T) {
	mockRepo := &mockRepository{
		detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
			return &radiologi.DetailPermintaanRadiologi{
				PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
					NoPermintaan:  noPermintaan,
					NoRawat:       "2026/09/05/000001",
					DokterPerujuk: radiologi.DokterInfo{KodeDokter: "DR02"},
					Status:        "Ralan",
				},
			}, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	req := radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat: "2026/09/05/000001",
	}

	_, err := svc.UpdatePermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD202609050001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatalf("Expected forbidden error, got nil")
	}
}

func TestService_HapusPermintaanRadiologi_LockedWhenSampleTaken(t *testing.T) {
	mockRepo := &mockRepository{
		detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
			return &radiologi.DetailPermintaanRadiologi{
				PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
					NoPermintaan:  noPermintaan,
					NoRawat:       "2026/09/05/000001",
					DokterPerujuk: radiologi.DokterInfo{KodeDokter: "DR01"},
					Status:        "Ralan",
					TanggalSampel: "2026-09-05",
				},
			}, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	err := svc.HapusPermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD202609050001", shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatalf("Expected locked error, got nil")
	}
}

func TestService_DetailPermintaanRadiologi_Success(t *testing.T) {
	mockRepo := &mockRepository{
		detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
			if noPermintaan == "RAD202609050001" {
				return &radiologi.DetailPermintaanRadiologi{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
						NoPermintaan:      noPermintaan,
						NoRawat:           "2026/09/05/000001",
						TanggalPermintaan: "2026-09-05",
						JamPermintaan:     "10:00:00",
						Status:            "Ralan",
					},
				}, nil
			}
			return nil, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	detail, err := svc.DetailPermintaanRadiologi(context.Background(), "RAD202609050001")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if detail.NoPermintaan != "RAD202609050001" {
		t.Errorf("Expected NoPermintaan RAD202609050001, got %s", detail.NoPermintaan)
	}
}

func TestService_DetailPermintaanRadiologi_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
			return nil, nil
		},
	}

	svc := createTestService(mockRepo, nil, nil, nil)
	_, err := svc.DetailPermintaanRadiologi(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Fatalf("Expected NotFound error, got nil")
	}
}

func TestService_UpdateAndHapusPermintaanRadiologi_ErrNoRowsMapsToNotFound(t *testing.T) {
	now := time.Now().Add(-1 * time.Hour)
	tgl := now.Format("2006-01-02")
	jam := now.Format("15:04:05")

	existingDetail := func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
		return &radiologi.DetailPermintaanRadiologi{
			PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
				NoPermintaan:      noPermintaan,
				NoRawat:           "2026/09/05/000001",
				TanggalPermintaan: tgl,
				JamPermintaan:     jam,
				DokterPerujuk:     radiologi.DokterInfo{KodeDokter: "DR01", NamaDokter: "dr. SpRad"},
				Status:            "Ralan",
				TanggalSampel:     "0000-00-00",
				TanggalHasil:      "0000-00-00",
			},
		}, nil
	}

	t.Run("UpdatePermintaanRadiologi maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		mockRepo := &mockRepository{
			detailPermintaanRadiologiFn: existingDetail,
			updatePermintaanRadiologiFn: func(ctx context.Context, noPermintaan string, req radiologi.SimpanPermintaanRadiologiRequest, kodeTindakanList []string) error {
				return sql.ErrNoRows
			},
		}
		svc := createTestService(mockRepo, nil, nil, nil)

		req := radiologi.SimpanPermintaanRadiologiRequest{
			NoRawat:           "2026/09/05/000001",
			TanggalPermintaan: tgl,
			JamPermintaan:     jam,
			DiagnosaKlinis:    "Pneumonia",
			Pemeriksaan: []radiologi.ItemPemeriksaanRadiologiRequest{
				{KodeTindakan: "RAD001"},
			},
		}

		_, err := svc.UpdatePermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD202609050001", shared.StatusLanjutRawatJalan, req)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})

	t.Run("HapusPermintaanRadiologi maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		mockRepo := &mockRepository{
			detailPermintaanRadiologiFn: existingDetail,
			hapusPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) error {
				return sql.ErrNoRows
			},
		}
		svc := createTestService(mockRepo, nil, nil, nil)

		err := svc.HapusPermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD202609050001", shared.StatusLanjutRawatJalan)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})
}

func TestService_Radiologi_Bug14WindowStatusAndMultiPenjamin(t *testing.T) {
	now := time.Now()
	oldReg := now.Add(-50 * time.Hour)
	checkout := now.Add(-12 * time.Hour)

	t.Run("Ralan ditolak ForbiddenError jika > 48 jam dari registrasi meskipun masih aktif di kamar inap", func(t *testing.T) {
		rj := &mockRawatJalanService{
			getInfoRegistrasiFn: func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
				return &rawatjalan.InfoRegistrasiPasien{
					TanggalRegistrasi: oldReg.Format("2006-01-02"),
					JamRegistrasi:     oldReg.Format("15:04:05"),
					KodePenjamin:      "UMU",
					StatusBayar:       "Belum Bayar",
				}, nil
			},
		}
		ri := &mockRawatInapService{
			cekStatusKamarInapFn: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return true, true, nil
			},
		}
		svc := createTestService(&mockRepository{}, rj, ri, nil)
		_, err := svc.SimpanPermintaanRadiologi(context.Background(), "DR01", shared.StatusLanjutRawatJalan, radiologi.SimpanPermintaanRadiologiRequest{
			NoRawat:           "2026/09/05/000001",
			TanggalPermintaan: now.Format("2006-01-02"),
			JamPermintaan:     now.Format("15:04:05"),
			Pemeriksaan:       []radiologi.ItemPemeriksaanRadiologiRequest{{KodeTindakan: "RAD001"}},
		})
		var fErr *apperror.ForbiddenError
		if !errors.As(err, &fErr) {
			t.Fatalf("expected ForbiddenError for ralan > 48h even when active in ranap, got %v", err)
		}
	})

	t.Run("Ranap checkout <= 48 jam diizinkan jika waktu_permintaan <= waktu_keluar, dan ditolak jika > waktu_keluar", func(t *testing.T) {
		rj := &mockRawatJalanService{
			getInfoRegistrasiFn: func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
				return &rawatjalan.InfoRegistrasiPasien{
					TanggalRegistrasi: oldReg.Format("2006-01-02"),
					JamRegistrasi:     oldReg.Format("15:04:05"),
					KodePenjamin:      "UMU",
					StatusBayar:       "Belum Bayar",
				}, nil
			},
		}
		ri := &mockRawatInapService{
			cekStatusKamarInapFn: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, true, nil
			},
			getWaktuKeluarTerakhirFn: func(ctx context.Context, noRawat string) (string, string, error) {
				return checkout.Format("2006-01-02"), checkout.Format("15:04:05"), nil
			},
		}
		validInput := checkout.Add(-1 * time.Hour)
		repo := &mockRepository{
			simpanPermintaanRadiologiFn: func(ctx context.Context, noRawat, kodeDokter string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest, kodeTindakanList []string) (string, error) {
				return "RAD202609050009", nil
			},
			detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
				return &radiologi.DetailPermintaanRadiologi{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{NoPermintaan: noPermintaan, Status: "Ranap"},
				}, nil
			},
		}
		svc := createTestService(repo, rj, ri, nil)
		if _, err := svc.SimpanPermintaanRadiologi(context.Background(), "DR01", shared.StatusLanjutRawatInap, radiologi.SimpanPermintaanRadiologiRequest{
			NoRawat:           "2026/09/05/000001",
			TanggalPermintaan: validInput.Format("2006-01-02"),
			JamPermintaan:     validInput.Format("15:04:05"),
			Pemeriksaan:       []radiologi.ItemPemeriksaanRadiologiRequest{{KodeTindakan: "RAD001"}},
		}); err != nil {
			t.Fatalf("expected allowed post-checkout <= 48h with input <= checkout, got %v", err)
		}

		afterInput := checkout.Add(30 * time.Minute)
		_, err := svc.SimpanPermintaanRadiologi(context.Background(), "DR01", shared.StatusLanjutRawatInap, radiologi.SimpanPermintaanRadiologiRequest{
			NoRawat:           "2026/09/05/000001",
			TanggalPermintaan: afterInput.Format("2006-01-02"),
			JamPermintaan:     afterInput.Format("15:04:05"),
			Pemeriksaan:       []radiologi.ItemPemeriksaanRadiologiRequest{{KodeTindakan: "RAD001"}},
		})
		var valErr apperror.ValidationError
		if !errors.As(err, &valErr) || valErr["tanggal_permintaan"] == "" {
			t.Fatalf("expected ValidationError on tanggal_permintaan when input > checkout, got %v", err)
		}
	})

	t.Run("Update dan Hapus ditolak NotFoundError jika status_lanjut URL tidak cocok, dan multi-penjamin mengunci UMU", func(t *testing.T) {
		repo := &mockRepository{
			detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
				return &radiologi.DetailPermintaanRadiologi{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
						NoPermintaan:  noPermintaan,
						NoRawat:       "2026/09/05/000001",
						DokterPerujuk: radiologi.DokterInfo{KodeDokter: "DR01"},
						Status:        "Ralan",
						TanggalSampel: "0000-00-00",
						TanggalHasil:  "0000-00-00",
					},
				}, nil
			},
		}
		svc := radiologi.NewService(repo, &mockRawatJalanService{
			getInfoRegistrasiFn: func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
				return &rawatjalan.InfoRegistrasiPasien{
					TanggalRegistrasi: now.Add(-1 * time.Hour).Format("2006-01-02"),
					JamRegistrasi:     now.Add(-1 * time.Hour).Format("15:04:05"),
					KodePenjamin:      "UMU",
					StatusBayar:       "Sudah Bayar",
				}, nil
			},
		}, &mockRawatInapService{}, &mockTindakanService{}, 48, logger.New(), "BPJ", "UMU")

		err := svc.HapusPermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD01", shared.StatusLanjutRawatInap)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected NotFoundError on status_lanjut mismatch, got %v", err)
		}

		err = svc.HapusPermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD01", shared.StatusLanjutRawatJalan)
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Fatalf("expected BusinessError when UMU configured in KODE_PENJAMIN_KUNCI_BAYAR, got %v", err)
		}
	})

	t.Run("TanggalSampel/Hasil kosong tidak memicu false lock dan JamSampel non-zero tetap mengunci", func(t *testing.T) {
		rjOk := &mockRawatJalanService{
			getInfoRegistrasiFn: func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
				return &rawatjalan.InfoRegistrasiPasien{
					TanggalRegistrasi: now.Add(-1 * time.Hour).Format("2006-01-02"),
					JamRegistrasi:     now.Add(-1 * time.Hour).Format("15:04:05"),
					KodePenjamin:      "UMU",
					StatusBayar:       "Belum Bayar",
				}, nil
			},
		}

		repoEmpty := &mockRepository{
			detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
				return &radiologi.DetailPermintaanRadiologi{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
						NoPermintaan:  noPermintaan,
						NoRawat:       "2026/09/05/000001",
						DokterPerujuk: radiologi.DokterInfo{KodeDokter: "DR01"},
						Status:        "Ralan",
						TanggalSampel: "",
						JamSampel:     "",
						TanggalHasil:  "",
						JamHasil:      "",
					},
				}, nil
			},
			hapusPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) error {
				return nil
			},
		}
		svcEmpty := createTestService(repoEmpty, rjOk, &mockRawatInapService{}, nil)
		if err := svcEmpty.HapusPermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD01", shared.StatusLanjutRawatJalan); err != nil {
			t.Fatalf("expected empty sampel/hasil fields not to false-lock, got %v", err)
		}

		repoJamOnly := &mockRepository{
			detailPermintaanRadiologiFn: func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
				return &radiologi.DetailPermintaanRadiologi{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
						NoPermintaan:  noPermintaan,
						NoRawat:       "2026/09/05/000001",
						DokterPerujuk: radiologi.DokterInfo{KodeDokter: "DR01"},
						Status:        "Ralan",
						TanggalSampel: "0000-00-00",
						JamSampel:     "00:05:00",
						TanggalHasil:  "0000-00-00",
						JamHasil:      "00:00:00",
					},
				}, nil
			},
		}
		svcJam := createTestService(repoJamOnly, rjOk, &mockRawatInapService{}, nil)
		err := svcJam.HapusPermintaanRadiologi(context.Background(), "DR01", "2026/09/05/000001", "RAD01", shared.StatusLanjutRawatJalan)
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Fatalf("expected ForbiddenError when JamSampel is non-zero, got %v", err)
		}
	})
}

func TestService_SimpanPermintaanRadiologi_DuplicateKodeTindakan(t *testing.T) {
	past := time.Now().Add(-5 * time.Minute)
	svc := createTestService(&mockRepository{}, &mockRawatJalanService{}, &mockRawatInapService{}, &mockTindakanService{})

	_, err := svc.SimpanPermintaanRadiologi(context.Background(), "DR01", shared.StatusLanjutRawatJalan, radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat:           "2026/09/05/000001",
		TanggalPermintaan: past.Format("2006-01-02"),
		JamPermintaan:     past.Format("15:04:05"),
		DiagnosaKlinis:    "Susp Pneumonia",
		Pemeriksaan: []radiologi.ItemPemeriksaanRadiologiRequest{
			{KodeTindakan: "RAD001"},
			{KodeTindakan: "RAD001"},
		},
	})
	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) || valErr["pemeriksaan[1].id_tindakan"] == "" {
		t.Fatalf("expected ValidationError on pemeriksaan[1].id_tindakan, got %v", err)
	}
}
