package radiologi

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"erm-dokter/internal/shared"
)

type txTrackerDriver struct {
	attempt    atomic.Int32
	rollbacks  atomic.Int32
	commits    atomic.Int32
	stmtOpens  atomic.Int32
	stmtCloses atomic.Int32

	queryFn  func(attempt int, query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)
	execFn   func(attempt int, query string, args []driver.NamedValue) error
	stmtExec func(attempt int, query string, args []driver.NamedValue) error
}

func (d *txTrackerDriver) Connect(context.Context) (driver.Conn, error) {
	return &txTrackerConn{drv: d}, nil
}
func (d *txTrackerDriver) Driver() driver.Driver { return nil }

type txTrackerConn struct {
	drv        *txTrackerDriver
	curAttempt int
}

func (c *txTrackerConn) Close() error { return nil }

func (c *txTrackerConn) Begin() (driver.Tx, error) {
	c.curAttempt = int(c.drv.attempt.Add(1))
	return &txTrackerTx{drv: c.drv}, nil
}

func (c *txTrackerConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.drv.queryFn(c.curAttempt, query, args)
	if err != nil {
		return nil, err
	}
	return &txTrackerRows{cols: cols, rows: rows}, nil
}

func (c *txTrackerConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.drv.execFn != nil {
		if err := c.drv.execFn(c.curAttempt, query, args); err != nil {
			return nil, err
		}
	}
	return driver.RowsAffected(1), nil
}

func (c *txTrackerConn) Prepare(query string) (driver.Stmt, error) {
	c.drv.stmtOpens.Add(1)
	return &txTrackerStmt{drv: c.drv, attempt: c.curAttempt, query: query}, nil
}

type txTrackerTx struct {
	drv  *txTrackerDriver
	done bool
}

func (t *txTrackerTx) Commit() error {
	t.done = true
	t.drv.commits.Add(1)
	return nil
}

func (t *txTrackerTx) Rollback() error {
	if !t.done {
		t.done = true
		t.drv.rollbacks.Add(1)
	}
	return nil
}

type txTrackerStmt struct {
	drv     *txTrackerDriver
	attempt int
	query   string
}

func (s *txTrackerStmt) Close() error {
	s.drv.stmtCloses.Add(1)
	return nil
}
func (s *txTrackerStmt) NumInput() int { return -1 }
func (s *txTrackerStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New("unexpected Exec without context")
}
func (s *txTrackerStmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, errors.New("unexpected Query")
}
func (s *txTrackerStmt) ExecContext(_ context.Context, args []driver.NamedValue) (driver.Result, error) {
	if s.drv.stmtExec != nil {
		if err := s.drv.stmtExec(s.attempt, s.query, args); err != nil {
			return nil, err
		}
	}
	return driver.RowsAffected(1), nil
}

type txTrackerRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *txTrackerRows) Columns() []string { return r.cols }
func (r *txTrackerRows) Close() error      { return nil }
func (r *txTrackerRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func TestSimpanPermintaanRadiologi_RetryDeadlockAndDuplicate_CleanRollbackAndStmtClose(t *testing.T) {
	t.Run("Attempt 1 Deadlock 1213, Attempt 2 Header Duplicate 1062, Attempt 3 Child Deadlock 1213, Attempt 4 Commits", func(t *testing.T) {
		drv := &txTrackerDriver{
			queryFn: func(attempt int, query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
				if strings.Contains(strings.ToUpper(query), "FOR UPDATE") {
					t.Fatalf("SELECT must not contain FOR UPDATE: %s", query)
				}
				if attempt == 1 {
					return nil, nil, &mysqlDriver.MySQLError{Number: 1213, Message: "Deadlock found"}
				}
				if attempt == 2 {
					return []string{"noorder"}, nil, nil
				}
				return []string{"noorder"}, [][]driver.Value{{"RAD202611120001"}}, nil
			},
			execFn: func(attempt int, query string, args []driver.NamedValue) error {
				if attempt == 2 {
					return &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry 'RAD202611120001' for key 'PRIMARY'"}
				}
				return nil
			},
			stmtExec: func(attempt int, query string, args []driver.NamedValue) error {
				if attempt == 3 {
					return &mysqlDriver.MySQLError{Number: 1213, Message: "Deadlock found on child insert"}
				}
				return nil
			},
		}

		db := sql.OpenDB(drv)
		defer db.Close()
		repo := NewRepository(db)

		orderNo, err := repo.SimpanPermintaanRadiologi(
			context.Background(),
			"2026/11/12/000001",
			"DR001",
			shared.StatusLanjutRawatJalan,
			SimpanPermintaanRadiologiRequest{
				TanggalPermintaan: "2026-11-12",
				JamPermintaan:     "10:00:00",
				DiagnosaKlinis:    "Batuk Kronis",
			},
			[]string{"RAD01", "RAD02"},
		)
		if err != nil {
			t.Fatalf("expected success on attempt 4, got error: %v", err)
		}
		if orderNo != "RAD202611120002" {
			t.Fatalf("expected RAD202611120002, got %q", orderNo)
		}
		if got := drv.attempt.Load(); got != 4 {
			t.Fatalf("expected 4 attempts, got %d", got)
		}
		if got := drv.rollbacks.Load(); got != 3 {
			t.Fatalf("expected 3 failed attempts rolled back, got %d", got)
		}
		if got := drv.commits.Load(); got != 1 {
			t.Fatalf("expected 1 commit, got %d", got)
		}
		if opens, closes := drv.stmtOpens.Load(), drv.stmtCloses.Load(); opens != 2 || closes != 2 {
			t.Fatalf("expected all prepared statements closed per attempt (opens=%d closes=%d)", opens, closes)
		}
	})

	t.Run("Child duplicate 1062 does NOT waste retries", func(t *testing.T) {
		drv := &txTrackerDriver{
			queryFn: func(attempt int, query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
				return []string{"noorder"}, nil, nil
			},
			stmtExec: func(attempt int, query string, args []driver.NamedValue) error {
				return &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry 'RAD202611120001-RAD01' for key 'PRIMARY'"}
			},
		}

		db := sql.OpenDB(drv)
		defer db.Close()
		repo := NewRepository(db)

		_, err := repo.SimpanPermintaanRadiologi(
			context.Background(),
			"2026/11/12/000001",
			"DR001",
			shared.StatusLanjutRawatJalan,
			SimpanPermintaanRadiologiRequest{
				TanggalPermintaan: "2026-11-12",
				JamPermintaan:     "10:00:00",
				DiagnosaKlinis:    "Batuk",
			},
			[]string{"RAD01", "RAD01"},
		)
		if err == nil {
			t.Fatal("expected error for child duplicate 1062")
		}
		if got := drv.attempt.Load(); got != 1 {
			t.Fatalf("expected child duplicate 1062 to fail immediately on attempt 1 without retry, got %d attempts", got)
		}
		if got := drv.rollbacks.Load(); got != 1 {
			t.Fatalf("expected 1 rollback, got %d", got)
		}
	})
}

