package laboratorium

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"erm-dokter/internal/shared"
)

type stubLabQueryFunc func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)

type stubLabConnector struct {
	fn stubLabQueryFunc
}

func (c *stubLabConnector) Connect(context.Context) (driver.Conn, error) {
	return &stubLabConn{fn: c.fn}, nil
}

func (c *stubLabConnector) Driver() driver.Driver { return nil }

type stubLabConn struct {
	fn stubLabQueryFunc
}

func (c *stubLabConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported in stub")
}
func (c *stubLabConn) Close() error              { return nil }
func (c *stubLabConn) Begin() (driver.Tx, error) { return nil, errors.New("tx not supported in stub") }

func (c *stubLabConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.fn(query, args)
	if err != nil {
		return nil, err
	}
	return &stubLabRows{cols: cols, rows: rows}, nil
}

type stubLabRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubLabRows) Columns() []string { return r.cols }
func (r *stubLabRows) Close() error      { return nil }
func (r *stubLabRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func newStubLabDB(t *testing.T, fn stubLabQueryFunc) *sql.DB {
	t.Helper()
	db := sql.OpenDB(&stubLabConnector{fn: fn})
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertDateFormatSQL(t *testing.T, query string, expectedExprs ...string) {
	t.Helper()
	if strings.Contains(query, "%%") || strings.Contains(query, "%!") {
		t.Fatalf("malformed format specifier in rendered SQL: %s", query)
	}
	for _, expr := range expectedExprs {
		if !strings.Contains(query, expr) {
			t.Fatalf("expected SQL to contain %q, got:\n%s", expr, query)
		}
	}
}

func TestRepository_DateFormattingOnLabQueries(t *testing.T) {
	ctx := context.Background()
	filter := FilterRiwayatLab{Page: 1, Limit: 10}

	t.Run("DaftarHasilLabPK & DetailHasilLabPK format pl.tgl_periksa", func(t *testing.T) {
		var capturedSelect string
		db := newStubLabDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "COUNT(*)") {
				return []string{"total"}, [][]driver.Value{{int64(1)}}, nil
			}
			capturedSelect = query
			return []string{
				"no_rawat", "kode_tindakan", "nama_tindakan", "kategori", "status",
				"tgl_periksa", "jam_periksa", "kd_dokter_perujuk", "nm_dokter_perujuk",
				"kd_dokter_pj", "nm_dokter_pj", "nip_petugas", "nama_petugas",
			}, nil, nil
		})
		repo := NewRepository(db)

		if _, _, err := repo.DaftarHasilLabPK(ctx, "2026/09/04/000001", shared.StatusLanjutRawatJalan, filter); err != nil {
			t.Fatalf("DaftarHasilLabPK unexpected error: %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, "COALESCE(DATE_FORMAT(pl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa")

		capturedSelect = ""
		_, err := repo.DetailHasilLabPK(ctx, "2026/09/04/000001", "J01", "2026-09-04", "10:00:00")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, "COALESCE(DATE_FORMAT(pl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa")
	})

	t.Run("DaftarHasilLabPA & DetailHasilLabPA format pl.tgl_periksa", func(t *testing.T) {
		var capturedSelect string
		db := newStubLabDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "COUNT(*)") {
				return []string{"total"}, [][]driver.Value{{int64(1)}}, nil
			}
			capturedSelect = query
			return []string{
				"no_rawat", "kode_tindakan", "nama_tindakan", "kategori", "status",
				"tgl_periksa", "jam_periksa", "kd_dokter_perujuk", "nm_dokter_perujuk",
				"kd_dokter_pj", "nm_dokter_pj", "nip_petugas", "nama_petugas",
			}, nil, nil
		})
		repo := NewRepository(db)

		if _, _, err := repo.DaftarHasilLabPA(ctx, "2026/09/04/000001", shared.StatusLanjutRawatJalan, filter); err != nil {
			t.Fatalf("DaftarHasilLabPA unexpected error: %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, "COALESCE(DATE_FORMAT(pl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa")

		capturedSelect = ""
		_, err := repo.DetailHasilLabPA(ctx, "2026/09/04/000001", "J01", "2026-09-04", "10:00:00")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, "COALESCE(DATE_FORMAT(pl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa")
	})

	t.Run("DaftarHasilLabMB & DetailHasilLabMB format pl.tgl_periksa", func(t *testing.T) {
		var capturedSelect string
		db := newStubLabDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "COUNT(*)") {
				return []string{"total"}, [][]driver.Value{{int64(1)}}, nil
			}
			capturedSelect = query
			return []string{
				"no_rawat", "kode_tindakan", "nama_tindakan", "kategori", "status",
				"tgl_periksa", "jam_periksa", "kd_dokter_perujuk", "nm_dokter_perujuk",
				"kd_dokter_pj", "nm_dokter_pj", "nip_petugas", "nama_petugas",
			}, nil, nil
		})
		repo := NewRepository(db)

		if _, _, err := repo.DaftarHasilLabMB(ctx, "2026/09/04/000001", shared.StatusLanjutRawatJalan, filter); err != nil {
			t.Fatalf("DaftarHasilLabMB unexpected error: %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, "COALESCE(DATE_FORMAT(pl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa")

		capturedSelect = ""
		_, err := repo.DetailHasilLabMB(ctx, "2026/09/04/000001", "J01", "2026-09-04", "10:00:00")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, "COALESCE(DATE_FORMAT(pl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa")
	})

	t.Run("Permintaan Lab PA queries format pengambilan_bahan & tanggal_pa_sebelumnya", func(t *testing.T) {
		var capturedSelect string
		cols := []string{
			"no_permintaan", "no_rawat", "tgl_permintaan", "jam_permintaan",
			"tgl_sampel", "jam_sampel", "tgl_hasil", "jam_hasil",
			"dokter_perujuk", "nama_dokter_perujuk", "status", "informasi_tambahan",
			"diagnosa_klinis", "pengambilan_bahan", "diperoleh_dengan", "lokasi_jaringan",
			"diawetkan_dengan", "pernah_dilakukan_di", "tanggal_pa_sebelumnya",
			"nomor_pa_sebelumnya", "diagnosa_pa_sebelumnya",
		}
		db := newStubLabDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "COUNT(") {
				return []string{"total"}, [][]driver.Value{{int64(1)}}, nil
			}
			capturedSelect = query
			return cols, nil, nil
		})
		repo := NewRepository(db)

		wantExprs := []string{
			"COALESCE(DATE_FORMAT(pl.pengambilan_bahan, '%Y-%m-%d'), '') AS pengambilan_bahan",
			"COALESCE(DATE_FORMAT(pl.tanggal_pa_sebelumnya, '%Y-%m-%d'), '') AS tanggal_pa_sebelumnya",
		}

		if _, err := repo.DaftarPermintaanLabPA(ctx, "2026/09/04/000001", shared.StatusLanjutRawatJalan); err != nil {
			t.Fatalf("DaftarPermintaanLabPA unexpected error: %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, wantExprs...)

		capturedSelect = ""
		if _, _, err := repo.DaftarPermintaanLabPAByRM(ctx, "000001", shared.StatusLanjutRawatJalan, filter); err != nil {
			t.Fatalf("DaftarPermintaanLabPAByRM unexpected error: %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, wantExprs...)

		capturedSelect = ""
		_, err := repo.DetailPermintaanLabPA(ctx, "PA202609040001")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
		assertDateFormatSQL(t, capturedSelect, wantExprs...)
	})
}

