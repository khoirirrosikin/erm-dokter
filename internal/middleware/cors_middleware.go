package middleware

import (
	"net/http"
)

func CORSMiddleware(allowOrigin string) func(http.Handler) http.Handler {
	if allowOrigin == "" {
		allowOrigin = "*"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Origin, Range, If-Range, X-Requested-With, X-Request-ID, X-API-Key")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, Accept-Ranges, Content-Range, Content-Length, Content-Disposition")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
