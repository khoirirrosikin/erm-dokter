package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"erm-dokter/internal/pkg/response"
)

const DefaultMaxRateLimitEntries = 10000

type clientBucket struct {
	tokens    int
	lastReset time.Time
}

type RateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*clientBucket
	limit       int
	window      time.Duration
	maxEntries  int
	lastCleanup time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return NewRateLimiterWithMaxEntries(limit, window, DefaultMaxRateLimitEntries)
}

func NewRateLimiterWithMaxEntries(limit int, window time.Duration, maxEntries int) *RateLimiter {
	if maxEntries <= 0 {
		maxEntries = DefaultMaxRateLimitEntries
	}
	return &RateLimiter{
		buckets:     make(map[string]*clientBucket),
		limit:       limit,
		window:      window,
		maxEntries:  maxEntries,
		lastCleanup: time.Now(),
	}
}

func (rl *RateLimiter) Len() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.buckets)
}

func (rl *RateLimiter) Allow(key string) bool {
	if rl.limit <= 0 {
		return false
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if now.Sub(rl.lastCleanup) > rl.window*2 {
		rl.sweepExpiredLocked(now)
		rl.lastCleanup = now
	}

	b, exists := rl.buckets[key]
	if !exists {
		if len(rl.buckets) >= rl.maxEntries {
			rl.sweepExpiredLocked(now)
			rl.lastCleanup = now
		}
		if len(rl.buckets) >= rl.maxEntries {
			rl.evictOldestLocked()
		}
		rl.buckets[key] = &clientBucket{
			tokens:    rl.limit - 1,
			lastReset: now,
		}
		return true
	}

	if now.Sub(b.lastReset) >= rl.window {
		b.tokens = rl.limit - 1
		b.lastReset = now
		return true
	}

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

func (rl *RateLimiter) sweepExpiredLocked(now time.Time) {
	for k, b := range rl.buckets {
		if now.Sub(b.lastReset) >= rl.window {
			delete(rl.buckets, k)
		}
	}
}

func (rl *RateLimiter) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, b := range rl.buckets {
		if first || b.lastReset.Before(oldestTime) {
			oldestKey = k
			oldestTime = b.lastReset
			first = false
		}
	}
	if !first {
		delete(rl.buckets, oldestKey)
	}
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
	return RateLimitMiddlewareWithProxies(rate, window, nil)
}

func RateLimitMiddlewareWithProxies(rate int, window time.Duration, trustedProxies []string) func(http.HandlerFunc) http.HandlerFunc {
	trustedPrefixes := ParseIPPrefixes(trustedProxies)
	return rateLimit(NewRateLimiter(rate, window), func(r *http.Request) string {
		return ExtractClientIP(r, trustedPrefixes)
	})
}

func LoginRateLimitMiddleware(rate int, window time.Duration, enabled bool) func(http.HandlerFunc) http.HandlerFunc {
	return LoginRateLimitMiddlewareWithProxies(rate, window, enabled, nil)
}

func LoginRateLimitMiddlewareWithProxies(rate int, window time.Duration, enabled bool, trustedProxies []string) func(http.HandlerFunc) http.HandlerFunc {
	if !enabled {
		return func(next http.HandlerFunc) http.HandlerFunc { return next }
	}

	trustedPrefixes := ParseIPPrefixes(trustedProxies)
	userLimiter := NewRateLimiter(rate, window)

	ipBurstLimit := rate * 5
	if ipBurstLimit < 30 {
		ipBurstLimit = 30
	}
	ipLimiter := NewRateLimiter(ipBurstLimit, window)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := ExtractClientIP(r, trustedPrefixes)
			if !ipLimiter.Allow(ip) {
				response.Error(w, http.StatusTooManyRequests, "Terlalu banyak percobaan. Silakan coba lagi nanti.", nil)
				return
			}

			username := peekLoginUsername(r)
			key := ip
			if username != "" {
				key = ip + ":" + username
			}
			if !userLimiter.Allow(key) {
				response.Error(w, http.StatusTooManyRequests, "Terlalu banyak percobaan. Silakan coba lagi nanti.", nil)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

type multiReadCloser struct {
	io.Reader
	closer io.Closer
}

func (m *multiReadCloser) Close() error {
	if m.closer != nil {
		return m.closer.Close()
	}
	return nil
}

func peekLoginUsername(r *http.Request) string {
	if r == nil || r.Body == nil {
		return ""
	}

	origBody := r.Body
	bodyBytes, err := io.ReadAll(io.LimitReader(origBody, 64*1024))
	if len(bodyBytes) > 0 {
		r.Body = &multiReadCloser{
			Reader: io.MultiReader(bytes.NewReader(bodyBytes), origBody),
			closer: origBody,
		}
	}
	if err != nil || len(bodyBytes) == 0 {
		return ""
	}

	var payload struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return ""
	}
	username := strings.ToLower(strings.TrimSpace(payload.Username))
	if len(username) > 128 {
		username = username[:128]
	}
	return username
}
