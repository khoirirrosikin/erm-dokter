package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"erm-dokter/internal/auth"
	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/diagnosa"
	"erm-dokter/internal/docs"
	"erm-dokter/internal/health"
	"erm-dokter/internal/laboratorium"
	"erm-dokter/internal/master"
	"erm-dokter/internal/obat"
	"erm-dokter/internal/pasien"
	"erm-dokter/internal/pemeriksaan"
	"erm-dokter/internal/penilaianmedis"
	"erm-dokter/internal/radiologi"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/resep"
	"erm-dokter/internal/resumepasien"
	"erm-dokter/internal/routes"
	"erm-dokter/internal/rujukaninternal"
	"erm-dokter/internal/tindakan"
)

func TestRouteConfig_SetupAndBuildHandler(t *testing.T) {
	routeCfg := routes.NewRouteConfig(
		health.NewHandler(),
		docs.NewHandler(),
		auth.NewHandler(nil),
		master.NewHandler(nil),
		obat.NewHandler(nil, "key"),
		rawatjalan.NewHandler(nil, "key"),
		rawatinap.NewHandler(nil, "key"),
		pemeriksaan.NewHandler(nil, "key"),
		resep.NewHandler(nil, "key"),
		rujukaninternal.NewHandler(nil, "key"),
		penilaianmedis.NewHandler(nil, "key"),
		tindakan.NewHandler(nil, "key"),
		laboratorium.NewHandler(nil, "key"),
		radiologi.NewHandler(nil, "key"),
		berkasdigital.NewHandler(nil, "key"),
		resumepasien.NewHandler(nil, "key"),
		pasien.NewHandler(nil, "key"),
		diagnosa.NewHandler(nil, "key"),
		"test-jwt-secret",
		"test-service-api-key",
		"development",
		[]string{"198.51.100.10", "198.51.100.20"},
	)

	routeCfg.Setup()

	handler := routeCfg.BuildHandler("*")
	if handler == nil {
		t.Fatal("Expected non-nil http.Handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 on /health, got %d", rr.Code)
	}

	rateLimitHit := false
	routeCfg.LoginRateLimitMiddleware = func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rateLimitHit = true
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}
	routeCfg.Mux = http.NewServeMux()
	routeCfg.Setup()
	handlerWithRateLimit := routeCfg.BuildHandler("*")

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	loginRR := httptest.NewRecorder()
	handlerWithRateLimit.ServeHTTP(loginRR, loginReq)

	if !rateLimitHit {
		t.Errorf("Expected LoginRateLimitMiddleware to be executed on /api/v1/auth/login")
	}
	if loginRR.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 on /api/v1/auth/login, got %d", loginRR.Code)
	}
}

func TestRouteConfig_Production_IPWhitelistAndDocsBlocking(t *testing.T) {
	routeCfg := routes.NewRouteConfig(
		health.NewHandler(),
		docs.NewHandler(),
		auth.NewHandler(nil),
		master.NewHandler(nil),
		obat.NewHandler(nil, "key"),
		rawatjalan.NewHandler(nil, "key"),
		rawatinap.NewHandler(nil, "key"),
		pemeriksaan.NewHandler(nil, "key"),
		resep.NewHandler(nil, "key"),
		rujukaninternal.NewHandler(nil, "key"),
		penilaianmedis.NewHandler(nil, "key"),
		tindakan.NewHandler(nil, "key"),
		laboratorium.NewHandler(nil, "key"),
		radiologi.NewHandler(nil, "key"),
		berkasdigital.NewHandler(nil, "key"),
		resumepasien.NewHandler(nil, "key"),
		pasien.NewHandler(nil, "key"),
		diagnosa.NewHandler(nil, "key"),
		"test-jwt-secret",
		"test-service-api-key",
		"production",
		[]string{"198.51.100.10", "198.51.100.20"},
	)

	routeCfg.Setup()
	handler := routeCfg.BuildHandler("*")

	// 1. Health check bypasses IP check
	reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
	reqHealth.RemoteAddr = "10.0.0.99:12345"
	recHealth := httptest.NewRecorder()
	handler.ServeHTTP(recHealth, reqHealth)
	if recHealth.Code != http.StatusOK {
		t.Errorf("Expected /health status 200 in production, got %d", recHealth.Code)
	}

	// 2. Allowed IP accessing API (e.g. 198.51.100.10) should pass whitelist (returns 401 due to missing token, NOT 403)
	reqAllowed := httptest.NewRequest(http.MethodGet, "/api/v1/master/poli", nil)
	reqAllowed.RemoteAddr = "198.51.100.10:54321"
	recAllowed := httptest.NewRecorder()
	handler.ServeHTTP(recAllowed, reqAllowed)
	if recAllowed.Code == http.StatusForbidden {
		t.Errorf("Allowed IP 198.51.100.10 should not be forbidden, got 403")
	}

	// 3. Disallowed IP accessing API (e.g. 203.0.113.99) should receive 403 Forbidden
	reqBlocked := httptest.NewRequest(http.MethodGet, "/api/v1/master/poli", nil)
	reqBlocked.RemoteAddr = "203.0.113.99:54321"
	recBlocked := httptest.NewRecorder()
	handler.ServeHTTP(recBlocked, reqBlocked)
	if recBlocked.Code != http.StatusForbidden {
		t.Errorf("Disallowed IP should receive 403 Forbidden, got %d", recBlocked.Code)
	}

	// 4. Docs in production must return 404 Not Found even from allowed IP
	reqDocsAllowed := httptest.NewRequest(http.MethodGet, "/docs", nil)
	reqDocsAllowed.RemoteAddr = "198.51.100.10:54321"
	recDocsAllowed := httptest.NewRecorder()
	handler.ServeHTTP(recDocsAllowed, reqDocsAllowed)
	if recDocsAllowed.Code != http.StatusNotFound {
		t.Errorf("Expected /docs in production to return 404 Not Found, got %d", recDocsAllowed.Code)
	}

	reqDocsSpec := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	reqDocsSpec.RemoteAddr = "198.51.100.20:54321"
	recDocsSpec := httptest.NewRecorder()
	handler.ServeHTTP(recDocsSpec, reqDocsSpec)
	if recDocsSpec.Code != http.StatusNotFound {
		t.Errorf("Expected /docs/openapi.yaml in production to return 404 Not Found, got %d", recDocsSpec.Code)
	}
}

func TestRouteConfig_Development_DocsEnabled(t *testing.T) {
	routeCfg := routes.NewRouteConfig(
		health.NewHandler(),
		docs.NewHandler(),
		auth.NewHandler(nil),
		master.NewHandler(nil),
		obat.NewHandler(nil, "key"),
		rawatjalan.NewHandler(nil, "key"),
		rawatinap.NewHandler(nil, "key"),
		pemeriksaan.NewHandler(nil, "key"),
		resep.NewHandler(nil, "key"),
		rujukaninternal.NewHandler(nil, "key"),
		penilaianmedis.NewHandler(nil, "key"),
		tindakan.NewHandler(nil, "key"),
		laboratorium.NewHandler(nil, "key"),
		radiologi.NewHandler(nil, "key"),
		berkasdigital.NewHandler(nil, "key"),
		resumepasien.NewHandler(nil, "key"),
		pasien.NewHandler(nil, "key"),
		diagnosa.NewHandler(nil, "key"),
		"test-jwt-secret",
		"test-service-api-key",
		"development",
		[]string{"198.51.100.10", "198.51.100.20"},
	)

	routeCfg.Setup()
	handler := routeCfg.BuildHandler("*")

	reqDocs := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recDocs := httptest.NewRecorder()
	handler.ServeHTTP(recDocs, reqDocs)
	if recDocs.Code != http.StatusOK {
		t.Errorf("Expected /docs in development to return 200 OK, got %d", recDocs.Code)
	}
}
