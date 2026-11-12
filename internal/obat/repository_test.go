package obat

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type stubQueryCall struct {
	query string
	args  []driver.NamedValue
}

type stubQueryHandler func(query string, args []driver.NamedValue) (driver.Rows, error)

type stubDriver struct {
	mu      sync.Mutex
	calls   []stubQueryCall
	handler stubQueryHandler
}

func (d *stubDriver) record(query string, args []driver.NamedValue) {
	copied := make([]driver.NamedValue, len(args))
	copy(copied, args)
	d.mu.Lock()
	d.calls = append(d.calls, stubQueryCall{query: query, args: copied})
	d.mu.Unlock()
}

func (d *stubDriver) getCalls() []stubQueryCall {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]stubQueryCall, len(d.calls))
	copy(out, d.calls)
	return out
}

func (d *stubDriver) Open(_ string) (driver.Conn, error) {
	return &stubConn{drv: d}, nil
}

type stubConn struct {
	drv *stubDriver
}

func (c *stubConn) Prepare(_ string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *stubConn) Close() error                          { return nil }
func (c *stubConn) Begin() (driver.Tx, error)             { return nil, driver.ErrSkip }

func (c *stubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.drv.record(query, args)
	return c.drv.handler(query, args)
}

type stubRows struct {
	cols []string
	data [][]driver.Value
	idx  int
}

func (r *stubRows) Columns() []string { return r.cols }
func (r *stubRows) Close() error      { return nil }
func (r *stubRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.idx])
	r.idx++
	return nil
}

var obatDriverSeq atomic.Uint64

func openStubObatDB(t *testing.T, handler stubQueryHandler) (*sql.DB, *stubDriver) {
	t.Helper()
	drv := &stubDriver{handler: handler}
	name := fmt.Sprintf("obat_stub_%d", obatDriverSeq.Add(1))
	sql.Register(name, drv)
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("failed to open stub db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, drv
}

var obatSelectColumns = []string{
	"kode_obat", "nama_obat", "komposisi", "harga", "satuan",
	"stok", "kapasitas", "kode_depo", "nama_depo",
	"kode_jenis", "nama_jenis", "kode_golongan", "nama_golongan",
	"kode_kategori", "nama_kategori",
}

func sampleObatRow(kode, nama string, stok float64, kodeDepo, namaDepo string) []driver.Value {
	return []driver.Value{
		kode, nama, "Paracetamol 500mg", "2500", "TAB",
		stok, float64(500), kodeDepo, namaDepo,
		"J01", "Tablet", "G01", "Obat Bebas", "K01", "Generik",
	}
}

func TestBuildOrderClause(t *testing.T) {
	tests := []struct {
		name      string
		orderBy   string
		sortOrder string
		hasDepo   bool
		want      string
	}{
		{
			name:      "nama_depo without depo falls back to dtb.nama_brng",
			orderBy:   "nama_depo",
			sortOrder: "DESC",
			hasDepo:   false,
			want:      "dtb.nama_brng DESC, dtb.kode_brng ASC",
		},
		{
			name:      "nama_depo with depo uses bg.nm_bangsal",
			orderBy:   "nama_depo",
			sortOrder: "ASC",
			hasDepo:   true,
			want:      "bg.nm_bangsal ASC, dtb.nama_brng ASC, dtb.kode_brng ASC",
		},
		{
			name:      "stok without depo uses alias stok",
			orderBy:   "stok",
			sortOrder: "DESC",
			hasDepo:   false,
			want:      "stok DESC, dtb.nama_brng ASC, dtb.kode_brng ASC",
		},
		{
			name:      "stok with depo uses gdb.stok",
			orderBy:   "stok",
			sortOrder: "ASC",
			hasDepo:   true,
			want:      "gdb.stok ASC, dtb.nama_brng ASC, dtb.kode_brng ASC",
		},
		{
			name:      "unknown column falls back to dtb.nama_brng",
			orderBy:   "unknown",
			sortOrder: "ASC",
			hasDepo:   false,
			want:      "dtb.nama_brng ASC, dtb.kode_brng ASC",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildOrderClause(tc.orderBy, tc.sortOrder, tc.hasDepo)
			if got != tc.want {
				t.Errorf("buildOrderClause(%q, %q, %v) = %q, want %q", tc.orderBy, tc.sortOrder, tc.hasDepo, got, tc.want)
			}
		})
	}
}

func TestDaftarObat_EmptyDepo_OrderByNamaDepo_NoUnknownColumn(t *testing.T) {
	db, drv := openStubObatDB(t, func(query string, _ []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.HasPrefix(query, "SELECT COUNT(*)"):
			return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(1)}}}, nil
		case strings.Contains(query, "FROM gudangbarang") && strings.Contains(query, "GROUP BY kode_brng"):
			return &stubRows{
				cols: []string{"kode_brng", "total_stok"},
				data: [][]driver.Value{{"OBT001", float64(42)}},
			}, nil
		default:
			return &stubRows{
				cols: obatSelectColumns,
				data: [][]driver.Value{sampleObatRow("OBT001", "Paracetamol", 0, "", "Semua Depo")},
			}, nil
		}
	})

	repo := NewRepository(db)
	filter := FilterDaftarObat{
		OrderBy:   "nama_depo",
		SortOrder: "DESC",
		Page:      1,
		Limit:     20,
	}
	filter.Sanitize()

	items, total, err := repo.DaftarObat(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item (total=1), got len=%d total=%d", len(items), total)
	}
	if items[0].Stok != 42 {
		t.Errorf("expected hydrated stok 42, got %v", items[0].Stok)
	}

	calls := drv.getCalls()
	var foundSelect, foundHydrate bool
	for _, c := range calls {
		if strings.Contains(c.query, "AS kode_obat") {
			foundSelect = true
			if strings.Contains(c.query, "bg.nm_bangsal") {
				t.Errorf("select query must not reference bg.nm_bangsal when depo is empty: %s", c.query)
			}
			if !strings.Contains(c.query, "ORDER BY dtb.nama_brng DESC, dtb.kode_brng ASC") {
				t.Errorf("expected fallback ORDER BY dtb.nama_brng DESC, got: %s", c.query)
			}
			if !strings.Contains(c.query, "0 AS stok") {
				t.Errorf("expected lightweight 0 AS stok when order_by != stok, got: %s", c.query)
			}
		}
		if strings.Contains(c.query, "GROUP BY kode_brng") {
			foundHydrate = true
		}
	}
	if !foundSelect || !foundHydrate {
		t.Errorf("expected both select and hydrateStok queries, got foundSelect=%v foundHydrate=%v", foundSelect, foundHydrate)
	}
}

func TestDaftarObat_EmptyDepo_OrderByStok_SortsGloballyAndSkipsHydrate(t *testing.T) {
	db, drv := openStubObatDB(t, func(query string, _ []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.HasPrefix(query, "SELECT COUNT(*)"):
			return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(2)}}}, nil
		case strings.Contains(query, "GROUP BY kode_brng"):
			t.Errorf("hydrateStok should not be called when order_by=stok: %s", query)
			return &stubRows{cols: []string{"kode_brng", "total_stok"}}, nil
		default:
			return &stubRows{
				cols: obatSelectColumns,
				data: [][]driver.Value{
					sampleObatRow("OBT002", "Cetirizine", 150.5, "", "Semua Depo"),
					sampleObatRow("OBT001", "Amoxicillin", 80, "", "Semua Depo"),
				},
			}, nil
		}
	})

	repo := NewRepository(db)
	filter := FilterDaftarObat{
		OrderBy:   "  STOK ",
		SortOrder: "desc",
		Page:      1,
		Limit:     20,
	}

	items, total, err := repo.DaftarObat(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 items (total=2), got len=%d total=%d", len(items), total)
	}
	if items[0].Stok != 150.5 || items[1].Stok != 80 {
		t.Errorf("unexpected stok values: got %v and %v", items[0].Stok, items[1].Stok)
	}

	calls := drv.getCalls()
	if len(calls) != 2 {
		t.Fatalf("expected exactly 2 queries (COUNT + SELECT, without hydrateStok), got %d", len(calls))
	}

	var foundSelect bool
	for _, c := range calls {
		if strings.Contains(c.query, "AS kode_obat") {
			foundSelect = true
			if !strings.Contains(c.query, "COALESCE((SELECT SUM(gdb.stok) FROM gudangbarang gdb WHERE gdb.kode_brng = dtb.kode_brng") {
				t.Errorf("expected correlated PK stok subquery in SELECT, got: %s", c.query)
			}
			if !strings.Contains(c.query, "ORDER BY stok DESC, dtb.nama_brng ASC, dtb.kode_brng ASC") {
				t.Errorf("expected ORDER BY stok DESC with tie-breakers, got: %s", c.query)
			}
		}
	}
	if !foundSelect {
		t.Error("expected main SELECT query to be recorded")
	}
}

func TestDaftarObat_WithDepo_UsesJoinedColumns(t *testing.T) {
	db, drv := openStubObatDB(t, func(query string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.HasPrefix(query, "SELECT COUNT(*)") {
			return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(1)}}}, nil
		}
		return &stubRows{
			cols: obatSelectColumns,
			data: [][]driver.Value{sampleObatRow("OBT001", "Paracetamol", 25, "DPRJ", "Depo Rawat Jalan")},
		}, nil
	})

	repo := NewRepository(db)
	filter := FilterDaftarObat{
		Depo:      "DPRJ",
		OrderBy:   "nama_depo",
		SortOrder: "DROP TABLE",
		Page:      1,
		Limit:     10,
	}

	items, total, err := repo.DaftarObat(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].KodeDepo != "DPRJ" {
		t.Fatalf("unexpected result: total=%d items=%+v", total, items)
	}

	calls := drv.getCalls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 queries when depo is specified, got %d", len(calls))
	}
	for _, c := range calls {
		if strings.Contains(c.query, "AS kode_obat") {
			if !strings.Contains(c.query, "ORDER BY bg.nm_bangsal ASC, dtb.nama_brng ASC, dtb.kode_brng ASC") {
				t.Errorf("expected bg.nm_bangsal ASC with tie-breakers when depo is set, got: %s", c.query)
			}
		}
	}
}

