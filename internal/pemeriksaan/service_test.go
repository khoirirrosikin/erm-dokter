package pemeriksaan_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"erm-dokter/internal/pemeriksaan"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	daftarPemeriksaanFunc func(ctx context.Context, listNoRawat []string, statusLanjut shared.StatusLanjut, filter pemeriksaan.FilterDaftarPemeriksaan) ([]pemeriksaan.Pemeriksaan, int, error)
	detailPemeriksaanFunc func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error)
	simpanPemeriksaanFunc func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req pemeriksaan.SimpanPemeriksaanRequest) error
	updatePemeriksaanFunc func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut, req pemeriksaan.UpdatePemeriksaanRequest) error
	hapusPemeriksaanFunc  func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) error
}

func (m *mockRepository) DaftarPemeriksaan(ctx context.Context, listNoRawat []string, statusLanjut shared.StatusLanjut, filter pemeriksaan.FilterDaftarPemeriksaan) ([]pemeriksaan.Pemeriksaan, int, error) {
	if m.daftarPemeriksaanFunc != nil {
		return m.daftarPemeriksaanFunc(ctx, listNoRawat, statusLanjut, filter)
	}
	return []pemeriksaan.Pemeriksaan{}, 0, nil
}

func (m *mockRepository) DetailPemeriksaan(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
	if m.detailPemeriksaanFunc != nil {
		return m.detailPemeriksaanFunc(ctx, id, statusLanjut)
	}
	return nil, nil
}

func (m *mockRepository) SimpanPemeriksaan(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req pemeriksaan.SimpanPemeriksaanRequest) error {
	if m.simpanPemeriksaanFunc != nil {
		return m.simpanPemeriksaanFunc(ctx, kodeDokter, statusLanjut, req)
	}
	return nil
}

func (m *mockRepository) UpdatePemeriksaan(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut, req pemeriksaan.UpdatePemeriksaanRequest) error {
	if m.updatePemeriksaanFunc != nil {
		return m.updatePemeriksaanFunc(ctx, id, statusLanjut, req)
	}
	return nil
}

func (m *mockRepository) HapusPemeriksaan(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) error {
	if m.hapusPemeriksaanFunc != nil {
		return m.hapusPemeriksaanFunc(ctx, id, statusLanjut)
	}
	return nil
}

type mockRawatJalanService struct {
	rawatjalan.Service
	riwayatKunjunganPasienFunc func(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error)
	getWaktuRegistrasiFunc     func(ctx context.Context, noRawat string) (string, string, bool, error)
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

func TestDaftarKesadaran(t *testing.T) {
	repo := &mockRepository{}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	daftar := svc.DaftarKesadaran(context.Background())
	if len(daftar) != 11 {
		t.Fatalf("expected 11 items in daftar kesadaran, got %d", len(daftar))
	}

	foundComposMentis := false
	for _, item := range daftar {
		if item.Value == "Compos Mentis" && item.Label == "Compos Mentis" {
			foundComposMentis = true
			break
		}
	}
	if !foundComposMentis {
		t.Error("expected 'Compos Mentis' to be in daftar kesadaran")
	}
}

func TestSimpanPemeriksaan_SuccessRalan(t *testing.T) {
	called := false
	repo := &mockRepository{
		simpanPemeriksaanFunc: func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req pemeriksaan.SimpanPemeriksaanRequest) error {
			called = true
			if kodeDokter != "DK001" {
				t.Errorf("expected kodeDokter 'DK001', got '%s'", kodeDokter)
			}
			if statusLanjut != shared.StatusLanjutRawatJalan {
				t.Errorf("expected statusLanjut 'Ralan', got '%s'", statusLanjut)
			}
			if req.NoRawat != "2026/04/22/036934" {
				t.Errorf("expected no_rawat '2026/04/22/036934', got '%s'", req.NoRawat)
			}
			return nil
		},
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: id.NoRawat,
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan:  id.TanggalPemeriksaan,
					JamPemeriksaan:      id.JamPemeriksaan,
					Kesadaran:           pemeriksaan.KesadaranComposMentis,
					Keluhan:             "Demam",
					Pemeriksaan:         "Normal",
					Penilaian:           "Febris",
					RencanaTindakLanjut: "Istirahat",
					Instruksi:           "Minum obat",
					Evaluasi:            "Stabil",
				},
				KodeDokterPetugas: "DK001",
				StatusLanjut:      statusLanjut,
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	now := time.Now()
	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	res, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Error("expected repository SimpanPemeriksaan to be called")
	}
	if res == nil {
		t.Fatal("expected returned Pemeriksaan object, got nil")
	}
	if res.NoRawat != "2026/04/22/036934" {
		t.Errorf("expected NoRawat '2026/04/22/036934', got '%s'", res.NoRawat)
	}
}

func TestSimpanPemeriksaan_ValidationError(t *testing.T) {

	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			Kesadaran: "KesadaranPalsu",
		},
	}

	valErr := req.Validate()
	if valErr == nil {
		t.Fatal("expected validation error, got nil")
	}

	if _, exists := valErr["no_rawat"]; !exists {
		t.Error("expected validation error for 'no_rawat'")
	}
	if _, exists := valErr["kesadaran"]; !exists {
		t.Error("expected validation error for 'kesadaran'")
	}
	if _, exists := valErr["keluhan"]; !exists {
		t.Error("expected validation error for 'keluhan'")
	}
}

func TestSimpanPemeriksaan_RepoError(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		simpanPemeriksaanFunc: func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req pemeriksaan.SimpanPemeriksaanRequest) error {
			return errors.New("db error")
		},
	}
	rjRepo := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			reg := now.Add(-1 * time.Hour)
			return reg.Format("2006-01-02"), reg.Format("15:04:05"), true, nil
		},
	}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	_, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected repo error, got nil")
	}
}

func TestSimpanPemeriksaan_SuhuTubuhValidation(t *testing.T) {
	baseReq := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  "2026-04-23",
			JamPemeriksaan:      "12:10:00",
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	reqNonNumeric := baseReq
	reqNonNumeric.SuhuTubuh = "36.A"
	if errs := reqNonNumeric.Validate(); errs == nil {
		t.Fatal("expected error for non-numeric suhu tubuh, got nil")
	}

	reqTooLow := baseReq
	reqTooLow.SuhuTubuh = "20.0"
	if errs := reqTooLow.Validate(); errs == nil {
		t.Fatal("expected error for too low suhu tubuh, got nil")
	}

	reqTooHigh := baseReq
	reqTooHigh.SuhuTubuh = "48.5"
	if errs := reqTooHigh.Validate(); errs == nil {
		t.Fatal("expected error for too high suhu tubuh, got nil")
	}

	reqValidComma := baseReq
	reqValidComma.SuhuTubuh = "36,8"
	reqValidComma.Sanitize()
	if errs := reqValidComma.Validate(); errs != nil {
		t.Fatalf("expected nil error for valid suhu tubuh with comma, got %v", errs)
	}
}

func TestSimpanPemeriksaan_TTVValidation(t *testing.T) {
	baseReq := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  "2026-04-23",
			JamPemeriksaan:      "12:10:00",
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	reqTensiNoSlash := baseReq
	reqTensiNoSlash.Tensi = "12080"
	if errs := reqTensiNoSlash.Validate(); errs == nil {
		t.Fatal("expected error for tensi without slash, got nil")
	}

	reqTensiInvalid := baseReq
	reqTensiInvalid.Tensi = "80/120"
	if errs := reqTensiInvalid.Validate(); errs == nil {
		t.Fatal("expected error for sistolik <= diastolik, got nil")
	}

	reqTensiValid := baseReq
	reqTensiValid.Tensi = "120/80"
	if errs := reqTensiValid.Validate(); errs != nil {
		t.Fatalf("expected nil error for valid tensi, got %v", errs)
	}

	reqNadiInvalid := baseReq
	reqNadiInvalid.Nadi = "400"
	if errs := reqNadiInvalid.Validate(); errs == nil {
		t.Fatal("expected error for nadi > 300, got nil")
	}

	reqRespInvalid := baseReq
	reqRespInvalid.Respirasi = "150"
	if errs := reqRespInvalid.Validate(); errs == nil {
		t.Fatal("expected error for respirasi > 100, got nil")
	}

	reqTBInvalid := baseReq
	reqTBInvalid.TinggiBadan = "350"
	if errs := reqTBInvalid.Validate(); errs == nil {
		t.Fatal("expected error for tinggi badan > 250, got nil")
	}

	reqBBValid := baseReq
	reqBBValid.BeratBadan = "65,5"
	reqBBValid.Sanitize()
	if errs := reqBBValid.Validate(); errs != nil {
		t.Fatalf("expected nil error for valid berat badan, got %v", errs)
	}
}

func TestSimpanPemeriksaan_WaktuMasaDepanValidation(t *testing.T) {
	baseReq := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	besok := time.Now().AddDate(0, 0, 1)
	reqFutureDate := baseReq
	reqFutureDate.TanggalPemeriksaan = besok.Format("2006-01-02")
	reqFutureDate.JamPemeriksaan = "10:00:00"

	valErr := reqFutureDate.Validate()
	if valErr == nil {
		t.Fatal("expected error for future examination date, got nil")
	}
	if _, exists := valErr["tanggal_pemeriksaan"]; !exists {
		t.Error("expected validation error on 'tanggal_pemeriksaan'")
	}

	nanti := time.Now().Add(2 * time.Hour)
	reqFutureTime := baseReq
	reqFutureTime.TanggalPemeriksaan = nanti.Format("2006-01-02")
	reqFutureTime.JamPemeriksaan = nanti.Format("15:04:05")

	valErr = reqFutureTime.Validate()
	if valErr == nil {
		t.Fatal("expected error for future examination time today, got nil")
	}
}

func TestSimpanPemeriksaan_WaktuSebelumRegistrasi(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{}
	rjRepo := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			reg := now.Add(-1 * time.Hour)
			return reg.Format("2006-01-02"), reg.Format("15:04:05"), true, nil
		},
	}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	periksa := now.Add(-2 * time.Hour)
	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  periksa.Format("2006-01-02"),
			JamPemeriksaan:      periksa.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	_, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected error when examination time is before registration time, got nil")
	}

	valErr, ok := err.(apperror.ValidationError)
	if !ok {
		t.Fatalf("expected apperror.ValidationError, got %T", err)
	}

	if _, exists := valErr["tanggal_pemeriksaan"]; !exists {
		t.Error("expected validation error on 'tanggal_pemeriksaan'")
	}
}

func TestSimpanPemeriksaan_RegistrasiNotFound(t *testing.T) {
	repo := &mockRepository{}
	rjRepo := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return "", "", false, nil
		},
	}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	now := time.Now()
	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	_, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected NotFoundError when registration not found, got nil")
	}
}

func TestSimpanPemeriksaan_DuplicateEntry(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		simpanPemeriksaanFunc: func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req pemeriksaan.SimpanPemeriksaanRequest) error {
			return errors.New("Error 1062 (23000): Duplicate entry '2026/04/22/036934-2026-04-23-12:10:10' for key 'PRIMARY'")
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	_, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected validation error for duplicate entry, got nil")
	}

	valErr, ok := err.(apperror.ValidationError)
	if !ok {
		t.Fatalf("expected apperror.ValidationError, got %T", err)
	}

	if _, exists := valErr["jam_pemeriksaan"]; !exists {
		t.Error("expected validation error for 'jam_pemeriksaan'")
	}
}

func TestHapusPemeriksaan_Success(t *testing.T) {
	hapusCalled := false
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
			}, nil
		},
		hapusPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) error {
			hapusCalled = true
			return nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	err := svc.HapusPemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !hapusCalled {
		t.Error("expected HapusPemeriksaan repository to be called")
	}
}

func TestHapusPemeriksaan_ForbiddenDifferentDoctor(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK999",
				NamaDokterPetugas: "dr. Lain",
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	err := svc.HapusPemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected ForbiddenError when doctor does not match creator, got nil")
	}

	forbiddenErr, ok := err.(*apperror.ForbiddenError)
	if !ok {
		t.Fatalf("expected *apperror.ForbiddenError, got %T", err)
	}

	if forbiddenErr.Message == "" {
		t.Error("expected non-empty forbidden error message")
	}
}

func TestHapusPemeriksaan_MelebihiBatasWaktu(t *testing.T) {

	oldTime := time.Now().Add(-50 * time.Hour)
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: oldTime.Format("2006-01-02"),
					JamPemeriksaan:     oldTime.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return oldTime.Format("2006-01-02"), oldTime.Format("15:04:05"), true, nil
		},
	}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: oldTime.Format("2006-01-02"),
		JamPemeriksaan:     oldTime.Format("15:04:05"),
	}

	err := svc.HapusPemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected ForbiddenError when examination is older than 48 hours, got nil")
	}

	forbiddenErr, ok := err.(*apperror.ForbiddenError)
	if !ok {
		t.Fatalf("expected *apperror.ForbiddenError, got %T", err)
	}

	if forbiddenErr.Message == "" {
		t.Error("expected non-empty forbidden error message")
	}
}

func TestHapusPemeriksaan_NotFound(t *testing.T) {
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return nil, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: "2026-04-23",
		JamPemeriksaan:     "12:10:10",
	}

	err := svc.HapusPemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected NotFoundError when examination does not exist, got nil")
	}
}

func TestUpdatePemeriksaan_SuccessRalan(t *testing.T) {
	updateCalled := false
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
					Keluhan:            "Keluhan awal",
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
				StatusLanjut:      shared.StatusLanjutRawatJalan,
			}, nil
		},
		updatePemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut, req pemeriksaan.UpdatePemeriksaanRequest) error {
			updateCalled = true
			if req.Keluhan != "Keluhan diperbarui" {
				t.Errorf("expected keluhan 'Keluhan diperbarui', got '%s'", req.Keluhan)
			}
			if req.TanggalPemeriksaan != now.Format("2006-01-02") {
				t.Errorf("expected tanggal_pemeriksaan '%s', got '%s'", now.Format("2006-01-02"), req.TanggalPemeriksaan)
			}
			return nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			SuhuTubuh:           "36.8",
			Tensi:               "120/80",
			Nadi:                "80",
			Respirasi:           "20",
			TinggiBadan:         "170",
			BeratBadan:          "65",
			SpO2:                "98",
			Gcs:                 "15",
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan diperbarui",
			Pemeriksaan:         "Pemeriksaan fisik normal",
			Alergi:              "Tidak ada",
			LingkarPerut:        "80",
			RencanaTindakLanjut: "Kontrol 1 minggu",
			Penilaian:           "Sehat",
			Instruksi:           "Minum air putih",
			Evaluasi:            "Stabil",
		},
	}

	res, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan, req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !updateCalled {
		t.Error("expected UpdatePemeriksaan repository to be called")
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestUpdatePemeriksaan_SuccessRanap(t *testing.T) {
	updateCalled := false
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
					Keluhan:            "Keluhan ranap awal",
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
				StatusLanjut:      shared.StatusLanjutRawatInap,
			}, nil
		},
		updatePemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut, req pemeriksaan.UpdatePemeriksaanRequest) error {
			updateCalled = true
			return nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return true, true, nil
		},
	}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			SuhuTubuh:           "37.0",
			Tensi:               "110/70",
			Nadi:                "78",
			Respirasi:           "18",
			TinggiBadan:         "170",
			BeratBadan:          "65",
			SpO2:                "99",
			Gcs:                 "15",
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan berkurang",
			Pemeriksaan:         "Abdomen supel",
			Alergi:              "Tidak ada",
			RencanaTindakLanjut: "Observasi TTV",
			Penilaian:           "Perbaikan",
			Instruksi:           "Terapi lanjut",
			Evaluasi:            "Segar",
		},
	}

	res, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatInap, req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !updateCalled {
		t.Error("expected UpdatePemeriksaan repository to be called")
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestUpdatePemeriksaan_ForbiddenDifferentDoctor(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK999",
				NamaDokterPetugas: "dr. Lain",
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan",
			Pemeriksaan:         "Pemeriksaan",
			Penilaian:           "Penilaian",
			RencanaTindakLanjut: "RTL",
			Instruksi:           "Instruksi",
			Evaluasi:            "Evaluasi",
		},
	}

	_, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected ForbiddenError when doctor does not match creator, got nil")
	}

	forbiddenErr, ok := err.(*apperror.ForbiddenError)
	if !ok {
		t.Fatalf("expected *apperror.ForbiddenError, got %T", err)
	}

	if forbiddenErr.Message == "" {
		t.Error("expected non-empty forbidden error message")
	}
}

func TestUpdatePemeriksaan_ForbiddenMelebihiBatasWaktu(t *testing.T) {
	oldTime := time.Now().Add(-50 * time.Hour)
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: oldTime.Format("2006-01-02"),
					JamPemeriksaan:     oldTime.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return oldTime.Format("2006-01-02"), oldTime.Format("15:04:05"), true, nil
		},
	}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: oldTime.Format("2006-01-02"),
		JamPemeriksaan:     oldTime.Format("15:04:05"),
	}

	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  time.Now().Format("2006-01-02"),
			JamPemeriksaan:      time.Now().Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan",
			Pemeriksaan:         "Pemeriksaan",
			Penilaian:           "Penilaian",
			RencanaTindakLanjut: "RTL",
			Instruksi:           "Instruksi",
			Evaluasi:            "Evaluasi",
		},
	}

	_, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected ForbiddenError when examination is older than 48 hours, got nil")
	}

	forbiddenErr, ok := err.(*apperror.ForbiddenError)
	if !ok {
		t.Fatalf("expected *apperror.ForbiddenError, got %T", err)
	}

	if forbiddenErr.Message == "" {
		t.Error("expected non-empty forbidden error message")
	}
}

func TestUpdatePemeriksaan_NotFound(t *testing.T) {
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return nil, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: "2026-04-23",
		JamPemeriksaan:     "12:10:10",
	}

	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  "2026-04-23",
			JamPemeriksaan:      "12:10:10",
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan",
			Pemeriksaan:         "Pemeriksaan",
			Penilaian:           "Penilaian",
			RencanaTindakLanjut: "RTL",
			Instruksi:           "Instruksi",
			Evaluasi:            "Evaluasi",
		},
	}

	_, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected NotFoundError when examination does not exist, got nil")
	}
}

func TestUpdatePemeriksaan_ValidationErrors(t *testing.T) {
	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan: "bukan-tanggal",
			JamPemeriksaan:     "bukan-jam",
			Kesadaran:          "KesadaranGhoib",
			SuhuTubuh:          "999.0",
			Tensi:              "100/120",
		},
	}

	valErr := req.Validate()
	if valErr == nil {
		t.Fatal("expected ValidationError, got nil")
	}

	if _, exists := valErr["tanggal_pemeriksaan"]; !exists {
		t.Error("expected error for 'tanggal_pemeriksaan'")
	}
	if _, exists := valErr["jam_pemeriksaan"]; !exists {
		t.Error("expected error for 'jam_pemeriksaan'")
	}
	if _, exists := valErr["kesadaran"]; !exists {
		t.Error("expected error for 'kesadaran'")
	}
	if _, exists := valErr["suhu_tubuh"]; !exists {
		t.Error("expected error for 'suhu_tubuh'")
	}
	if _, exists := valErr["tensi"]; !exists {
		t.Error("expected error for 'tensi'")
	}
	if _, exists := valErr["keluhan"]; !exists {
		t.Error("expected error for 'keluhan'")
	}
}

func TestUpdatePemeriksaan_WaktuSebelumRegistrasi(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			reg := now.Add(-1 * time.Hour)
			return reg.Format("2006-01-02"), reg.Format("15:04:05"), true, nil
		},
	}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	periksa := now.Add(-2 * time.Hour)
	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  periksa.Format("2006-01-02"),
			JamPemeriksaan:      periksa.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan",
			Pemeriksaan:         "Pemeriksaan",
			Penilaian:           "Penilaian",
			RencanaTindakLanjut: "RTL",
			Instruksi:           "Instruksi",
			Evaluasi:            "Evaluasi",
		},
	}

	_, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected ValidationError for time before registration, got nil")
	}

	valErr, ok := err.(apperror.ValidationError)
	if !ok {
		t.Fatalf("expected apperror.ValidationError, got %T", err)
	}

	if _, exists := valErr["tanggal_pemeriksaan"]; !exists {
		t.Error("expected error for 'tanggal_pemeriksaan' when before registration time")
	}
}

func TestUpdatePemeriksaan_DuplicateEntry(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
			}, nil
		},
		updatePemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut, req pemeriksaan.UpdatePemeriksaanRequest) error {
			return errors.New("Error 1062 (23000): Duplicate entry '2026/04/22/036934-2026-04-23 12:10:10' for key 'PRIMARY'")
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan",
			Pemeriksaan:         "Pemeriksaan",
			Penilaian:           "Penilaian",
			RencanaTindakLanjut: "RTL",
			Instruksi:           "Instruksi",
			Evaluasi:            "Evaluasi",
		},
	}

	_, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected ValidationError for duplicate entry, got nil")
	}

	valErr, ok := err.(apperror.ValidationError)
	if !ok {
		t.Fatalf("expected apperror.ValidationError, got %T", err)
	}

	if _, exists := valErr["jam_pemeriksaan"]; !exists {
		t.Error("expected error for 'jam_pemeriksaan' on duplicate entry")
	}
}

func TestDetailPemeriksaan_Success(t *testing.T) {
	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: "2026-04-23",
		JamPemeriksaan:     "12:10:00",
	}

	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: id.NoRawat,
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: id.TanggalPemeriksaan,
					JamPemeriksaan:     id.JamPemeriksaan,
					Kesadaran:          pemeriksaan.KesadaranComposMentis,
				},
				StatusLanjut: statusLanjut,
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	res, err := svc.DetailPemeriksaan(context.Background(), id, shared.StatusLanjutRawatJalan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.NoRawat != id.NoRawat {
		t.Errorf("expected pemeriksaan with NoRawat '%s', got %+v", id.NoRawat, res)
	}
}

func TestDetailPemeriksaan_NotFound(t *testing.T) {
	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: "2026-04-23",
		JamPemeriksaan:     "12:10:00",
	}

	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return nil, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	res, err := svc.DetailPemeriksaan(context.Background(), id, shared.StatusLanjutRawatJalan)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("expected nil result, got %+v", res)
	}

	var notFoundErr *apperror.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected *apperror.NotFoundError, got %T (%v)", err, err)
	}
}

func TestSimpanPemeriksaan_MelebihiBatas48JamRalan(t *testing.T) {
	now := time.Now()
	oldReg := now.Add(-50 * time.Hour)
	repo := &mockRepository{}
	rjRepo := &mockRawatJalanService{
		getWaktuRegistrasiFunc: func(ctx context.Context, noRawat string) (string, string, bool, error) {
			return oldReg.Format("2006-01-02"), oldReg.Format("15:04:05"), true, nil
		},
	}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	_, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatJalan, req)
	if err == nil {
		t.Fatal("expected ForbiddenError when registration exceeds 48 hours, got nil")
	}
	var forbiddenErr *apperror.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("expected *apperror.ForbiddenError, got %T (%v)", err, err)
	}
}

func TestSimpanPemeriksaan_CheckoutRanap(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	_, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatInap, req)
	if err == nil {
		t.Fatal("expected BusinessError when patient checked out, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected *apperror.BusinessError, got %T (%v)", err, err)
	}
	if bErr.Message != "Pasien sudah keluar dari kamar inap" {
		t.Errorf("unexpected error message: %s", bErr.Message)
	}
}

func TestSimpanPemeriksaan_RanapTanpaKamar(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, false, nil
		},
	}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	req := pemeriksaan.SimpanPemeriksaanRequest{
		NoRawat: "2026/04/22/036934",
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Demam",
			Pemeriksaan:         "Normal",
			Penilaian:           "Febris",
			RencanaTindakLanjut: "Istirahat",
			Instruksi:           "Minum obat",
			Evaluasi:            "Stabil",
		},
	}

	_, err := svc.SimpanPemeriksaan(context.Background(), "DK001", shared.StatusLanjutRawatInap, req)
	if err == nil {
		t.Fatal("expected BusinessError when saving Ranap without room, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected *apperror.BusinessError, got %T (%v)", err, err)
	}
}

func TestUpdatePemeriksaan_CheckoutRanap(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
				StatusLanjut:      shared.StatusLanjutRawatInap,
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	req := pemeriksaan.UpdatePemeriksaanRequest{
		DataPemeriksaan: pemeriksaan.DataPemeriksaan{
			TanggalPemeriksaan:  now.Format("2006-01-02"),
			JamPemeriksaan:      now.Format("15:04:05"),
			Kesadaran:           pemeriksaan.KesadaranComposMentis,
			Keluhan:             "Keluhan",
			Pemeriksaan:         "Pemeriksaan",
			Penilaian:           "Penilaian",
			RencanaTindakLanjut: "RTL",
			Instruksi:           "Instruksi",
			Evaluasi:            "Evaluasi",
		},
	}

	_, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatInap, req)
	if err == nil {
		t.Fatal("expected BusinessError when updating checked out patient, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected *apperror.BusinessError, got %T (%v)", err, err)
	}
}

func TestHapusPemeriksaan_CheckoutRanap(t *testing.T) {
	now := time.Now()
	repo := &mockRepository{
		detailPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
			return &pemeriksaan.Pemeriksaan{
				NoRawat: "2026/04/22/036934",
				DataPemeriksaan: pemeriksaan.DataPemeriksaan{
					TanggalPemeriksaan: now.Format("2006-01-02"),
					JamPemeriksaan:     now.Format("15:04:05"),
				},
				KodeDokterPetugas: "DK001",
				NamaDokterPetugas: "dr. Budi",
				StatusLanjut:      shared.StatusLanjutRawatInap,
			}, nil
		},
	}
	rjRepo := &mockRawatJalanService{}
	mockRI := &mockRawatInapService{
		cekStatusKamarInapFunc: func(ctx context.Context, noRawat string) (bool, bool, error) {
			return false, true, nil
		},
	}
	log := logger.New()
	svc := pemeriksaan.NewService(repo, rjRepo, mockRI, 48, log)

	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	err := svc.HapusPemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatInap)
	if err == nil {
		t.Fatal("expected BusinessError when deleting checked out patient examination, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected *apperror.BusinessError, got %T (%v)", err, err)
	}
}

func TestUpdateAndHapusPemeriksaan_ErrNoRowsMapsToNotFound(t *testing.T) {
	now := time.Now()
	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/04/22/036934",
		TanggalPemeriksaan: now.Format("2006-01-02"),
		JamPemeriksaan:     now.Format("15:04:05"),
	}

	detailExisting := func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) (*pemeriksaan.Pemeriksaan, error) {
		return &pemeriksaan.Pemeriksaan{
			NoRawat: id.NoRawat,
			DataPemeriksaan: pemeriksaan.DataPemeriksaan{
				TanggalPemeriksaan: id.TanggalPemeriksaan,
				JamPemeriksaan:     id.JamPemeriksaan,
			},
			KodeDokterPetugas: "DK001",
			NamaDokterPetugas: "dr. Budi",
			StatusLanjut:      shared.StatusLanjutRawatJalan,
		}, nil
	}

	t.Run("UpdatePemeriksaan maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		repo := &mockRepository{
			detailPemeriksaanFunc: detailExisting,
			updatePemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut, req pemeriksaan.UpdatePemeriksaanRequest) error {
				return sql.ErrNoRows
			},
		}
		svc := pemeriksaan.NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, logger.New())

		req := pemeriksaan.UpdatePemeriksaanRequest{
			DataPemeriksaan: pemeriksaan.DataPemeriksaan{
				TanggalPemeriksaan:  id.TanggalPemeriksaan,
				JamPemeriksaan:      id.JamPemeriksaan,
				Kesadaran:           pemeriksaan.KesadaranComposMentis,
				Keluhan:             "Keluhan",
				Pemeriksaan:         "Pemeriksaan",
				Penilaian:           "Penilaian",
				RencanaTindakLanjut: "RTL",
				Instruksi:           "Instruksi",
				Evaluasi:            "Evaluasi",
			},
		}

		_, err := svc.UpdatePemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan, req)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})

	t.Run("HapusPemeriksaan maps sql.ErrNoRows to NotFoundError", func(t *testing.T) {
		repo := &mockRepository{
			detailPemeriksaanFunc: detailExisting,
			hapusPemeriksaanFunc: func(ctx context.Context, id pemeriksaan.IdPemeriksaan, statusLanjut shared.StatusLanjut) error {
				return sql.ErrNoRows
			},
		}
		svc := pemeriksaan.NewService(repo, &mockRawatJalanService{}, &mockRawatInapService{}, 48, logger.New())

		err := svc.HapusPemeriksaan(context.Background(), "DK001", id, shared.StatusLanjutRawatJalan)
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *apperror.NotFoundError, got %T (%v)", err, err)
		}
	})
}
