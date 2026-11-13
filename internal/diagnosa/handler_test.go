package diagnosa_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"erm-dokter/internal/diagnosa"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/shared"
)

const testEncKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockService struct {
	daftarDiagnosaProsedurFunc func(ctx context.Context, noRawat string, status shared.StatusLanjut) (*diagnosa.DaftarDiagnosaProsedurResponse, error)
	riwayatPasienFunc          func(ctx context.Context, noRekamMedis string, status shared.StatusLanjut) (*diagnosa.RiwayatPasienResponse, error)
	tambahDiagnosaFunc         func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.TambahDiagnosaRequest) (*diagnosa.DiagnosaPasien, error)
	updateDiagnosaFunc         func(ctx context.Context, id diagnosa.IdDiagnosa, req diagnosa.UpdateDiagnosaRequest) error
	hapusDiagnosaFunc          func(ctx context.Context, id diagnosa.IdDiagnosa) error
	reorderDiagnosaFunc        func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error
	tambahProsedurFunc         func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.TambahProsedurRequest) (*diagnosa.ProsedurPasien, error)
	updateProsedurFunc         func(ctx context.Context, id diagnosa.IdProsedur, req diagnosa.UpdateProsedurRequest) error
	hapusProsedurFunc          func(ctx context.Context, id diagnosa.IdProsedur) error
	reorderProsedurFunc        func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error
	simulasiEklaimFunc         func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, req diagnosa.SimulasiEklaimRequest) (*diagnosa.SimulasiEklaimResponse, error)
}

func (m *mockService) DaftarDiagnosaProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut) (*diagnosa.DaftarDiagnosaProsedurResponse, error) {
	if m.daftarDiagnosaProsedurFunc != nil {
		return m.daftarDiagnosaProsedurFunc(ctx, noRawat, status)
	}
	return &diagnosa.DaftarDiagnosaProsedurResponse{}, nil
}

func (m *mockService) RiwayatPasien(ctx context.Context, noRekamMedis string, status shared.StatusLanjut) (*diagnosa.RiwayatPasienResponse, error) {
	if m.riwayatPasienFunc != nil {
		return m.riwayatPasienFunc(ctx, noRekamMedis, status)
	}
	return &diagnosa.RiwayatPasienResponse{}, nil
}

func (m *mockService) TambahDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.TambahDiagnosaRequest) (*diagnosa.DiagnosaPasien, error) {
	if m.tambahDiagnosaFunc != nil {
		return m.tambahDiagnosaFunc(ctx, noRawat, status, req)
	}
	return &diagnosa.DiagnosaPasien{NoRawat: noRawat, Kode: req.Kode, Status: status}, nil
}

func (m *mockService) UpdateDiagnosa(ctx context.Context, id diagnosa.IdDiagnosa, req diagnosa.UpdateDiagnosaRequest) error {
	if m.updateDiagnosaFunc != nil {
		return m.updateDiagnosaFunc(ctx, id, req)
	}
	return nil
}

func (m *mockService) HapusDiagnosa(ctx context.Context, id diagnosa.IdDiagnosa) error {
	if m.hapusDiagnosaFunc != nil {
		return m.hapusDiagnosaFunc(ctx, id)
	}
	return nil
}

func (m *mockService) ReorderDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error {
	if m.reorderDiagnosaFunc != nil {
		return m.reorderDiagnosaFunc(ctx, noRawat, status, req)
	}
	return nil
}

func (m *mockService) TambahProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.TambahProsedurRequest) (*diagnosa.ProsedurPasien, error) {
	if m.tambahProsedurFunc != nil {
		return m.tambahProsedurFunc(ctx, noRawat, status, req)
	}
	return &diagnosa.ProsedurPasien{NoRawat: noRawat, Kode: req.Kode, Status: status}, nil
}

func (m *mockService) UpdateProsedur(ctx context.Context, id diagnosa.IdProsedur, req diagnosa.UpdateProsedurRequest) error {
	if m.updateProsedurFunc != nil {
		return m.updateProsedurFunc(ctx, id, req)
	}
	return nil
}

func (m *mockService) HapusProsedur(ctx context.Context, id diagnosa.IdProsedur) error {
	if m.hapusProsedurFunc != nil {
		return m.hapusProsedurFunc(ctx, id)
	}
	return nil
}

func (m *mockService) ReorderProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error {
	if m.reorderProsedurFunc != nil {
		return m.reorderProsedurFunc(ctx, noRawat, status, req)
	}
	return nil
}

func (m *mockService) SimulasiEklaim(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, req diagnosa.SimulasiEklaimRequest) (*diagnosa.SimulasiEklaimResponse, error) {
	if m.simulasiEklaimFunc != nil {
		return m.simulasiEklaimFunc(ctx, noRawat, statusLanjut, req)
	}
	return &diagnosa.SimulasiEklaimResponse{
		KodeCBG:      "I-4-17-I",
		DeskripsiCBG: "GAGAL JANTUNG & SYOK KARDIOGENIK RINGAN",
		Tarif:        4250000,
	}, nil
}

func noopMw(next http.HandlerFunc) http.HandlerFunc {
	return next
}

func setupTestServer(svc diagnosa.Service) *http.ServeMux {
	mux := http.NewServeMux()
	handler := diagnosa.NewHandler(svc, testEncKey)
	handler.RegisterRoutes(mux, noopMw, noopMw)
	return mux
}

func TestHandler_DaftarDiagnosaProsedur(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/17/000001", testEncKey)

	t.Run("Sukses get data dan enkripsi ID", func(t *testing.T) {
		svc := &mockService{
			daftarDiagnosaProsedurFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut) (*diagnosa.DaftarDiagnosaProsedurResponse, error) {
				return &diagnosa.DaftarDiagnosaProsedurResponse{
					Diagnosa: []diagnosa.DiagnosaPasien{
						{NoRawat: noRawat, Kode: "I63.9", Nama: "Stroke", Status: status, Prioritas: 1, StatusPenyakit: "Baru"},
					},
					Prosedur: []diagnosa.ProsedurPasien{
						{NoRawat: noRawat, Kode: "87.03", Nama: "CT Scan", Status: status, Prioritas: 1},
					},
				}, nil
			},
		}

		mux := setupTestServer(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnosa/ralan/"+encKunjungan, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Success bool `json:"success"`
			Data    struct {
				Diagnosa []diagnosa.DiagnosaPasien `json:"diagnosa"`
				Prosedur []diagnosa.ProsedurPasien `json:"prosedur"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse json: %v", err)
		}
		if len(resp.Data.Diagnosa) != 1 || resp.Data.Diagnosa[0].Id == "" {
			t.Errorf("Expected populated diagnosa with encrypted ID, got %+v", resp.Data.Diagnosa)
		}
		if len(resp.Data.Prosedur) != 1 || resp.Data.Prosedur[0].Id == "" {
			t.Errorf("Expected populated prosedur with encrypted ID, got %+v", resp.Data.Prosedur)
		}
	})

	t.Run("Status lanjut tidak valid", func(t *testing.T) {
		mux := setupTestServer(&mockService{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnosa/invalid_status/"+encKunjungan, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rec.Code)
		}
	})

	t.Run("ID kunjungan token tidak valid", func(t *testing.T) {
		mux := setupTestServer(&mockService{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnosa/ralan/invalid_token", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandler_RiwayatPasien(t *testing.T) {
	encRM, _ := crypto.Encrypt("123456", testEncKey)

	t.Run("Sukses get riwayat pasien", func(t *testing.T) {
		svc := &mockService{
			riwayatPasienFunc: func(ctx context.Context, noRekamMedis string, status shared.StatusLanjut) (*diagnosa.RiwayatPasienResponse, error) {
				return &diagnosa.RiwayatPasienResponse{
					Diagnosa: []diagnosa.DiagnosaPasien{
						{NoRawat: "RAWAT-01", Kode: "I10", Nama: "Hipertensi", Status: status, Prioritas: 1, StatusPenyakit: "Baru"},
					},
					Prosedur: []diagnosa.ProsedurPasien{},
				}, nil
			},
		}

		mux := setupTestServer(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnosa/ralan/pasien/"+encRM, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Sukses get riwayat pasien dengan status semua", func(t *testing.T) {
		svc := &mockService{
			riwayatPasienFunc: func(ctx context.Context, noRekamMedis string, status shared.StatusLanjut) (*diagnosa.RiwayatPasienResponse, error) {
				if status != "Semua" {
					t.Errorf("Expected status 'Semua', got '%s'", status)
				}
				return &diagnosa.RiwayatPasienResponse{
					Diagnosa: []diagnosa.DiagnosaPasien{
						{NoRawat: "RAWAT-01", Kode: "I10", Nama: "Hipertensi", Status: shared.StatusLanjutRawatJalan, Prioritas: 1, StatusPenyakit: "Baru"},
						{NoRawat: "RAWAT-02", Kode: "I63.9", Nama: "Stroke", Status: shared.StatusLanjutRawatInap, Prioritas: 1, StatusPenyakit: "Baru"},
					},
					Prosedur: []diagnosa.ProsedurPasien{},
				}, nil
			},
		}

		mux := setupTestServer(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnosa/semua/pasien/"+encRM, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Data struct {
				Diagnosa []diagnosa.DiagnosaPasien `json:"diagnosa"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		if len(resp.Data.Diagnosa) != 2 {
			t.Errorf("Expected 2 diagnosa items, got %d", len(resp.Data.Diagnosa))
		}
	})
}

func TestHandler_TambahDiagnosa(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/17/000001", testEncKey)

	t.Run("Sukses tambah diagnosa", func(t *testing.T) {
		svc := &mockService{
			tambahDiagnosaFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.TambahDiagnosaRequest) (*diagnosa.DiagnosaPasien, error) {
				return &diagnosa.DiagnosaPasien{
					NoRawat:        noRawat,
					Kode:           req.Kode,
					Nama:           "Stroke",
					Status:         status,
					Prioritas:      1,
					StatusPenyakit: "Baru",
				}, nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(`{"kode":"I63.9"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/diagnosa/ralan/"+encKunjungan, body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Gagal validasi input kode kosong", func(t *testing.T) {
		mux := setupTestServer(&mockService{})
		body := bytes.NewBufferString(`{"kode":""}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/diagnosa/ralan/"+encKunjungan, body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandler_UpdateDiagnosa(t *testing.T) {
	noRawat := "2026/09/17/000001"
	kode := "I63.9"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncKey)
	encId, _ := crypto.Encrypt(fmt.Sprintf("%s~%s~Ralan", noRawat, kode), testEncKey)

	t.Run("Sukses update diagnosa", func(t *testing.T) {
		svc := &mockService{
			updateDiagnosaFunc: func(ctx context.Context, id diagnosa.IdDiagnosa, req diagnosa.UpdateDiagnosaRequest) error {
				return nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(`{"prioritas":2,"status_penyakit":"Lama"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/diagnosa/ralan/"+encKunjungan+"/"+encId, body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Gagal ID diagnosa tidak sesuai dengan ID kunjungan", func(t *testing.T) {
		mux := setupTestServer(&mockService{})
		body := bytes.NewBufferString(`{"prioritas":2}`)
		differentKunjungan, _ := crypto.Encrypt("2026/09/17/999999", testEncKey)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/diagnosa/ralan/"+differentKunjungan+"/"+encId, body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandler_HapusDiagnosa(t *testing.T) {
	noRawat := "2026/09/17/000001"
	kode := "I63.9"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncKey)
	encId, _ := crypto.Encrypt(fmt.Sprintf("%s~%s~Ralan", noRawat, kode), testEncKey)

	t.Run("Sukses hapus diagnosa", func(t *testing.T) {
		svc := &mockService{
			hapusDiagnosaFunc: func(ctx context.Context, id diagnosa.IdDiagnosa) error {
				return nil
			},
		}

		mux := setupTestServer(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/diagnosa/ralan/"+encKunjungan+"/"+encId, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestHandler_ReorderDiagnosa(t *testing.T) {
	noRawat := "2026/09/17/000001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncKey)

	t.Run("Sukses reorder diagnosa dengan kode (kontrak FE)", func(t *testing.T) {
		svc := &mockService{
			reorderDiagnosaFunc: func(ctx context.Context, gotNoRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error {
				if len(req.Items) != 2 || req.Items[0].Kode != "I63.9" || req.Items[1].Kode != "I10" {
					t.Fatalf("unexpected items: %+v", req.Items)
				}
				return nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(`{"items":[{"kode":"I63.9","prioritas":1},{"kode":"I10","prioritas":2}]}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/diagnosa/ralan/"+encKunjungan+"/reorder", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Sukses reorder diagnosa dengan id terenkripsi", func(t *testing.T) {
		encId1, _ := crypto.Encrypt(noRawat+"~I63.9~Ralan", testEncKey)
		encId2, _ := crypto.Encrypt(noRawat+"~I10~Ralan", testEncKey)

		svc := &mockService{
			reorderDiagnosaFunc: func(ctx context.Context, gotNoRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error {
				if len(req.Items) != 2 || req.Items[0].Kode != "I63.9" || req.Items[1].Kode != "I10" {
					t.Fatalf("expected decrypted kode I63.9 and I10, got %+v", req.Items)
				}
				return nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(fmt.Sprintf(`{"items":[{"id":%q,"prioritas":1},{"id":%q,"prioritas":2}]}`, encId1, encId2))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/diagnosa/ralan/"+encKunjungan+"/reorder", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Gagal jika prioritas duplikat", func(t *testing.T) {
		mux := setupTestServer(&mockService{})
		body := bytes.NewBufferString(`{"items":[{"kode":"I63.9","prioritas":1},{"kode":"I10","prioritas":1}]}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/diagnosa/ralan/"+encKunjungan+"/reorder", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Gagal jika id berasal dari status lanjut berbeda", func(t *testing.T) {
		encIdRanap, _ := crypto.Encrypt(noRawat+"~I63.9~Ranap", testEncKey)
		mux := setupTestServer(&mockService{})
		body := bytes.NewBufferString(fmt.Sprintf(`{"items":[{"id":%q,"prioritas":1}]}`, encIdRanap))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/diagnosa/ralan/"+encKunjungan+"/reorder", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestHandler_ProsedurCRUD(t *testing.T) {
	noRawat := "2026/09/17/000001"
	kode := "87.03"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncKey)
	encId, _ := crypto.Encrypt(fmt.Sprintf("%s~%s~Ralan", noRawat, kode), testEncKey)

	t.Run("Sukses tambah prosedur", func(t *testing.T) {
		svc := &mockService{
			tambahProsedurFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.TambahProsedurRequest) (*diagnosa.ProsedurPasien, error) {
				return &diagnosa.ProsedurPasien{
					NoRawat:   noRawat,
					Kode:      req.Kode,
					Nama:      "CT Scan",
					Status:    status,
					Prioritas: 1,
				}, nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(`{"kode":"87.03"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/prosedur/ralan/"+encKunjungan, body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Sukses update prosedur", func(t *testing.T) {
		svc := &mockService{
			updateProsedurFunc: func(ctx context.Context, id diagnosa.IdProsedur, req diagnosa.UpdateProsedurRequest) error {
				return nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(`{"prioritas":3}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/prosedur/ralan/"+encKunjungan+"/"+encId, body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Sukses hapus prosedur", func(t *testing.T) {
		svc := &mockService{
			hapusProsedurFunc: func(ctx context.Context, id diagnosa.IdProsedur) error {
				return nil
			},
		}

		mux := setupTestServer(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/prosedur/ralan/"+encKunjungan+"/"+encId, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Sukses reorder prosedur", func(t *testing.T) {
		svc := &mockService{
			reorderProsedurFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error {
				if len(req.Items) != 1 || req.Items[0].Kode != "87.03" {
					t.Fatalf("unexpected items: %+v", req.Items)
				}
				return nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(`{"items":[{"kode":"87.03","prioritas":1}]}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/prosedur/ralan/"+encKunjungan+"/reorder", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Sukses reorder prosedur dengan id terenkripsi", func(t *testing.T) {
		svc := &mockService{
			reorderProsedurFunc: func(ctx context.Context, noRawat string, status shared.StatusLanjut, req diagnosa.ReorderRequest) error {
				if len(req.Items) != 1 || req.Items[0].Kode != kode {
					t.Fatalf("expected decrypted kode %q, got %+v", kode, req.Items)
				}
				return nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(fmt.Sprintf(`{"items":[{"id":%q,"prioritas":1}]}`, encId))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/prosedur/ralan/"+encKunjungan+"/reorder", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestSimulasiEklaim_Handler(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/18/000001", testEncKey)

	t.Run("Sukses simulasi E-Klaim via handler", func(t *testing.T) {
		svc := &mockService{
			simulasiEklaimFunc: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, req diagnosa.SimulasiEklaimRequest) (*diagnosa.SimulasiEklaimResponse, error) {
				if noRawat != "2026/09/18/000001" {
					t.Errorf("Unexpected noRawat: %s", noRawat)
				}
				if statusLanjut != shared.StatusLanjutRawatJalan {
					t.Errorf("Expected statusLanjut Ralan, got %s", statusLanjut)
				}
				return &diagnosa.SimulasiEklaimResponse{
					KodeCBG:      "I-4-17-I",
					DeskripsiCBG: "GAGAL JANTUNG & SYOK KARDIOGENIK RINGAN",
					Tarif:        4250000,
				}, nil
			},
		}

		mux := setupTestServer(svc)
		body := bytes.NewBufferString(`{"diagnosa":["I50.9"],"prosedur":["88.72"]}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/diagnosa/ralan/"+encKunjungan+"/simulasi-eklaim", body)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp["message"] != "Simulasi biaya E-Klaim berhasil" {
			t.Errorf("Unexpected message: %v", resp["message"])
		}
	})

	t.Run("Gagal jika status_lanjut tidak valid", func(t *testing.T) {
		svc := &mockService{}
		mux := setupTestServer(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/diagnosa/invalid/"+encKunjungan+"/simulasi-eklaim", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Gagal jika ID kunjungan tidak valid", func(t *testing.T) {
		svc := &mockService{}
		mux := setupTestServer(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/diagnosa/ralan/invalid-token/simulasi-eklaim", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", rec.Code)
		}
	})
}
