package middleware

import (
	"net/http"
	"net/netip"
	"strings"

	"erm-dokter/internal/pkg/response"
)

func IPWhitelistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	prefixes := make([]netip.Prefix, 0, len(allowedIPs))
	for _, raw := range allowedIPs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if prefix, err := netip.ParsePrefix(raw); err == nil {
			prefixes = append(prefixes, prefix)
			continue
		}

		if addr, err := netip.ParseAddr(raw); err == nil {
			prefixes = append(prefixes, netip.PrefixFrom(addr.Unmap(), addr.BitLen()))
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			clientAddr, err := netip.ParseAddr(GetClientIP(r))
			if err != nil {
				response.Error(w, http.StatusForbidden, "Akses ditolak: IP Anda tidak memiliki izin untuk mengakses layanan ini", nil)
				return
			}

			clientAddr = clientAddr.Unmap()
			for _, prefix := range prefixes {
				if prefix.Contains(clientAddr) {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Error(w, http.StatusForbidden, "Akses ditolak: IP Anda tidak memiliki izin untuk mengakses layanan ini", nil)
		})
	}
}
