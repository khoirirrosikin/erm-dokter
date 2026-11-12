package resumepasien

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type stubResumeQueryFunc func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)

type stubResumeConnector struct {
	fn stubResumeQueryFunc
}

func (c *stubResumeConnector) Connect(context.Context) (driver.Conn, error) {
	return &stubResumeConn{fn: c.fn}, nil
}

func (c *stubResumeConnector) Driver() driver.Driver { return nil }

type stubResumeConn struct {
	fn stubResumeQueryFunc
}

func (c *stubResumeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}
func (c *stubResumeConn) Close() error              { return nil }
func (c *stubResumeConn) Begin() (driver.Tx, error) { return nil, errors.New("tx not supported") }

func (c *stubResumeConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.fn(query, args)
	if err != nil {
		return nil, err
	}
	return &stubResumeRows{cols: cols, rows: rows}, nil
}

type stubResumeRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubResumeRows) Columns() []string { return r.cols }
func (r *stubResumeRows) Close() error      { return nil }
func (r *stubResumeRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func newStubResumeDB(t *testing.T, fn stubResumeQueryFunc) *sql.DB {
	t.Helper()
	db := sql.OpenDB(&stubResumeConnector{fn: fn})
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func resumeRanapHeaderCols() []string {
	return []string{
		"no_rawat", "kd_dokter", "nama_dokter",
		"diagnosa_awal", "alasan", "keluhan_utama", "pemeriksaan_fisik",
		"jalannya_penyakit", "pemeriksaan_penunjang", "hasil_laborat", "tindakan_dan_operasi",
		"obat_di_rs", "diagnosa_utama", "kd_diagnosa_utama", "diagnosa_sekunder", "kd_diagnosa_sekunder",
		"diagnosa_sekunder2", "kd_diagnosa_sekunder2", "diagnosa_sekunder3", "kd_diagnosa_sekunder3",
		"diagnosa_sekunder4", "kd_diagnosa_sekunder4", "prosedur_utama", "kd_prosedur_utama",
		"prosedur_sekunder", "kd_prosedur_sekunder", "prosedur_sekunder2", "kd_prosedur_sekunder2",
		"prosedur_sekunder3", "kd_prosedur_sekunder3", "alergi", "diet", "lab_belum", "edukasi",
		"cara_keluar", "ket_keluar", "keadaan", "ket_keadaan", "dilanjutkan", "ket_dilanjutkan",
		"kontrol", "obat_pulang",
	}
}

func makeResumeRanapHeaderRow(noRawat string) []driver.Value {
	return []driver.Value{
		noRawat, "DR01", "dr. Budi",
		"Diagnosa Awal", "Alasan", "Keluhan", "Fisik",
		"Jalannya", "Penunjang", "Lab", "Tindakan",
		"Obat RS", "Diagnosa Utama", "I21", "", "",
		"", "", "", "",
		"", "", "", "",
		"", "", "", "",
		"", "", "-", "-", "-", "Edukasi",
		"Atas Izin Dokter", "-", "Membaik", "-", "Kembali Ke RS", "-",
		"2026-09-12 09:00:00", "Obat Pulang",
	}
}

func TestRepositoryRanap_MultiRoomTransferAndBatchHydration(t *testing.T) {
	kamarCols := []string{
		"no_rawat", "tgl_masuk", "jam_masuk", "tgl_keluar", "jam_keluar", "stts_pulang", "kd_kamar", "nm_bangsal",
	}

	var kamarQueryCount int
	db := newStubResumeDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
		if strings.Contains(query, "GROUP BY no_rawat") {
			t.Fatalf("unexpected full-table GROUP BY no_rawat in query:\n%s", query)
		}
		if strings.Contains(query, "FROM kamar_inap ki") {
			kamarQueryCount++
			return kamarCols, [][]driver.Value{
				{"2026/09/01/000001", "2026-09-01", "22:15:00", "2026-09-03", "08:30:00", "Pindah Kamar", "ICCU01", "Ruang ICCU"},
				{"2026/09/01/000001", "2026-09-03", "08:30:00", "0000-00-00", "00:00:00", "-", "V01", "VIP A"},

				{"2026/08/10/000002", "2026-08-10", "23:45:00", "2026-08-12", "07:00:00", "Pindah Kamar", "M01", "Melati 1"},
				{"2026/08/10/000002", "2026-08-12", "07:00:00", "2026-08-15", "00:00:00", "Sehat", "M02", "Melati 2"},
			}, nil
		}
		return resumeRanapHeaderCols(), [][]driver.Value{
			makeResumeRanapHeaderRow("2026/09/01/000001"),
			makeResumeRanapHeaderRow("2026/08/10/000002"),
		}, nil
	})
	db.SetMaxOpenConns(1)

	repo := NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	list, err := repo.RiwayatResumePasienRanapByNoRM(ctx, "000001")
	if err != nil {
		t.Fatalf("RiwayatResumePasienRanapByNoRM unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
	if kamarQueryCount != 1 {
		t.Fatalf("expected 1 batch kamar_inap query, got %d", kamarQueryCount)
	}

	gotActive := list[0]
	if gotActive.TanggalMasuk != "2026-09-01" || gotActive.JamMasuk != "22:15:00" {
		t.Fatalf("expected masuk 2026-09-01 22:15:00, got %s %s", gotActive.TanggalMasuk, gotActive.JamMasuk)
	}
	if gotActive.TanggalKeluar != "" || gotActive.JamKeluar != "" {
		t.Fatalf("expected empty keluar for active transferred patient, got %q %q", gotActive.TanggalKeluar, gotActive.JamKeluar)
	}
	if gotActive.Kamar != "V01 - VIP A" {
		t.Fatalf("expected Kamar='V01 - VIP A', got %q", gotActive.Kamar)
	}

	gotDischarged := list[1]
	if gotDischarged.TanggalMasuk != "2026-08-10" || gotDischarged.JamMasuk != "23:45:00" {
		t.Fatalf("expected masuk 2026-08-10 23:45:00, got %s %s", gotDischarged.TanggalMasuk, gotDischarged.JamMasuk)
	}
	if gotDischarged.TanggalKeluar != "2026-08-15" || gotDischarged.JamKeluar != "00:00:00" {
		t.Fatalf("expected keluar 2026-08-15 00:00:00 preserved, got %q %q", gotDischarged.TanggalKeluar, gotDischarged.JamKeluar)
	}
	if gotDischarged.Kamar != "M02 - Melati 2" {
		t.Fatalf("expected Kamar='M02 - Melati 2', got %q", gotDischarged.Kamar)
	}
	if gotDischarged.WaktuKontrol != "2026-09-12 09:00:00" {
		t.Fatalf("expected valid WaktuKontrol preserved, got %q", gotDischarged.WaktuKontrol)
	}
}

func TestRepositoryRanap_ZeroDateTimeKontrolNormalizedOnMeninggal(t *testing.T) {
	kamarCols := []string{
		"no_rawat", "tgl_masuk", "jam_masuk", "tgl_keluar", "jam_keluar", "stts_pulang", "kd_kamar", "nm_bangsal",
	}

	deceasedRow := makeResumeRanapHeaderRow("2026/05/29/049994")
	deceasedRow[36] = "Meninggal"
	deceasedRow[40] = "0000-00-00 00:00:00"

	db := newStubResumeDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
		if strings.Contains(query, "FROM kamar_inap ki") {
			return kamarCols, [][]driver.Value{
				{"2026/05/29/049994", "2026-05-29", "10:00:00", "2026-05-30", "14:20:00", "+", "ICU01", "Ruang ICU"},
			}, nil
		}
		return resumeRanapHeaderCols(), [][]driver.Value{deceasedRow}, nil
	})
	db.SetMaxOpenConns(1)

	repo := NewRepository(db)
	got, err := repo.DetailResumePasienRanap(context.Background(), "2026/05/29/049994")
	if err != nil {
		t.Fatalf("DetailResumePasienRanap unexpected error: %v", err)
	}
	if got.KeadaanPulang != KeadaanPulangMeninggal {
		t.Fatalf("expected KeadaanPulang=%q, got %q", KeadaanPulangMeninggal, got.KeadaanPulang)
	}
	if got.WaktuKontrol != "" {
		t.Fatalf("expected zero-datetime WaktuKontrol normalized to empty string, got %q", got.WaktuKontrol)
	}

	updateReq := UpdateResumePasienRanapRequest{
		DataResumePasienRanap: got.DataResumePasienRanap,
	}
	updateReq.Sanitize()
	if errs := updateReq.Validate(); errs != nil {
		t.Fatalf("expected round-trip update validation to pass for Meninggal with empty WaktuKontrol, got %v", errs)
	}
	if updateReq.WaktuKontrol != "0000-00-00 00:00:00" {
		t.Fatalf("expected Sanitize() to restore DB sentinel 0000-00-00 00:00:00 for Meninggal, got %q", updateReq.WaktuKontrol)
	}
}

