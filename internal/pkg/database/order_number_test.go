package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func TestMySQLErrorHelpers(t *testing.T) {
	dupMySQL := &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry '202611120001' for key 'PRIMARY'"}
	deadlockMySQL := &mysqlDriver.MySQLError{Number: 1213, Message: "Deadlock found when trying to get lock"}
	lockWaitMySQL := &mysqlDriver.MySQLError{Number: 1205, Message: "Lock wait timeout exceeded"}
	fkMySQL := &mysqlDriver.MySQLError{Number: 1452, Message: "Cannot add or update a child row: foreign key constraint fails (`sik`.`resep_obat` PRIMARY KEY)"}

	if !IsDuplicateEntry(fmt.Errorf("wrapped: %w", dupMySQL)) {
		t.Fatal("expected wrapped 1062 MySQLError to be duplicate entry")
	}
	if IsDuplicateEntry(fkMySQL) {
		t.Fatal("expected 1452 FK error mentioning PRIMARY KEY NOT to match IsDuplicateEntry")
	}
	if !IsDeadlockOrLockTimeout(fmt.Errorf("wrapped: %w", deadlockMySQL)) {
		t.Fatal("expected wrapped 1213 MySQLError to match deadlock")
	}
	if !IsDeadlockOrLockTimeout(lockWaitMySQL) {
		t.Fatal("expected 1205 MySQLError to match lock timeout")
	}
	if !IsRetryableOrderError(dupMySQL) || !IsRetryableOrderError(deadlockMySQL) || !IsRetryableOrderError(lockWaitMySQL) {
		t.Fatal("expected 1062, 1213, 1205 to all be retryable order errors")
	}
	if IsRetryableOrderError(errors.New("driver: bad connection")) || IsRetryableOrderError(nil) {
		t.Fatal("expected network/nil errors NOT to be retryable order errors")
	}
}

type stubOrderDBConnector struct {
	queryFn func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)
}

func (c *stubOrderDBConnector) Connect(context.Context) (driver.Conn, error) {
	return &stubOrderDBConn{queryFn: c.queryFn}, nil
}
func (c *stubOrderDBConnector) Driver() driver.Driver { return nil }

type stubOrderDBConn struct {
	queryFn func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)
}

func (c *stubOrderDBConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (c *stubOrderDBConn) Close() error                        { return nil }
func (c *stubOrderDBConn) Begin() (driver.Tx, error)           { return stubOrderTx{}, nil }
func (c *stubOrderDBConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.queryFn(query, args)
	if err != nil {
		return nil, err
	}
	return &stubOrderRows{cols: cols, rows: rows}, nil
}

type stubOrderTx struct{}

func (stubOrderTx) Commit() error   { return nil }
func (stubOrderTx) Rollback() error { return nil }

type stubOrderRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubOrderRows) Columns() []string { return r.cols }
func (r *stubOrderRows) Close() error      { return nil }
func (r *stubOrderRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func TestGenerateDailyOrderNumber_AllSpecsAndFallbackWidths(t *testing.T) {
	fixedNow := time.Date(2026, 11, 12, 13, 45, 9, 0, time.Local)

	specs := []struct {
		name        string
		spec        OrderNumberSpec
		maxColWidth int
		wantFirst   string
		lastNormal  string
		wantNext    string
		maxReached  string
		wantFB1     string
		wantFB2     string
	}{
		{
			name:        "Resep (VARCHAR(14))",
			spec:        OrderSpecResep,
			maxColWidth: 14,
			wantFirst:   "202611120001",
			lastNormal:  "202611120041",
			wantNext:    "202611120042",
			maxReached:  "202611129999",
			wantFB1:     "261112134509",
			wantFB2:     "261112134510",
		},
		{
			name:        "Radiologi (VARCHAR(15))",
			spec:        OrderSpecRadiologi,
			maxColWidth: 15,
			wantFirst:   "RAD202611120001",
			lastNormal:  "RAD202611120009",
			wantNext:    "RAD202611120010",
			maxReached:  "RAD202611129999",
			wantFB1:     "RAD261112134509",
			wantFB2:     "RAD261112134510",
		},
		{
			name:        "Lab PK (VARCHAR(15))",
			spec:        OrderSpecLabPK,
			maxColWidth: 15,
			wantFirst:   "PK202611120001",
			lastNormal:  "PK202611120199",
			wantNext:    "PK202611120200",
			maxReached:  "PK202611129999",
			wantFB1:     "PK261112134509",
			wantFB2:     "PK261112134510",
		},
		{
			name:        "Lab PA (VARCHAR(15))",
			spec:        OrderSpecLabPA,
			maxColWidth: 15,
			wantFirst:   "PA202611120001",
			lastNormal:  "PA202611120001",
			wantNext:    "PA202611120002",
			maxReached:  "PA202611129999",
			wantFB1:     "PA261112134509",
			wantFB2:     "PA261112134510",
		},
		{
			name:        "Lab MB (VARCHAR(15))",
			spec:        OrderSpecLabMB,
			maxColWidth: 15,
			wantFirst:   "MB202611120001",
			lastNormal:  "MB202611120005",
			wantNext:    "MB202611120006",
			maxReached:  "MB202611129999",
			wantFB1:     "MB261112134509",
			wantFB2:     "MB261112134510",
		},
	}

	for _, tc := range specs {
		t.Run(tc.name, func(t *testing.T) {
			var simulatedRows [][]driver.Value

			db := sql.OpenDB(&stubOrderDBConnector{
				queryFn: func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
					if strings.Contains(strings.ToUpper(query), "FOR UPDATE") {
						t.Fatalf("query must NOT contain FOR UPDATE: %s", query)
					}
					if !strings.Contains(query, "CHAR_LENGTH(") {
						t.Fatalf("query must filter by CHAR_LENGTH: %s", query)
					}
					if len(args) != 3 {
						t.Fatalf("expected 3 query args (min, max, expectedLen), got %d", len(args))
					}
					return []string{tc.spec.Column}, simulatedRows, nil
				},
			})
			defer db.Close()

			simulatedRows = nil
			tx1, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			gotFirst, err := GenerateDailyOrderNumber(context.Background(), tx1, tc.spec, "2026-11-12", 1, fixedNow)
			_ = tx1.Rollback()
			if err != nil || gotFirst != tc.wantFirst {
				t.Fatalf("expected first=%q, got %q (err=%v)", tc.wantFirst, gotFirst, err)
			}

			simulatedRows = [][]driver.Value{{tc.lastNormal}}
			tx2, _ := db.BeginTx(context.Background(), nil)
			gotNext, err := GenerateDailyOrderNumber(context.Background(), tx2, tc.spec, "2026-11-12", 1, fixedNow)
			_ = tx2.Rollback()
			if err != nil || gotNext != tc.wantNext {
				t.Fatalf("expected next=%q, got %q (err=%v)", tc.wantNext, gotNext, err)
			}

			simulatedRows = [][]driver.Value{{tc.maxReached}}
			tx3, _ := db.BeginTx(context.Background(), nil)
			gotFB1, err := GenerateDailyOrderNumber(context.Background(), tx3, tc.spec, "2026-11-12", 1, fixedNow)
			_ = tx3.Rollback()
			if err != nil || gotFB1 != tc.wantFB1 {
				t.Fatalf("expected fallback1=%q, got %q (err=%v)", tc.wantFB1, gotFB1, err)
			}
			if len(gotFB1) > tc.maxColWidth {
				t.Fatalf("fallback %q length %d exceeds column width %d!", gotFB1, len(gotFB1), tc.maxColWidth)
			}

			tx4, _ := db.BeginTx(context.Background(), nil)
			gotFB2, err := GenerateDailyOrderNumber(context.Background(), tx4, tc.spec, "2026-11-12", 2, fixedNow)
			_ = tx4.Rollback()
			if err != nil || gotFB2 != tc.wantFB2 {
				t.Fatalf("expected fallback2 (+1s)=%q, got %q (err=%v)", tc.wantFB2, gotFB2, err)
			}
		})
	}

	t.Run("Rejects unknown table/column spec", func(t *testing.T) {
		db := sql.OpenDB(&stubOrderDBConnector{})
		defer db.Close()
		tx, _ := db.BeginTx(context.Background(), nil)
		defer tx.Rollback()

		_, err := GenerateDailyOrderNumber(context.Background(), tx, OrderNumberSpec{Table: "users; DROP TABLE", Column: "id"}, "2026-11-12", 1, fixedNow)
		if err == nil {
			t.Fatal("expected invalid OrderNumberSpec to fail")
		}
	})
}

func TestRunOrderTxWithRetry(t *testing.T) {
	t.Run("Recovers after Deadlock 1213 then Duplicate 1062", func(t *testing.T) {
		calls := 0
		got, err := RunOrderTxWithRetry(context.Background(), 5, func(attempt int) (string, bool, error) {
			calls++
			switch attempt {
			case 1:
				err := &mysqlDriver.MySQLError{Number: 1213, Message: "Deadlock found"}
				return "", IsRetryableOrderError(err), err
			case 2:
				err := &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry"}
				return "", IsRetryableOrderError(err), err
			default:
				return "RAD202611120003", false, nil
			}
		})
		if err != nil {
			t.Fatalf("expected retry recovery, got err=%v", err)
		}
		if calls != 3 || got != "RAD202611120003" {
			t.Fatalf("expected 3 calls and RAD202611120003, got calls=%d order=%q", calls, got)
		}
	})

	t.Run("Stops immediately on non-retryable error (e.g. network or child FK/duplicate)", func(t *testing.T) {
		calls := 0
		nonRetryErr := errors.New("Error 1452: foreign key constraint fails")
		_, err := RunOrderTxWithRetry(context.Background(), 5, func(attempt int) (string, bool, error) {
			calls++
			return "", false, nonRetryErr
		})
		if !errors.Is(err, nonRetryErr) || calls != 1 {
			t.Fatalf("expected immediate stop on attempt 1, got calls=%d err=%v", calls, err)
		}
	})

	t.Run("Aborts immediately when context is cancelled during backoff", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		_, err := RunOrderTxWithRetry(ctx, 5, func(attempt int) (string, bool, error) {
			calls++
			cancel()
			dup := &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry"}
			return "", true, dup
		})
		if !errors.Is(err, context.Canceled) || calls != 1 {
			t.Fatalf("expected context.Canceled after attempt 1, got calls=%d err=%v", calls, err)
		}
	})
}

