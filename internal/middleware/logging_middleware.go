package middleware

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/pkg/token"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	errorDetail string
	message     string
	user        string
}

func (sw *statusResponseWriter) WriteHeader(code int) {
	sw.statusCode = code
	if errDetail := sw.ResponseWriter.Header().Get("X-Error-Detail"); errDetail != "" {
		sw.errorDetail = errDetail
		sw.ResponseWriter.Header().Del("X-Error-Detail")
	}
	if logMsg := sw.ResponseWriter.Header().Get("X-Log-Message"); logMsg != "" {
		sw.message = logMsg
		sw.ResponseWriter.Header().Del("X-Log-Message")
	}
	if u := sw.ResponseWriter.Header().Get("X-User-ID"); u != "" {
		sw.user = u
		sw.ResponseWriter.Header().Del("X-User-ID")
	}
	sw.ResponseWriter.Header().Del("X-Logging-Middleware")
	sw.ResponseWriter.Header().Del("X-Request-Method")
	sw.ResponseWriter.Header().Del("X-Request-Path")
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusResponseWriter) Write(b []byte) (int, error) {
	if sw.statusCode == 0 {
		sw.WriteHeader(http.StatusOK)
	}
	return sw.ResponseWriter.Write(b)
}

func (sw *statusResponseWriter) Flush() {
	if flusher, ok := sw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (sw *statusResponseWriter) Unwrap() http.ResponseWriter {
	return sw.ResponseWriter
}

func (sw *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := sw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return LoggingMiddlewareWithLogger(logger.New())(next)
}

func LoggingMiddlewareWithLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			sw := &statusResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			sw.ResponseWriter.Header().Set("X-Logging-Middleware", "true")

			user := extractUserFromRequest(r)

			next.ServeHTTP(sw, r)

			sw.ResponseWriter.Header().Del("X-Logging-Middleware")
			sw.ResponseWriter.Header().Del("X-Error-Detail")
			sw.ResponseWriter.Header().Del("X-Log-Message")
			sw.ResponseWriter.Header().Del("X-User-ID")
			sw.ResponseWriter.Header().Del("X-Request-Method")
			sw.ResponseWriter.Header().Del("X-Request-Path")

			if sw.user != "" && (user == "-" || user == "") {
				user = sw.user
			}

			duration := time.Since(start)
			durationMs := float64(duration.Microseconds()) / 1000.0

			clientIP := GetClientIP(r)

			requestID := logger.GetRequestID(r.Context())
			if requestID == "" {
				requestID = sw.Header().Get("X-Request-ID")
			}

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.statusCode,
				"duration_ms", durationMs,
				"client_ip", clientIP,
				"user", user,
			}
			if requestID != "" {
				attrs = append(attrs, "request_id", requestID)
			}

			if r.URL.RawQuery != "" {
				attrs = append(attrs, "query", r.URL.RawQuery)
			}
			if ua := r.UserAgent(); ua != "" {
				attrs = append(attrs, "user_agent", ua)
			}

			msg := "HTTP Request"
			if sw.errorDetail != "" {
				msg = sw.errorDetail
			} else if sw.message != "" {
				msg = sw.message
			}

			switch {
			case sw.statusCode >= 500:
				log.ErrorContext(r.Context(), msg, attrs...)
			case sw.statusCode >= 400:
				log.WarnContext(r.Context(), msg, attrs...)
			default:
				log.InfoContext(r.Context(), msg, attrs...)
			}
		})
	}
}

func extractUserFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "-"
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "-"
	}

	claims, err := token.ExtractClaimsUnverified(parts[1])
	if err != nil || claims.KodeDokter == "" {
		return "-"
	}

	return claims.KodeDokter
}
