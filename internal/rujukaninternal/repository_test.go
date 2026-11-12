package rujukaninternal

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
)

type stubRujukanDriver struct {
	queryFn func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)
	execFn  func(query string, args []driver.NamedValue) (int64, error)
}

func (d *stubRujukanDriver) Connect(context.Context) (driver.Conn, error) {
	return &stubRujukanConn{drv: d}, nil
}
func (d *stubRujukanDriver) Driver() driver.Driver { return nil }

type stubRujukanConn struct {
	drv *stubRujukanDriver
}

func (c *stubRujukanConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (c *stubRujukanConn) Close() error                        { return nil }
func (c *stubRujukanConn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }

func (c *stubRujukanConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.drv.queryFn(query, args)
	if err != nil {
		return nil, err
	}
	return &stubRujukanRows{cols: cols, rows: rows}, nil
}

func (c *stubRujukanConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	affected, err := c.drv.execFn(query, args)
	if err != nil {
		return nil, err
	}
	return driver.RowsAffected(affected), nil
}

type stubRujukanRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubRujukanRows) Columns() []string { return r.cols }
func (r *stubRujukanRows) Close() error      { return nil }
func (r *stubRujukanRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func TestRepository_QueriesAndDeleteWithKodePoli(t *testing.T) {
	t.Run("HapusRujukanInternal includes kd_poli in WHERE clause", func(t *testing.T) {
		var capturedSQL string
		var capturedArgs []driver.NamedValue

		db := sql.OpenDB(&stubRujukanDriver{
			execFn: func(query string, args []driver.NamedValue) (int64, error) {
				capturedSQL = query
				capturedArgs = args
				return 1, nil
			},
		})
		defer db.Close()
		repo := NewRepository(db)

		if err := repo.HapusRujukanInternal(context.Background(), "2026/11/12/000001", "INT", "DR002"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(capturedSQL, "kd_poli = ?") || len(capturedArgs) != 3 {
			t.Fatalf("expected DELETE with no_rawat, kd_poli, kd_dokter (3 args), got len=%d sql=%s", len(capturedArgs), capturedSQL)
		}
		if capturedArgs[0].Value != "2026/11/12/000001" || capturedArgs[1].Value != "INT" || capturedArgs[2].Value != "DR002" {
			t.Fatalf("unexpected DELETE args order: %+v", capturedArgs)
		}
	})

	t.Run("CekAktivitasDokterKonsulen checks SOAPIE, all 3 Penilaian Medis tables, and Resep", func(t *testing.T) {
		var capturedSQL string
		var capturedArgs []driver.NamedValue

		db := sql.OpenDB(&stubRujukanDriver{
			queryFn: func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
				capturedSQL = query
				capturedArgs = args
				return []string{"exists"}, [][]driver.Value{{int64(1)}}, nil
			},
		})
		defer db.Close()
		repo := NewRepository(db)

		aktif, err := repo.CekAktivitasDokterKonsulen(context.Background(), "2026/11/12/000001", "DR002")
		if err != nil || !aktif {
			t.Fatalf("expected aktif=true, got %v (err=%v)", aktif, err)
		}
		for _, table := range []string{
			"pemeriksaan_ralan",
			"penilaian_medis_ralan",
			"penilaian_medis_ralan_kandungan",
			"penilaian_medis_igd",
			"resep_obat",
		} {
			if !strings.Contains(capturedSQL, table) {
				t.Fatalf("expected query to check %s, got: %s", table, capturedSQL)
			}
		}
		if len(capturedArgs) != 10 {
			t.Fatalf("expected 10 query args (5 pairs of no_rawat, kd_dokter), got %d", len(capturedArgs))
		}
	})

	t.Run("CekOpsiPoliDokterAktif checks jadwal, poliklinik.status='1', and dokter.status='1'", func(t *testing.T) {
		var capturedSQL string
		db := sql.OpenDB(&stubRujukanDriver{
			queryFn: func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
				capturedSQL = query
				return []string{"exists"}, [][]driver.Value{{int64(1)}}, nil
			},
		})
		defer db.Close()
		repo := NewRepository(db)

		ok, err := repo.CekOpsiPoliDokterAktif(context.Background(), "INT", "DR002")
		if err != nil || !ok {
			t.Fatalf("expected ok=true, got %v (err=%v)", ok, err)
		}
		if !strings.Contains(capturedSQL, "FROM jadwal") || !strings.Contains(capturedSQL, "p.status = '1'") || !strings.Contains(capturedSQL, "d.status = '1'") {
			t.Fatalf("unexpected CekOpsiPoliDokterAktif SQL: %s", capturedSQL)
		}
	})
}

