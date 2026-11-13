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

	t.Run("StatusKunjungan tanggal_mrs dengan StatusPulang Sembuh dan rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganTanggalMRS,
			StatusPulang:    StatusPulangSembuh,
		}
		cond, args := buildStatusFilter(filter, "2026-09-01", "2026-09-09")
		if cond != "ki.stts_pulang = ? AND ki.tgl_masuk BETWEEN ? AND ?" {
			t.Errorf("Unexpected cond: %s", cond)
		}
		if len(args) != 3 || args[0] != "Sembuh" || args[1] != "2026-09-01" || args[2] != "2026-09-09" {
			t.Errorf("Unexpected args: %v", args)
		}
	})

	t.Run("StatusKunjungan tanggal_mrs dengan StatusPulang belum_pulang dan rentang tanggal", func(t *testing.T) {
		filter := FilterPasienRawatInap{
			StatusKunjungan: StatusKunjunganTanggalMRS,
			StatusPulang:    StatusPulangBelumPulang,
		}
		cond, args := buildStatusFilter(filter, "2026-09-01", "2026-09-09")
		wantCond := sqlKamarInapAktif + " AND ki.tgl_masuk BETWEEN ? AND ?"
		if cond != wantCond {
			t.Errorf("Expected %q, got %q", wantCond, cond)
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
		{OrderByWaktuMRS, "ASC", "ki.tgl_masuk ASC, ki.jam_masuk ASC, ki.no_rawat ASC"},
		{OrderByWaktuMRS, "DESC", "ki.tgl_masuk DESC, ki.jam_masuk DESC, ki.no_rawat DESC"},
		{OrderByWaktuKRS, "ASC", "ki.tgl_keluar ASC, ki.jam_keluar ASC, ki.tgl_masuk ASC, ki.jam_masuk ASC, ki.no_rawat ASC"},
		{OrderByNamaPasien, "ASC", "p.nm_pasien ASC, ki.tgl_masuk ASC, ki.jam_masuk ASC, ki.no_rawat ASC"},
		{OrderByKamar, "DESC", "ki.kd_kamar DESC, ki.tgl_masuk DESC, ki.jam_masuk DESC, ki.no_rawat DESC"},
		{OrderByKelas, "ASC", "k.kelas ASC, ki.tgl_masuk ASC, ki.jam_masuk ASC, ki.no_rawat ASC"},
		{OrderByBangsal, "DESC", "b.nm_bangsal DESC, ki.tgl_masuk DESC, ki.jam_masuk DESC, ki.no_rawat DESC"},
		{OrderByPenjamin, "ASC", "pj.png_jawab ASC, ki.tgl_masuk ASC, ki.jam_masuk ASC, ki.no_rawat ASC"},
		{OrderByStatusPulang, "DESC", "ki.stts_pulang DESC, ki.tgl_masuk DESC, ki.jam_masuk DESC, ki.no_rawat DESC"},
		{"invalid", "ASC", "ki.tgl_masuk ASC, ki.jam_masuk ASC, ki.no_rawat ASC"},
	}

	for _, tt := range tests {
		got := buildOrderClause(tt.orderBy, tt.sortOrder)
		if got != tt.want {
			t.Errorf("buildOrderClause(%s, %s) = %q, want %q", tt.orderBy, tt.sortOrder, got, tt.want)
		}
	}
}

func TestBuildJoinHelpers(t *testing.T) {
	t.Run("Default belum_krs + semua + waktu_mrs menghasilkan 0 join pada page_keys dan count", func(t *testing.T) {
		filter := FilterPasienRawatInap{}
		filter.Sanitize()

		if got := buildFilterJoins("D001", filter); got != "" {
			t.Errorf("expected empty filterJoins, got %q", got)
		}
		if got := buildPageKeyJoins("D001", filter); got != "" {
			t.Errorf("expected empty pageKeyJoins, got %q", got)
		}
	})

	t.Run("OrderByPenjamin menyertakan reg_periksa dan penjab hanya di page_keys", func(t *testing.T) {
		filter := FilterPasienRawatInap{OrderBy: OrderByPenjamin}
		filter.Sanitize()

		if got := buildFilterJoins("D001", filter); got != "" {
			t.Errorf("expected empty filterJoins, got %q", got)
		}
		pageJoins := buildPageKeyJoins("D001", filter)
		if !strings.Contains(pageJoins, "LEFT JOIN reg_periksa rp") || !strings.Contains(pageJoins, "LEFT JOIN penjab pj") {
			t.Errorf("expected reg_periksa and penjab joins, got %q", pageJoins)
		}
	})

	t.Run("Filter Bangsal menyertakan kamar dan bangsal", func(t *testing.T) {
		filter := FilterPasienRawatInap{Bangsal: "B01"}
		filter.Sanitize()

		filterJoins := buildFilterJoins("D001", filter)
		if !strings.Contains(filterJoins, "LEFT JOIN kamar k") || !strings.Contains(filterJoins, "LEFT JOIN bangsal b") {
			t.Errorf("expected kamar and bangsal joins, got %q", filterJoins)
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
			if strings.Contains(query, "DATE_FORMAT") {
				t.Errorf("expected sargable query without DATE_FORMAT, got: %s", query)
			}
			if !strings.Contains(query, "tgl_keluar < '1000-01-01'") {
				t.Errorf("expected zero-date safe < '1000-01-01' check, got: %s", query)
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

func TestRepository_GetWaktuKeluarTerakhir(t *testing.T) {
	t.Run("Mengembalikan tanggal dan jam keluar non-Pindah Kamar terbaru", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if !strings.Contains(query, "ki.tgl_keluar >= '1000-01-01'") {
				t.Errorf("expected sargable >= '1000-01-01' in WHERE, got: %s", query)
			}
			if !strings.Contains(query, "(COALESCE(ki.stts_pulang, '-') <> 'Pindah Kamar') DESC") {
				t.Errorf("expected non-Pindah Kamar priority in ORDER BY, got: %s", query)
			}
			return []string{"tgl_keluar", "jam_keluar"}, [][]driver.Value{{"2026-11-12", "14:30:00"}}, nil
		})

		repo := NewRepository(db)
		tgl, jam, err := repo.GetWaktuKeluarTerakhir(context.Background(), "2026/11/10/000001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tgl != "2026-11-12" || jam != "14:30:00" {
			t.Fatalf("expected (2026-11-12, 14:30:00), got (%q, %q)", tgl, jam)
		}
	})

	t.Run("Tidak ada baris tgl_keluar valid mengembalikan string kosong", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			return []string{"tgl_keluar", "jam_keluar"}, nil, nil
		})

		repo := NewRepository(db)
		tgl, jam, err := repo.GetWaktuKeluarTerakhir(context.Background(), "2026/11/10/000099")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tgl != "" || jam != "" {
			t.Fatalf("expected empty strings, got (%q, %q)", tgl, jam)
		}
	})
}

func stubRanapColumns() []string {
	return []string{
		"no_rawat", "no_registrasi", "tanggal_registrasi", "jam_registrasi",
		"no_rkm_medis", "nama_pasien", "jenis_kelamin", "tanggal_lahir", "alamat",
		"kode_kamar", "kelas_kamar", "kode_bangsal", "nama_bangsal",
		"tanggal_masuk", "jam_masuk", "tanggal_keluar", "jam_keluar",
		"status_pulang", "lama_inap", "kode_penjamin", "nama_penjamin",
		"status_bayar", "kode_dokter_rawat_jalan", "dokter_rawat_jalan",
		"diagnosa_awal", "diagnosa_akhir", "golongan_darah", "agama",
		"no_telepon", "no_peserta", "no_ktp", "penanggung_jawab",
		"hubungan_penanggung_jawab", "alamat_penanggung_jawab",
	}
}

func stubRanapRow(noRawat string, jk string, lama int64) []driver.Value {
	return []driver.Value{
		noRawat, "001", "2026-06-01", "21:12:39",
		"235338", "MUHAMMAD ELLIF", jk, "2006-02-19", "DUSUN II",
		"EDELWEIS-A-1-B", "Kelas 1", "ri42", "EDELWEIS A",
		"2026-06-02", "11:49:02", nil, nil,
		"-", lama, "BPJ", "BPJS",
		"Belum Bayar", "D001", "dr. Umum",
		"Obs Febris", nil, "O", "ISLAM",
		"08123456789", "000123", "350001", "AYAH",
		"AYAH", "DUSUN II",
	}
}

func TestRepository_DaftarPasienRawatInap(t *testing.T) {
	t.Run("Halaman 1 dengan hasil < limit melewati query COUNT dan batch-hydrate DPJP terurut", func(t *testing.T) {
		var executedQueries []string
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			executedQueries = append(executedQueries, query)
			switch {
			case strings.Contains(query, "AS page_keys"):
				if strings.Contains(query, "GROUP_CONCAT") {
					t.Errorf("unexpected correlated GROUP_CONCAT in main query: %s", query)
				}
				return stubRanapColumns(), [][]driver.Value{
					stubRanapRow("2026/06/01/050495", "L", 3),
					stubRanapRow("2026/06/01/050496", "P", 0),
				}, nil
			case strings.Contains(query, "FROM dpjp_ranap dr"):
				return []string{"no_rawat", "nm_dokter"}, [][]driver.Value{
					{"2026/06/01/050495", "dr. Budi, Sp.PD"},
					{"2026/06/01/050495", "dr. Andi, Sp.JP"},
					{"2026/06/01/050495", "dr. Budi, Sp.PD "},
				}, nil
			default:
				t.Fatalf("unexpected extra query (COUNT should be skipped!): %s", query)
				return nil, nil, nil
			}
		})

		repo := NewRepository(db)
		filter := FilterPasienRawatInap{Page: 1, Limit: 100}
		filter.Sanitize()

		list, total, err := repo.DaftarPasienRawatInap(context.Background(), "D001", filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(executedQueries) != 2 {
			t.Fatalf("expected 2 queries (page_keys + dpjp_ranap), got %d", len(executedQueries))
		}
		if total != 2 || len(list) != 2 {
			t.Fatalf("expected total=2 len=2, got total=%d len=%d", total, len(list))
		}
		if list[0].JenisKelamin != "L" {
			t.Errorf("expected jenis_kelamin='L' per API contract, got %q", list[0].JenisKelamin)
		}
		if list[0].LamaInap != 3 {
			t.Errorf("expected lama_inap=3, got %d", list[0].LamaInap)
		}
		if len(list[0].DPJP) != 2 || list[0].DPJP[0] != "dr. Andi, Sp.JP" || list[0].DPJP[1] != "dr. Budi, Sp.PD" {
			t.Errorf("expected sorted deduplicated DPJP, got %v", list[0].DPJP)
		}
		if list[1].DPJP == nil || len(list[1].DPJP) != 0 {
			t.Errorf("expected non-nil empty DPJP slice for row without dpjp_ranap, got %v", list[1].DPJP)
		}
	})

	t.Run("Halaman penuh (len == limit) menjalankan COUNT minimal-join", func(t *testing.T) {
		var countCalled bool
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			switch {
			case strings.Contains(query, "AS page_keys"):
				return stubRanapColumns(), [][]driver.Value{
					stubRanapRow("2026/06/01/050495", "L", 1),
				}, nil
			case strings.Contains(query, "FROM dpjp_ranap dr"):
				return []string{"no_rawat", "nm_dokter"}, nil, nil
			case strings.Contains(query, "SELECT COUNT(*)"):
				countCalled = true
				if strings.Contains(query, "JOIN") {
					t.Errorf("expected 0 joins on default belum_krs COUNT query, got: %s", query)
				}
				return []string{"count"}, [][]driver.Value{{int64(42)}}, nil
			default:
				t.Fatalf("unexpected query: %s", query)
				return nil, nil, nil
			}
		})

		repo := NewRepository(db)
		filter := FilterPasienRawatInap{Page: 1, Limit: 1}
		filter.Sanitize()

		list, total, err := repo.DaftarPasienRawatInap(context.Background(), "D001", filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !countCalled {
			t.Fatalf("expected COUNT query when len(list) == limit")
		}
		if total != 42 || len(list) != 1 {
			t.Fatalf("expected total=42 len=1, got total=%d len=%d", total, len(list))
		}
	})
}

func TestRepository_DetailPasienRawatInap(t *testing.T) {
	t.Run("Data ditemukan menghidrasi DPJP", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "FROM dpjp_ranap dr") {
				return []string{"no_rawat", "nm_dokter"}, [][]driver.Value{
					{"2026/06/01/050495", "dr. Citra, Sp.A"},
				}, nil
			}
			return stubRanapColumns(), [][]driver.Value{
				stubRanapRow("2026/06/01/050495", "L", 2),
			}, nil
		})

		repo := NewRepository(db)
		detail, err := repo.DetailPasienRawatInap(context.Background(), "2026/06/01/050495", "2026-06-02", "11:49:02")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if detail == nil || len(detail.DPJP) != 1 || detail.DPJP[0] != "dr. Citra, Sp.A" {
			t.Fatalf("unexpected detail: %+v", detail)
		}
	})

	t.Run("Data tidak ditemukan mengembalikan (nil, nil)", func(t *testing.T) {
		db := newStubDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			return stubRanapColumns(), nil, nil
		})

		repo := NewRepository(db)
		detail, err := repo.DetailPasienRawatInap(context.Background(), "2026/06/01/999999", "2026-06-02", "11:49:02")
		if err != nil || detail != nil {
			t.Fatalf("expected (nil, nil), got (%+v, %v)", detail, err)
		}
	})
}
