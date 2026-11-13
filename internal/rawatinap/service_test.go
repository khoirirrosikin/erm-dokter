package rawatinap_test

import (
	"context"
	"errors"
	"testing"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
)

type mockRepository struct {
	cekStatusKamarInapFn     func(ctx context.Context, noRawat string) (bool, bool, error)
	getWaktuKeluarTerakhirFn func(ctx context.Context, noRawat string) (string, string, error)
	daftarPasienRawatInapFn  func(ctx context.Context, kodeDokterLogin string, filter rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, int, error)
	detailPasienRawatInapFn  func(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*rawatinap.KunjunganRawatInap, error)
}

func (m *mockRepository) GetWaktuKeluarTerakhir(ctx context.Context, noRawat string) (string, string, error) {
	if m.getWaktuKeluarTerakhirFn != nil {
		return m.getWaktuKeluarTerakhirFn(ctx, noRawat)
	}
	return "", "", nil
}

func (m *mockRepository) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	if m.cekStatusKamarInapFn != nil {
		return m.cekStatusKamarInapFn(ctx, noRawat)
	}
	return false, false, nil
}

func (m *mockRepository) DaftarPasienRawatInap(ctx context.Context, kodeDokterLogin string, filter rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, int, error) {
	if m.daftarPasienRawatInapFn != nil {
		return m.daftarPasienRawatInapFn(ctx, kodeDokterLogin, filter)
	}
	return nil, 0, nil
}

func (m *mockRepository) DetailPasienRawatInap(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*rawatinap.KunjunganRawatInap, error) {
	if m.detailPasienRawatInapFn != nil {
		return m.detailPasienRawatInapFn(ctx, noRawat, tglMasuk, jamMasuk)
	}
	return nil, nil
}

func (m *mockRepository) GetKelasRawat(ctx context.Context, noRawat string) (string, error) {
	return "", nil
}

func TestService_CekStatusKamarInap(t *testing.T) {
	log := logger.New()

	t.Run("Empty no_rawat returns false, false, nil without calling repo", func(t *testing.T) {
		mockRepo := &mockRepository{
			cekStatusKamarInapFn: func(ctx context.Context, noRawat string) (bool, bool, error) {
				t.Fatal("Repo should not be called for empty no_rawat")
				return false, false, nil
			},
		}
		svc := rawatinap.NewService(mockRepo, log)

		isAktif, hasRecord, err := svc.CekStatusKamarInap(context.Background(), "   ")
		if err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}
		if isAktif || hasRecord {
			t.Fatalf("Expected false, false, got %v, %v", isAktif, hasRecord)
		}
	})

	t.Run("Active inpatient returns true, true", func(t *testing.T) {
		mockRepo := &mockRepository{
			cekStatusKamarInapFn: func(ctx context.Context, noRawat string) (bool, bool, error) {
				if noRawat != "2026/09/04/000001" {
					t.Fatalf("Unexpected noRawat %s", noRawat)
				}
				return true, true, nil
			},
		}
		svc := rawatinap.NewService(mockRepo, log)

		isAktif, hasRecord, err := svc.CekStatusKamarInap(context.Background(), "2026/09/04/000001")
		if err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}
		if !isAktif || !hasRecord {
			t.Fatalf("Expected true, true, got %v, %v", isAktif, hasRecord)
		}
	})

	t.Run("Checked-out inpatient returns false, true", func(t *testing.T) {
		mockRepo := &mockRepository{
			cekStatusKamarInapFn: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, true, nil
			},
		}
		svc := rawatinap.NewService(mockRepo, log)

		isAktif, hasRecord, err := svc.CekStatusKamarInap(context.Background(), "2026/09/04/000001")
		if err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}
		if isAktif || !hasRecord {
			t.Fatalf("Expected false, true, got %v, %v", isAktif, hasRecord)
		}
	})

	t.Run("Non-inpatient returns false, false", func(t *testing.T) {
		mockRepo := &mockRepository{
			cekStatusKamarInapFn: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, false, nil
			},
		}
		svc := rawatinap.NewService(mockRepo, log)

		isAktif, hasRecord, err := svc.CekStatusKamarInap(context.Background(), "2026/09/04/000001")
		if err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}
		if isAktif || hasRecord {
			t.Fatalf("Expected false, false, got %v, %v", isAktif, hasRecord)
		}
	})

	t.Run("Database error propagated", func(t *testing.T) {
		dbErr := errors.New("db connection failure")
		mockRepo := &mockRepository{
			cekStatusKamarInapFn: func(ctx context.Context, noRawat string) (bool, bool, error) {
				return false, false, dbErr
			},
		}
		svc := rawatinap.NewService(mockRepo, log)

		_, _, err := svc.CekStatusKamarInap(context.Background(), "2026/09/04/000001")
		if !errors.Is(err, dbErr) {
			t.Fatalf("Expected %v, got %v", dbErr, err)
		}
	})
}

func TestService_DaftarPasienRawatInap(t *testing.T) {
	log := logger.New()

	t.Run("Success returns list and pagination meta", func(t *testing.T) {
		filter := rawatinap.FilterPasienRawatInap{
			Page:  1,
			Limit: 10,
		}
		mockList := []rawatinap.KunjunganRawatInap{
			{
				NoRawat:      "2026/09/01/000001",
				NamaPasien:   "Pasien A",
				KodeKamar:    "K01",
				TanggalMasuk: "2026-09-01",
				JamMasuk:     "08:00:00",
			},
		}

		mockRepo := &mockRepository{
			daftarPasienRawatInapFn: func(ctx context.Context, kodeDokterLogin string, f rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, int, error) {
				if kodeDokterLogin != "D001" {
					t.Errorf("expected kodeDokterLogin D001, got %s", kodeDokterLogin)
				}
				return mockList, 25, nil
			},
		}

		svc := rawatinap.NewService(mockRepo, log)
		list, meta, err := svc.DaftarPasienRawatInap(context.Background(), " D001 ", filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 item, got %d", len(list))
		}
		if meta.TotalRecords != 25 {
			t.Errorf("expected total records 25, got %d", meta.TotalRecords)
		}
		if meta.TotalPages != 3 {
			t.Errorf("expected total pages 3, got %d", meta.TotalPages)
		}
	})

	t.Run("Repository error propagated", func(t *testing.T) {
		repoErr := errors.New("query failure")
		mockRepo := &mockRepository{
			daftarPasienRawatInapFn: func(ctx context.Context, kodeDokterLogin string, f rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, int, error) {
				return nil, 0, repoErr
			},
		}

		svc := rawatinap.NewService(mockRepo, log)
		_, _, err := svc.DaftarPasienRawatInap(context.Background(), "D001", rawatinap.FilterPasienRawatInap{Page: 1, Limit: 10})
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}

func TestService_DaftarStatusPulang(t *testing.T) {
	log := logger.New()
	svc := rawatinap.NewService(&mockRepository{}, log)
	list := svc.DaftarStatusPulang(context.Background())
	if len(list) != len(rawatinap.ListStatusPulang) {
		t.Errorf("expected %d items, got %d", len(rawatinap.ListStatusPulang), len(list))
	}
	if len(list) != 14 {
		t.Errorf("expected 14 items, got %d", len(list))
	}
}

func TestService_DetailPasienRawatInap(t *testing.T) {
	log := logger.New()

	t.Run("Validation errors on empty params", func(t *testing.T) {
		svc := rawatinap.NewService(&mockRepository{}, log)

		_, err := svc.DetailPasienRawatInap(context.Background(), "", "2026-09-01", "10:00:00")
		if err == nil || err.Error() != "Nomor rawat wajib diisi" {
			t.Fatalf("expected Nomor rawat wajib diisi, got %v", err)
		}

		_, err = svc.DetailPasienRawatInap(context.Background(), "2026/09/01/000001", "", "10:00:00")
		if err == nil || err.Error() != "Tanggal masuk kamar inap wajib diisi" {
			t.Fatalf("expected Tanggal masuk kamar inap wajib diisi, got %v", err)
		}

		_, err = svc.DetailPasienRawatInap(context.Background(), "2026/09/01/000001", "2026-09-01", "")
		if err == nil || err.Error() != "Jam masuk kamar inap wajib diisi" {
			t.Fatalf("expected Jam masuk kamar inap wajib diisi, got %v", err)
		}
	})

	t.Run("Success detail query", func(t *testing.T) {
		mockRepo := &mockRepository{
			detailPasienRawatInapFn: func(ctx context.Context, noRawat, tglMasuk, jamMasuk string) (*rawatinap.KunjunganRawatInap, error) {
				return &rawatinap.KunjunganRawatInap{
					NoRawat:      noRawat,
					TanggalMasuk: tglMasuk,
					JamMasuk:     jamMasuk,
					NamaPasien:   "Budi",
				}, nil
			},
		}

		svc := rawatinap.NewService(mockRepo, log)
		item, err := svc.DetailPasienRawatInap(context.Background(), "2026/09/01/000001", "2026-09-01", "10:00:00")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item == nil || item.NamaPasien != "Budi" {
			t.Fatalf("unexpected item: %+v", item)
		}
	})

	t.Run("Repository error propagated", func(t *testing.T) {
		repoErr := errors.New("db error")
		mockRepo := &mockRepository{
			detailPasienRawatInapFn: func(ctx context.Context, noRawat, tglMasuk, jamMasuk string) (*rawatinap.KunjunganRawatInap, error) {
				return nil, repoErr
			},
		}

		svc := rawatinap.NewService(mockRepo, log)
		_, err := svc.DetailPasienRawatInap(context.Background(), "2026/09/01/000001", "2026-09-01", "10:00:00")
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}

func TestService_GetWaktuKeluarTerakhir(t *testing.T) {
	log := logger.New()

	t.Run("Empty no_rawat returns empty without calling repo", func(t *testing.T) {
		mockRepo := &mockRepository{
			getWaktuKeluarTerakhirFn: func(ctx context.Context, noRawat string) (string, string, error) {
				t.Fatal("repo should not be called")
				return "", "", nil
			},
		}
		svc := rawatinap.NewService(mockRepo, log)
		tgl, jam, err := svc.GetWaktuKeluarTerakhir(context.Background(), "   ")
		if err != nil || tgl != "" || jam != "" {
			t.Fatalf("expected empty result, got (%q, %q, %v)", tgl, jam, err)
		}
	})

	t.Run("Returns trimmed discharge time from repo", func(t *testing.T) {
		mockRepo := &mockRepository{
			getWaktuKeluarTerakhirFn: func(ctx context.Context, noRawat string) (string, string, error) {
				if noRawat != "2026/11/12/000001" {
					t.Fatalf("unexpected noRawat: %s", noRawat)
				}
				return "2026-11-12", "15:00:00", nil
			},
		}
		svc := rawatinap.NewService(mockRepo, log)
		tgl, jam, err := svc.GetWaktuKeluarTerakhir(context.Background(), " 2026/11/12/000001 ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tgl != "2026-11-12" || jam != "15:00:00" {
			t.Fatalf("expected (2026-11-12, 15:00:00), got (%q, %q)", tgl, jam)
		}
	})
}
