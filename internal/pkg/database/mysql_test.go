package database

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func TestBuildDSN(t *testing.T) {
	t.Run("Standard configuration", func(t *testing.T) {
		dsn := buildDSN("127.0.0.1", "3306", "root", "secret", "sik")

		if !strings.Contains(dsn, "clientFoundRows=true") {
			t.Fatalf("expected DSN to contain clientFoundRows=true, got: %s", dsn)
		}

		cfg, err := mysqlDriver.ParseDSN(dsn)
		if err != nil {
			t.Fatalf("expected valid MySQL DSN, got error: %v", err)
		}

		if !cfg.ClientFoundRows {
			t.Error("expected ClientFoundRows to be true")
		}
		if cfg.ParseTime {
			t.Error("expected ParseTime to be false")
		}
		if cfg.Loc == nil || cfg.Loc.String() != time.Local.String() {
			t.Errorf("expected Loc %q, got %v", time.Local.String(), cfg.Loc)
		}
		if cfg.Timeout != 5*time.Second {
			t.Errorf("expected Timeout 5s, got %v", cfg.Timeout)
		}
		if cfg.ReadTimeout != 30*time.Second {
			t.Errorf("expected ReadTimeout 30s, got %v", cfg.ReadTimeout)
		}
		if cfg.WriteTimeout != 30*time.Second {
			t.Errorf("expected WriteTimeout 30s, got %v", cfg.WriteTimeout)
		}
		if cfg.Addr != "127.0.0.1:3306" {
			t.Errorf("expected Addr '127.0.0.1:3306', got '%s'", cfg.Addr)
		}
		if cfg.DBName != "sik" {
			t.Errorf("expected DBName 'sik', got '%s'", cfg.DBName)
		}
	})

	t.Run("Special characters in password and IPv6 host", func(t *testing.T) {
		specialPass := "p@ss/w?rd#1:x"
		dsn := buildDSN("::1", "3306", "dokter_user", specialPass, "sik_prod")

		cfg, err := mysqlDriver.ParseDSN(dsn)
		if err != nil {
			t.Fatalf("expected valid MySQL DSN for special password and IPv6, got error: %v (dsn=%s)", err, dsn)
		}

		if cfg.User != "dokter_user" {
			t.Errorf("expected User 'dokter_user', got %q", cfg.User)
		}
		if cfg.Passwd != specialPass {
			t.Errorf("expected Passwd %q, got %q", specialPass, cfg.Passwd)
		}
		if cfg.Addr != "[::1]:3306" {
			t.Errorf("expected Addr '[::1]:3306', got %q", cfg.Addr)
		}
		if cfg.DBName != "sik_prod" {
			t.Errorf("expected DBName 'sik_prod', got %q", cfg.DBName)
		}
	})

	t.Run("Named timezone location Asia/Jakarta", func(t *testing.T) {
		origLocal := time.Local
		t.Cleanup(func() { time.Local = origLocal })

		jakarta, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			t.Fatalf("failed to load Asia/Jakarta: %v", err)
		}
		time.Local = jakarta

		dsn := buildDSN("127.0.0.1", "3306", "root", "p@ss/w?rd#1:x", "sik")
		cfg, err := mysqlDriver.ParseDSN(dsn)
		if err != nil {
			t.Fatalf("expected valid MySQL DSN with Asia/Jakarta, got error: %v (dsn=%s)", err, dsn)
		}
		if cfg.Loc == nil || cfg.Loc.String() != "Asia/Jakarta" {
			t.Errorf("expected Loc 'Asia/Jakarta', got %v", cfg.Loc)
		}
	})
}

type stubSQLResult struct {
	rows int64
	err  error
}

func (s stubSQLResult) LastInsertId() (int64, error) { return 0, nil }
func (s stubSQLResult) RowsAffected() (int64, error) { return s.rows, s.err }

func TestCheckRowsAffected(t *testing.T) {
	if err := CheckRowsAffected(stubSQLResult{rows: 1}); err != nil {
		t.Fatalf("expected nil error for rows=1, got: %v", err)
	}

	if err := CheckRowsAffected(stubSQLResult{rows: 0}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for rows=0, got: %v", err)
	}

	driverErr := errors.New("driver failure")
	if err := CheckRowsAffected(stubSQLResult{err: driverErr}); !errors.Is(err, driverErr) {
		t.Fatalf("expected driver failure error, got: %v", err)
	}
}
