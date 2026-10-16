package rawatjalan_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/token"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
)

const testEncKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockRawatJalanService struct {
	daftarAntreanDokterFn    func(ctx context.Context, kodeDokter string, filter rawatjalan.FilterAntreanDokter) ([]rawatjalan.KunjunganRawatJalan, shared.PaginationMeta, error)
	detailKunjunganFn        func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error)
	riwayatKunjunganPasienFn func(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error)
	getWaktuRegistrasiFn     func(ctx context.Context, noRawat string) (tanggal string, jam string, exists bool, err error)
	getInfoRegistrasiFn      func(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error)
	panggilAntreanFn         func(ctx context.Context, noRawat string, kdPoli string, kodeDokter string) (*rawatjalan.HasilPanggilAntrean, error)
}

func (m *mockRawatJalanService) DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter rawatjalan.FilterAntreanDokter) ([]rawatjalan.KunjunganRawatJalan, shared.PaginationMeta, error) {
	if m.daftarAntreanDokterFn != nil {
		return m.daftarAntreanDokterFn(ctx, kodeDokter, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockRawatJalanService) DetailKunjungan(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
	if m.detailKunjunganFn != nil {
		return m.detailKunjunganFn(ctx, noRawat, kodeDokter)
	}
	return nil, nil
}

func (m *mockRawatJalanService) RiwayatKunjunganPasien(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
	if m.riwayatKunjunganPasienFn != nil {
		return m.riwayatKunjunganPasienFn(ctx, noRM)
	}
	return nil, nil
}

func (m *mockRawatJalanService) GetWaktuRegistrasi(ctx context.Context, noRawat string) (tanggal string, jam string, exists bool, err error) {
	if m.getWaktuRegistrasiFn != nil {
		return m.getWaktuRegistrasiFn(ctx, noRawat)
	}
	return "2026-09-03", "08:00:00", true, nil
}

func (m *mockRawatJalanService) GetInfoRegistrasi(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
	if m.getInfoRegistrasiFn != nil {
		return m.getInfoRegistrasiFn(ctx, noRawat)
	}
	return &rawatjalan.InfoRegistrasiPasien{
		TanggalRegistrasi: "2026-09-03",
		JamRegistrasi:     "08:00:00",
		KodePenjamin:      "UMU",
		StatusBayar:       "Belum Bayar",
	}, nil
}

func (m *mockRawatJalanService) PanggilAntrean(ctx context.Context, noRawat string, kdPoli string, kodeDokter string) (*rawatjalan.HasilPanggilAntrean, error) {
	if m.panggilAntreanFn != nil {
		return m.panggilAntreanFn(ctx, noRawat, kdPoli, kodeDokter)
	}
	return &rawatjalan.HasilPanggilAntrean{
		NoRawat:  noRawat,
		KodePoli: kdPoli,
	}, nil
}


func (m *mockRawatJalanService) DaftarStatusPemeriksaan(ctx context.Context) []rawatjalan.OpsiReferensi {
	return []rawatjalan.OpsiReferensi{{Value: "Belum", Label: "Belum Diperiksa"}}
}

func (m *mockRawatJalanService) DaftarStatusLanjut(ctx context.Context) []rawatjalan.OpsiReferensi {
	return []rawatjalan.OpsiReferensi{{Value: "Ralan", Label: "Rawat Jalan"}}
}

func (m *mockRawatJalanService) DaftarStatusBayar(ctx context.Context) []rawatjalan.OpsiReferensi {
	return []rawatjalan.OpsiReferensi{{Value: "Sudah Bayar", Label: "Sudah Bayar"}}
}

func (m *mockRawatJalanService) DaftarJenisAntrean(ctx context.Context) []rawatjalan.OpsiReferensi {
	return []rawatjalan.OpsiReferensi{{Value: "Semua", Label: "Semua Pasien"}}
}

func authMwForTest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{KodeDokter: "DR01"})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func TestRawatJalanHandler_DaftarAntreanDokter(t *testing.T) {
	mockSvc := &mockRawatJalanService{
		daftarAntreanDokterFn: func(ctx context.Context, kodeDokter string, filter rawatjalan.FilterAntreanDokter) ([]rawatjalan.KunjunganRawatJalan, shared.PaginationMeta, error) {
			if kodeDokter != "DR01" {
				t.Errorf("Expected DR01, got %s", kodeDokter)
			}
			return []rawatjalan.KunjunganRawatJalan{
				{
					NoRawat:           "2026/09/03/000001",
					NoRekamMedis:      "00123456",
					NamaPasien:        "Budi",
					NoRegistrasi:      "001",
					StatusPemeriksaan: rawatjalan.StatusBelum,
				},
			}, shared.NewPaginationMeta(1, 1, 20), nil
		},
	}

	handler := rawatjalan.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-jalan/antrean?page=1&limit=20", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Success bool                             `json:"success"`
		Data    []rawatjalan.KunjunganRawatJalan `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].NoRekamMedis != "123456" {
		t.Errorf("Unexpected formatted data: %+v", resp.Data)
	}
	if resp.Data[0].Id == "" || resp.Data[0].IdPasien == "" {
		t.Errorf("Expected encrypted Id and IdPasien")
	}
}

func TestRawatJalanHandler_DetailKunjungan(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncKey)

	mockSvc := &mockRawatJalanService{
		detailKunjunganFn: func(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
			return &rawatjalan.KunjunganRawatJalan{
				NoRawat:      noRawat,
				NoRekamMedis: "00123456",
				NamaPasien:   "Budi",
			}, nil
		},
	}

	handler := rawatjalan.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-jalan/"+encKunjungan, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}
}

func TestRawatJalanHandler_ReferenceRoutes(t *testing.T) {
	handler := rawatjalan.NewHandler(&mockRawatJalanService{}, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, noOpMw, noOpMw, noOpMw)

	routes := []string{
		"/api/v1/rawat-jalan/status-pemeriksaan",
		"/api/v1/rawat-jalan/status-lanjut",
		"/api/v1/rawat-jalan/status-bayar",
		"/api/v1/rawat-jalan/jenis-antrean",
	}

	for _, rt := range routes {
		req := httptest.NewRequest(http.MethodGet, rt, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Route %s returned status %d, expected 200", rt, rr.Code)
		}
	}
}

func TestRawatJalanHandler_ServiceRole(t *testing.T) {
	serviceAuthMw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{
				Role:     middleware.RoleService,
				NamaUser: "External Service",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}

	t.Run("Daftar antrean oleh service mengosongkan kode dokter (mengambil semua)", func(t *testing.T) {
		mockSvc := &mockRawatJalanService{
			daftarAntreanDokterFn: func(ctx context.Context, kodeDokter string, filter rawatjalan.FilterAntreanDokter) ([]rawatjalan.KunjunganRawatJalan, shared.PaginationMeta, error) {
				if kodeDokter != "" {
					t.Errorf("Expected empty kodeDokter for service, got %s", kodeDokter)
				}
				return []rawatjalan.KunjunganRawatJalan{
					{
						NoRawat:           "2026/09/03/000001",
						NoRekamMedis:      "00123456",
						NamaPasien:        "Pasien Umum",
						StatusPemeriksaan: rawatjalan.StatusBelum,
					},
				}, shared.NewPaginationMeta(1, 1, 20), nil
			},
		}

		handler := rawatjalan.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
		handler.RegisterRoutes(mux, authMwForTest, serviceAuthMw, noOpMw)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-jalan/antrean", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}
	})

	t.Run("Detail kunjungan TIDAK boleh diakses oleh service (hanya dokter)", func(t *testing.T) {
		encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncKey)
		mockSvc := &mockRawatJalanService{}

		handler := rawatjalan.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }

		handler.RegisterRoutes(mux, serviceAuthMw, serviceAuthMw, noOpMw)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-jalan/"+encKunjungan, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("Expected status 401 for service accessing DetailKunjungan, got %d", rr.Code)
		}
	})
}

func TestRawatJalanHandler_PanggilAntrean(t *testing.T) {
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }

	t.Run("Sukses panggil via path id_kunjungan", func(t *testing.T) {
		encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncKey)
		mockSvc := &mockRawatJalanService{
			panggilAntreanFn: func(ctx context.Context, noRawat, kdPoli, kodeDokter string) (*rawatjalan.HasilPanggilAntrean, error) {
				if noRawat != "2026/09/03/000001" {
					t.Errorf("expected no_rawat '2026/09/03/000001', got %s", noRawat)
				}
				if kdPoli != "U0001" {
					t.Errorf("expected kd_poli 'U0001', got %s", kdPoli)
				}
				return &rawatjalan.HasilPanggilAntrean{
					NoRawat:  noRawat,
					KodePoli: "U0001",
				}, nil
			},
		}

		handler := rawatjalan.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

		body := `{"kd_poli":"U0001"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rawat-jalan/"+encKunjungan+"/panggil", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]interface{}
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp["success"] != true {
			t.Errorf("Expected success true, got %v", resp["success"])
		}
	})

	t.Run("Error ID kunjungan tidak valid", func(t *testing.T) {
		mockSvc := &mockRawatJalanService{}
		handler := rawatjalan.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rawat-jalan/invalid-encrypted-id/panggil", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400 for invalid ID, got %d", rr.Code)
		}
	})
}


