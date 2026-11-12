package rujukaninternal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/token"
	"erm-dokter/internal/rujukaninternal"
)

const testEncKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockRujukanService struct {
	daftarOpsiFn    func(ctx context.Context, kodeDokterLogin string, keyword string) ([]rujukaninternal.OpsiPoliDokter, error)
	daftarRujukanFn func(ctx context.Context, noRawat string) ([]rujukaninternal.RujukanInternal, error)
	simpanRujukanFn func(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) (*rujukaninternal.RujukanInternal, error)
	hapusRujukanFn  func(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) error
}

func (m *mockRujukanService) DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin string, keyword string) ([]rujukaninternal.OpsiPoliDokter, error) {
	if m.daftarOpsiFn != nil {
		return m.daftarOpsiFn(ctx, kodeDokterLogin, keyword)
	}
	return nil, nil
}

func (m *mockRujukanService) DaftarRujukanInternal(ctx context.Context, noRawat string) ([]rujukaninternal.RujukanInternal, error) {
	if m.daftarRujukanFn != nil {
		return m.daftarRujukanFn(ctx, noRawat)
	}
	return nil, nil
}

func (m *mockRujukanService) SimpanRujukanInternal(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) (*rujukaninternal.RujukanInternal, error) {
	if m.simpanRujukanFn != nil {
		return m.simpanRujukanFn(ctx, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter)
	}
	return nil, nil
}

func (m *mockRujukanService) HapusRujukanInternal(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) error {
	if m.hapusRujukanFn != nil {
		return m.hapusRujukanFn(ctx, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter)
	}
	return nil
}

func authMw(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{KodeDokter: "DR01"})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func TestRujukanInternalHandler(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncKey)
	encRujukan, _ := crypto.Encrypt("2026/09/03/000001~INT~DR02", testEncKey)
	encTujuan, _ := crypto.Encrypt("INT~DR02", testEncKey)

	var capturedDeletePoli, capturedDeleteDokter string

	mockSvc := &mockRujukanService{
		daftarOpsiFn: func(ctx context.Context, kodeDokterLogin string, keyword string) ([]rujukaninternal.OpsiPoliDokter, error) {
			return []rujukaninternal.OpsiPoliDokter{
				{
					InfoPoliDokter: rujukaninternal.InfoPoliDokter{
						KodePoli:   "INT",
						NamaPoli:   "Poli Dalam",
						KodeDokter: "DR02",
						NamaDokter: "dr. Budi",
					},
				},
			}, nil
		},
		daftarRujukanFn: func(ctx context.Context, noRawat string) ([]rujukaninternal.RujukanInternal, error) {
			return []rujukaninternal.RujukanInternal{
				{
					NoRawat: noRawat,
					InfoPoliDokter: rujukaninternal.InfoPoliDokter{
						KodePoli:   "INT",
						NamaPoli:   "Poli Dalam",
						KodeDokter: "DR02",
						NamaDokter: "dr. Budi",
					},
				},
			}, nil
		},
		simpanRujukanFn: func(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) (*rujukaninternal.RujukanInternal, error) {
			return &rujukaninternal.RujukanInternal{
				NoRawat: noRawat,
				InfoPoliDokter: rujukaninternal.InfoPoliDokter{
					KodePoli:   targetKodePoli,
					KodeDokter: targetKodeDokter,
				},
			}, nil
		},
		hapusRujukanFn: func(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) error {
			capturedDeletePoli = targetKodePoli
			capturedDeleteDokter = targetKodeDokter
			return nil
		},
	}

	handler := rujukaninternal.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, authMw, noOpMw)

	t.Run("DaftarRujukanInternal", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rujukan-internal/"+encKunjungan, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}

		var resp struct {
			Success bool                              `json:"success"`
			Data    []rujukaninternal.RujukanInternal `json:"data"`
		}
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		if len(resp.Data) != 1 || resp.Data[0].Id == "" {
			t.Errorf("Expected encrypted Id in response, got %+v", resp.Data)
		}
	})

	t.Run("SimpanRujukanInternal", func(t *testing.T) {
		body, _ := json.Marshal(rujukaninternal.SimpanRujukanRequest{
			IdTujuan: encTujuan,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rujukan-internal/"+encKunjungan, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("HapusRujukanInternal forwards KodePoli and KodeDokter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/rujukan-internal/"+encKunjungan+"/"+encRujukan, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}
		if capturedDeletePoli != "INT" || capturedDeleteDokter != "DR02" {
			t.Fatalf("expected KodePoli=INT and KodeDokter=DR02, got poli=%q dokter=%q", capturedDeletePoli, capturedDeleteDokter)
		}
	})
}
