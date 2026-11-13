package middleware

import (
	"net/http"

	"erm-dokter/internal/pkg/response"
)

func IPWhitelistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	return IPWhitelistMiddlewareWithProxies(allowedIPs, nil)
}

func IPWhitelistMiddlewareWithProxies(allowedIPs, trustedProxies []string) func(http.Handler) http.Handler {
	allowedPrefixes := ParseIPPrefixes(allowedIPs)
	proxyPrefixes := ParseIPPrefixes(trustedProxies)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			peerAddr, ok := ExtractPeerAddr(r, proxyPrefixes)
			if !ok || !containsAddr(allowedPrefixes, peerAddr) {
				response.Error(w, http.StatusForbidden, "Akses ditolak: IP Anda tidak memiliki izin untuk mengakses layanan ini", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
