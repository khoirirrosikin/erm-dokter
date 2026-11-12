package rawatjalan

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

var rjDriverSeq atomic.Uint64

func openStubRawatJalanDB(t *testing.T, handler stubQueryHandler) (*sql.DB, *stubDriver) {
	t.Helper()
	drv := &stubDriver{handler: handler}
	name := fmt.Sprintf("rawatjalan_stub_%d", rjDriverSeq.Add(1))
	sql.Register(name, drv)
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("failed to open stub db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, drv
}

var kunjunganColumns = []string{
	"no_rawat", "no_registrasi", "tanggal_registrasi", "jam_registrasi",
	"no_rekam_medis", "nama_pasien", "jenis_kelamin", "tanggal_lahir", "alamat",
	"kode_poli_asal", "nama_poli_asal", "kode_dokter_asal", "nama_dokter_asal",
	"kode_poli_rujukan", "nama_poli_rujukan", "kode_dokter_rujukan", "nama_dokter_rujukan",
	"kode_penjamin", "nama_penjamin", "status_pemeriksaan", "status_lanjut", "status_bayar",
	"jenis_antrean", "golongan_darah", "agama", "no_telepon", "no_peserta", "no_ktp",
	"penanggung_jawab", "hubungan_penanggung_jawab", "alamat_penanggung_jawab",
}

func sampleKunjunganRow(
	noRawat, tgl, jam, namaPasien, namaPenjamin, jenisAntrean, kdPoliRujukan, kdDokterRujukan string,
) []driver.Value {
	return []driver.Value{
		noRawat, "001", tgl, jam,
		"000123", namaPasien, "L", "1990-05-10", "Jl. Mawar 1",
		"INT", "Poli Dalam", "DR001", "dr. Budi",
		kdPoliRujukan, "Poli Paru", kdDokterRujukan, "dr. Andy, Sp.P",
		"BPJ", namaPenjamin, "Belum", "Ralan", "Sudah Bayar",
		jenisAntrean, "O", "ISLAM", "08123456789", "0001", "357901",
		"Keluarga", "SAUDARA", "Jl. Mawar 1",
	}
}

func TestBuildNoRawatBounds(t *testing.T) {
	lower, upper, ok := buildNoRawatBounds("2026-04-22", "2026-04-22")
	if !ok {
		t.Fatal("expected valid bounds")
	}
	if lower != "2026/01/22/" {
		t.Errorf("expected lower bound 2026/01/22/, got %q", lower)
	}
	if upper != "2026/07/23/" {
		t.Errorf("expected upper bound 2026/07/23/, got %q", upper)
	}
}

func TestDaftarAntreanDokter_ServiceKey_MergesRujukanAndAppliesPKWindow(t *testing.T) {
	db, drv := openStubRawatJalanDB(t, func(query string, _ []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.HasPrefix(query, "SELECT COUNT(*) FROM reg_periksa r"):
			return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(1)}}}, nil
		case strings.HasPrefix(query, "SELECT COUNT(*) FROM rujukan_internal_poli rip"):
			return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(1)}}}, nil
		case strings.Contains(query, "ON page_r.no_rawat = r.no_rawat"):
			return &stubRows{
				cols: kunjunganColumns,
				data: [][]driver.Value{
					sampleKunjunganRow("2026/04/22/000002", "2026-04-22", "08:30:00", "Siti", "BPJS", "Bukan Rujukan", "", ""),
				},
			}, nil
		case strings.Contains(query, "ON page_rip.no_rawat = rip.no_rawat"):
			return &stubRows{
				cols: kunjunganColumns,
				data: [][]driver.Value{
					sampleKunjunganRow("2026/04/22/000001", "2026-04-22", "08:00:00", "Budi", "BPJS", "Rujukan", "PAR", "DRANDY"),
				},
			}, nil
		default:
			t.Errorf("unexpected query: %s", query)
			return &stubRows{cols: kunjunganColumns}, nil
		}
	})

	repo := NewRepository(db)
	filter := FilterAntreanDokter{
		Tanggal:   "2026-04-22,2026-04-22",
		OrderBy:   "waktu_registrasi",
		SortOrder: "ASC",
		Page:      1,
		Limit:     20,
	}

	items, total, err := repo.DaftarAntreanDokter(context.Background(), "", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected total=2 len=2, got total=%d len=%d", total, len(items))
	}
	if items[0].JenisAntrean != JenisAntreanRujukan || items[0].KodeDokterRujukan != "DRANDY" {
		t.Errorf("expected first item to be Rujukan (08:00:00), got %+v", items[0])
	}
	if items[1].JenisAntrean != JenisAntreanTidakRujukan {
		t.Errorf("expected second item to be Bukan Rujukan (08:30:00), got %+v", items[1])
	}

	for _, c := range drv.getCalls() {
		if !strings.Contains(c.query, "CONVERT(? USING latin1)") {
			t.Errorf("expected PK no_rawat window with CONVERT(? USING latin1) in service mode, got: %s", c.query)
		}
		if len(c.args) < 2 || c.args[0].Value != "2026/01/22/" || c.args[1].Value != "2026/07/23/" {
			t.Errorf("expected ±3 months bounds in args, got: %+v", c.args)
		}
	}
}

func TestDaftarAntreanDokter_ServiceKey_OrderByPenjaminAndNamaPasien_JoinsRequiredTables(t *testing.T) {
	for _, orderCol := range []struct {
		orderBy      string
		expectedJoin string
		expectedSort string
	}{
		{
			orderBy:      "penjamin",
			expectedJoin: "INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj",
			expectedSort: "ORDER BY pj.png_jawab DESC",
		},
		{
			orderBy:      "nama_pasien",
			expectedJoin: "INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis",
			expectedSort: "ORDER BY p.nm_pasien DESC",
		},
	} {
		t.Run(orderCol.orderBy, func(t *testing.T) {
			db, drv := openStubRawatJalanDB(t, func(query string, _ []driver.NamedValue) (driver.Rows, error) {
				switch {
				case strings.HasPrefix(query, "SELECT COUNT(*) FROM reg_periksa r"):
					return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(1)}}}, nil
				case strings.HasPrefix(query, "SELECT COUNT(*) FROM rujukan_internal_poli rip"):
					return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(0)}}}, nil
				default:
					return &stubRows{
						cols: kunjunganColumns,
						data: [][]driver.Value{
							sampleKunjunganRow("2026/04/22/000001", "2026-04-22", "08:00:00", "Andi", "UMUM", "Bukan Rujukan", "", ""),
						},
					}, nil
				}
			})

			repo := NewRepository(db)
			filter := FilterAntreanDokter{
				Tanggal:   "2026-04-22,2026-04-22",
				OrderBy:   orderCol.orderBy,
				SortOrder: "DESC",
				Page:      1,
				Limit:     20,
			}

			items, total, err := repo.DaftarAntreanDokter(context.Background(), "", filter)
			if err != nil {
				t.Fatalf("unexpected error for order_by=%s: %v", orderCol.orderBy, err)
			}
			if total != 1 || len(items) != 1 {
				t.Fatalf("expected 1 item, got total=%d len=%d", total, len(items))
			}

			var checkedSubquery bool
			for _, c := range drv.getCalls() {
				if strings.Contains(c.query, ") AS page_r") {
					checkedSubquery = true
					subPart := strings.Split(c.query, ") AS page_r")[0]
					if !strings.Contains(subPart, orderCol.expectedJoin) {
						t.Errorf("inner subquery missing %q for order_by=%s: %s", orderCol.expectedJoin, orderCol.orderBy, subPart)
					}
					if !strings.Contains(subPart, orderCol.expectedSort) {
						t.Errorf("inner subquery missing %q: %s", orderCol.expectedSort, subPart)
					}
				}
			}
			if !checkedSubquery {
				t.Error("expected deferred join page_r query to be executed")
			}
		})
	}
}

func TestDaftarAntreanDokter_DoctorLogin_DoesNotApplyNoRawatWindow(t *testing.T) {
	db, drv := openStubRawatJalanDB(t, func(query string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.HasPrefix(query, "SELECT COUNT(*)") {
			return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(1)}}}, nil
		}
		return &stubRows{
			cols: kunjunganColumns,
			data: [][]driver.Value{
				sampleKunjunganRow("2025/01/10/000001", "2026-04-22", "09:00:00", "Rina", "BPJS", "Bukan Rujukan", "", ""),
			},
		}, nil
	})

	repo := NewRepository(db)
	filter := FilterAntreanDokter{
		Tanggal:      "2026-04-22,2026-04-22",
		JenisAntrean: JenisAntreanTidakRujukan,
		OrderBy:      "waktu_registrasi",
		SortOrder:    "ASC",
		Page:         1,
		Limit:        20,
	}

	items, total, err := repo.DaftarAntreanDokter(context.Background(), "DR001", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, got total=%d len=%d", total, len(items))
	}

	for _, c := range drv.getCalls() {
		if strings.Contains(c.query, "CONVERT(? USING latin1)") {
			t.Errorf("doctor login mode should use kd_dokter index without no_rawat window, got: %s", c.query)
		}
		if !strings.Contains(c.query, "r.kd_dokter = ?") {
			t.Errorf("expected r.kd_dokter = ? filter, got: %s", c.query)
		}
	}
}

func TestDaftarAntreanDokter_MultiRujukanSameNoRawat_PreservesCompositePK(t *testing.T) {
	db, drv := openStubRawatJalanDB(t, func(query string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.HasPrefix(query, "SELECT COUNT(*)") {
			return &stubRows{cols: []string{"count"}, data: [][]driver.Value{{int64(2)}}}, nil
		}
		return &stubRows{
			cols: kunjunganColumns,
			data: [][]driver.Value{
				sampleKunjunganRow("2026/04/22/000005", "2026-04-22", "10:00:00", "Tono", "BPJS", "Rujukan", "JTG", "DRJTG"),
				sampleKunjunganRow("2026/04/22/000005", "2026-04-22", "10:00:00", "Tono", "BPJS", "Rujukan", "PAR", "DRPAR"),
			},
		}, nil
	})

	repo := NewRepository(db)
	filter := FilterAntreanDokter{
		Tanggal:      "2026-04-22,2026-04-22",
		JenisAntrean: JenisAntreanRujukan,
		Page:         1,
		Limit:        20,
	}

	items, total, err := repo.DaftarAntreanDokter(context.Background(), "", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 referrals on same no_rawat, got total=%d len=%d", total, len(items))
	}

	var verifiedCompositeJoin bool
	for _, c := range drv.getCalls() {
		if strings.Contains(c.query, ") AS page_rip") {
			verifiedCompositeJoin = true
			if !strings.Contains(c.query, "ON page_rip.no_rawat = rip.no_rawat AND page_rip.kd_dokter = rip.kd_dokter") {
				t.Errorf("expected composite PK join on page_rip, got: %s", c.query)
			}
		}
	}
	if !verifiedCompositeJoin {
		t.Error("expected deferred join page_rip query to be executed")
	}
}

func TestSortAntrean_DeterministicTieBreaker(t *testing.T) {
	items := []KunjunganRawatJalan{
		{
			NoRawat:           "2026/04/22/000002",
			TanggalRegistrasi: "2026-04-22",
			JamRegistrasi:     "09:00:00",
			StatusPemeriksaan: StatusBelum,
			JenisAntrean:      JenisAntreanTidakRujukan,
		},
		{
			NoRawat:           "2026/04/22/000001",
			TanggalRegistrasi: "2026-04-22",
			JamRegistrasi:     "08:00:00",
			StatusPemeriksaan: StatusBelum,
			JenisAntrean:      JenisAntreanRujukan,
			KodeDokterRujukan: "DRANDY",
		},
		{
			NoRawat:           "2026/04/22/000003",
			TanggalRegistrasi: "2026-04-22",
			JamRegistrasi:     "07:30:00",
			StatusPemeriksaan: StatusSudah,
			JenisAntrean:      JenisAntreanTidakRujukan,
		},
	}

	sortAntrean(items, "status_pemeriksaan", "DESC")

	if items[0].NoRawat != "2026/04/22/000003" {
		t.Errorf("expected Sudah item first, got %s", items[0].NoRawat)
	}
	if items[1].NoRawat != "2026/04/22/000001" || items[2].NoRawat != "2026/04/22/000002" {
		t.Errorf("expected tied Belum items ordered by earliest registration time, got %s then %s", items[1].NoRawat, items[2].NoRawat)
	}
}

func TestGetInfoRegistrasi_ScansExtendedVisitMetadata(t *testing.T) {
	db, _ := openStubRawatJalanDB(t, func(_ string, _ []driver.NamedValue) (driver.Rows, error) {
		return &stubRows{
			cols: []string{"tgl_registrasi", "jam_reg", "kd_pj", "status_bayar", "kd_dokter", "stts", "status_lanjut"},
			data: [][]driver.Value{
				{"2026-11-12", "08:30:00", "BPJ", "Belum Bayar", "DR001", "Belum", "Ranap"},
			},
		}, nil
	})

	repo := NewRepository(db)
	info, err := repo.GetInfoRegistrasi(context.Background(), "2026/11/12/000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil || info.KodeDokter != "DR001" || info.StatusPemeriksaan != "Belum" || info.StatusLanjut != "Ranap" {
		t.Fatalf("unexpected InfoRegistrasiPasien: %+v", info)
	}
}
