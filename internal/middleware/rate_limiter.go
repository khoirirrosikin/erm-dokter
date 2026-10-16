package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"erm-dokter/internal/pkg/response"
)

type clientBucket struct {
	tokens    int
	lastReset time.Time
}

type RateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*clientBucket
	limit       int
	window      time.Duration
	lastCleanup time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:     make(map[string]*clientBucket),
		limit:       limit,
		window:      window,
		lastCleanup: time.Now(),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if now.Sub(rl.lastCleanup) > rl.window*2 {
		for k, b := range rl.buckets {
			if now.Sub(b.lastReset) > rl.window*2 {
				delete(rl.buckets, k)
			}
		}
		rl.lastCleanup = now
	}

	b, exists := rl.buckets[key]
	if !exists || now.Sub(b.lastReset) >= rl.window {
		rl.buckets[key] = &clientBucket{
			tokens:    rl.limit - 1,
			lastReset: now,
		}
		return true
	}

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

func rateLimit(limiter *RateLimiter, keyFunc func(r *http.Request) string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(keyFunc(r)) {
				response.Error(w, http.StatusTooManyRequests, "Terlalu banyak percobaan. Silakan coba lagi nanti.", nil)
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

func RateLimitMiddleware(rate int, window time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return rateLimit(NewRateLimiter(rate, window), GetClientIP)
}

func LoginRateLimitMiddleware(rate int, window time.Duration, enabled bool) func(http.HandlerFunc) http.HandlerFunc {
	if !enabled {
		return func(next http.HandlerFunc) http.HandlerFunc { return next }
	}

	limiter := NewRateLimiter(rate, window)
	return rateLimit(limiter, func(r *http.Request) string {
		ip := GetClientIP(r)
		username := peekLoginUsername(r)
		if username != "" {
			return ip + ":" + username
		}
		return ip
	})
}

func peekLoginUsername(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil || len(bodyBytes) == 0 {
		return ""
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var payload struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Username)
}

func GetClientIP(r *http.Request) string {
	cleanIP := func(raw string) string {
		raw = strings.TrimSpace(raw)
		if host, _, err := net.SplitHostPort(raw); err == nil && host != "" {
			return host
		}
		return raw
	}

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first := strings.TrimSpace(strings.Split(forwarded, ",")[0]); first != "" {
			return cleanIP(first)
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return cleanIP(realIP)
	}
	return cleanIP(r.RemoteAddr)
}
