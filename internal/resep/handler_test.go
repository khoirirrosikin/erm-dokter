package resep_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/token"
	"erm-dokter/internal/resep"
	"erm-dokter/internal/shared"
)

const testEncKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockResepService struct {
	daftarResepFn     func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, shared.PaginationMeta, error)
	daftarResepByRMFn func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, shared.PaginationMeta, error)
	detailResepFn     func(ctx context.Context, noResep string) (*resep.Resep, error)
	simpanResepFn     func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error)
	hapusResepFn      func(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut) error
	updateResepFn     func(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error)
	daftarAturanFn    func(ctx context.Context, keyword string) ([]resep.AturanPakai, error)
	daftarMetodeFn    func(ctx context.Context) ([]resep.MetodeRacik, error)
}

func (m *mockResepService) DaftarResep(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, shared.PaginationMeta, error) {
	if m.daftarResepFn != nil {
		return m.daftarResepFn(ctx, noRawat, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockResepService) DaftarResepByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, shared.PaginationMeta, error) {
	if m.daftarResepByRMFn != nil {
		return m.daftarResepByRMFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockResepService) DetailResep(ctx context.Context, noResep string) (*resep.Resep, error) {
	if m.detailResepFn != nil {
		return m.detailResepFn(ctx, noResep)
	}
	return nil, nil
}

func (m *mockResepService) SimpanResep(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error) {
	if m.simpanResepFn != nil {
		return m.simpanResepFn(ctx, kodeDokter, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockResepService) HapusResep(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut) error {
	if m.hapusResepFn != nil {
		return m.hapusResepFn(ctx, kodeDokter, noRawat, noResep, statusLanjut)
	}
	return nil
}

func (m *mockResepService) UpdateResep(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error) {
	if m.updateResepFn != nil {
		return m.updateResepFn(ctx, kodeDokter, noRawat, noResep, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockResepService) DaftarAturanPakai(ctx context.Context, keyword string) ([]resep.AturanPakai, error) {
	if m.daftarAturanFn != nil {
		return m.daftarAturanFn(ctx, keyword)
	}
	return []resep.AturanPakai{{AturanPakai: "3x1 Sehari"}}, nil
}

func (m *mockResepService) DaftarMetodeRacik(ctx context.Context) ([]resep.MetodeRacik, error) {
	if m.daftarMetodeFn != nil {
		return m.daftarMetodeFn(ctx)
	}
	return []resep.MetodeRacik{{Kode: "PULV", Nama: "Puyer"}}, nil
}

func authMwForResep(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{KodeDokter: "DR01"})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func TestResepHandler_DaftarResep(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncKey)
	mockSvc := &mockResepService{
		daftarResepFn: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter resep.FilterDaftarResep) ([]resep.Resep, shared.PaginationMeta, error) {
			return []resep.Resep{
				{
					NoResep:          "202609030001",
					NoRawat:          noRawat,
					TanggalPeresepan: "2026-09-03",
					JamPeresepan:     "10:00:00",
					Status:           "ralan",
				},
			}, shared.NewPaginationMeta(1, 1, 20), nil
		},
	}

	handler := resep.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, authMwForResep, noOpMw)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resep/Ralan/"+encKunjungan+"?page=1&limit=20", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Success bool          `json:"success"`
		Data    []resep.Resep `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].Id == "" {
		t.Errorf("Expected encrypted Id in resep response, got %+v", resp.Data)
	}
}

func TestResepHandler_SimpanResep(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/03/000001", testEncKey)
	encObat, _ := crypto.Encrypt("OB001", testEncKey)
	now := time.Now()

	mockSvc := &mockResepService{
		simpanResepFn: func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error) {
			return &resep.Resep{
				NoResep:          "202609030001",
				NoRawat:          req.NoRawat,
				TanggalPeresepan: req.TanggalPeresepan,
				JamPeresepan:     req.JamPeresepan,
				Status:           "ralan",
			}, nil
		},
	}

	handler := resep.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, authMwForResep, noOpMw)

	body, _ := json.Marshal(resep.SimpanResepRequest{
		NoRawat:          "2026/09/03/000001",
		TanggalPeresepan: now.Format("2006-01-02"),
		JamPeresepan:     now.Format("15:04:05"),
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					IdObat: encObat,
					Jumlah: 10,
				},
				AturanPakai: "3x1",
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/resep/Ralan/"+encKunjungan, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	reqInvalid := httptest.NewRequest(http.MethodGet, "/api/v1/resep/invalid_status/"+encKunjungan, nil)
	rrInvalid := httptest.NewRecorder()
	mux.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for invalid status lanjut, got %d", rrInvalid.Code)
	}
}

func TestResepHandler_DetailResep(t *testing.T) {
	encResep, _ := crypto.Encrypt("202609030001", testEncKey)

	mockSvc := &mockResepService{
		detailResepFn: func(ctx context.Context, noResep string) (*resep.Resep, error) {
			if noResep == "202609030001" {
				return &resep.Resep{
					NoResep:          "202609030001",
					NoRawat:          "2026/09/03/000001",
					TanggalPeresepan: "2026-09-03",
					JamPeresepan:     "10:00:00",
					Status:           "ralan",
				}, nil
			}
			return nil, nil
		},
	}

	handler := resep.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, authMwForResep, noOpMw)

	t.Run("Sukses detail resep via route ringkas /api/v1/resep/{id_resep}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resep/"+encResep, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Success bool        `json:"success"`
			Data    resep.Resep `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		decId, err := crypto.Decrypt(resp.Data.Id, testEncKey)
		if err != nil || decId != "202609030001" {
			t.Errorf("Expected decrypted Id '202609030001', got '%s' (err: %v)", decId, err)
		}

		decKunjungan, err := crypto.Decrypt(resp.Data.IdKunjungan, testEncKey)
		if err != nil || decKunjungan != "2026/09/03/000001" {
			t.Errorf("Expected decrypted IdKunjungan '2026/09/03/000001', got '%s' (err: %v)", decKunjungan, err)
		}
	})

	t.Run("Error ID resep tidak valid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resep/invalid-encrypted-id", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Expected status 400 for invalid id_resep, got %d", rr.Code)
		}
	})
}

func TestResepHandler_DuplicateDecryptedObat(t *testing.T) {
	noRawat := "2026/09/03/000001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncKey)
	encObatA, _ := crypto.Encrypt("OB001", testEncKey)
	encObatB, _ := crypto.Encrypt("OB001", testEncKey)
	if encObatA == encObatB {
		t.Fatal("expected distinct AES-GCM ciphertexts for same plaintext")
	}
	past := time.Now().Add(-5 * time.Minute)

	mockSvc := &mockResepService{
		simpanResepFn: func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req resep.SimpanResepRequest) (*resep.Resep, error) {
			return &resep.Resep{NoResep: "202609030001", NoRawat: req.NoRawat, Status: "ralan"}, nil
		},
	}

	handler := resep.NewHandler(mockSvc, testEncKey)
	mux := http.NewServeMux()
	noOpMw := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, authMwForResep, noOpMw)

	t.Run("Dua ciphertext berbeda untuk obat sama pada resep_dokter ditolak 400", func(t *testing.T) {
		body, _ := json.Marshal(resep.SimpanResepRequest{
			NoRawat:          noRawat,
			TanggalPeresepan: past.Format("2006-01-02"),
			JamPeresepan:     past.Format("15:04:05"),
			ResepDokter: []resep.ResepDokterInput{
				{ItemObatInput: resep.ItemObatInput{IdObat: encObatA, Jumlah: 5}, AturanPakai: "3x1"},
				{ItemObatInput: resep.ItemObatInput{IdObat: encObatB, Jumlah: 5}, AturanPakai: "2x1"},
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resep/Ralan/"+encKunjungan, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Dua ciphertext berbeda untuk obat sama dalam satu racikan ditolak 400", func(t *testing.T) {
		body, _ := json.Marshal(resep.SimpanResepRequest{
			NoRawat:          noRawat,
			TanggalPeresepan: past.Format("2006-01-02"),
			JamPeresepan:     past.Format("15:04:05"),
			ResepRacikan: []resep.ResepRacikanInput{
				{
					NamaRacik:     "Puyer",
					KodeRacik:     "R01",
					JumlahRacikan: 10,
					AturanPakai:   "3x1",
					Detail: []resep.ResepRacikanDetailInput{
						{ItemObatInput: resep.ItemObatInput{IdObat: encObatA, Jumlah: 5}},
						{ItemObatInput: resep.ItemObatInput{IdObat: encObatB, Jumlah: 2}},
					},
				},
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resep/Ralan/"+encKunjungan, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Obat sama antara resep_dokter dan resep_racikan tetap diizinkan 201", func(t *testing.T) {
		body, _ := json.Marshal(resep.SimpanResepRequest{
			NoRawat:          noRawat,
			TanggalPeresepan: past.Format("2006-01-02"),
			JamPeresepan:     past.Format("15:04:05"),
			ResepDokter: []resep.ResepDokterInput{
				{ItemObatInput: resep.ItemObatInput{IdObat: encObatA, Jumlah: 5}, AturanPakai: "3x1"},
			},
			ResepRacikan: []resep.ResepRacikanInput{
				{
					NamaRacik:     "Puyer",
					KodeRacik:     "R01",
					JumlahRacikan: 10,
					AturanPakai:   "3x1",
					Detail: []resep.ResepRacikanDetailInput{
						{ItemObatInput: resep.ItemObatInput{IdObat: encObatB, Jumlah: 5}},
					},
				},
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resep/Ralan/"+encKunjungan, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}
