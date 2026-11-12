package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

const DefaultOrderMaxAttempts = 6

type OrderNumberSpec struct {
	Table      string
	Column     string
	CodePrefix string
}

var (
	OrderSpecResep     = OrderNumberSpec{Table: "resep_obat", Column: "no_resep", CodePrefix: ""}
	OrderSpecRadiologi = OrderNumberSpec{Table: "permintaan_radiologi", Column: "noorder", CodePrefix: "RAD"}
	OrderSpecLabPK     = OrderNumberSpec{Table: "permintaan_lab", Column: "noorder", CodePrefix: "PK"}
	OrderSpecLabPA     = OrderNumberSpec{Table: "permintaan_labpa", Column: "noorder", CodePrefix: "PA"}
	OrderSpecLabMB     = OrderNumberSpec{Table: "permintaan_labmb", Column: "noorder", CodePrefix: "MB"}
)

func isValidOrderSpec(spec OrderNumberSpec) bool {
	switch spec {
	case OrderSpecResep, OrderSpecRadiologi, OrderSpecLabPK, OrderSpecLabPA, OrderSpecLabMB:
		return true
	default:
		return false
	}
}

func IsDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "1062") ||
		strings.Contains(errStr, "duplicate entry") ||
		strings.Contains(errStr, "duplicate key")
}

func IsDeadlockOrLockTimeout(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1213 || mysqlErr.Number == 1205
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "1213") ||
		strings.Contains(errStr, "deadlock") ||
		strings.Contains(errStr, "1205") ||
		strings.Contains(errStr, "lock wait timeout")
}

func IsRetryableOrderError(err error) bool {
	return IsDuplicateEntry(err) || IsDeadlockOrLockTimeout(err)
}

func GenerateDailyOrderNumber(
	ctx context.Context,
	tx *sql.Tx,
	spec OrderNumberSpec,
	tanggal string,
	attempt int,
	now time.Time,
) (string, error) {
	if !isValidOrderSpec(spec) {
		return "", fmt.Errorf("database: spesifikasi tabel/kolom order tidak valid (%s.%s)", spec.Table, spec.Column)
	}
	if now.IsZero() {
		now = time.Now()
	}

	prefixDate := now.Format("20060102")
	if parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(tanggal)); err == nil {
		prefixDate = parsedDate.Format("20060102")
	}

	prefix := spec.CodePrefix + prefixDate
	expectedLen := len(prefix) + 4
	minOrder := prefix + "0000"
	maxOrder := prefix + "9999"

	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE %s BETWEEN ? AND ? AND CHAR_LENGTH(%s) = ? ORDER BY %s DESC LIMIT 1`,
		spec.Column, spec.Table, spec.Column, spec.Column, spec.Column,
	)

	var lastToday string
	err := tx.QueryRowContext(ctx, query, minOrder, maxOrder, expectedLen).Scan(&lastToday)
	if errors.Is(err, sql.ErrNoRows) {
		return prefix + "0001", nil
	}
	if err != nil {
		return "", err
	}

	if len(lastToday) != expectedLen || !strings.HasPrefix(lastToday, prefix) {
		return formatFallbackOrderNumber(spec.CodePrefix, attempt, now), nil
	}

	tail := lastToday[len(lastToday)-4:]
	next, err := strconv.Atoi(tail)
	if err != nil || next < 0 || next >= 9999 {
		return formatFallbackOrderNumber(spec.CodePrefix, attempt, now), nil
	}

	return fmt.Sprintf("%s%04d", prefix, next+1), nil
}

func formatFallbackOrderNumber(codePrefix string, attempt int, now time.Time) string {
	shift := 0
	if attempt > 1 {
		shift = attempt - 1
	}
	return codePrefix + now.Add(time.Duration(shift)*time.Second).Format("060102150405")
}

func RunOrderTxWithRetry(
	ctx context.Context,
	maxAttempts int,
	attemptFn func(attempt int) (orderNo string, retry bool, err error),
) (string, error) {
	if maxAttempts <= 0 {
		maxAttempts = DefaultOrderMaxAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			backoff := time.Duration(5+rand.IntN(21)) * time.Millisecond
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return "", ctx.Err()
			case <-timer.C:
			}
		}

		orderNo, retry, err := attemptFn(attempt)
		if err == nil {
			return orderNo, nil
		}
		if !retry {
			return "", err
		}
		lastErr = err
	}

	return "", fmt.Errorf("gagal membuat nomor order unik setelah %d percobaan: %w", maxAttempts, lastErr)
}

