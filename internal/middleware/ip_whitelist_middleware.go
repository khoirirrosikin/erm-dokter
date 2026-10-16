package middleware

import (
	"net"
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

			remoteHost := strings.TrimSpace(r.RemoteAddr)
			if host, _, err := net.SplitHostPort(remoteHost); err == nil && host != "" {
				remoteHost = host
			}

			remoteAddr, err := netip.ParseAddr(remoteHost)
			if err != nil {
				response.Error(w, http.StatusForbidden, "Akses ditolak: IP Anda tidak memiliki izin untuk mengakses layanan ini", nil)
				return
			}

			remoteAddr = remoteAddr.Unmap()
			for _, prefix := range prefixes {
				if prefix.Contains(remoteAddr) {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Error(w, http.StatusForbidden, "Akses ditolak: IP Anda tidak memiliki izin untuk mengakses layanan ini", nil)
		})
	}
}
