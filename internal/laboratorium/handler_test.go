package laboratorium_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/laboratorium"
	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/token"
	"erm-dokter/internal/shared"
)

const testEncryptionKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockService struct {
	getRiwayatLabKunjunganFn        func(ctx context.Context, kategori shared.KategoriLab, noRawat string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) (*laboratorium.HasilLaboratoriumKunjungan, shared.PaginationMeta, error)
	getRiwayatLabPasienFn           func(ctx context.Context, kategori shared.KategoriLab, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.HasilLaboratorium, shared.PaginationMeta, error)
	getDetailHasilLabFn             func(ctx context.Context, kategori shared.KategoriLab, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa string) (*laboratorium.HasilLaboratorium, error)
	simpanPermintaanLabPKFn         func(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPKRequest) (*laboratorium.DetailPermintaanLabPK, error)
	getDaftarPermintaanLabPKFn      func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabPK, error)
	getRiwayatPermintaanLabPKByRMFn func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabPK, shared.PaginationMeta, error)
	getDetailPermintaanLabPKFn      func(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabPK, error)
	hapusPermintaanLabPKFn          func(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error
	updatePermintaanLabPKFn         func(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPKRequest) (*laboratorium.DetailPermintaanLabPK, error)

	simpanPermintaanLabPAFn         func(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPARequest) (*laboratorium.DetailPermintaanLabPA, error)
	getDaftarPermintaanLabPAFn      func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabPA, error)
	getRiwayatPermintaanLabPAByRMFn func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabPA, shared.PaginationMeta, error)
	getDetailPermintaanLabPAFn      func(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabPA, error)
	hapusPermintaanLabPAFn          func(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error
	updatePermintaanLabPAFn         func(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPARequest) (*laboratorium.DetailPermintaanLabPA, error)

	simpanPermintaanLabMBFn         func(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabMBRequest) (*laboratorium.DetailPermintaanLabMB, error)
	getDaftarPermintaanLabMBFn      func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabMB, error)
	getRiwayatPermintaanLabMBByRMFn func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabMB, shared.PaginationMeta, error)
	getDetailPermintaanLabMBFn      func(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabMB, error)
	hapusPermintaanLabMBFn          func(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error
	updatePermintaanLabMBFn         func(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabMBRequest) (*laboratorium.DetailPermintaanLabMB, error)
}

func (m *mockService) GetRiwayatLabKunjungan(ctx context.Context, kategori shared.KategoriLab, noRawat string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) (*laboratorium.HasilLaboratoriumKunjungan, shared.PaginationMeta, error) {
	if m.getRiwayatLabKunjunganFn != nil {
		return m.getRiwayatLabKunjunganFn(ctx, kategori, noRawat, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockService) GetRiwayatLabPasien(ctx context.Context, kategori shared.KategoriLab, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.HasilLaboratorium, shared.PaginationMeta, error) {
	if m.getRiwayatLabPasienFn != nil {
		return m.getRiwayatLabPasienFn(ctx, kategori, noRM, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockService) GetDetailHasilLab(ctx context.Context, kategori shared.KategoriLab, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa string) (*laboratorium.HasilLaboratorium, error) {
	if m.getDetailHasilLabFn != nil {
		return m.getDetailHasilLabFn(ctx, kategori, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa)
	}
	return nil, nil
}

func (m *mockService) SimpanPermintaanLabPK(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPKRequest) (*laboratorium.DetailPermintaanLabPK, error) {
	if m.simpanPermintaanLabPKFn != nil {
		return m.simpanPermintaanLabPKFn(ctx, kodeDokterLogin, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockService) GetDaftarPermintaanLabPK(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabPK, error) {
	if m.getDaftarPermintaanLabPKFn != nil {
		return m.getDaftarPermintaanLabPKFn(ctx, noRawat, statusLanjut)
	}
	return nil, nil
}

func (m *mockService) GetRiwayatPermintaanLabPKByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabPK, shared.PaginationMeta, error) {
	if m.getRiwayatPermintaanLabPKByRMFn != nil {
		return m.getRiwayatPermintaanLabPKByRMFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockService) GetDetailPermintaanLabPK(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabPK, error) {
	if m.getDetailPermintaanLabPKFn != nil {
		return m.getDetailPermintaanLabPKFn(ctx, noPermintaan)
	}
	return nil, nil
}

func (m *mockService) HapusPermintaanLabPK(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
	if m.hapusPermintaanLabPKFn != nil {
		return m.hapusPermintaanLabPKFn(ctx, noRawat, noPermintaan, statusLanjut, kodeDokterLogin)
	}
	return nil
}

func (m *mockService) SimpanPermintaanLabPA(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPARequest) (*laboratorium.DetailPermintaanLabPA, error) {
	if m.simpanPermintaanLabPAFn != nil {
		return m.simpanPermintaanLabPAFn(ctx, kodeDokterLogin, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockService) GetDaftarPermintaanLabPA(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabPA, error) {
	if m.getDaftarPermintaanLabPAFn != nil {
		return m.getDaftarPermintaanLabPAFn(ctx, noRawat, statusLanjut)
	}
	return nil, nil
}

func (m *mockService) GetRiwayatPermintaanLabPAByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabPA, shared.PaginationMeta, error) {
	if m.getRiwayatPermintaanLabPAByRMFn != nil {
		return m.getRiwayatPermintaanLabPAByRMFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockService) GetDetailPermintaanLabPA(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabPA, error) {
	if m.getDetailPermintaanLabPAFn != nil {
		return m.getDetailPermintaanLabPAFn(ctx, noPermintaan)
	}
	return nil, nil
}

func (m *mockService) HapusPermintaanLabPA(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
	if m.hapusPermintaanLabPAFn != nil {
		return m.hapusPermintaanLabPAFn(ctx, noRawat, noPermintaan, statusLanjut, kodeDokterLogin)
	}
	return nil
}

func (m *mockService) SimpanPermintaanLabMB(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabMBRequest) (*laboratorium.DetailPermintaanLabMB, error) {
	if m.simpanPermintaanLabMBFn != nil {
		return m.simpanPermintaanLabMBFn(ctx, kodeDokterLogin, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockService) GetDaftarPermintaanLabMB(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabMB, error) {
	if m.getDaftarPermintaanLabMBFn != nil {
		return m.getDaftarPermintaanLabMBFn(ctx, noRawat, statusLanjut)
	}
	return nil, nil
}

func (m *mockService) GetRiwayatPermintaanLabMBByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabMB, shared.PaginationMeta, error) {
	if m.getRiwayatPermintaanLabMBByRMFn != nil {
		return m.getRiwayatPermintaanLabMBByRMFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockService) GetDetailPermintaanLabMB(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabMB, error) {
	if m.getDetailPermintaanLabMBFn != nil {
		return m.getDetailPermintaanLabMBFn(ctx, noPermintaan)
	}
	return nil, nil
}

func (m *mockService) HapusPermintaanLabMB(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
	if m.hapusPermintaanLabMBFn != nil {
		return m.hapusPermintaanLabMBFn(ctx, noRawat, noPermintaan, statusLanjut, kodeDokterLogin)
	}
	return nil
}

func (m *mockService) UpdatePermintaanLabPK(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPKRequest) (*laboratorium.DetailPermintaanLabPK, error) {
	if m.updatePermintaanLabPKFn != nil {
		return m.updatePermintaanLabPKFn(ctx, kodeDokterLogin, noRawat, noPermintaan, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockService) UpdatePermintaanLabPA(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPARequest) (*laboratorium.DetailPermintaanLabPA, error) {
	if m.updatePermintaanLabPAFn != nil {
		return m.updatePermintaanLabPAFn(ctx, kodeDokterLogin, noRawat, noPermintaan, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockService) UpdatePermintaanLabMB(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabMBRequest) (*laboratorium.DetailPermintaanLabMB, error) {
	if m.updatePermintaanLabMBFn != nil {
		return m.updatePermintaanLabMBFn(ctx, kodeDokterLogin, noRawat, noPermintaan, statusLanjut, req)
	}
	return nil, nil
}

func authMiddlewareForTest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{
			KodeDokter: "DR01",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func TestHandler_DaftarHasilLab_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	mockSvc := &mockService{
		getRiwayatLabKunjunganFn: func(ctx context.Context, kategori shared.KategoriLab, noRawat string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) (*laboratorium.HasilLaboratoriumKunjungan, shared.PaginationMeta, error) {
			if noRawat != "2026/04/22/000001" {
				t.Errorf("Expected noRawat 2026/04/22/000001, got %s", noRawat)
			}
			return &laboratorium.HasilLaboratoriumKunjungan{
				HasilPemeriksaan: []laboratorium.HasilLaboratorium{
					{
						NoRawat:      "2026/04/22/000001",
						KodeTindakan: "LAB001",
						NamaTindakan: "DARAH LENGKAP",
						Kategori:     "PK",
					},
				},
				BerkasDigital: []berkasdigital.BerkasDigital{
					{
						Kode:       "005",
						NamaBerkas: "HASIL LAB PK",
						IdBerkas:   "http://berkasrawat.example.com/webapps/berkasrawat/pages/upload/pk.pdf",
					},
				},
			}, shared.NewPaginationMeta(1, 1, 5), nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pk/hasil/Semua/"+encKunjungan+"?page=1&limit=5", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Code    int                                     `json:"code"`
		Status  string                                  `json:"status"`
		Message string                                  `json:"message"`
		Data    laboratorium.HasilLaboratoriumKunjungan `json:"data"`
		Meta    shared.PaginationMeta                   `json:"meta"`
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.Data.HasilPemeriksaan) != 1 || resp.Data.HasilPemeriksaan[0].NamaTindakan != "DARAH LENGKAP" {
		t.Errorf("Unexpected response data: %+v", resp.Data)
	}
	if resp.Data.HasilPemeriksaan[0].Id == "" {
		t.Errorf("Expected encrypted Id in response")
	}

	if len(resp.Data.BerkasDigital) != 1 {
		t.Fatalf("Expected 1 berkas digital, got %d", len(resp.Data.BerkasDigital))
	}
	if resp.Data.BerkasDigital[0].IdBerkas == "" || !strings.HasPrefix(resp.Data.BerkasDigital[0].UrlBerkas, "/api/v1/berkas-digital/") {
		t.Errorf("Expected encrypted IdBerkas and valid UrlBerkas, got %+v", resp.Data.BerkasDigital[0])
	}
}

func TestHandler_DaftarHasilLab_InvalidKategori(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	handler := laboratorium.NewHandler(&mockService{}, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/radiologi/hasil/Semua/"+encKunjungan, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for invalid kategori, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarHasilLab_InvalidStatusLanjut(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	handler := laboratorium.NewHandler(&mockService{}, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pk/hasil/InvalidStatus/"+encKunjungan, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for invalid status lanjut, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarHasilLabByRM_Success(t *testing.T) {
	encPasien, _ := crypto.Encrypt("123456", testEncryptionKey)
	mockSvc := &mockService{
		getRiwayatLabPasienFn: func(ctx context.Context, kategori shared.KategoriLab, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.HasilLaboratorium, shared.PaginationMeta, error) {
			if noRM != "123456" {
				t.Errorf("Expected noRM 123456, got %s", noRM)
			}
			return []laboratorium.HasilLaboratorium{
				{
					NoRawat:      "2026/04/22/000001",
					KodeTindakan: "PA001",
					NamaTindakan: "BIOPSI",
					Kategori:     "PA",
				},
			}, shared.NewPaginationMeta(1, 1, 5), nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pa/hasil/Semua/pasien/"+encPasien, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailHasilLab_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	encHasil, _ := crypto.Encrypt("2026/04/22/000001~LAB001~2026-04-22~10:00:00", testEncryptionKey)

	mockSvc := &mockService{
		getDetailHasilLabFn: func(ctx context.Context, kategori shared.KategoriLab, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa string) (*laboratorium.HasilLaboratorium, error) {
			return &laboratorium.HasilLaboratorium{
				NoRawat:        noRawat,
				KodeTindakan:   kodeTindakan,
				TanggalPeriksa: tanggalPeriksa,
				JamPeriksa:     jamPeriksa,
				NamaTindakan:   "DARAH LENGKAP",
				Kategori:       "PK",
				Status:         "Ralan",
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pk/hasil/Ralan/"+encKunjungan+"/"+encHasil, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_SimpanPermintaanLabPK_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("TND001", testEncryptionKey)
	encTemplate, _ := crypto.Encrypt("101", testEncryptionKey)

	mockSvc := &mockService{
		simpanPermintaanLabPKFn: func(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPKRequest) (*laboratorium.DetailPermintaanLabPK, error) {
			if req.NoRawat != "2026/09/03/000001" {
				t.Errorf("Expected NoRawat 2026/09/03/000001, got %s", req.NoRawat)
			}
			if statusLanjut != shared.StatusLanjutRawatJalan {
				t.Errorf("Expected StatusLanjut Ralan, got %s", statusLanjut)
			}
			if len(req.Pemeriksaan) != 1 || req.Pemeriksaan[0].KodeTindakan != "TND001" {
				t.Errorf("Expected decrypted KodeTindakan TND001, got %+v", req.Pemeriksaan)
			}
			return &laboratorium.DetailPermintaanLabPK{
				PermintaanLabPK: laboratorium.PermintaanLabPK{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan: "PK202609030001",
						NoRawat:      req.NoRawat,
						Status:       "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabPKItem{},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body, _ := json.Marshal(laboratorium.SimpanPermintaanLabPKRequest{
		PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
			NoRawat:           "2026/09/03/000001",
			TanggalPermintaan: "2026-09-03",
			JamPermintaan:     "10:00:00",
			DiagnosaKlinis:    "Febris H-3",
		},
		Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
			{
				IdTindakan: encTindakan,
				IdTemplate: []string{encTemplate},
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/laboratorium/pk/permintaan/Ralan/"+encKunjungan, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_SimpanPermintaanLabPK_MismatchNoRawat(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncryptionKey)

	handler := laboratorium.NewHandler(&mockService{}, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body, _ := json.Marshal(laboratorium.SimpanPermintaanLabPKRequest{
		PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
			NoRawat:           "2026/09/03/999999",
			TanggalPermintaan: "2026-09-03",
			JamPermintaan:     "10:00:00",
			DiagnosaKlinis:    "Febris H-3",
		},
		Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
			{IdTindakan: "enc-tindakan-1"},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/laboratorium/pk/permintaan/Ralan/"+encKunjungan, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request for mismatched NoRawat, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_SimpanPermintaanLabPK_InvalidTindakanId(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncryptionKey)

	handler := laboratorium.NewHandler(&mockService{}, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body, _ := json.Marshal(laboratorium.SimpanPermintaanLabPKRequest{
		PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
			NoRawat:           "2026/09/03/000001",
			TanggalPermintaan: "2026-09-03",
			JamPermintaan:     "10:00:00",
			DiagnosaKlinis:    "Febris H-3",
		},
		Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
			{IdTindakan: "invalid-encrypted-token"},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/laboratorium/pk/permintaan/Ralan/"+encKunjungan, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request for invalid IdTindakan, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarPermintaanLabPK_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncryptionKey)
	mockSvc := &mockService{
		getDaftarPermintaanLabPKFn: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabPK, error) {
			return []laboratorium.PermintaanLabPK{
				{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan: "PK202609030001",
						NoRawat:      noRawat,
						Status:       "ralan",
						StatusProses: "Menunggu Sampel",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pk/permintaan/Semua/"+encKunjungan, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarPermintaanLabPKByRM_Success(t *testing.T) {
	encPasien, _ := crypto.Encrypt("123456", testEncryptionKey)
	mockSvc := &mockService{
		getRiwayatPermintaanLabPKByRMFn: func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabPK, shared.PaginationMeta, error) {
			return []laboratorium.PermintaanLabPK{
				{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan: "PK202609030001",
						NoRawat:      "2026/09/03/000001",
						Status:       "ralan",
					},
				},
			}, shared.NewPaginationMeta(1, 1, 5), nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pk/permintaan/Semua/pasien/"+encPasien, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailPermintaanLabPK_Success(t *testing.T) {
	encOrder, _ := crypto.Encrypt("PK202609030001", testEncryptionKey)

	mockSvc := &mockService{
		getDetailPermintaanLabPKFn: func(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabPK, error) {
			return &laboratorium.DetailPermintaanLabPK{
				PermintaanLabPK: laboratorium.PermintaanLabPK{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan: noPermintaan,
						NoRawat:      "2026/09/03/000001",
						Status:       "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabPKItem{},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pk/permintaan/"+encOrder, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_HapusPermintaanLabPK_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncryptionKey)
	encOrder, _ := crypto.Encrypt("PK202609030001", testEncryptionKey)

	mockSvc := &mockService{
		hapusPermintaanLabPKFn: func(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
			return nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/laboratorium/pk/permintaan/Ralan/"+encKunjungan+"/"+encOrder, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_SimpanPermintaanLabPA_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("PA00001", testEncryptionKey)

	mockSvc := &mockService{
		simpanPermintaanLabPAFn: func(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPARequest) (*laboratorium.DetailPermintaanLabPA, error) {
			if req.Pemeriksaan[0].KodeTindakan != "PA00001" {
				t.Errorf("Expected decrypted KodeTindakan PA00001, got %s", req.Pemeriksaan[0].KodeTindakan)
			}
			return &laboratorium.DetailPermintaanLabPA{
				PermintaanLabPA: laboratorium.PermintaanLabPA{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      "PA202609040001",
						NoRawat:           "2026/09/04/000001",
						TanggalPermintaan: "2026-09-04",
						JamPermintaan:     "10:00:00",
						DiagnosaKlinis:    "Tumor Mammae",
						Status:            "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabPAItem{
					{
						KodeTindakan: "PA00001",
						NamaTindakan: "Pemeriksaan PA Sediaan Kecil",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	payload := map[string]any{
		"no_rawat":           "2026/09/04/000001",
		"tanggal_permintaan": "2026-09-03",
		"jam_permintaan":     "10:00:00",
		"diagnosa_klinis":    "Tumor Mammae",
		"informasi_tambahan": "Teraba benjolan",
		"pengambilan_bahan":  "2026-09-03",
		"diperoleh_dengan":   "Biopsi",
		"lokasi_jaringan":    "Mammae Dextra",
		"diawetkan_dengan":   "Formalin 10%",
		"pemeriksaan": []map[string]any{
			{
				"id_tindakan": encTindakan,
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/laboratorium/pa/permintaan/Ralan/"+encKunjungan, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarPermintaanLabPA_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)

	mockSvc := &mockService{
		getDaftarPermintaanLabPAFn: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabPA, error) {
			return []laboratorium.PermintaanLabPA{
				{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      "PA202609040001",
						NoRawat:           noRawat,
						TanggalPermintaan: "2026-09-04",
						JamPermintaan:     "10:00:00",
						DiagnosaKlinis:    "Tumor Mammae",
						Status:            "ralan",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pa/permintaan/Semua/"+encKunjungan, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarPermintaanLabPAByRM_Success(t *testing.T) {
	encPasien, _ := crypto.Encrypt("000001", testEncryptionKey)

	mockSvc := &mockService{
		getRiwayatPermintaanLabPAByRMFn: func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabPA, shared.PaginationMeta, error) {
			return []laboratorium.PermintaanLabPA{
				{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      "PA202609040001",
						NoRawat:           "2026/09/04/000001",
						TanggalPermintaan: "2026-09-04",
						JamPermintaan:     "10:00:00",
						Status:            "ralan",
					},
				},
			}, shared.NewPaginationMeta(1, 1, 10), nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pa/permintaan/Semua/pasien/"+encPasien, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailPermintaanLabPA_Success(t *testing.T) {
	encOrder, _ := crypto.Encrypt("PA202609040001", testEncryptionKey)

	mockSvc := &mockService{
		getDetailPermintaanLabPAFn: func(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabPA, error) {
			return &laboratorium.DetailPermintaanLabPA{
				PermintaanLabPA: laboratorium.PermintaanLabPA{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      noPermintaan,
						NoRawat:           "2026/09/04/000001",
						TanggalPermintaan: "2026-09-04",
						JamPermintaan:     "10:00:00",
						DiagnosaKlinis:    "Tumor Mammae",
						Status:            "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabPAItem{
					{
						KodeTindakan: "PA00001",
						NamaTindakan: "Pemeriksaan PA Sediaan Kecil",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/pa/permintaan/"+encOrder, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_HapusPermintaanLabPA_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encOrder, _ := crypto.Encrypt("PA202609040001", testEncryptionKey)

	mockSvc := &mockService{
		hapusPermintaanLabPAFn: func(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
			return nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/laboratorium/pa/permintaan/Ralan/"+encKunjungan+"/"+encOrder, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_SimpanPermintaanLabMB_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("MB0001", testEncryptionKey)
	encTemplate, _ := crypto.Encrypt("101", testEncryptionKey)

	mockSvc := &mockService{
		simpanPermintaanLabMBFn: func(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabMBRequest) (*laboratorium.DetailPermintaanLabMB, error) {
			return &laboratorium.DetailPermintaanLabMB{
				PermintaanLabMB: laboratorium.PermintaanLabMB{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      "MB202609040001",
						NoRawat:           req.NoRawat,
						TanggalPermintaan: req.TanggalPermintaan,
						JamPermintaan:     req.JamPermintaan,
						DiagnosaKlinis:    req.DiagnosaKlinis,
						Status:            "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabMBItem{
					{
						KodeTindakan: "MB0001",
						NamaTindakan: "Kultur dan Resistensi Mikroorganisme",
						DetailTemplate: []laboratorium.DetailTemplateLabMBItem{
							{
								IdTemplate:      "101",
								NamaPemeriksaan: "Bakteri Batang Gram Negatif",
							},
						},
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body := map[string]any{
		"no_rawat":           "2026/09/04/000001",
		"tanggal_permintaan": "2026-09-04",
		"jam_permintaan":     "10:00:00",
		"diagnosa_klinis":    "Sepsis Curiga Bakteremia",
		"pemeriksaan": []map[string]any{
			{
				"id_tindakan": encTindakan,
				"id_template": []string{encTemplate},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/laboratorium/mb/permintaan/Ralan/"+encKunjungan, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_SimpanPermintaanLabMB_MismatchNoRawat(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("MB0001", testEncryptionKey)

	handler := laboratorium.NewHandler(&mockService{}, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body := map[string]any{
		"no_rawat":           "2026/09/04/999999",
		"tanggal_permintaan": "2026-09-04",
		"jam_permintaan":     "10:00:00",
		"diagnosa_klinis":    "Sepsis Curiga Bakteremia",
		"pemeriksaan": []map[string]any{
			{
				"id_tindakan": encTindakan,
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/laboratorium/mb/permintaan/Ralan/"+encKunjungan, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for mismatch no_rawat, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarPermintaanLabMB_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)

	mockSvc := &mockService{
		getDaftarPermintaanLabMBFn: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]laboratorium.PermintaanLabMB, error) {
			return []laboratorium.PermintaanLabMB{
				{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      "MB202609040001",
						NoRawat:           noRawat,
						TanggalPermintaan: "2026-09-04",
						JamPermintaan:     "10:00:00",
						Status:            "ralan",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/mb/permintaan/Ralan/"+encKunjungan, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_RiwayatPermintaanLabMBByRM_Success(t *testing.T) {
	encPasien, _ := crypto.Encrypt("RM123456", testEncryptionKey)

	mockSvc := &mockService{
		getRiwayatPermintaanLabMBByRMFn: func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter laboratorium.FilterRiwayatLab) ([]laboratorium.PermintaanLabMB, shared.PaginationMeta, error) {
			return []laboratorium.PermintaanLabMB{
				{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      "MB202609040001",
						NoRawat:           "2026/09/04/000001",
						TanggalPermintaan: "2026-09-04",
						JamPermintaan:     "10:00:00",
						Status:            "ralan",
					},
				},
			}, shared.NewPaginationMeta(1, 1, 10), nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/mb/permintaan/Semua/pasien/"+encPasien, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailPermintaanLabMB_Success(t *testing.T) {
	encOrder, _ := crypto.Encrypt("MB202609040001", testEncryptionKey)

	mockSvc := &mockService{
		getDetailPermintaanLabMBFn: func(ctx context.Context, noPermintaan string) (*laboratorium.DetailPermintaanLabMB, error) {
			return &laboratorium.DetailPermintaanLabMB{
				PermintaanLabMB: laboratorium.PermintaanLabMB{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      noPermintaan,
						NoRawat:           "2026/09/04/000001",
						TanggalPermintaan: "2026-09-04",
						JamPermintaan:     "10:00:00",
						DiagnosaKlinis:    "Sepsis Curiga Bakteremia",
						Status:            "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabMBItem{
					{
						KodeTindakan: "MB0001",
						NamaTindakan: "Kultur dan Resistensi Mikroorganisme",
						DetailTemplate: []laboratorium.DetailTemplateLabMBItem{
							{
								IdTemplate:      "101",
								NamaPemeriksaan: "Bakteri Batang Gram Negatif",
							},
						},
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/laboratorium/mb/permintaan/"+encOrder, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_HapusPermintaanLabMB_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encOrder, _ := crypto.Encrypt("MB202609040001", testEncryptionKey)

	mockSvc := &mockService{
		hapusPermintaanLabMBFn: func(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
			return nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/laboratorium/mb/permintaan/Ralan/"+encKunjungan+"/"+encOrder, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_UpdatePermintaanLabPK_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encOrder, _ := crypto.Encrypt("PK202609040001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("TND001", testEncryptionKey)

	mockSvc := &mockService{
		updatePermintaanLabPKFn: func(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPKRequest) (*laboratorium.DetailPermintaanLabPK, error) {
			return &laboratorium.DetailPermintaanLabPK{
				PermintaanLabPK: laboratorium.PermintaanLabPK{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      noPermintaan,
						NoRawat:           noRawat,
						TanggalPermintaan: req.TanggalPermintaan,
						JamPermintaan:     req.JamPermintaan,
						DiagnosaKlinis:    req.DiagnosaKlinis,
						Status:            "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabPKItem{
					{
						KodeTindakan: req.Pemeriksaan[0].KodeTindakan,
						NamaTindakan: "Darah Lengkap",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body := `{
		"no_rawat": "2026/09/04/000001",
		"tanggal_permintaan": "2026-09-04",
		"jam_permintaan": "10:00:00",
		"diagnosa_klinis": "Febris Update",
		"pemeriksaan": [
			{"id_tindakan": "` + encTindakan + `"}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/laboratorium/pk/permintaan/Ralan/"+encKunjungan+"/"+encOrder, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_UpdatePermintaanLabPA_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encOrder, _ := crypto.Encrypt("PA202609040001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("PA001", testEncryptionKey)

	mockSvc := &mockService{
		updatePermintaanLabPAFn: func(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabPARequest) (*laboratorium.DetailPermintaanLabPA, error) {
			return &laboratorium.DetailPermintaanLabPA{
				PermintaanLabPA: laboratorium.PermintaanLabPA{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      noPermintaan,
						NoRawat:           noRawat,
						TanggalPermintaan: req.TanggalPermintaan,
						JamPermintaan:     req.JamPermintaan,
						DiagnosaKlinis:    req.DiagnosaKlinis,
						Status:            "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabPAItem{
					{
						KodeTindakan: req.Pemeriksaan[0].KodeTindakan,
						NamaTindakan: "Biopsi PA",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body := `{
		"no_rawat": "2026/09/04/000001",
		"tanggal_permintaan": "2026-09-04",
		"jam_permintaan": "10:00:00",
		"diagnosa_klinis": "Tumor Mammae Update",
		"pemeriksaan": [
			{"id_tindakan": "` + encTindakan + `"}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/laboratorium/pa/permintaan/Ralan/"+encKunjungan+"/"+encOrder, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_UpdatePermintaanLabMB_Success(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/04/000001", testEncryptionKey)
	encOrder, _ := crypto.Encrypt("MB202609040001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("MB001", testEncryptionKey)

	mockSvc := &mockService{
		updatePermintaanLabMBFn: func(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req laboratorium.SimpanPermintaanLabMBRequest) (*laboratorium.DetailPermintaanLabMB, error) {
			return &laboratorium.DetailPermintaanLabMB{
				PermintaanLabMB: laboratorium.PermintaanLabMB{
					PermintaanLabHeader: laboratorium.PermintaanLabHeader{
						NoPermintaan:      noPermintaan,
						NoRawat:           noRawat,
						TanggalPermintaan: req.TanggalPermintaan,
						JamPermintaan:     req.JamPermintaan,
						DiagnosaKlinis:    req.DiagnosaKlinis,
						Status:            "ralan",
					},
				},
				Pemeriksaan: []laboratorium.PemeriksaanLabMBItem{
					{
						KodeTindakan: req.Pemeriksaan[0].KodeTindakan,
						NamaTindakan: "Kultur MB",
					},
				},
			}, nil
		},
	}

	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	body := `{
		"no_rawat": "2026/09/04/000001",
		"tanggal_permintaan": "2026-09-04",
		"jam_permintaan": "10:00:00",
		"diagnosa_klinis": "Sepsis Update",
		"pemeriksaan": [
			{"id_tindakan": "` + encTindakan + `"}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/laboratorium/mb/permintaan/Ralan/"+encKunjungan+"/"+encOrder, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailEndpoints_RejectSemua(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	encHasil, _ := crypto.Encrypt("2026/04/22/000001~LAB001~2026-04-22~10:00:00", testEncryptionKey)
	encOrder, _ := crypto.Encrypt("PK001", testEncryptionKey)

	mockSvc := &mockService{}
	handler := laboratorium.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	testCases := []struct {
		name   string
		method string
		url    string
	}{
		{
			name:   "DetailHasilLab_RejectSemua",
			method: http.MethodGet,
			url:    "/api/v1/laboratorium/pk/hasil/Semua/" + encKunjungan + "/" + encHasil,
		},
		{
			name:   "HapusPermintaanLabPK_RejectSemua",
			method: http.MethodDelete,
			url:    "/api/v1/laboratorium/pk/permintaan/Semua/" + encKunjungan + "/" + encOrder,
		},
		{
			name:   "HapusPermintaanLabPA_RejectSemua",
			method: http.MethodDelete,
			url:    "/api/v1/laboratorium/pa/permintaan/Semua/" + encKunjungan + "/" + encOrder,
		},
		{
			name:   "HapusPermintaanLabMB_RejectSemua",
			method: http.MethodDelete,
			url:    "/api/v1/laboratorium/mb/permintaan/Semua/" + encKunjungan + "/" + encOrder,
		},
		{
			name:   "UpdatePermintaanLabPK_RejectSemua",
			method: http.MethodPut,
			url:    "/api/v1/laboratorium/pk/permintaan/Semua/" + encKunjungan + "/" + encOrder,
		},
		{
			name:   "SimpanPermintaanLabPK_RejectSemua",
			method: http.MethodPost,
			url:    "/api/v1/laboratorium/pk/permintaan/Semua/" + encKunjungan,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected 400 Bad Request for %s, got %d: %s", tc.url, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandler_SimpanPermintaanLab_DuplicateDecryptedItems(t *testing.T) {
	noRawat := "2026/04/22/000001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncryptionKey)
	encTindakanA, _ := crypto.Encrypt("LAB01", testEncryptionKey)
	encTindakanB, _ := crypto.Encrypt("LAB01", testEncryptionKey)
	encTmplA, _ := crypto.Encrypt("101", testEncryptionKey)
	encTmplB, _ := crypto.Encrypt("101", testEncryptionKey)

	handler := laboratorium.NewHandler(&mockService{}, testEncryptionKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMiddlewareForTest, func(next http.HandlerFunc) http.HandlerFunc { return next })

	cases := []struct {
		name string
		url  string
		body any
	}{
		{
			name: "PK duplikat tindakan dari dua ciphertext berbeda",
			url:  "/api/v1/laboratorium/pk/permintaan/Ralan/" + encKunjungan,
			body: laboratorium.SimpanPermintaanLabPKRequest{
				PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
					NoRawat:           noRawat,
					TanggalPermintaan: "2026-04-22",
					JamPermintaan:     "10:00:00",
					DiagnosaKlinis:    "Febris",
				},
				Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
					{IdTindakan: encTindakanA},
					{IdTindakan: encTindakanB},
				},
			},
		},
		{
			name: "PK duplikat template dari dua ciphertext berbeda",
			url:  "/api/v1/laboratorium/pk/permintaan/Ralan/" + encKunjungan,
			body: laboratorium.SimpanPermintaanLabPKRequest{
				PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
					NoRawat:           noRawat,
					TanggalPermintaan: "2026-04-22",
					JamPermintaan:     "10:00:00",
					DiagnosaKlinis:    "Febris",
				},
				Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
					{IdTindakan: encTindakanA, IdTemplate: []string{encTmplA, encTmplB}},
				},
			},
		},
		{
			name: "PA duplikat tindakan dari dua ciphertext berbeda",
			url:  "/api/v1/laboratorium/pa/permintaan/Ralan/" + encKunjungan,
			body: laboratorium.SimpanPermintaanLabPARequest{
				PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
					NoRawat:           noRawat,
					TanggalPermintaan: "2026-04-22",
					JamPermintaan:     "10:00:00",
					DiagnosaKlinis:    "Tumor",
				},
				Pemeriksaan: []laboratorium.ItemPemeriksaanLabPARequest{
					{IdTindakan: encTindakanA},
					{IdTindakan: encTindakanB},
				},
			},
		},
		{
			name: "MB duplikat template dari dua ciphertext berbeda",
			url:  "/api/v1/laboratorium/mb/permintaan/Ralan/" + encKunjungan,
			body: laboratorium.SimpanPermintaanLabMBRequest{
				PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
					NoRawat:           noRawat,
					TanggalPermintaan: "2026-04-22",
					JamPermintaan:     "10:00:00",
					DiagnosaKlinis:    "Sepsis",
				},
				Pemeriksaan: []laboratorium.ItemPemeriksaanLabMBRequest{
					{IdTindakan: encTindakanA, IdTemplate: []string{encTmplA, encTmplB}},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, tc.url, bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 Bad Request, got %d: %s", rr.Code, rr.Body.String())
			}
		})
	}
}
