package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func ParseIPPrefixes(entries []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(entries))
	for _, raw := range entries {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if prefix, err := netip.ParsePrefix(raw); err == nil {
			addr := prefix.Addr().WithZone("")
			bits := prefix.Bits()
			if addr.Is4In6() && bits >= 96 {
				unmapped := addr.Unmap()
				if p := netip.PrefixFrom(unmapped, bits-96); p.IsValid() {
					prefixes = append(prefixes, p.Masked())
				}
				continue
			}
			if p := netip.PrefixFrom(addr, bits); p.IsValid() {
				prefixes = append(prefixes, p.Masked())
			}
			continue
		}

		if addr, err := netip.ParseAddr(raw); err == nil {
			unmapped := addr.WithZone("").Unmap()
			if p := netip.PrefixFrom(unmapped, unmapped.BitLen()); p.IsValid() {
				prefixes = append(prefixes, p)
			}
		}
	}
	return prefixes
}

func ExtractPeerAddr(r *http.Request, trustedProxies []netip.Prefix) (netip.Addr, bool) {
	if r == nil {
		return netip.Addr{}, false
	}

	remoteIP, ok := parseNormalizedAddr(r.RemoteAddr)
	if !ok {
		return netip.Addr{}, false
	}

	if len(trustedProxies) == 0 || !containsAddr(trustedProxies, remoteIP) {
		return remoteIP, true
	}

	if forwarded := strings.TrimSpace(strings.Join(r.Header.Values("X-Forwarded-For"), ",")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		var leftmostValid netip.Addr
		for i := len(parts) - 1; i >= 0; i-- {
			ip, valid := parseNormalizedAddr(parts[i])
			if !valid {
				return netip.Addr{}, false
			}
			leftmostValid = ip
			if !containsAddr(trustedProxies, ip) {
				return ip, true
			}
		}
		if leftmostValid.IsValid() {
			return leftmostValid, true
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		ip, valid := parseNormalizedAddr(realIP)
		if !valid {
			return netip.Addr{}, false
		}
		return ip, true
	}

	return remoteIP, true
}

func ExtractClientIP(r *http.Request, trustedProxies []netip.Prefix) string {
	if r == nil {
		return ""
	}

	remoteIP, ok := parseNormalizedAddr(r.RemoteAddr)
	if !ok {
		return cleanHost(r.RemoteAddr)
	}

	if len(trustedProxies) == 0 || !containsAddr(trustedProxies, remoteIP) {
		return remoteIP.String()
	}

	if forwarded := strings.TrimSpace(strings.Join(r.Header.Values("X-Forwarded-For"), ",")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		var leftmostValid netip.Addr
		for i := len(parts) - 1; i >= 0; i-- {
			ip, valid := parseNormalizedAddr(parts[i])
			if !valid {
				break
			}
			leftmostValid = ip
			if !containsAddr(trustedProxies, ip) {
				return ip.String()
			}
		}
		if leftmostValid.IsValid() {
			return leftmostValid.String()
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if ip, valid := parseNormalizedAddr(realIP); valid {
			return ip.String()
		}
	}

	return remoteIP.String()
}

func GetClientIP(r *http.Request) string {
	return ExtractClientIP(r, nil)
}

func cleanHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if host, _, err := net.SplitHostPort(raw); err == nil && host != "" {
		return host
	}
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		return raw[1 : len(raw)-1]
	}
	return raw
}

func parseNormalizedAddr(raw string) (netip.Addr, bool) {
	host := cleanHost(raw)
	if host == "" {
		return netip.Addr{}, false
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.WithZone("").Unmap(), true
}

func containsAddr(prefixes []netip.Prefix, addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

