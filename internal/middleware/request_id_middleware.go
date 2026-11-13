package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"

	"erm-dokter/internal/pkg/logger"
)

func GenerateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if reqID == "" {
			reqID = GenerateRequestID()
		}

		w.Header().Set("X-Request-ID", reqID)
		w.Header().Set("X-Request-Method", r.Method)
		w.Header().Set("X-Request-Path", r.URL.Path)

		ctx := logger.WithRequestID(r.Context(), reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
