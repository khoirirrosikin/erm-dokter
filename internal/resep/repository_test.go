package resep

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

