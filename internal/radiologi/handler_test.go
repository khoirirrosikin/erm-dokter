package radiologi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/token"
	"erm-dokter/internal/radiologi"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

const testEncryptionKey = "bafaa956-751d-4f59-98cc-574ee9dfe9f6"

type mockRadiologiService struct {
	daftarHasilRadiologiFn     func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, shared.PaginationMeta, error)
	daftarHasilRadiologiByRMFn func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, shared.PaginationMeta, error)
	detailHasilRadiologiFn     func(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error)

	simpanPermintaanRadiologiFn     func(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest) (*radiologi.DetailPermintaanRadiologi, error)
	daftarPermintaanRadiologiFn     func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]radiologi.DetailPermintaanRadiologi, error)
	daftarPermintaanRadiologiByRMFn func(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatPermintaanRadiologi) ([]radiologi.DetailPermintaanRadiologi, shared.PaginationMeta, error)
	detailPermintaanRadiologiFn     func(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error)
	updatePermintaanRadiologiFn     func(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest) (*radiologi.DetailPermintaanRadiologi, error)
	hapusPermintaanRadiologiFn      func(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut) error
}

func (m *mockRadiologiService) DaftarHasilRadiologi(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, shared.PaginationMeta, error) {
	if m.daftarHasilRadiologiFn != nil {
		return m.daftarHasilRadiologiFn(ctx, noRawat, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockRadiologiService) DaftarHasilRadiologiByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, shared.PaginationMeta, error) {
	if m.daftarHasilRadiologiByRMFn != nil {
		return m.daftarHasilRadiologiByRMFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockRadiologiService) DetailHasilRadiologi(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error) {
	if m.detailHasilRadiologiFn != nil {
		return m.detailHasilRadiologiFn(ctx, idHasil)
	}
	return nil, nil
}

func (m *mockRadiologiService) SimpanPermintaanRadiologi(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest) (*radiologi.DetailPermintaanRadiologi, error) {
	if m.simpanPermintaanRadiologiFn != nil {
		return m.simpanPermintaanRadiologiFn(ctx, kodeDokter, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockRadiologiService) DaftarPermintaanRadiologi(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]radiologi.DetailPermintaanRadiologi, error) {
	if m.daftarPermintaanRadiologiFn != nil {
		return m.daftarPermintaanRadiologiFn(ctx, noRawat, statusLanjut)
	}
	return nil, nil
}

func (m *mockRadiologiService) DaftarPermintaanRadiologiByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatPermintaanRadiologi) ([]radiologi.DetailPermintaanRadiologi, shared.PaginationMeta, error) {
	if m.daftarPermintaanRadiologiByRMFn != nil {
		return m.daftarPermintaanRadiologiByRMFn(ctx, noRM, statusLanjut, filter)
	}
	return nil, shared.PaginationMeta{}, nil
}

func (m *mockRadiologiService) DetailPermintaanRadiologi(ctx context.Context, noPermintaan string) (*radiologi.DetailPermintaanRadiologi, error) {
	if m.detailPermintaanRadiologiFn != nil {
		return m.detailPermintaanRadiologiFn(ctx, noPermintaan)
	}
	return nil, nil
}

func (m *mockRadiologiService) UpdatePermintaanRadiologi(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest) (*radiologi.DetailPermintaanRadiologi, error) {
	if m.updatePermintaanRadiologiFn != nil {
		return m.updatePermintaanRadiologiFn(ctx, kodeDokter, noRawat, noPermintaan, statusLanjut, req)
	}
	return nil, nil
}

func (m *mockRadiologiService) HapusPermintaanRadiologi(ctx context.Context, kodeDokter, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut) error {
	if m.hapusPermintaanRadiologiFn != nil {
		return m.hapusPermintaanRadiologiFn(ctx, kodeDokter, noRawat, noPermintaan, statusLanjut)
	}
	return nil
}

func setupRadiologiRouter(svc radiologi.Service) *http.ServeMux {
	mux := http.NewServeMux()
	handler := radiologi.NewHandler(svc, testEncryptionKey)
	passthroughMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
	handler.RegisterRoutes(mux, passthroughMiddleware, passthroughMiddleware)
	return mux
}

func TestHandler_DaftarHasilRadiologi_Success(t *testing.T) {
	noRawat := "2026/04/22/000001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncryptionKey)

	svc := &mockRadiologiService{
		daftarHasilRadiologiFn: func(ctx context.Context, nr string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, shared.PaginationMeta, error) {
			if nr != noRawat {
				t.Errorf("Expected noRawat %s, got %s", noRawat, nr)
			}
			return []radiologi.HasilRadiologi{
				{
					NoRawat:        nr,
					KodeTindakan:   "RAD001",
					NamaTindakan:   "Rontgen Thorax",
					Status:         "Ralan",
					TanggalPeriksa: "2026-04-22",
					JamPeriksa:     "10:00:00",
					Hasil:          "Cor dan Pulmo normal",
					GambarPACS:     []string{"http://pacs.example.com/viewer?token=abc"},
				},
			}, shared.NewPaginationMeta(1, 1, 10), nil
		},
	}

	router := setupRadiologiRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/Ralan/"+encKunjungan+"?page=1&limit=10", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Message string                     `json:"message"`
		Data    []radiologi.HasilRadiologi `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(resp.Data))
	}
	if resp.Data[0].Id == "" || resp.Data[0].IdKunjungan == "" {
		t.Errorf("Expected encrypted Id and IdKunjungan, got id=%s, idKunjungan=%s", resp.Data[0].Id, resp.Data[0].IdKunjungan)
	}
}

func TestHandler_DaftarHasilRadiologi_InvalidToken(t *testing.T) {
	svc := &mockRadiologiService{}
	router := setupRadiologiRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/Ralan/invalid-token", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarHasilRadiologi_InvalidStatusLanjut(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	svc := &mockRadiologiService{}
	router := setupRadiologiRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/BukanStatus/"+encKunjungan, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarHasilRadiologi_InvalidFilter(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	svc := &mockRadiologiService{}
	router := setupRadiologiRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/Ralan/"+encKunjungan+"?tanggal=bukan-tanggal", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarHasilRadiologi_ServiceError(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	svc := &mockRadiologiService{
		daftarHasilRadiologiFn: func(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, shared.PaginationMeta, error) {
			return nil, shared.PaginationMeta{}, errors.New("db error")
		},
	}
	router := setupRadiologiRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/Ralan/"+encKunjungan, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status 500, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarHasilRadiologiByRM_Success(t *testing.T) {
	noRM := "123456"
	encPasien, _ := crypto.Encrypt(noRM, testEncryptionKey)

	svc := &mockRadiologiService{
		daftarHasilRadiologiByRMFn: func(ctx context.Context, rm string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatRadiologi) ([]radiologi.HasilRadiologi, shared.PaginationMeta, error) {
			if rm != noRM {
				t.Errorf("Expected noRM %s, got %s", noRM, rm)
			}
			return []radiologi.HasilRadiologi{
				{
					NoRawat:        "2026/04/22/000001",
					KodeTindakan:   "RAD001",
					NamaTindakan:   "Rontgen Thorax",
					Status:         "Ralan",
					TanggalPeriksa: "2026-04-22",
					JamPeriksa:     "10:00:00",
					Hasil:          "Normal",
				},
			}, shared.NewPaginationMeta(1, 1, 10), nil
		},
	}

	router := setupRadiologiRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/Semua/pasien/"+encPasien, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarHasilRadiologiByRM_InvalidToken(t *testing.T) {
	svc := &mockRadiologiService{}
	router := setupRadiologiRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/Semua/pasien/invalid-token", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailHasilRadiologi_Success(t *testing.T) {
	noRawat := "2026/04/22/000001"
	kodeTindakan := "RAD001"
	tanggalPeriksa := "2026-04-22"
	jamPeriksa := "10:00:00"

	composite := noRawat + "~" + kodeTindakan + "~" + tanggalPeriksa + "~" + jamPeriksa
	encRadiologi, _ := crypto.Encrypt(composite, testEncryptionKey)

	svc := &mockRadiologiService{
		detailHasilRadiologiFn: func(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error) {
			return &radiologi.HasilRadiologi{
				NoRawat:        idHasil.NoRawat,
				KodeTindakan:   idHasil.KodeTindakan,
				NamaTindakan:   "Rontgen Thorax",
				Status:         "Ralan",
				TanggalPeriksa: idHasil.TanggalPeriksa,
				JamPeriksa:     idHasil.JamPeriksa,
				Hasil:          "Cor dan Pulmo normal",
				GambarPACS:     []string{"http://pacs.example.com/view"},
			}, nil
		},
	}

	router := setupRadiologiRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/"+encRadiologi, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool                     `json:"success"`
		Data    radiologi.HasilRadiologi `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Data.Hasil != "Cor dan Pulmo normal" {
		t.Errorf("Expected hasil Cor dan Pulmo normal, got %s", resp.Data.Hasil)
	}
}

func TestHandler_DetailHasilRadiologi_MalformedCompositeID(t *testing.T) {
	encRadiologi, _ := crypto.Encrypt("invalid~composite", testEncryptionKey)

	svc := &mockRadiologiService{}
	router := setupRadiologiRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/"+encRadiologi, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailHasilRadiologi_NotFound(t *testing.T) {
	noRawat := "2026/04/22/000001"
	composite := noRawat + "~RAD001~2026-04-22~10:00:00"
	encRadiologi, _ := crypto.Encrypt(composite, testEncryptionKey)

	svc := &mockRadiologiService{
		detailHasilRadiologiFn: func(ctx context.Context, idHasil radiologi.IdHasilRadiologi) (*radiologi.HasilRadiologi, error) {
			return nil, apperror.NewNotFoundError("Data hasil pemeriksaan radiologi tidak ditemukan")
		},
	}

	router := setupRadiologiRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/hasil/"+encRadiologi, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func authMiddlewareForTest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserClaimKey, &token.Claims{
			KodeDokter: "DR01",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func setupRadiologiRouterWithAuth(svc radiologi.Service) *http.ServeMux {
	mux := http.NewServeMux()
	handler := radiologi.NewHandler(svc, testEncryptionKey)
	dummyTimeout := func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
	handler.RegisterRoutes(mux, authMiddlewareForTest, dummyTimeout)
	return mux
}

func TestHandler_SimpanPermintaanRadiologi_Success(t *testing.T) {
	noRawat := "2026/09/05/000001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("RAD001", testEncryptionKey)

	svc := &mockRadiologiService{
		simpanPermintaanRadiologiFn: func(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req radiologi.SimpanPermintaanRadiologiRequest) (*radiologi.DetailPermintaanRadiologi, error) {
			if kodeDokterLogin != "DR01" {
				t.Errorf("Expected kodeDokter DR01, got %s", kodeDokterLogin)
			}
			if req.NoRawat != noRawat {
				t.Errorf("Expected noRawat %s, got %s", noRawat, req.NoRawat)
			}
			if len(req.Pemeriksaan) != 1 || req.Pemeriksaan[0].KodeTindakan != "RAD001" {
				t.Errorf("Expected decrypted KodeTindakan RAD001, got %+v", req.Pemeriksaan)
			}
			return &radiologi.DetailPermintaanRadiologi{
				PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
					NoPermintaan:      "RAD202609050001",
					NoRawat:           noRawat,
					TanggalPermintaan: req.TanggalPermintaan,
					JamPermintaan:     req.JamPermintaan,
					DokterPerujuk:     radiologi.DokterInfo{KodeDokter: "DR01", NamaDokter: "dr. SpRad"},
					Status:            "Ralan",
					InformasiTambahan: req.InformasiTambahan,
					DiagnosaKlinis:    req.DiagnosaKlinis,
				},
				Pemeriksaan: []radiologi.PemeriksaanRadiologiItem{
					{KodeTindakan: "RAD001", NamaTindakan: "Thorax AP", StatusBayar: "Belum"},
				},
			}, nil
		},
	}

	router := setupRadiologiRouterWithAuth(svc)

	body := `{
		"no_rawat": "` + noRawat + `",
		"tanggal_permintaan": "2026-09-05",
		"jam_permintaan": "10:00:00",
		"informasi_tambahan": "Thorax AP",
		"diagnosa_klinis": "Batuk kronis",
		"pemeriksaan": [
			{"id_tindakan": "` + encTindakan + `"}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/radiologi/permintaan/Ralan/"+encKunjungan, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool                                `json:"success"`
		Message string                              `json:"message"`
		Data    radiologi.DetailPermintaanRadiologi `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Data.NoPermintaan == "RAD202609050001" {
		t.Errorf("Expected encrypted NoPermintaan, got plain: %s", resp.Data.NoPermintaan)
	}
}

func TestHandler_SimpanPermintaanRadiologi_MismatchedNoRawat(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/05/000001", testEncryptionKey)
	encTindakan, _ := crypto.Encrypt("RAD001", testEncryptionKey)

	svc := &mockRadiologiService{}
	router := setupRadiologiRouterWithAuth(svc)

	body := `{
		"no_rawat": "2026/09/05/999999",
		"tanggal_permintaan": "2026-09-05",
		"jam_permintaan": "10:00:00",
		"pemeriksaan": [{"id_tindakan": "` + encTindakan + `"}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/radiologi/permintaan/Ralan/"+encKunjungan, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_SimpanPermintaanRadiologi_RejectSemua(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/05/000001", testEncryptionKey)
	svc := &mockRadiologiService{}
	router := setupRadiologiRouterWithAuth(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/radiologi/permintaan/Semua/"+encKunjungan, strings.NewReader("{}"))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for status lanjut 'Semua' on create, got %d", rr.Code)
	}
}

func TestHandler_DaftarPermintaanRadiologi_Success(t *testing.T) {
	noRawat := "2026/09/05/000001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncryptionKey)

	svc := &mockRadiologiService{
		daftarPermintaanRadiologiFn: func(ctx context.Context, nr string, statusLanjut shared.StatusLanjut) ([]radiologi.DetailPermintaanRadiologi, error) {
			return []radiologi.DetailPermintaanRadiologi{
				{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
						NoPermintaan:      "RAD202609050001",
						NoRawat:           nr,
						TanggalPermintaan: "2026-09-05",
						Status:            "Ralan",
					},
					Pemeriksaan: []radiologi.PemeriksaanRadiologiItem{
						{KodeTindakan: "RAD001", NamaTindakan: "Thorax AP"},
					},
				},
			}, nil
		},
	}

	router := setupRadiologiRouterWithAuth(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/permintaan/Semua/"+encKunjungan, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DaftarPermintaanRadiologiByRM_Success(t *testing.T) {
	noRM := "123456"
	encPasien, _ := crypto.Encrypt(noRM, testEncryptionKey)

	svc := &mockRadiologiService{
		daftarPermintaanRadiologiByRMFn: func(ctx context.Context, rm string, statusLanjut shared.StatusLanjut, filter radiologi.FilterRiwayatPermintaanRadiologi) ([]radiologi.DetailPermintaanRadiologi, shared.PaginationMeta, error) {
			return []radiologi.DetailPermintaanRadiologi{
				{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
						NoPermintaan:      "RAD202609050001",
						TanggalPermintaan: "2026-09-05",
						Status:            "Ralan",
					},
				},
			}, shared.NewPaginationMeta(1, 1, 5), nil
		},
	}

	router := setupRadiologiRouterWithAuth(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/permintaan/Semua/pasien/"+encPasien+"?page=1&limit=5", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailPermintaanRadiologi_Success(t *testing.T) {
	noRawat := "2026/09/05/000001"
	noOrder := "RAD202609050001"
	encPermintaan, _ := crypto.Encrypt(noOrder, testEncryptionKey)

	svc := &mockRadiologiService{
		detailPermintaanRadiologiFn: func(ctx context.Context, np string) (*radiologi.DetailPermintaanRadiologi, error) {
			return &radiologi.DetailPermintaanRadiologi{
				PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
					NoPermintaan:      np,
					NoRawat:           noRawat,
					TanggalPermintaan: "2026-09-05",
					Status:            "Ralan",
				},
			}, nil
		},
	}

	router := setupRadiologiRouterWithAuth(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/permintaan/"+encPermintaan, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_UpdatePermintaanRadiologi_RejectSemua(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/05/000001", testEncryptionKey)
	encPermintaan, _ := crypto.Encrypt("RAD202609050001", testEncryptionKey)

	svc := &mockRadiologiService{}
	router := setupRadiologiRouterWithAuth(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/radiologi/permintaan/Semua/"+encKunjungan+"/"+encPermintaan, strings.NewReader("{}"))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for status lanjut 'Semua' on update, got %d", rr.Code)
	}
}

func TestHandler_HapusPermintaanRadiologi_RejectSemua(t *testing.T) {
	encKunjungan, _ := crypto.Encrypt("2026/09/05/000001", testEncryptionKey)
	encPermintaan, _ := crypto.Encrypt("RAD202609050001", testEncryptionKey)

	svc := &mockRadiologiService{}
	router := setupRadiologiRouterWithAuth(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/radiologi/permintaan/Semua/"+encKunjungan+"/"+encPermintaan, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for status lanjut 'Semua' on delete, got %d", rr.Code)
	}
}

func TestHandler_HapusPermintaanRadiologi_Success(t *testing.T) {
	noRawat := "2026/09/05/000001"
	noOrder := "RAD202609050001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncryptionKey)
	encPermintaan, _ := crypto.Encrypt(noOrder, testEncryptionKey)

	svc := &mockRadiologiService{
		hapusPermintaanRadiologiFn: func(ctx context.Context, kodeDokterLogin, nr, np string, statusLanjut shared.StatusLanjut) error {
			if kodeDokterLogin != "DR01" {
				t.Errorf("Expected DR01, got %s", kodeDokterLogin)
			}
			return nil
		},
	}

	router := setupRadiologiRouterWithAuth(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/radiologi/permintaan/Ralan/"+encKunjungan+"/"+encPermintaan, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DetailPermintaanRadiologi(t *testing.T) {
	noRawat := "2026/09/05/000001"
	noOrder := "RAD202609050001"
	encPermintaan, _ := crypto.Encrypt(noOrder, testEncryptionKey)

	svc := &mockRadiologiService{
		detailPermintaanRadiologiFn: func(ctx context.Context, np string) (*radiologi.DetailPermintaanRadiologi, error) {
			if np == noOrder {
				return &radiologi.DetailPermintaanRadiologi{
					PermintaanRadiologiHeader: radiologi.PermintaanRadiologiHeader{
						NoPermintaan:      noOrder,
						NoRawat:           noRawat,
						TanggalPermintaan: "2026-09-05",
						JamPermintaan:     "10:00:00",
						Status:            "Ralan",
					},
				}, nil
			}
			return nil, nil
		},
	}

	router := setupRadiologiRouterWithAuth(svc)

	t.Run("Sukses detail permintaan via canonical route ringkas", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/permintaan/"+encPermintaan, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Success bool                                `json:"success"`
			Data    radiologi.DetailPermintaanRadiologi `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		decOrder, err := crypto.Decrypt(resp.Data.NoPermintaan, testEncryptionKey)
		if err != nil || decOrder != noOrder {
			t.Errorf("Expected decrypted NoPermintaan %s, got %s (err: %v)", noOrder, decOrder, err)
		}
	})

	t.Run("Error ID permintaan tidak valid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/permintaan/invalid-encrypted-id", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400 Bad Request, got %d", rr.Code)
		}
	})
}

func TestHandler_SimpanPermintaanRadiologi_DuplicateDecryptedTindakan(t *testing.T) {
	noRawat := "2026/09/05/000001"
	encKunjungan, _ := crypto.Encrypt(noRawat, testEncryptionKey)
	encTindakanA, _ := crypto.Encrypt("RAD001", testEncryptionKey)
	encTindakanB, _ := crypto.Encrypt("RAD001", testEncryptionKey)
	if encTindakanA == encTindakanB {
		t.Fatal("expected distinct AES-GCM ciphertexts")
	}

	router := setupRadiologiRouterWithAuth(&mockRadiologiService{})
	body, _ := json.Marshal(radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat:           noRawat,
		TanggalPermintaan: "2026-09-05",
		JamPermintaan:     "10:00:00",
		DiagnosaKlinis:    "Susp Pneumonia",
		Pemeriksaan: []radiologi.ItemPemeriksaanRadiologiRequest{
			{IdTindakan: encTindakanA},
			{IdTindakan: encTindakanB},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/radiologi/permintaan/Ralan/"+encKunjungan, strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for duplicate decrypted tindakan, got %d: %s", rr.Code, rr.Body.String())
	}
}
