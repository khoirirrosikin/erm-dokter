package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/pkg/token"
)

func TestLoggingMiddleware_AuthenticatedAndAnonymous(t *testing.T) {
	var buf bytes.Buffer
	testLogger := logger.NewWithOptions(&buf, "json", slog.LevelDebug, "erm-dokter")

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	loggedHandler := middleware.LoggingMiddlewareWithLogger(testLogger)(dummyHandler)

	req1 := httptest.NewRequest("GET", "/health", nil)
	rr1 := httptest.NewRecorder()
	loggedHandler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr1.Code)
	}

	secret := "test-secret-key-1234567890123456"
	tkn, err := token.GenerateToken("DR001", "dr. Budi", secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req2 := httptest.NewRequest("GET", "/api/v1/master/depo", nil)
	req2.Header.Set("Authorization", "Bearer "+tkn)
	rr2 := httptest.NewRecorder()
	loggedHandler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}

	req3 := httptest.NewRequest("OPTIONS", "/api/v1/pemeriksaan/Semua/pasien/test", nil)
	rr3 := httptest.NewRecorder()
	loggedHandler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr3.Code)
	}

	defaultHandler := middleware.LoggingMiddleware(dummyHandler)
	req4 := httptest.NewRequest("GET", "/health", nil)
	rr4 := httptest.NewRecorder()
	defaultHandler.ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr4.Code)
	}
}

func TestLoggingMiddleware_StatusCodesAndOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	testLogger := logger.NewWithOptions(&buf, "json", slog.LevelDebug, "erm-dokter")

	tests := []struct {
		name          string
		returnCode    int
		expectedLevel string
	}{
		{"200 OK -> INFO", http.StatusOK, "INFO"},
		{"404 Not Found -> WARN", http.StatusNotFound, "WARN"},
		{"500 Internal Error -> ERROR", http.StatusInternalServerError, "ERROR"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.returnCode)
			})
			wrapped := middleware.LoggingMiddlewareWithOptions(middleware.LoggingOptions{
				Logger:         testLogger,
				TrustedProxies: []string{"192.0.2.1", "198.51.100.1"},
			})(handler)

			req := httptest.NewRequest("POST", "/api/v1/test?filter=aktif", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.195, 198.51.100.1")
			req.Header.Set("User-Agent", "GoTestClient/1.0")

			rr := httptest.NewRecorder()
			wrapped.ServeHTTP(rr, req)

			var parsed map[string]any
			if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
				t.Fatalf("Failed to parse JSON: %v, raw: %s", err, buf.String())
			}

			if parsed["level"] != tc.expectedLevel {
				t.Errorf("expected level %s, got %v", tc.expectedLevel, parsed["level"])
			}
			if parsed["method"] != "POST" {
				t.Errorf("expected method POST, got %v", parsed["method"])
			}
			if parsed["path"] != "/api/v1/test" {
				t.Errorf("expected path /api/v1/test, got %v", parsed["path"])
			}
			if parsed["query"] != "filter=aktif" {
				t.Errorf("expected query filter=aktif, got %v", parsed["query"])
			}
			if parsed["client_ip"] != "203.0.113.195" {
				t.Errorf("expected client_ip 203.0.113.195, got %v", parsed["client_ip"])
			}
			if parsed["status"] != float64(tc.returnCode) {
				t.Errorf("expected status %d, got %v", tc.returnCode, parsed["status"])
			}
			if parsed["user_agent"] != "GoTestClient/1.0" {
				t.Errorf("expected user_agent GoTestClient/1.0, got %v", parsed["user_agent"])
			}
		})
	}

	t.Run("Untrusted RemoteAddr ignores spoofed X-Forwarded-For in logs", func(t *testing.T) {
		buf.Reset()
		wrapped := middleware.LoggingMiddlewareWithOptions(middleware.LoggingOptions{
			Logger:         testLogger,
			TrustedProxies: []string{"198.51.100.10"},
		})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))

		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.RemoteAddr = "192.0.2.99:54321"
		req.Header.Set("X-Forwarded-For", "198.51.100.10")

		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		var parsed map[string]any
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		if parsed["client_ip"] != "192.0.2.99" {
			t.Errorf("expected real attacker IP 192.0.2.99, got %v", parsed["client_ip"])
		}
	})
}

func TestLoggingMiddleware_WithErrorDetailAndUser(t *testing.T) {
	var buf bytes.Buffer
	testLogger := logger.NewWithOptions(&buf, "json", slog.LevelDebug, "erm-dokter")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-User-ID", "DRBUDI")
		w.Header().Set("X-Error-Detail", "Autentikasi ditolak: Kredensial login tidak cocok untuk username 'DRBUDI'")
		w.WriteHeader(http.StatusUnauthorized)
	})

	wrapped := middleware.LoggingMiddlewareWithLogger(testLogger)(handler)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}

	if rr.Header().Get("X-Error-Detail") != "" {
		t.Error("expected X-Error-Detail to be deleted from response")
	}
	if rr.Header().Get("X-User-ID") != "" {
		t.Error("expected X-User-ID to be deleted from response")
	}
	if rr.Header().Get("X-Logging-Middleware") != "" {
		t.Error("expected X-Logging-Middleware to be deleted from response")
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Failed to parse log JSON: %v, raw: %s", err, buf.String())
	}

	if parsed["level"] != "WARN" {
		t.Errorf("expected level WARN, got %v", parsed["level"])
	}
	expectedMsg := "Autentikasi ditolak: Kredensial login tidak cocok untuk username 'DRBUDI'"
	if parsed["msg"] != expectedMsg {
		t.Errorf("expected msg '%s', got '%v'", expectedMsg, parsed["msg"])
	}
	if parsed["user"] != "DRBUDI" {
		t.Errorf("expected user 'DRBUDI', got '%v'", parsed["user"])
	}
	if parsed["status"] != float64(401) {
		t.Errorf("expected status 401, got %v", parsed["status"])
	}
}

func TestLoggingMiddleware_WithSuccessMessage(t *testing.T) {
	var buf bytes.Buffer
	testLogger := logger.NewWithOptions(&buf, "json", slog.LevelDebug, "erm-dokter")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-User-ID", "DRBUDI")
		response.Success(w, "Login berhasil", map[string]string{"token": "dummy-token"})
	})

	wrapped := middleware.LoggingMiddlewareWithLogger(testLogger)(handler)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if rr.Header().Get("X-Log-Message") != "" {
		t.Error("expected X-Log-Message to be deleted from client response")
	}
	if rr.Header().Get("X-User-ID") != "" {
		t.Error("expected X-User-ID to be deleted from client response")
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Failed to parse log JSON: %v, raw: %s", err, buf.String())
	}

	if parsed["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", parsed["level"])
	}
	if parsed["msg"] != "Login berhasil" {
		t.Errorf("expected msg 'Login berhasil', got '%v'", parsed["msg"])
	}
	if parsed["user"] != "DRBUDI" {
		t.Errorf("expected user 'DRBUDI', got '%v'", parsed["user"])
	}
	if parsed["status"] != float64(200) {
		t.Errorf("expected status 200, got %v", parsed["status"])
	}
}

func TestLoggingMiddlewareWithOptions_ContextualRefsAndHygiene(t *testing.T) {
	const encKey = "12345678901234567890123456789012"
	const jwtSecret = "real-jwt-secret-1234567890123456"

	var buf bytes.Buffer
	testLogger := logger.NewWithOptions(&buf, "json", slog.LevelDebug, "erm-dokter")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/resep/{status_lanjut}/{id_kunjungan}", func(w http.ResponseWriter, r *http.Request) {
		middleware.SetLogRef(w, middleware.HeaderLogNoResep, "202611130001")
		response.Created(w, "Berhasil menyimpan resep obat", nil)
	})
	mux.HandleFunc("GET /api/v1/rawat-inap/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true}`))
	})
	mux.HandleFunc("GET /api/v1/radiologi/{status_lanjut}/pasien/{id_pasien}", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "OK", nil)
	})
	mux.HandleFunc("DELETE /api/v1/laboratorium/permintaan/pk/{status_lanjut}/{id_kunjungan}/{id_permintaan}", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "Berhasil menghapus permintaan", nil)
	})

	handler := middleware.RequestIDMiddleware(
		middleware.LoggingMiddlewareWithOptions(middleware.LoggingOptions{
			Logger:        testLogger,
			EncryptionKey: encKey,
			JWTSecret:     jwtSecret,
		})(mux),
	)

	encKunjungan, err := crypto.Encrypt("2026/11/13/000042", encKey)
	if err != nil {
		t.Fatalf("encrypt no_rawat: %v", err)
	}

	t.Run("POST resep mencatat no_rawat & no_resep tanpa membocorkan scratch header atau duplikasi request_id", func(t *testing.T) {
		buf.Reset()
		validToken, _ := token.GenerateToken("DR001", "dr. Budi", jwtSecret, time.Hour)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resep/ralan/"+encKunjungan, nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rr.Code)
		}
		if rr.Header().Get(middleware.HeaderLogNoResep) != "" {
			t.Error("expected X-Log-No-Resep scratch header to be stripped")
		}

		rawLog := buf.String()
		if strings.Count(rawLog, `"request_id":`) != 1 {
			t.Errorf("expected single request_id in JSON log, got: %s", rawLog)
		}

		var parsed map[string]any
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Fatalf("unmarshal log: %v", err)
		}
		if parsed["no_rawat"] != "2026/11/13/000042" {
			t.Errorf("expected no_rawat 2026/11/13/000042, got %v", parsed["no_rawat"])
		}
		if parsed["no_resep"] != "202611130001" {
			t.Errorf("expected no_resep 202611130001, got %v", parsed["no_resep"])
		}
		if parsed["user"] != "DR001" {
			t.Errorf("expected verified user DR001, got %v", parsed["user"])
		}
	})

	t.Run("Direct w.Write tanpa WriteHeader tetap membersihkan X-Logging-Middleware dan mengekstrak no_rawat ranap", func(t *testing.T) {
		buf.Reset()
		encRanapID, _ := crypto.Encrypt("2026/11/13/000099~2026-11-10~08:00:00", encKey)
		forgedToken, _ := token.GenerateToken("DR999", "Forged", "wrong-secret-key-1234567890123456", time.Hour)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rawat-inap/"+encRanapID, nil)
		req.Header.Set("Authorization", "Bearer "+forgedToken)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("X-Logging-Middleware") != "" {
			t.Error("expected X-Logging-Middleware header not to leak on direct Write")
		}

		var parsed map[string]any
		if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
			t.Fatalf("unmarshal log: %v", err)
		}
		if parsed["no_rawat"] != "2026/11/13/000099" {
			t.Errorf("expected no_rawat 2026/11/13/000099 from composite ranap id, got %v", parsed["no_rawat"])
		}
		if parsed["user"] != "-" {
			t.Errorf("expected forged JWT user to be '-', got %v", parsed["user"])
		}
	})

	t.Run("Path id_pasien dan id_permintaan didekripsi menjadi no_rkm_medis dan no_permintaan", func(t *testing.T) {
		buf.Reset()
		encRM, _ := crypto.Encrypt("001234", encKey)
		req1 := httptest.NewRequest(http.MethodGet, "/api/v1/radiologi/Semua/pasien/"+encRM, nil)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)

		var parsed1 map[string]any
		if err := json.Unmarshal(buf.Bytes(), &parsed1); err != nil {
			t.Fatalf("unmarshal log 1: %v", err)
		}
		if parsed1["no_rkm_medis"] != "001234" {
			t.Errorf("expected no_rkm_medis 001234, got %v", parsed1["no_rkm_medis"])
		}

		buf.Reset()
		encOrder, _ := crypto.Encrypt("PK202611130005", encKey)
		req2 := httptest.NewRequest(http.MethodDelete, "/api/v1/laboratorium/permintaan/pk/ralan/"+encKunjungan+"/"+encOrder, nil)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		var parsed2 map[string]any
		if err := json.Unmarshal(buf.Bytes(), &parsed2); err != nil {
			t.Fatalf("unmarshal log 2: %v", err)
		}
		if parsed2["no_rawat"] != "2026/11/13/000042" || parsed2["no_permintaan"] != "PK202611130005" {
			t.Errorf("expected no_rawat & no_permintaan in log, got %+v", parsed2)
		}
	})
}
