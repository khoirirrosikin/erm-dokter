package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/pkg/token"
)

const testSecret = "my-secret-key-12345678901234567890"

func TestJWTMiddleware(t *testing.T) {
	jwtMw := middleware.JWTMiddleware(testSecret)

	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		kodeDokter, err := middleware.GetKodeDokter(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		w.Write([]byte("OK: " + kodeDokter))
	}

	wrapped := jwtMw(dummyHandler)

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Invalid Bearer Prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic abcdef")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Invalid Token String", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.payload")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Valid Token", func(t *testing.T) {
		validToken, err := token.GenerateToken("DR001", "dr. Ahmad", testSecret, 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
		if rr.Body.String() != "OK: DR001" {
			t.Errorf("Unexpected body: %s", rr.Body.String())
		}
	})

	t.Run("X-API-Key Rejected by JWTMiddleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-Key", "valid-service-api-key")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 when X-API-Key is sent to JWTMiddleware, got %d", rr.Code)
		}
	})
}

func TestAuthMiddleware_APIKey(t *testing.T) {
	const serviceKey = "valid-service-api-key"
	authMw := middleware.AuthMiddleware(testSecret, serviceKey)

	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		kodeDokter, err := middleware.GetKodeDokter(r.Context(), true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if middleware.IsService(r.Context()) {
			w.Write([]byte("SERVICE_OK"))
			return
		}
		w.Write([]byte("DOKTER_OK: " + kodeDokter))
	}

	wrapped := authMw(dummyHandler)

	t.Run("Valid X-API-Key Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-Key", serviceKey)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
		if rr.Body.String() != "SERVICE_OK" {
			t.Errorf("Unexpected body: %s", rr.Body.String())
		}
	})

	t.Run("Invalid X-API-Key Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Valid Bearer Token via AuthMiddleware", func(t *testing.T) {
		validToken, err := token.GenerateToken("DR002", "dr. Budi", testSecret, 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
		if rr.Body.String() != "DOKTER_OK: DR002" {
			t.Errorf("Unexpected body: %s", rr.Body.String())
		}
	})

	t.Run("Missing Both Headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})
}

func TestGetKodeDokter(t *testing.T) {
	t.Run("Service Context with allowService=true", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserClaimKey, &token.Claims{
			Role: middleware.RoleService,
		})
		kode, err := middleware.GetKodeDokter(ctx, true)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if kode != "" {
			t.Errorf("Expected empty kode, got %s", kode)
		}
		if !middleware.IsService(ctx) {
			t.Errorf("Expected IsService true")
		}
	})

	t.Run("Service Context with default allowService=false (Rejected)", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserClaimKey, &token.Claims{
			Role: middleware.RoleService,
		})
		_, err := middleware.GetKodeDokter(ctx)
		if err == nil {
			t.Errorf("Expected error when service calls doctor-only endpoint, got nil")
		}
	})

	t.Run("Dokter Context Valid", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserClaimKey, &token.Claims{
			Role:       middleware.RoleDokter,
			KodeDokter: "DR001",
		})
		kode, err := middleware.GetKodeDokter(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if kode != "DR001" {
			t.Errorf("Expected DR001, got %s", kode)
		}
		if middleware.IsService(ctx) {
			t.Errorf("Expected IsService false")
		}
	})

	t.Run("Dokter Context Empty Kode", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserClaimKey, &token.Claims{
			Role:       middleware.RoleDokter,
			KodeDokter: "   ",
		})
		_, err := middleware.GetKodeDokter(ctx)
		if err == nil {
			t.Errorf("Expected error for empty doctor code, got nil")
		}
	})

	t.Run("Nil Claims Context", func(t *testing.T) {
		_, err := middleware.GetKodeDokter(context.Background())
		if err == nil {
			t.Errorf("Expected error for nil claims, got nil")
		}
	})
}

func TestCORSMiddleware(t *testing.T) {
	cors := middleware.CORSMiddleware("http://localhost:3000")
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("cors passed"))
	}))

	t.Run("Preflight OPTIONS", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("Expected status 204 No Content for OPTIONS, got %d", rr.Code)
		}
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Errorf("Unexpected CORS origin: %s", rr.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("GET Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})
}

func TestRateLimiter(t *testing.T) {
	mw := middleware.RateLimitMiddleware(2, 500*time.Millisecond)

	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.50:12345"

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("Req 1: expected 200, got %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Errorf("Req 2: expected 200, got %d", rr2.Code)
	}

	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req)
	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("Req 3: expected 429 Too Many Requests, got %d", rr3.Code)
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	timeoutMw := middleware.TimeoutMiddleware(50 * time.Millisecond)

	t.Run("Within Timeout", func(t *testing.T) {
		handler := timeoutMw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fast response"))
		})

		req := httptest.NewRequest(http.MethodGet, "/fast", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})

	t.Run("Exceeded Timeout", func(t *testing.T) {
		handler := timeoutMw(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/slow", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusGatewayTimeout {
			t.Errorf("Expected status 504 Gateway Timeout, got %d", rr.Code)
		}
	})

	t.Run("Preserves Outer & Handler Headers Under LoggingMiddleware", func(t *testing.T) {
		inner := timeoutMw(func(w http.ResponseWriter, r *http.Request) {
			if w.Header().Get("X-Logging-Middleware") != "true" {
				t.Errorf("Expected X-Logging-Middleware=true inside handler")
			}
			w.Header().Set("X-Custom-Header", "erm-ok")
			response.Created(w, "Data tersimpan", map[string]string{"id": "1"})
		})
		wrapped := middleware.LoggingMiddleware(inner)

		req := httptest.NewRequest(http.MethodPost, "/items", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", rr.Code)
		}
		if rr.Header().Get("X-Custom-Header") != "erm-ok" {
			t.Fatalf("Expected X-Custom-Header='erm-ok', got %q", rr.Header().Get("X-Custom-Header"))
		}
		if rr.Header().Get("X-Logging-Middleware") != "" {
			t.Fatalf("Expected internal X-Logging-Middleware header stripped, got %q", rr.Header().Get("X-Logging-Middleware"))
		}
	})

	t.Run("Partial Write & Concurrent Header Mutation On Timeout Never Corrupts 504 JSON", func(t *testing.T) {
		shortTimeout := middleware.TimeoutMiddleware(15 * time.Millisecond)
		handler := shortTimeout(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"partial":`))
			deadline := time.Now().Add(35 * time.Millisecond)
			for time.Now().Before(deadline) {
				w.Header().Set("X-Race-Test", "value")
				_, _ = w.Write([]byte("x"))
				time.Sleep(time.Millisecond)
			}
		})

		var wg sync.WaitGroup
		for i := 0; i < 30; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodGet, "/partial-timeout", nil)
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusGatewayTimeout {
					t.Errorf("Expected status 504, got %d", rr.Code)
					return
				}
				var parsed response.Response
				if err := json.Unmarshal(rr.Body.Bytes(), &parsed); err != nil {
					t.Errorf("Expected pure valid 504 JSON, got corrupted body %q: %v", rr.Body.String(), err)
					return
				}
				if parsed.Success || parsed.Message != "Request timeout" {
					t.Errorf("Unexpected 504 payload: %+v", parsed)
				}
			}()
		}
		wg.Wait()
	})

	t.Run("Panic In Child Goroutine Recovered As Clean 500 JSON", func(t *testing.T) {
		handler := middleware.LoggingMiddleware(timeoutMw(func(w http.ResponseWriter, r *http.Request) {
			panic("unexpected nil pointer in handler")
		}))

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("Expected status 500 on recovered panic, got %d", rr.Code)
		}
		var parsed response.Response
		if err := json.Unmarshal(rr.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("Expected valid JSON 500 response, got %q: %v", rr.Body.String(), err)
		}
		if parsed.Success || parsed.Message != "Terjadi kesalahan pada server" {
			t.Fatalf("Unexpected 500 payload: %+v", parsed)
		}
	})

	t.Run("Client Context Canceled Does Not Write False 504", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		handler := timeoutMw(func(w http.ResponseWriter, r *http.Request) {
			cancel()
			<-r.Context().Done()
			time.Sleep(10 * time.Millisecond)
		})

		req := httptest.NewRequest(http.MethodGet, "/canceled", nil).WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code == http.StatusGatewayTimeout {
			t.Fatalf("Expected client cancellation not to emit 504, got %d", rr.Code)
		}
		if rr.Body.Len() != 0 {
			t.Fatalf("Expected empty body on client cancellation, got %q", rr.Body.String())
		}
	})
}
