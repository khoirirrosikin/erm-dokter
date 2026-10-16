package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/response"
)

func TestIPWhitelistMiddleware(t *testing.T) {
	allowedIPs := []string{"198.51.100.10", "198.51.100.20", "10.10.0.0/16"}
	mw := middleware.IPWhitelistMiddleware(allowedIPs)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	tests := []struct {
		name           string
		path           string
		remoteAddr     string
		forwardedFor   string
		realIP         string
		expectedStatus int
	}{
		{
			name:           "Allowed IP 1 via RemoteAddr",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "198.51.100.10:51234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed IP 2 via RemoteAddr",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "198.51.100.20:41234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed IP 1 with IPv4-mapped IPv6 RemoteAddr",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "[::ffff:198.51.100.10]:51234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed IP via X-Forwarded-For",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "127.0.0.1:8080",
			forwardedFor:   "198.51.100.10, 10.0.0.1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed IP with port in X-Forwarded-For",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "127.0.0.1:8080",
			forwardedFor:   "198.51.100.20:8080",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed IP via X-Real-IP",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "127.0.0.1:8080",
			realIP:         "198.51.100.20",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed CIDR subnet IP",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "10.10.5.20:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Rejected IP not in whitelist",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "192.0.2.100:51234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Rejected IP from different subnet",
			path:           "/api/v1/rawat-jalan",
			remoteAddr:     "172.16.0.5:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Bypass health check endpoint regardless of IP",
			path:           "/health",
			remoteAddr:     "192.0.2.100:51234",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.expectedStatus == http.StatusForbidden {
				var errResp response.Response
				if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
					t.Fatalf("Failed to parse forbidden response JSON: %v", err)
				}
				if errResp.Success {
					t.Errorf("Expected success to be false on forbidden")
				}
			}
		})
	}
}
