package rawatinap_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/token"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/shared"
)

const testEncKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockRawatInapService struct {
	cekStatusKamarInapFn    func(ctx context.Context, noRawat string) (bool, bool, error)
	daftarPasienRawatInapFn func(ctx context.Context, kodeDokterLogin string, filter rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, shared.PaginationMeta, error)
	daftarStatusPulangFn    func(ctx context.Context) []rawatinap.OpsiReferensi
	detailPasienRawatInapFn func(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*rawatinap.KunjunganRawatInap, error)
}

func (m *mockRawatInapService) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	if m.cekStatusKamarInapFn != nil {
		return m.cekStatusKamarInapFn(ctx, noRawat)
	}
	return false, false, nil
}

func (m *mockRawatInapService) DaftarPasienRawatInap(ctx context.Context, kodeDokterLogin string, filter rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, shared.PaginationMeta, error) {
	if m.daftarPasienRawatInapFn != nil {
		return m.daftarPasienRawatInapFn(ctx, kodeDokterLogin, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockRawatInapService) DaftarStatusPulang(ctx context.Context) []rawatinap.OpsiReferensi {
	if m.daftarStatusPulangFn != nil {
		return m.daftarStatusPulangFn(ctx)
	}
	return []rawatinap.OpsiReferensi{{Value: "-", Label: "Belum Pulang"}}
}

func (m *mockRawatInapService) DetailPasienRawatInap(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*rawatinap.KunjunganRawatInap, error) {
	if m.detailPasienRawatInapFn != nil {
		return m.detailPasienRawatInapFn(ctx, noRawat, tglMasuk, jamMasuk)
	}
	return nil, nil
}

func (m *mockRawatInapService) GetKelasRawat(ctx context.Context, noRawat string) (string, error) {
	return "", nil
}

func (m *mockRawatInapService) GetWaktuKeluarTerakhir(ctx context.Context, noRawat string) (string, string, error) {
	return "", "", nil
}

func authMwForTest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{KodeDokter: "DR01"})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func noOpMw(next http.HandlerFunc) http.HandlerFunc {
	return next
}

func TestRawatInapHandler_DaftarPasienRawatInap_Success(t *testing.T) {
	mockSvc := &mockRawatInapService{
		daftarPasienRawatInapFn: func(ctx context.Context, kodeDokterLogin string, filter rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, shared.PaginationMeta, error) {
			if kodeDokterLogin != "DR01" {
				t.Errorf("expected kodeDokterLogin DR01, got %s", kodeDokterLogin)
			}
			if filter.Bangsal != "B01" {
				t.Errorf("expected Bangsal B01, got %s", filter.Bangsal)
			}
			return []rawatinap.KunjunganRawatInap{
				{
					NoRawat:          "2026/09/01/000001",
					NoRekamMedis:     "00123456",
					NamaPasien:       "Pasien Ranap",
					TanggalMasuk:     "2026-09-01",
					JamMasuk:         "10:00:00",
					KodeKamar:        "K01",
					KelasKamar:       "Kelas 1",
					KodeBangsal:      "B01",
					DokterRawatJalan: "dr. Hendra",
					DPJP:             []string{"dr. Budi, Sp.PD"},
				},
			}, shared.NewPaginationMeta(1, 1, 20), nil
		},
	}

	handler := rawatinap.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/pasien?bangsal=B01&page=1&limit=20", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Success bool                           `json:"success"`
		Message string                         `json:"message"`
		Data    []rawatinap.KunjunganRawatInap `json:"data"`
		Meta    shared.PaginationMeta          `json:"meta"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Data))
	}

	item := resp.Data[0]
	if item.NoRekamMedis != "123456" {
		t.Errorf("expected formatted NoRekamMedis 123456, got %s", item.NoRekamMedis)
	}

	decryptedId, err := crypto.Decrypt(item.Id, testEncKey)
	if err != nil {
		t.Fatalf("failed to decrypt item.Id: %v", err)
	}
	if decryptedId != "2026/09/01/000001~2026-09-01~10:00:00" {
		t.Errorf("decryptedId got %s, want 2026/09/01/000001~2026-09-01~10:00:00", decryptedId)
	}

	decryptedIdKunjungan, err := crypto.Decrypt(item.IdKunjungan, testEncKey)
	if err != nil {
		t.Fatalf("failed to decrypt item.IdKunjungan: %v", err)
	}
	if decryptedIdKunjungan != "2026/09/01/000001" {
		t.Errorf("decryptedIdKunjungan got %s, want 2026/09/01/000001", decryptedIdKunjungan)
	}

	decryptedIdPasien, err := crypto.Decrypt(item.IdPasien, testEncKey)
	if err != nil {
		t.Fatalf("failed to decrypt item.IdPasien: %v", err)
	}
	if decryptedIdPasien != "00123456" {
		t.Errorf("decryptedIdPasien got %s, want 00123456", decryptedIdPasien)
	}
}

func TestRawatInapHandler_DaftarPasienRawatInap_ValidationError(t *testing.T) {
	mockSvc := &mockRawatInapService{}
	handler := rawatinap.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/pasien?sort_order=INVALID", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestRawatInapHandler_DaftarPasienRawatInap_ServiceError(t *testing.T) {
	mockSvc := &mockRawatInapService{
		daftarPasienRawatInapFn: func(ctx context.Context, kodeDokterLogin string, filter rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, shared.PaginationMeta, error) {
			return nil, shared.PaginationMeta{}, errors.New("db error")
		},
	}
	handler := rawatinap.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/pasien", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

func TestRawatInapHandler_DaftarStatusPulang(t *testing.T) {
	mockSvc := &mockRawatInapService{
		daftarStatusPulangFn: func(ctx context.Context) []rawatinap.OpsiReferensi {
			return []rawatinap.OpsiReferensi{
				{Value: "Sehat", Label: "Sehat"},
				{Value: "-", Label: "Belum Pulang (-)"},
			}
		},
	}
	handler := rawatinap.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/status-pulang", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Success bool                      `json:"success"`
		Message string                    `json:"message"`
		Data    []rawatinap.OpsiReferensi `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Data))
	}
	if resp.Data[0].Value != "Sehat" {
		t.Errorf("expected first item Sehat, got %s", resp.Data[0].Value)
	}
}

func TestRawatInapHandler_DetailPasienRawatInap(t *testing.T) {
	t.Run("Success with valid composite ID", func(t *testing.T) {
		mockSvc := &mockRawatInapService{
			detailPasienRawatInapFn: func(ctx context.Context, noRawat, tglMasuk, jamMasuk string) (*rawatinap.KunjunganRawatInap, error) {
				if noRawat != "2026/09/01/000001" || tglMasuk != "2026-09-01" || jamMasuk != "10:00:00" {
					t.Fatalf("unexpected params: %s %s %s", noRawat, tglMasuk, jamMasuk)
				}
				return &rawatinap.KunjunganRawatInap{
					NoRawat:         noRawat,
					TanggalMasuk:    tglMasuk,
					JamMasuk:        jamMasuk,
					NoRekamMedis:    "00123456",
					NamaPasien:      "Budi Santoso",
					KodeKamar:       "K01",
					NamaBangsal:     "Melati",
					GolonganDarah:   "O",
					Agama:           "Islam",
					PenanggungJawab: "Siti",
				}, nil
			},
		}

		handler := rawatinap.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

		encId, err := crypto.Encrypt("2026/09/01/000001~2026-09-01~10:00:00", testEncKey)
		if err != nil {
			t.Fatalf("encrypt error: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/"+encId, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Success bool                         `json:"success"`
			Message string                       `json:"message"`
			Data    rawatinap.KunjunganRawatInap `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response error: %v", err)
		}

		item := resp.Data
		if item.NamaPasien != "Budi Santoso" {
			t.Errorf("expected Budi Santoso, got %s", item.NamaPasien)
		}
		if item.NoRekamMedis != "123456" {
			t.Errorf("expected formatted NoRekamMedis 123456, got %s", item.NoRekamMedis)
		}

		decId, err := crypto.Decrypt(item.Id, testEncKey)
		if err != nil || decId != "2026/09/01/000001~2026-09-01~10:00:00" {
			t.Errorf("expected decrypted Id to match composite, got %s, err: %v", decId, err)
		}

		decIdKunjungan, err := crypto.Decrypt(item.IdKunjungan, testEncKey)
		if err != nil || decIdKunjungan != "2026/09/01/000001" {
			t.Errorf("expected decrypted IdKunjungan to match NoRawat, got %s, err: %v", decIdKunjungan, err)
		}

		decIdPasien, err := crypto.Decrypt(item.IdPasien, testEncKey)
		if err != nil || decIdPasien != "00123456" {
			t.Errorf("expected decrypted IdPasien to match 00123456, got %s, err: %v", decIdPasien, err)
		}
	})

	t.Run("Invalid token format returns 400", func(t *testing.T) {
		handler := rawatinap.NewHandler(&mockRawatInapService{}, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

		encInvalid, _ := crypto.Encrypt("2026/09/01/000001", testEncKey)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/"+encInvalid, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for non-composite ID, got %d", rr.Code)
		}
	})

	t.Run("Bad encrypted token returns 400", func(t *testing.T) {
		handler := rawatinap.NewHandler(&mockRawatInapService{}, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/invalid-token-123", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for bad token, got %d", rr.Code)
		}
	})

	t.Run("Not found returns 404", func(t *testing.T) {
		mockSvc := &mockRawatInapService{
			detailPasienRawatInapFn: func(ctx context.Context, noRawat, tglMasuk, jamMasuk string) (*rawatinap.KunjunganRawatInap, error) {
				return nil, nil
			},
		}

		handler := rawatinap.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, authMwForTest, noOpMw)

		encId, _ := crypto.Encrypt("2026/09/01/000001~2026-09-01~10:00:00", testEncKey)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/"+encId, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rr.Code)
		}
	})
}

func TestRawatInapHandler_ServiceRole(t *testing.T) {
	serviceAuthMw := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{
				Role:     middleware.RoleService,
				NamaUser: "External Service",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }

	t.Run("Service role default scope_dpjp menjadi semua dan kodeDokter kosong", func(t *testing.T) {
		mockSvc := &mockRawatInapService{
			daftarPasienRawatInapFn: func(ctx context.Context, kodeDokterLogin string, filter rawatinap.FilterPasienRawatInap) ([]rawatinap.KunjunganRawatInap, shared.PaginationMeta, error) {
				if kodeDokterLogin != "" {
					t.Errorf("expected empty kodeDokterLogin for service, got %s", kodeDokterLogin)
				}
				if filter.ScopeDPJP != rawatinap.ScopeDPJPSemua {
					t.Errorf("expected ScopeDPJP to default to 'semua', got %s", filter.ScopeDPJP)
				}
				return []rawatinap.KunjunganRawatInap{
					{
						NoRawat:      "2026/09/01/000001",
						NoRekamMedis: "00123456",
						NamaPasien:   "Pasien Ranap",
						TanggalMasuk: "2026-09-01",
						JamMasuk:     "10:00:00",
					},
				}, shared.NewPaginationMeta(1, 1, 20), nil
			},
		}

		handler := rawatinap.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, serviceAuthMw, noOpMw)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/pasien", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})
}
