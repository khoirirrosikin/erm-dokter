package rawatinap

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestBuildStatusFilter(t *testing.T) {
	t.Run("StatusPulang belum_pulang", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusPulang: StatusPulangBelumPulang,
		}
		cond, args := buildStatusFilter(filter, "", "")
		if cond != sqlKamarInapAktif || len(args) != 0 {
			t.Errorf("Expected sqlKamarInapAktif, got cond=%q args=%v", cond, args)
		}
	})

	t.Run("StatusPulang dengan rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusPulang: StatusPulangSembuh,
		}
		cond, args := buildStatusFilter(filter, "2026-09-01", "2026-09-09")
		if !strings.Contains(cond, "ki.stts_pulang = ? AND ki.tgl_keluar BETWEEN ? AND ?") {
			t.Errorf("Unexpected cond: %s", cond)
		}
		if len(args) != 3 || args[0] != "Sembuh" || args[1] != "2026-09-01" || args[2] != "2026-09-09" {
			t.Errorf("Unexpected args: %v", args)
		}
	})

	t.Run("StatusPulang tanpa rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusPulang: StatusPulangSembuh,
		}
		cond, args := buildStatusFilter(filter, "", "")
		if cond != "ki.stts_pulang = ?" {
			t.Errorf("Unexpected cond: %s", cond)
		}
		if len(args) != 1 || args[0] != "Sembuh" {
			t.Errorf("Unexpected args: %v", args)
		}
	})

	t.Run("StatusKunjungan belum_krs", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganBelumKRS,
		}
		cond, args := buildStatusFilter(filter, "", "")
		if cond != sqlKamarInapAktif || len(args) != 0 {
			t.Errorf("Expected sqlKamarInapAktif, got cond=%q args=%v", cond, args)
		}
	})

	t.Run("StatusKunjungan tanggal_mrs dengan rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganTanggalMRS,
		}
		cond, args := buildStatusFilter(filter, "2026-09-01", "2026-09-09")
		if cond != "ki.tgl_masuk BETWEEN ? AND ?" {
			t.Errorf("Unexpected cond: %s", cond)
		}
		if len(args) != 2 || args[0] != "2026-09-01" || args[1] != "2026-09-09" {
			t.Errorf("Unexpected args: %v", args)
		}
	})

	t.Run("StatusKunjungan tanggal_mrs tanpa rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganTanggalMRS,
		}
		cond, args := buildStatusFilter(filter, "", "")
		if cond != "" || len(args) != 0 {
			t.Errorf("Expected empty cond, got cond=%q args=%v", cond, args)
		}
	})

	t.Run("StatusKunjungan tanggal_krs dengan rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganTanggalKRS,
		}
		cond, args := buildStatusFilter(filter, "2026-09-01", "2026-09-09")
		if !strings.Contains(cond, "BETWEEN ? AND ?") {
			t.Errorf("Expected rentang tanggal in cond, got cond=%q", cond)
		}
		if len(args) != 2 || args[0] != "2026-09-01" || args[1] != "2026-09-09" {
			t.Errorf("Unexpected args: %v", args)
		}
	})

	t.Run("StatusKunjungan tanggal_krs tanpa rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganTanggalKRS,
		}
		cond, args := buildStatusFilter(filter, "", "")
		if cond != sqlKamarInapKRS || len(args) != 0 {
			t.Errorf("Expected sqlKamarInapKRS, got cond=%q args=%v", cond, args)
		}
	})
}

func TestBuildDokterFilter(t *testing.T) {
	t.Run("Scope DPJP Dokter Login", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			ScopeDPJP: ScopeDPJPDokter,
		}
		cond, args := buildDokterFilter("D001", filter)
		if !strings.Contains(cond, "NOT EXISTS (SELECT 1 FROM dpjp_ranap") {
			t.Errorf("Unexpected cond: %s", cond)
		}
		if len(args) != 2 || args[0] != "D001" || args[1] != "D001" {
			t.Errorf("Unexpected args: %v", args)
		}
	})

	t.Run("Scope DPJP Semua", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			ScopeDPJP: ScopeDPJPSemua,
		}
		cond, args := buildDokterFilter("D001", filter)
		if cond != "" || len(args) != 0 {
			t.Errorf("Expected empty cond, got cond=%q args=%v", cond, args)
		}
	})
}

func TestBuildOrderClause(t *testing.T) {
	tests := []struct {
		orderBy   OrderByRanap
		sortOrder string
		want      string
	}{
		{OrderByWaktuMRS, "ASC", "ki.tgl_masuk ASC, ki.jam_masuk ASC"},
		{OrderByWaktuMRS, "DESC", "ki.tgl_masuk DESC, ki.jam_masuk DESC"},
		{OrderByWaktuKRS, "ASC", "ki.tgl_keluar ASC, ki.jam_keluar ASC"},
		{OrderByNamaPasien, "ASC", "p.nm_pasien ASC"},
		{OrderByKamar, "DESC", "ki.kd_kamar DESC"},
		{OrderByKelas, "ASC", "k.kelas ASC"},
		{OrderByBangsal, "DESC", "b.nm_bangsal DESC"},
		{OrderByPenjamin, "ASC", "pj.png_jawab ASC"},
		{OrderByStatusPulang, "DESC", "ki.stts_pulang DESC"},
		{"invalid", "ASC", "ki.tgl_masuk ASC, ki.jam_masuk ASC"},
	}

	for _, tt := range tests {
		got := buildOrderClause(tt.orderBy, tt.sortOrder)
		if got != tt.want {
			t.Errorf("buildOrderClause(%s, %s) = %q, want %q", tt.orderBy, tt.sortOrder, got, tt.want)
		}
	}
}

func TestParseDPJPList(t *testing.T) {
	t.Run("Null string", func(t *testing.T) {
		res := parseDPJPList(sql.NullString{Valid: false})
		if len(res) != 0 {
			t.Errorf("Expected empty slice, got %v", res)
		}
	})

	t.Run("Empty string", func(t *testing.T) {
		res := parseDPJPList(sql.NullString{Valid: true, String: "  "})
		if len(res) != 0 {
			t.Errorf("Expected empty slice, got %v", res)
		}
	})

	t.Run("Multiple DPJP string", func(t *testing.T) {
		res := parseDPJPList(sql.NullString{Valid: true, String: "dr. Andi, Sp.PD;;;dr. Budi, Sp.JP"})
		if len(res) != 2 || res[0] != "dr. Andi, Sp.PD" || res[1] != "dr. Budi, Sp.JP" {
			t.Errorf("Unexpected result: %v", res)
		}
	})
}

func TestBuildFilterConditions(t *testing.T) {
	repo := &repository{}

	t.Run("Default ScopeDPJPSemua (Semua Pasien)", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganBelumKRS,
			Bangsal:         "B01",
			Kelas:           "Kelas 1",
			Penjamin:        "BPJ",
			Keyword:         "Siti",
		}
		filter.Sanitize()

		conditions, args := repo.buildFilterConditions("D001", filter)
		if len(conditions) != 5 {
			t.Fatalf("Expected 5 conditions, got %d: %v", len(conditions), conditions)
		}

		if len(args) != 8 {
			t.Fatalf("Expected 8 args, got %d: %v", len(args), args)
		}
	})

	t.Run("Eksplisit ScopeDPJPDokter", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganBelumKRS,
			Bangsal:         "B01",
			Kelas:           "Kelas 1",
			Penjamin:        "BPJ",
			Keyword:         "Siti",
			ScopeDPJP:       ScopeDPJPDokter,
		}
		filter.Sanitize()

		conditions, args := repo.buildFilterConditions("D001", filter)
		if len(conditions) != 6 {
			t.Fatalf("Expected 6 conditions, got %d: %v", len(conditions), conditions)
		}

		if len(args) != 10 {
			t.Fatalf("Expected 10 args, got %d: %v", len(args), args)
		}
	})
}

type stubQueryFunc func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)

type stubConnector struct {
	fn stubQueryFunc
}

func (c *stubConnector) Connect(context.Context) (driver.Conn, error) {
	return &stubConn{fn: c.fn}, nil
}

func (c *stubConnector) Driver() driver.Driver { return nil }

type stubConn struct {
	fn stubQueryFunc
}

func (c *stubConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported in stub")
}
func (c *stubConn) Close() error              { return nil }
func (c *stubConn) Begin() (driver.Tx, error) { return nil, errors.New("tx not supported in stub") }

func (c *stubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.fn(query, args)
	if err != nil {
		return nil, err
	}
	return &stubRows{cols: cols, rows: rows}, nil
}

type stubRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubRows) Columns() []string { return r.cols }
func (r *stubRows) Close() error      { return nil }
func (r *stubRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func newStubDB(t *testing.T, fn stubQueryFunc) *sql.DB {
	t.Helper()
	db := sql.OpenDB(&stubConnector{fn: fn})
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestRepository_CekStatusKamarInap(t *testing.T) {
	t.Run("Belum pernah rawat inap (has_record=0, is_aktif=0)", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if !strings.Contains(query, "DATE_FORMAT(tgl_keluar, '%Y-%m-%d')") {
				t.Errorf("expected query to use DATE_FORMAT on tgl_keluar, got: %s", query)
			}
			if len(args) != 1 || args[0].Value != "2026/09/04/000001" {
				t.Errorf("unexpected args: %+v", args)
			}
			return []string{"has_record", "is_aktif"}, [][]driver.Value{{int64(0), int64(0)}}, nil
		})

		repo := NewRepository(db)
		isAktif, hasRecord, err := repo.CekStatusKamarInap(context.Background(), "2026/09/04/000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isAktif || hasRecord {
			t.Fatalf("expected (false, false), got (%v, %v)", isAktif, hasRecord)
		}
	})

	t.Run("Pasien aktif dirawat (termasuk pasca Pindah Kamar: has_record=1, is_aktif=1)", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			return []string{"has_record", "is_aktif"}, [][]driver.Value{{[]byte("1"), []byte("1")}}, nil
		})

		repo := NewRepository(db)
		isAktif, hasRecord, err := repo.CekStatusKamarInap(context.Background(), "2026/09/04/000002")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isAktif || !hasRecord {
			t.Fatalf("expected (true, true), got (%v, %v)", isAktif, hasRecord)
		}
	})

	t.Run("Pasien sudah KRS (has_record=1, is_aktif=0)", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			return []string{"has_record", "is_aktif"}, [][]driver.Value{{int64(1), int64(0)}}, nil
		})

		repo := NewRepository(db)
		isAktif, hasRecord, err := repo.CekStatusKamarInap(context.Background(), "2026/09/04/000003")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isAktif || !hasRecord {
			t.Fatalf("expected (false, true), got (%v, %v)", isAktif, hasRecord)
		}
	})

	t.Run("Database error diteruskan", func(t *testing.T) {
		wantErr := errors.New("connection lost")
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			return nil, nil, wantErr
		})

		repo := NewRepository(db)
		_, _, err := repo.CekStatusKamarInap(context.Background(), "2026/09/04/000004")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
	})
}

func TestRepository_GetKelasRawat(t *testing.T) {
	t.Run("Memprioritaskan kamar aktif dan trim spasi", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if !strings.Contains(query, "(COALESCE(ki.stts_pulang, '-') = '-') DESC") {
				t.Errorf("expected active room tie-breaker in ORDER BY, got: %s", query)
			}
			return []string{"kelas"}, [][]driver.Value{{"  Kelas 1  "}}, nil
		})

		repo := NewRepository(db)
		kelas, err := repo.GetKelasRawat(context.Background(), "2026/09/04/000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kelas != "Kelas 1" {
			t.Fatalf("expected 'Kelas 1', got %q", kelas)
		}
	})

	t.Run("Tidak ada record kamar mengembalikan string kosong tanpa error", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			return []string{"kelas"}, nil, nil
		})

		repo := NewRepository(db)
		kelas, err := repo.GetKelasRawat(context.Background(), "2026/09/04/000099")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kelas != "" {
			t.Fatalf("expected empty string, got %q", kelas)
		}
	})
}
