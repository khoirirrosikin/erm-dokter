package pasien_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pasien"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/token"
	"erm-dokter/internal/shared"
)

const testEncKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockPasienService struct {
	detailPasienFn           func(ctx context.Context, noRM string) (*pasien.Pasien, error)
	cariPasienFn             func(ctx context.Context, kataKunci string) ([]pasien.Pasien, error)
	getNoRMByNoRawatFn       func(ctx context.Context, noRawat string) (string, error)
	riwayatKunjunganPasienFn func(ctx context.Context, noRM string, filter pasien.FilterRiwayatKunjungan) ([]pasien.RiwayatKunjungan, shared.PaginationMeta, error)
}

func (m *mockPasienService) DetailPasien(ctx context.Context, noRM string) (*pasien.Pasien, error) {
	if m.detailPasienFn != nil {
		return m.detailPasienFn(ctx, noRM)
	}
	return nil, nil
}

func (m *mockPasienService) CariPasien(ctx context.Context, kataKunci string) ([]pasien.Pasien, error) {
	if m.cariPasienFn != nil {
		return m.cariPasienFn(ctx, kataKunci)
	}
	return nil, nil
}

func (m *mockPasienService) GetNoRMByNoRawat(ctx context.Context, noRawat string) (string, error) {
	if m.getNoRMByNoRawatFn != nil {
		return m.getNoRMByNoRawatFn(ctx, noRawat)
	}
	return "", nil
}

func (m *mockPasienService) RiwayatKunjunganPasien(ctx context.Context, noRM string, filter pasien.FilterRiwayatKunjungan) ([]pasien.RiwayatKunjungan, shared.PaginationMeta, error) {
	if m.riwayatKunjunganPasienFn != nil {
		return m.riwayatKunjunganPasienFn(ctx, noRM, filter)
	}
	return nil, shared.PaginationMeta{}, nil
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

func TestPasienHandler_RiwayatKunjungan(t *testing.T) {
	t.Run("Success via encrypted id_pasien in path", func(t *testing.T) {
		mockSvc := &mockPasienService{
			riwayatKunjunganPasienFn: func(ctx context.Context, noRM string, filter pasien.FilterRiwayatKunjungan) ([]pasien.RiwayatKunjungan, shared.PaginationMeta, error) {
				if noRM != "00123456" {
					t.Fatalf("unexpected noRM %s", noRM)
				}
				if filter.Limit != 3 {
					t.Fatalf("expected default limit 3, got %d", filter.Limit)
				}
				return []pasien.RiwayatKunjungan{
					{
						NoRawat:      "2026/09/01/000001",
						NoRekamMedis: "00123456",
						NamaPasien:   "Budi Santoso",
						StatusLanjut: shared.StatusLanjutRawatJalan,
						NamaPoli:     "Poli Penyakit Dalam",
						NamaDokter:   "dr. Budi",
					},
					{
						NoRawat:      "2026/08/10/000005",
						NoRekamMedis: "00123456",
						NamaPasien:   "Budi Santoso",
						StatusLanjut: shared.StatusLanjutRawatInap,
						NamaDokter:   "dr. Budi, Sp.PD",
						KamarInap: &pasien.RiwayatRawatInap{
							DPJP: []string{"dr. Budi, Sp.PD"},
							Kamar: []pasien.RiwayatKamarInap{
								{
									KodeKamar:    "K01",
									NamaBangsal:  "Bangsal Melati",
									Kelas:        "Kelas 1",
									TanggalMasuk: "2026-08-10",
									JamMasuk:     "08:00:00",
								},
							},
						},
					},
				}, shared.NewPaginationMeta(2, 1, 3), nil
			},
		}

		handler := pasien.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, noOpMw)

		encPasien, err := crypto.Encrypt("00123456", testEncKey)
		if err != nil {
			t.Fatalf("encrypt error: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/pasien/"+encPasien+"/riwayat-kunjungan", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Success bool                      `json:"success"`
			Message string                    `json:"message"`
			Data    []pasien.RiwayatKunjungan `json:"data"`
			Meta    shared.PaginationMeta     `json:"meta"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response error: %v", err)
		}

		if len(resp.Data) != 2 {
			t.Fatalf("expected 2 items, got %d", len(resp.Data))
		}

		item0 := resp.Data[0]
		if item0.NoRekamMedis != "123456" {
			t.Errorf("expected formatted NoRekamMedis 123456, got %s", item0.NoRekamMedis)
		}

		decRawat, err := crypto.Decrypt(item0.IdKunjungan, testEncKey)
		if err != nil || decRawat != "2026/09/01/000001" {
			t.Errorf("expected decrypted IdKunjungan 2026/09/01/000001, got %s, err: %v", decRawat, err)
		}

		decPasien, err := crypto.Decrypt(item0.IdPasien, testEncKey)
		if err != nil || decPasien != "00123456" {
			t.Errorf("expected decrypted IdPasien 00123456, got %s, err: %v", decPasien, err)
		}
	})

	t.Run("Rejects plain no_rm in path to prevent medical record enumeration", func(t *testing.T) {
		handler := pasien.NewHandler(&mockPasienService{}, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, noOpMw)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/pasien/123456/riwayat-kunjungan", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for plaintext no_rm, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Rejects request without valid KodeDokter in context", func(t *testing.T) {
		handler := pasien.NewHandler(&mockPasienService{}, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, noOpMw, noOpMw)

		encPasien, _ := crypto.Encrypt("00123456", testEncKey)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/pasien/"+encPasien+"/riwayat-kunjungan", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 without doctor claim, got %d", rr.Code)
		}
	})

	t.Run("Error invalid encrypted id_pasien token", func(t *testing.T) {
		handler := pasien.NewHandler(&mockPasienService{}, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, noOpMw)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/pasien/invalid-long-token-that-is-not-valid-aes-gcm-base64/riwayat-kunjungan", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for bad id_pasien token, got %d", rr.Code)
		}
	})

	t.Run("Validation error on invalid tanggal range", func(t *testing.T) {
		handler := pasien.NewHandler(&mockPasienService{}, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, noOpMw)

		encPasien, _ := crypto.Encrypt("00123456", testEncKey)
		for _, badTanggal := range []string{"invalid-date", "abc,xyz", "2026-12-31,2026-01-01"} {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/pasien/"+encPasien+"/riwayat-kunjungan?tanggal="+badTanggal, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400 for tanggal=%s, got %d", badTanggal, rr.Code)
			}
		}
	})

	t.Run("Service error returns 500", func(t *testing.T) {
		mockSvc := &mockPasienService{
			riwayatKunjunganPasienFn: func(ctx context.Context, noRM string, filter pasien.FilterRiwayatKunjungan) ([]pasien.RiwayatKunjungan, shared.PaginationMeta, error) {
				return nil, shared.PaginationMeta{}, errors.New("db down")
			},
		}

		handler := pasien.NewHandler(mockSvc, testEncKey)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, authMwForTest, noOpMw)

		encPasien, _ := crypto.Encrypt("00123456", testEncKey)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/pasien/"+encPasien+"/riwayat-kunjungan", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500 for service error, got %d", rr.Code)
		}
	})
}
