package middleware

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"time"

	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/pkg/token"
)

const (
	HeaderLogNoRawat      = "X-Log-No-Rawat"
	HeaderLogNoResep      = "X-Log-No-Resep"
	HeaderLogNoPermintaan = "X-Log-No-Permintaan"
)

func SetLogRef(w http.ResponseWriter, headerKey, value string) {
	if w == nil || strings.TrimSpace(value) == "" {
		return
	}
	if w.Header().Get("X-Logging-Middleware") == "true" {
		w.Header().Set(headerKey, strings.TrimSpace(value))
	}
}

type LoggingOptions struct {
	Logger         *logger.Logger
	EncryptionKey  string
	JWTSecret      string
	TrustedProxies []string
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	wroteHeader  bool
	errorDetail  string
	message      string
	user         string
	noRawat      string
	noResep      string
	noPermintaan string
}

func (sw *statusResponseWriter) consumeScratchHeaders() {
	h := sw.ResponseWriter.Header()
	if errDetail := h.Get("X-Error-Detail"); errDetail != "" {
		sw.errorDetail = errDetail
	}
	if logMsg := h.Get("X-Log-Message"); logMsg != "" {
		sw.message = logMsg
	}
	if u := h.Get("X-User-ID"); u != "" {
		sw.user = u
	}
	if nr := h.Get(HeaderLogNoRawat); nr != "" {
		sw.noRawat = strings.TrimSpace(nr)
	}
	if nres := h.Get(HeaderLogNoResep); nres != "" {
		sw.noResep = strings.TrimSpace(nres)
	}
	if np := h.Get(HeaderLogNoPermintaan); np != "" {
		sw.noPermintaan = strings.TrimSpace(np)
	}
	h.Del("X-Error-Detail")
	h.Del("X-Log-Message")
	h.Del("X-User-ID")
	h.Del(HeaderLogNoRawat)
	h.Del(HeaderLogNoResep)
	h.Del(HeaderLogNoPermintaan)
	h.Del("X-Logging-Middleware")
	h.Del("X-Request-Method")
	h.Del("X-Request-Path")
}

func (sw *statusResponseWriter) WriteHeader(code int) {
	if sw.wroteHeader {
		return
	}
	sw.statusCode = code
	sw.wroteHeader = true
	sw.consumeScratchHeaders()
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusResponseWriter) Write(b []byte) (int, error) {
	if !sw.wroteHeader {
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
	return LoggingMiddlewareWithOptions(LoggingOptions{Logger: log})
}

func LoggingMiddlewareWithOptions(opts LoggingOptions) func(http.Handler) http.Handler {
	log := opts.Logger
	if log == nil {
		log = logger.New()
	}
	trustedPrefixes := ParseIPPrefixes(opts.TrustedProxies)

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

			next.ServeHTTP(sw, r)

			sw.consumeScratchHeaders()

			user := sw.user
			if user == "" || user == "-" {
				user = extractUserFromRequest(r, opts.JWTSecret)
			}

			duration := time.Since(start)
			durationMs := float64(duration.Microseconds()) / 1000.0

			clientIP := ExtractClientIP(r, trustedPrefixes)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.statusCode,
				"duration_ms", durationMs,
				"client_ip", clientIP,
				"user", user,
			}

			if logger.GetRequestID(r.Context()) == "" {
				if headerReqID := sw.Header().Get("X-Request-ID"); headerReqID != "" {
					attrs = append(attrs, "request_id", headerReqID)
				}
			}

			attrs = appendLogReferences(attrs, r, sw, opts.EncryptionKey)

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

func appendLogReferences(attrs []any, r *http.Request, sw *statusResponseWriter, encryptionKey string) []any {
	noRawat := sw.noRawat
	noRM := ""
	noResep := sw.noResep
	noPermintaan := sw.noPermintaan

	if encryptionKey != "" {
		if noRawat == "" {
			if encKunjungan := strings.TrimSpace(r.PathValue("id_kunjungan")); encKunjungan != "" {
				if plain, err := crypto.Decrypt(encKunjungan, encryptionKey); err == nil {
					noRawat = strings.TrimSpace(plain)
				}
			} else if strings.HasPrefix(r.URL.Path, "/api/v1/rawat-inap/") {
				if encID := strings.TrimSpace(r.PathValue("id")); encID != "" {
					if plain, err := crypto.Decrypt(encID, encryptionKey); err == nil {
						parts := strings.Split(plain, "~")
						if len(parts) > 0 {
							noRawat = strings.TrimSpace(parts[0])
						}
					}
				}
			}
		}

		if encPasien := strings.TrimSpace(r.PathValue("id_pasien")); encPasien != "" {
			if plain, err := crypto.Decrypt(encPasien, encryptionKey); err == nil {
				noRM = strings.TrimSpace(plain)
			}
		}

		if noResep == "" {
			if encResep := strings.TrimSpace(r.PathValue("id_resep")); encResep != "" {
				if plain, err := crypto.Decrypt(encResep, encryptionKey); err == nil {
					noResep = strings.TrimSpace(plain)
				}
			}
		}

		if noPermintaan == "" {
			if encPermintaan := strings.TrimSpace(r.PathValue("id_permintaan")); encPermintaan != "" {
				if plain, err := crypto.Decrypt(encPermintaan, encryptionKey); err == nil {
					noPermintaan = strings.TrimSpace(plain)
				}
			}
		}
	}

	if noRawat != "" {
		attrs = append(attrs, "no_rawat", noRawat)
	}
	if noRM != "" {
		attrs = append(attrs, "no_rkm_medis", noRM)
	}
	if noResep != "" {
		attrs = append(attrs, "no_resep", noResep)
	}
	if noPermintaan != "" {
		attrs = append(attrs, "no_permintaan", noPermintaan)
	}

	return attrs
}

func extractUserFromRequest(r *http.Request, jwtSecret string) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "-"
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "-"
	}

	if jwtSecret != "" {
		claims, err := token.ValidateToken(parts[1], jwtSecret)
		if err != nil || claims == nil {
			return "-"
		}
		if claims.KodeDokter != "" {
			return claims.KodeDokter
		}
		if claims.Role == RoleService {
			return "service"
		}
		return "-"
	}

	claims, err := token.ExtractClaimsUnverified(parts[1])
	if err != nil || claims.KodeDokter == "" {
		return "-"
	}

	return claims.KodeDokter
}
