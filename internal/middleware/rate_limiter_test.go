package middleware_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"erm-dokter/internal/middleware"
)

func TestGetClientIP(t *testing.T) {
	t.Run("RemoteAddr with port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:54321"

		ip := middleware.GetClientIP(req)
		if ip != "192.168.1.100" {
			t.Errorf("Expected 192.168.1.100, got %s", ip)
		}
	})

	t.Run("RemoteAddr IPv6 with port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "[::1]:54321"

		ip := middleware.GetClientIP(req)
		if ip != "::1" {
			t.Errorf("Expected ::1, got %s", ip)
		}
	})

	t.Run("RemoteAddr IPv4-mapped IPv6 unmaps to canonical IPv4", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "[::ffff:192.168.1.100]:54321"

		ip := middleware.GetClientIP(req)
		if ip != "192.168.1.100" {
			t.Errorf("Expected 192.168.1.100, got %s", ip)
		}
	})

	t.Run("Default GetClientIP ignores spoofed X-Forwarded-For and X-Real-IP from untrusted peer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:54321"
		req.Header.Set("X-Real-IP", "10.10.10.5")
		req.Header.Set("X-Forwarded-For", "172.16.0.2")

		ip := middleware.GetClientIP(req)
		if ip != "192.168.1.100" {
			t.Errorf("Expected 192.168.1.100, got %s", ip)
		}
	})

	trusted := middleware.ParseIPPrefixes([]string{"198.51.100.10", "127.0.0.1"})

	t.Run("ExtractClientIP trusts X-Real-IP from trusted FE server", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "198.51.100.10:54321"
		req.Header.Set("X-Real-IP", "192.168.30.153")

		ip := middleware.ExtractClientIP(req, trusted)
		if ip != "192.168.30.153" {
			t.Errorf("Expected 192.168.30.153, got %s", ip)
		}
	})

	t.Run("ExtractClientIP walks X-Forwarded-For right-to-left and defeats browser spoofing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		req.Header.Add("X-Forwarded-For", "8.8.8.8")
		req.Header.Add("X-Forwarded-For", "192.168.30.153, 198.51.100.10")

		ip := middleware.ExtractClientIP(req, trusted)
		if ip != "192.168.30.153" {
			t.Errorf("Expected 192.168.30.153, got %s", ip)
		}
	})
}

func TestRateLimiter_BoundedEntries(t *testing.T) {
	limiter := middleware.NewRateLimiterWithMaxEntries(2, 200*time.Millisecond, 5)
	for i := 0; i < 20; i++ {
		if !limiter.Allow(fmt.Sprintf("key-%d", i)) {
			t.Fatalf("expected new key-%d to be allowed", i)
		}
		if limiter.Len() > 5 {
			t.Fatalf("expected bucket count <= 5, got %d", limiter.Len())
		}
	}
}

func TestLoginRateLimitMiddleware(t *testing.T) {
	t.Run("Same IP and same username hits limit", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddleware(2, 500*time.Millisecond, true)

		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		sendLoginReq := func() *httptest.ResponseRecorder {
			body := bytes.NewBufferString(`{"username":"DR001","password":"secret"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "192.168.30.1:12345"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			return rr
		}

		if rr := sendLoginReq(); rr.Code != http.StatusOK {
			t.Fatalf("Req 1: expected 200, got %d", rr.Code)
		}

		if rr := sendLoginReq(); rr.Code != http.StatusOK {
			t.Fatalf("Req 2: expected 200, got %d", rr.Code)
		}

		if rr := sendLoginReq(); rr.Code != http.StatusTooManyRequests {
			t.Fatalf("Req 3: expected 429 Too Many Requests, got %d", rr.Code)
		}
	})

	t.Run("Case-insensitive username prevents uppercase/lowercase bypass", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddleware(2, 500*time.Millisecond, true)

		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		sendLoginReq := func(username string) *httptest.ResponseRecorder {
			body := bytes.NewBufferString(`{"username":"` + username + `","password":"secret"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "192.168.30.2:12345"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			return rr
		}

		if rr := sendLoginReq("DR001"); rr.Code != http.StatusOK {
			t.Fatalf("Req 1: expected 200, got %d", rr.Code)
		}
		if rr := sendLoginReq("dr001"); rr.Code != http.StatusOK {
			t.Fatalf("Req 2: expected 200, got %d", rr.Code)
		}
		if rr := sendLoginReq("Dr001"); rr.Code != http.StatusTooManyRequests {
			t.Fatalf("Req 3: expected 429 for case variant Dr001, got %d", rr.Code)
		}
	})

	t.Run("Spoofed X-Forwarded-For from untrusted peer cannot bypass rate limit", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddlewareWithProxies(2, 500*time.Millisecond, true, []string{"198.51.100.10"})

		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		for i := 1; i <= 3; i++ {
			body := bytes.NewBufferString(`{"username":"DR001","password":"secret"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "192.0.2.55:12345"
			req.Header.Set("X-Forwarded-For", fmt.Sprintf("203.0.113.%d", i))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if i <= 2 && rr.Code != http.StatusOK {
				t.Fatalf("Req %d: expected 200, got %d", i, rr.Code)
			}
			if i == 3 && rr.Code != http.StatusTooManyRequests {
				t.Fatalf("Req 3: expected 429 despite spoofed X-Forwarded-For, got %d", rr.Code)
			}
		}
	})

	t.Run("Trusted FE server isolates rate limits per forwarded client IP", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddlewareWithProxies(2, 500*time.Millisecond, true, []string{"198.51.100.10"})

		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		sendViaFE := func(clientIP string) int {
			body := bytes.NewBufferString(`{"username":"DR001","password":"secret"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "198.51.100.10:50000"
			req.Header.Set("X-Forwarded-For", clientIP)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			return rr.Code
		}

		if code := sendViaFE("192.168.30.101"); code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if code := sendViaFE("192.168.30.101"); code != http.StatusOK {
			t.Fatalf("expected 200, got %d", code)
		}
		if code := sendViaFE("192.168.30.101"); code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 for first workstation IP, got %d", code)
		}
		if code := sendViaFE("192.168.30.102"); code != http.StatusOK {
			t.Fatalf("expected 200 for second workstation IP behind same FE server, got %d", code)
		}
	})

	t.Run("Same IP but different usernames have independent quotas up to IP spray ceiling", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddleware(2, 500*time.Millisecond, true)

		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		sendLoginReq := func(username string) *httptest.ResponseRecorder {
			body := bytes.NewBufferString(`{"username":"` + username + `","password":"secret"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "192.168.30.1:12345"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			return rr
		}

		if rr := sendLoginReq("dr_ahmad"); rr.Code != http.StatusOK {
			t.Fatalf("Doctor A Req 1: expected 200, got %d", rr.Code)
		}
		if rr := sendLoginReq("dr_ahmad"); rr.Code != http.StatusOK {
			t.Fatalf("Doctor A Req 2: expected 200, got %d", rr.Code)
		}

		if rr := sendLoginReq("dr_ahmad"); rr.Code != http.StatusTooManyRequests {
			t.Fatalf("Doctor A Req 3: expected 429, got %d", rr.Code)
		}

		if rr := sendLoginReq("dr_budi"); rr.Code != http.StatusOK {
			t.Fatalf("Doctor B Req 1: expected 200, got %d (should NOT be blocked by Doctor A)", rr.Code)
		}
		if rr := sendLoginReq("dr_budi"); rr.Code != http.StatusOK {
			t.Fatalf("Doctor B Req 2: expected 200, got %d", rr.Code)
		}

		if rr := sendLoginReq("dr_budi"); rr.Code != http.StatusTooManyRequests {
			t.Fatalf("Doctor B Req 3: expected 429, got %d", rr.Code)
		}
	})

	t.Run("Aggregate per-IP ceiling blocks password spraying across many usernames", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddleware(2, 500*time.Millisecond, true)
		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		for i := 0; i < 31; i++ {
			body := bytes.NewBufferString(fmt.Sprintf(`{"username":"spray_user_%d","password":"secret"}`, i))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "192.168.30.77:12345"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if i < 30 && rr.Code != http.StatusOK {
				t.Fatalf("Spray req %d: expected 200, got %d", i+1, rr.Code)
			}
			if i == 30 && rr.Code != http.StatusTooManyRequests {
				t.Fatalf("Spray req 31: expected 429 from aggregate IP ceiling, got %d", rr.Code)
			}
		}
	})

	t.Run("Preserves r.Body including payloads larger than 64KB for downstream handler", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddleware(5, 1*time.Minute, true)

		var receivedLen int
		var capturedUsername string
		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			var payload struct {
				Username string `json:"username"`
			}
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}
			if err := json.Unmarshal(bodyBytes, &payload); err != nil {
				t.Fatalf("Failed to unmarshal body downstream: %v", err)
			}
			receivedLen = len(bodyBytes)
			capturedUsername = payload.Username
			w.WriteHeader(http.StatusOK)
		})

		largePadding := strings.Repeat("x", 80*1024)
		rawJSON := `{"username":"DR007","password":"password123","extra":"` + largePadding + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(rawJSON))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", rr.Code)
		}
		if receivedLen != len(rawJSON) || capturedUsername != "DR007" {
			t.Fatalf("Expected downstream handler to receive full %d bytes and DR007, got %d bytes and %q", len(rawJSON), receivedLen, capturedUsername)
		}
	})

	t.Run("Fallback to IP when username is missing", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddleware(2, 500*time.Millisecond, true)

		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		sendEmptyReq := func() *httptest.ResponseRecorder {
			body := bytes.NewBufferString(`{"password":"secret"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "10.0.0.99:9999"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			return rr
		}

		sendEmptyReq()
		sendEmptyReq()

		if rr := sendEmptyReq(); rr.Code != http.StatusTooManyRequests {
			t.Fatalf("Expected 429 when username is empty, got %d", rr.Code)
		}
	})

	t.Run("Disabled rate limit bypasses throttling", func(t *testing.T) {
		mw := middleware.LoginRateLimitMiddleware(1, 500*time.Millisecond, false)

		handler := mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		sendReq := func() *httptest.ResponseRecorder {
			body := bytes.NewBufferString(`{"username":"dr_bypass","password":"123"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
			req.RemoteAddr = "10.0.0.1:1234"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			return rr
		}

		for i := 0; i < 5; i++ {
			if rr := sendReq(); rr.Code != http.StatusOK {
				t.Fatalf("Req %d: expected 200 when disabled, got %d", i+1, rr.Code)
			}
		}
	})
}
