package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/pkg/token"
)

var (
	reUpdate     = regexp.MustCompile(`(?i)\bUPDATE\b`)
	reSet        = regexp.MustCompile(`(?i)\bSET\b`)
	reWhere      = regexp.MustCompile(`(?i)\bWHERE\b`)
	reDeleteFrom = regexp.MustCompile(`(?i)\bDELETE\s+FROM\b`)
	reAnd        = regexp.MustCompile(`(?i)\bAND\b`)
	reOr         = regexp.MustCompile(`(?i)\bOR\b`)
	reInsertInto = regexp.MustCompile("(?i)^\\s*insert\\s+into\\s+[`'\"]?([a-zA-Z0-9_]+)[`'\"]?")
)

type TrackerHook struct {
	mu      sync.RWMutex
	db      *sql.DB
	logger  *logger.Logger
	onTrack func(sqle, usere string)
}

func NewTrackerHook(log *logger.Logger) *TrackerHook {
	return &TrackerHook{
		logger: log,
	}
}

func (h *TrackerHook) SetDB(db *sql.DB) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.db = db
}

func (h *TrackerHook) SetLogger(log *logger.Logger) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger = log
}

func (h *TrackerHook) SetOnTrack(fn func(sqle, usere string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onTrack = fn
}

func (h *TrackerHook) Before(ctx context.Context, query string, args ...any) (context.Context, error) {
	return ctx, nil
}
func (h *TrackerHook) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	sqle, isCUD := FormatTrackerSQL(query, args)
	if !isCUD {
		return ctx, nil
	}

	usere := GetUserFromContext(ctx)

	h.mu.RLock()
	activeDB := h.db
	log := h.logger
	callback := h.onTrack
	h.mu.RUnlock()

	if callback != nil {
		callback(sqle, usere)
	}

	if activeDB == nil {
		return ctx, nil
	}

	go func() {
		ctxTracker, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := activeDB.ExecContext(
			ctxTracker,
			"INSERT INTO trackersql (tanggal, sqle, usere) VALUES (NOW(), ?, ?)",
			sqle, usere,
		)
		if err != nil && log != nil {
			log.Warn("Gagal mencatat query tracker ke trackersql: %v", err)
		}
	}()

	return ctx, nil
}

func FormatTrackerSQL(query string, args []any) (string, bool) {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)
	if strings.Contains(strings.ToLower(trimmed), "trackersql") {
		return "", false
	}

	if strings.HasPrefix(upper, "INSERT") {
		matches := reInsertInto.FindStringSubmatch(trimmed)
		if len(matches) >= 2 {
			tableName := strings.ToLower(matches[1])
			formattedArgs := make([]string, len(args))
			for i, arg := range args {
				formattedArgs[i] = formatValue(arg)
			}
			return fmt.Sprintf("insert into %s values(|%s)", tableName, strings.Join(formattedArgs, "|")), true
		}
		return replacePlaceholders(normalizeSQLTemplate(trimmed), args), true
	}

	if strings.HasPrefix(upper, "UPDATE") {
		cleaned := normalizeSQLTemplate(trimmed)
		return replacePlaceholders(cleaned, args), true
	}

	if strings.HasPrefix(upper, "DELETE") {
		cleaned := normalizeSQLTemplate(trimmed)
		return replacePlaceholders(cleaned, args), true
	}

	return "", false
}

func GetUserFromContext(ctx context.Context) string {
	if ctx == nil {
		return "-"
	}
	claims, ok := ctx.Value(middleware.UserClaimKey).(*token.Claims)
	if ok && claims != nil {
		if strings.TrimSpace(claims.KodeDokter) != "" {
			return strings.TrimSpace(claims.KodeDokter)
		}
		if strings.TrimSpace(claims.NamaUser) != "" {
			return strings.TrimSpace(claims.NamaUser)
		}
		if claims.Role != "" {
			return claims.Role
		}
	}
	return "-"
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}
	if nv, ok := v.(driver.NamedValue); ok {
		v = nv.Value
	} else if pnv, ok := v.(*driver.NamedValue); ok && pnv != nil {
		v = pnv.Value
	}
	if valuer, ok := v.(driver.Valuer); ok {
		if val, err := valuer.Value(); err == nil {
			v = val
		}
	}
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", val)
	}
}

func replacePlaceholders(query string, args []any) string {
	var sb strings.Builder
	argIdx := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' && argIdx < len(args) {
			sb.WriteString("|")
			sb.WriteString(formatValue(args[argIdx]))
			sb.WriteString("|")
			argIdx++
		} else {
			sb.WriteByte(query[i])
		}
	}
	return sb.String()
}

func cleanWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func normalizeSQLTemplate(s string) string {
	s = cleanWhitespace(s)
	s = reUpdate.ReplaceAllString(s, "update")
	s = reSet.ReplaceAllString(s, "set")
	s = reWhere.ReplaceAllString(s, "where")
	s = reDeleteFrom.ReplaceAllString(s, "delete from")
	s = reAnd.ReplaceAllString(s, "and")
	s = reOr.ReplaceAllString(s, "or")
	return s
}
