package resep

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

type stubResepQueryFunc func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)

type stubResepConnector struct {
	fn stubResepQueryFunc
}

func (c *stubResepConnector) Connect(context.Context) (driver.Conn, error) {
	return &stubResepConn{fn: c.fn}, nil
}

func (c *stubResepConnector) Driver() driver.Driver { return nil }

type stubResepConn struct {
	fn stubResepQueryFunc
}

func (c *stubResepConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported in stub")
}
func (c *stubResepConn) Close() error              { return nil }
func (c *stubResepConn) Begin() (driver.Tx, error) { return nil, errors.New("tx not supported in stub") }

func (c *stubResepConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.fn(query, args)
	if err != nil {
		return nil, err
	}
	return &stubResepRows{cols: cols, rows: rows}, nil
}

type stubResepRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubResepRows) Columns() []string { return r.cols }
func (r *stubResepRows) Close() error      { return nil }
func (r *stubResepRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func newStubResepDB(t *testing.T, fn stubResepQueryFunc) *sql.DB {
	t.Helper()
	db := sql.OpenDB(&stubResepConnector{fn: fn})
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestRepository_LoadResepRacikan_PreservesAllRacikanDetailsOnRealloc(t *testing.T) {
	t.Run("DetailResep dengan 3 racikan tetap menyimpan seluruh detail bahan obat", func(t *testing.T) {
		db := newStubResepDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			switch {
			case strings.Contains(query, "FROM resep_obat ro"):
				return []string{
						"no_resep", "no_rawat", "tanggal_peresepan", "jam_peresepan",
						"tanggal_perawatan", "jam_perawatan", "tanggal_penyerahan", "jam_penyerahan",
						"status", "kd_dokter", "nama_dokter",
					}, [][]driver.Value{{
						"202609040001", "2026/09/04/000001", "2026-09-04", "09:00:00",
						"0000-00-00", "00:00:00", "0000-00-00", "00:00:00",
						"ralan", "D001", "dr. Budi",
					}}, nil

			case strings.Contains(query, "FROM resep_dokter rd"):
				return []string{"no_resep", "kode_obat", "nama_obat", "jml", "satuan", "aturan_pakai"}, nil, nil

			case strings.Contains(query, "FROM resep_dokter_racikan rdr"):
				if !strings.Contains(query, "CAST(rdr.no_racik AS UNSIGNED)") {
					t.Errorf("expected numeric ordering on no_racik, got: %s", query)
				}
				return []string{
						"no_resep", "no_racik", "nama_racik", "kd_racik", "metode_racik", "jml_dr", "aturan_pakai", "keterangan",
					}, [][]driver.Value{
						{"202609040001", "1", "Puyer Batuk", "R01", "Puyer", int64(10), "3x1", "Habiskan"},
						{"202609040001", "2", "Salep Kulit", "R02", "Salep", int64(1), "2x1", "Oles tipis"},
						{"202609040001", "3", "Kapsul Nyeri", "R03", "Kapsul", int64(15), "3x1", "Sesudah makan"},
					}, nil

			case strings.Contains(query, "FROM resep_dokter_racikan_detail rdrd"):
				return []string{
						"no_resep", "no_racik", "kode_brng", "nama_obat", "kandungan", "jml", "satuan",
					}, [][]driver.Value{
						{"202609040001", "1", "OB001", "Ambroxol", "15 mg", float64(5), "TAB"},
						{"202609040001", "1", "OB002", "CTM", "2 mg", float64(5), "TAB"},
						{"202609040001", "2", "OB003", "Hydrocortisone", "1%", float64(1), "TUBE"},
						{"202609040001", "2", "OB004", "Gentamicin", "0.1%", float64(1), "TUBE"},
						{"202609040001", "3", "OB005", "Paracetamol", "250 mg", float64(7.5), "TAB"},
					}, nil
			default:
				t.Fatalf("unexpected query: %s", query)
				return nil, nil, nil
			}
		})

		repo := NewRepository(db)
		got, err := repo.DetailResep(context.Background(), "202609040001")
		if err != nil {
			t.Fatalf("DetailResep unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("expected non-nil Resep")
		}
		if len(got.ResepDokterRacikan) != 3 {
			t.Fatalf("expected 3 racikan, got %d", len(got.ResepDokterRacikan))
		}

		wantCounts := []int{2, 2, 1}
		wantFirstDrugs := []string{"OB001", "OB003", "OB005"}
		for i, wantLen := range wantCounts {
			detail := got.ResepDokterRacikan[i].DetailRacikan
			if len(detail) != wantLen {
				t.Fatalf("racikan[%d] (%s): expected %d detail items, got %d (%+v)",
					i, got.ResepDokterRacikan[i].NamaRacik, wantLen, len(detail), detail)
			}
			if detail[0].KodeObat != wantFirstDrugs[i] {
				t.Fatalf("racikan[%d]: expected first drug %q, got %q", i, wantFirstDrugs[i], detail[0].KodeObat)
			}
		}
	})

	t.Run("DaftarResep dengan banyak resep dan banyak racikan memetakan detail secara akurat", func(t *testing.T) {
		db := newStubResepDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			switch {
			case strings.Contains(query, "SELECT COUNT(*)"):
				return []string{"total"}, [][]driver.Value{{int64(2)}}, nil

			case strings.Contains(query, "FROM resep_obat ro"):
				return []string{
						"no_resep", "no_rawat", "tanggal_peresepan", "jam_peresepan", "status", "kd_dokter", "nama_dokter",
					}, [][]driver.Value{
						{"R001", "2026/09/04/000001", "2026-09-04", "10:00:00", "ralan", "D001", "dr. Budi"},
						{"R002", "2026/09/04/000001", "2026-09-04", "08:00:00", "ralan", "D001", "dr. Budi"},
					}, nil

			case strings.Contains(query, "FROM resep_dokter rd"):
				return []string{"no_resep", "kode_obat", "nama_obat", "jml", "satuan", "aturan_pakai"}, nil, nil

			case strings.Contains(query, "FROM resep_dokter_racikan rdr"):
				return []string{
						"no_resep", "no_racik", "nama_racik", "kd_racik", "metode_racik", "jml_dr", "aturan_pakai", "keterangan",
					}, [][]driver.Value{
						{"R001", "1", "Racik A1", "R01", "Puyer", int64(10), "3x1", "-"},
						{"R001", "2", "Racik A2", "R01", "Puyer", int64(10), "2x1", "-"},
						{"R002", "1", "Racik B1", "R02", "Kapsul", int64(5), "1x1", "-"},
						{"R002", "2", "Racik B2", "R02", "Kapsul", int64(5), "2x1", "-"},
					}, nil

			case strings.Contains(query, "FROM resep_dokter_racikan_detail rdrd"):
				return []string{
						"no_resep", "no_racik", "kode_brng", "nama_obat", "kandungan", "jml", "satuan",
					}, [][]driver.Value{
						{"R001", "1", "A1_OB1", "Obat A1-1", "10mg", float64(5), "TAB"},
						{"R001", "2", "A2_OB1", "Obat A2-1", "20mg", float64(5), "TAB"},
						{"R002", "1", "B1_OB1", "Obat B1-1", "5mg", float64(5), "TAB"},
						{"R002", "2", "B2_OB1", "Obat B2-1", "15mg", float64(5), "TAB"},
					}, nil
			default:
				t.Fatalf("unexpected query: %s", query)
				return nil, nil, nil
			}
		})

		repo := NewRepository(db)
		filter := FilterDaftarResep{Page: 1, Limit: 10}
		list, total, err := repo.DaftarResep(context.Background(), "2026/09/04/000001", shared.StatusLanjutRawatJalan, filter)
		if err != nil {
			t.Fatalf("DaftarResep unexpected error: %v", err)
		}
		if total != 2 || len(list) != 2 {
			t.Fatalf("expected 2 resep, got total=%d len=%d", total, len(list))
		}
		if len(list[0].ResepDokterRacikan[0].DetailRacikan) != 1 || list[0].ResepDokterRacikan[0].DetailRacikan[0].KodeObat != "A1_OB1" {
			t.Fatalf("R001 racikan 1 detail lost or mismatched: %+v", list[0].ResepDokterRacikan[0].DetailRacikan)
		}
		if len(list[1].ResepDokterRacikan[0].DetailRacikan) != 1 || list[1].ResepDokterRacikan[0].DetailRacikan[0].KodeObat != "B1_OB1" {
			t.Fatalf("R002 racikan 1 detail lost or mismatched: %+v", list[1].ResepDokterRacikan[0].DetailRacikan)
		}
	})
}

type resepTxDriver struct {
	attempt    atomic.Int32
	rollbacks  atomic.Int32
	commits    atomic.Int32
	stmtOpens  atomic.Int32
	stmtCloses atomic.Int32

	queryFn  func(attempt int, inTx bool, query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)
	execFn   func(attempt int, query string, args []driver.NamedValue) error
	stmtExec func(attempt int, query string, args []driver.NamedValue) error
}

func (d *resepTxDriver) Connect(context.Context) (driver.Conn, error) {
	return &resepTxConn{drv: d}, nil
}
func (d *resepTxDriver) Driver() driver.Driver { return nil }

type resepTxConn struct {
	drv        *resepTxDriver
	curAttempt int
	inTx       bool
}

func (c *resepTxConn) Close() error { return nil }

func (c *resepTxConn) Begin() (driver.Tx, error) {
	c.curAttempt = int(c.drv.attempt.Add(1))
	c.inTx = true
	return &resepTx{conn: c}, nil
}

func (c *resepTxConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.drv.queryFn(c.curAttempt, c.inTx, query, args)
	if err != nil {
		return nil, err
	}
	return &stubResepRows{cols: cols, rows: rows}, nil
}

func (c *resepTxConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.drv.execFn != nil {
		if err := c.drv.execFn(c.curAttempt, query, args); err != nil {
			return nil, err
		}
	}
	return driver.RowsAffected(1), nil
}

func (c *resepTxConn) Prepare(query string) (driver.Stmt, error) {
	c.drv.stmtOpens.Add(1)
	return &resepTxStmt{drv: c.drv, attempt: c.curAttempt, query: query}, nil
}

type resepTx struct {
	conn *resepTxConn
	done bool
}

func (t *resepTx) Commit() error {
	t.done = true
	t.conn.inTx = false
	t.conn.drv.commits.Add(1)
	return nil
}

func (t *resepTx) Rollback() error {
	if !t.done {
		t.done = true
		t.conn.inTx = false
		t.conn.drv.rollbacks.Add(1)
	}
	return nil
}

type resepTxStmt struct {
	drv     *resepTxDriver
	attempt int
	query   string
}

func (s *resepTxStmt) Close() error {
	s.drv.stmtCloses.Add(1)
	return nil
}
func (s *resepTxStmt) NumInput() int                                 { return -1 }
func (s *resepTxStmt) Exec([]driver.Value) (driver.Result, error)    { return nil, errors.New("unexpected Exec") }
func (s *resepTxStmt) Query([]driver.Value) (driver.Rows, error)     { return nil, errors.New("unexpected Query") }
func (s *resepTxStmt) ExecContext(_ context.Context, args []driver.NamedValue) (driver.Result, error) {
	if s.drv.stmtExec != nil {
		if err := s.drv.stmtExec(s.attempt, s.query, args); err != nil {
			return nil, err
		}
	}
	return driver.RowsAffected(1), nil
}

func TestSimpanResep_RetryDeadlockAndDuplicate_CleanRollbackAndStmtClose(t *testing.T) {
	t.Run("Recovers from Deadlock 1213 and Header Duplicate 1062, closes statements per attempt", func(t *testing.T) {
		drv := &resepTxDriver{
			queryFn: func(attempt int, inTx bool, query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
				if inTx {
					if strings.Contains(strings.ToUpper(query), "FOR UPDATE") {
						t.Fatalf("order generator query must NOT contain FOR UPDATE: %s", query)
					}
					if attempt == 1 {
						return nil, nil, &mysqlDriver.MySQLError{Number: 1213, Message: "Deadlock found"}
					}
					if attempt == 2 {
						return []string{"no_resep"}, [][]driver.Value{{"202611120005"}}, nil
					}
					return []string{"no_resep"}, [][]driver.Value{{"202611120006"}}, nil
				}
				switch {
				case strings.Contains(query, "FROM resep_obat ro"):
					return []string{
						"no_resep", "no_rawat", "tanggal_peresepan", "jam_peresepan",
						"tanggal_perawatan", "jam_perawatan", "tanggal_penyerahan", "jam_penyerahan",
						"status", "kd_dokter", "nama_dokter",
					}, [][]driver.Value{{
						"202611120007", "2026/11/12/000001", "2026-11-12", "09:00:00",
						"0000-00-00", "00:00:00", "0000-00-00", "00:00:00",
						"ralan", "DR001", "dr. Budi",
					}}, nil
				case strings.Contains(query, "FROM resep_dokter rd"):
					return []string{"no_resep", "kode_obat", "nama_obat", "jml", "satuan", "aturan_pakai"},
						[][]driver.Value{{"202611120007", "OB01", "Paracetamol", float64(10), "TAB", "3x1"}}, nil
				case strings.Contains(query, "FROM resep_dokter_racikan rdr"):
					return []string{"no_resep", "no_racik", "nama_racik", "kd_racik", "metode_racik", "jml_dr", "aturan_pakai", "keterangan"}, nil, nil
				default:
					t.Fatalf("unexpected query: %s", query)
					return nil, nil, nil
				}
			},
			execFn: func(attempt int, _ string, _ []driver.NamedValue) error {
				if attempt == 2 {
					return &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry '202611120006' for key 'PRIMARY'"}
				}
				return nil
			},
			stmtExec: func(attempt int, _ string, _ []driver.NamedValue) error {
				if attempt == 3 {
					return &mysqlDriver.MySQLError{Number: 1213, Message: "Deadlock on child resep_dokter"}
				}
				return nil
			},
		}

		db := sql.OpenDB(drv)
		defer db.Close()
		repo := NewRepository(db)

		got, err := repo.SimpanResep(context.Background(), "DR001", shared.StatusLanjutRawatJalan, SimpanResepRequest{
			NoRawat:          "2026/11/12/000001",
			TanggalPeresepan: "2026-11-12",
			JamPeresepan:     "09:00:00",
			ResepDokter: []ResepDokterInput{
				{ItemObatInput: ItemObatInput{KodeObat: "OB01", Jumlah: 10}, AturanPakai: "3x1"},
			},
		})
		if err != nil {
			t.Fatalf("expected SimpanResep to recover on attempt 4, got error: %v", err)
		}
		if got == nil || got.NoResep != "202611120007" {
			t.Fatalf("expected NoResep 202611120007, got %+v", got)
		}
		if drv.attempt.Load() != 4 || drv.rollbacks.Load() != 3 || drv.commits.Load() != 1 {
			t.Fatalf("unexpected tx counts: attempts=%d rollbacks=%d commits=%d", drv.attempt.Load(), drv.rollbacks.Load(), drv.commits.Load())
		}
		if opens, closes := drv.stmtOpens.Load(), drv.stmtCloses.Load(); opens != 2 || closes != 2 {
			t.Fatalf("expected all prepared statements closed (opens=%d closes=%d)", opens, closes)
		}
	})

	t.Run("Child duplicate 1062 on resep_dokter fails immediately without retry", func(t *testing.T) {
		drv := &resepTxDriver{
			queryFn: func(_ int, _ bool, _ string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
				return []string{"no_resep"}, nil, nil
			},
			stmtExec: func(_ int, _ string, _ []driver.NamedValue) error {
				return &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry '202611120001-OB01' for key 'PRIMARY'"}
			},
		}

		db := sql.OpenDB(drv)
		defer db.Close()
		repo := NewRepository(db)

		_, err := repo.SimpanResep(context.Background(), "DR001", shared.StatusLanjutRawatJalan, SimpanResepRequest{
			NoRawat:          "2026/11/12/000001",
			TanggalPeresepan: "2026-11-12",
			JamPeresepan:     "09:00:00",
			ResepDokter: []ResepDokterInput{
				{ItemObatInput: ItemObatInput{KodeObat: "OB01", Jumlah: 10}, AturanPakai: "3x1"},
				{ItemObatInput: ItemObatInput{KodeObat: "OB01", Jumlah: 5}, AturanPakai: "2x1"},
			},
		})
		if err == nil {
			t.Fatal("expected child duplicate 1062 error")
		}
		if drv.attempt.Load() != 1 || drv.rollbacks.Load() != 1 {
			t.Fatalf("expected 1 attempt & 1 rollback without wasted retries, got attempts=%d rollbacks=%d", drv.attempt.Load(), drv.rollbacks.Load())
		}
	})
}

