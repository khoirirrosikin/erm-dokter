package laboratorium

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"erm-dokter/internal/pkg/crypto"
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

func TestDaftarHasilLabMB_BatchHydratesDetailPKWithoutNestedConn(t *testing.T) {
	headerCols := []string{
		"no_rawat", "kode_tindakan", "nama_tindakan", "kategori", "status",
		"tgl_periksa", "jam_periksa", "kd_dokter_perujuk", "nm_dokter_perujuk",
		"kd_dokter_pj", "nm_dokter_pj", "nip_petugas", "nama_petugas",
	}
	detailCols := []string{
		"no_rawat", "kode_tindakan", "tgl_periksa", "jam_periksa",
		"id_template", "nama_pemeriksaan", "nilai", "satuan", "nilai_rujukan", "keterangan",
	}

	var detailQueryCount int
	var capturedDetailSQL string
	var capturedDetailArgs []driver.NamedValue

	db := newStubLabDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
		switch {
		case strings.Contains(query, "COUNT(*)"):
			return []string{"total"}, [][]driver.Value{{int64(2)}}, nil
		case strings.Contains(query, "FROM detail_periksa_lab dpl"):
			detailQueryCount++
			capturedDetailSQL = query
			capturedDetailArgs = args
			return detailCols, [][]driver.Value{
				{
					"2026/09/04/000001", "MB001", "2026-09-04", "10:15:00",
					int64(101), "Bakteri Aerob", "Klebsiella pneumoniae", "", "Negatif", "Positif",
				},
				{
					"2026/09/04/000001", "MB001", "2026-09-04", "10:15:00",
					int64(0), "Sensitivitas Meropenem", "Sensitif", "mm", ">= 23", "S",
				},
			}, nil
		default:
			return headerCols, [][]driver.Value{
				{
					"2026/09/04/000001", "MB001", "Kultur Sputum", "MB", "Ralan",
					"2026-09-04", "10:15:00", "D001", "dr. Perujuk", "D002", "dr. PJ", "P001", "Analis A",
				},
				{
					"2026/09/04/000002", "MB002", "Pewarnaan Gram", "MB", "Ralan",
					"2026-09-04", "11:00:00", "D001", "dr. Perujuk", "D002", "dr. PJ", "P001", "Analis A",
				},
			}, nil
		}
	})
	db.SetMaxOpenConns(1)

	repo := NewRepository(db)
	filter := FilterRiwayatLab{Page: 1, Limit: 10}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	list, total, err := repo.DaftarHasilLabMB(ctx, "2026/09/04/000001", shared.StatusLanjutRawatJalan, filter)
	if err != nil {
		t.Fatalf("DaftarHasilLabMB unexpected error: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected total=2 len=2, got total=%d len=%d", total, len(list))
	}
	if detailQueryCount != 1 {
		t.Fatalf("expected 1 batch detail query, got %d", detailQueryCount)
	}
	assertDateFormatSQL(t, capturedDetailSQL, "COALESCE(DATE_FORMAT(dpl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa")
	if len(capturedDetailArgs) != 8 {
		t.Fatalf("expected 8 composite PK args for 2 headers, got %d", len(capturedDetailArgs))
	}

	if len(list[0].DetailPK) != 2 {
		t.Fatalf("expected 2 detail items for first MB header, got %d", len(list[0].DetailPK))
	}
	if list[0].DetailPK[0].IdTemplate != "101" || list[0].DetailPK[0].Nilai != "Klebsiella pneumoniae" {
		t.Fatalf("unexpected first detail item: %+v", list[0].DetailPK[0])
	}
	if list[0].DetailPK[1].IdTemplate != "" || list[0].DetailPK[1].Nilai != "Sensitif" {
		t.Fatalf("unexpected second detail item: %+v", list[0].DetailPK[1])
	}
	if list[1].DetailPK == nil || len(list[1].DetailPK) != 0 {
		t.Fatalf("expected non-nil empty DetailPK slice for header without details, got %#v", list[1].DetailPK)
	}

	detailQueryCount = 0
	listByRM, totalByRM, err := repo.DaftarHasilLabMBByRM(ctx, "000001", shared.StatusLanjut("Semua"), filter)
	if err != nil {
		t.Fatalf("DaftarHasilLabMBByRM unexpected error: %v", err)
	}
	if totalByRM != 2 || len(listByRM) != 2 || detailQueryCount != 1 || len(listByRM[0].DetailPK) != 2 {
		t.Fatalf("unexpected DaftarHasilLabMBByRM result: total=%d len=%d detailQueries=%d", totalByRM, len(listByRM), detailQueryCount)
	}
}

func TestDetailHasilLabMB_LoadsDetailPK(t *testing.T) {
	headerCols := []string{
		"no_rawat", "kode_tindakan", "nama_tindakan", "kategori", "status",
		"tgl_periksa", "jam_periksa", "kd_dokter_perujuk", "nm_dokter_perujuk",
		"kd_dokter_pj", "nm_dokter_pj", "nip_petugas", "nama_petugas",
	}

	t.Run("loads detail_periksa_lab and encrypts IdTemplate on handler", func(t *testing.T) {
		db := newStubLabDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "FROM detail_periksa_lab dpl") {
				return []string{"id_template", "nama_pemeriksaan", "nilai", "satuan", "nilai_rujukan", "keterangan"},
					[][]driver.Value{
						{int64(205), "Kultur Darah", "Tidak ditemukan pertumbuhan kuman", "", "Steril", "Normal"},
					}, nil
			}
			return headerCols, [][]driver.Value{
				{
					"2026/09/04/000001", "MB010", "Kultur Darah Aerob", "MB", "Ranap",
					"2026-09-04", "14:30:00", "D001", "dr. Perujuk", "D002", "dr. PJ", "P001", "Analis B",
				},
			}, nil
		})
		db.SetMaxOpenConns(1)

		repo := NewRepository(db)
		item, err := repo.DetailHasilLabMB(context.Background(), "2026/09/04/000001", "MB010", "2026-09-04", "14:30:00")
		if err != nil {
			t.Fatalf("DetailHasilLabMB unexpected error: %v", err)
		}
		if len(item.DetailPK) != 1 || item.DetailPK[0].IdTemplate != "205" {
			t.Fatalf("expected DetailPK with IdTemplate=205, got %+v", item.DetailPK)
		}

		encKey := "0123456789abcdef0123456789abcdef"
		h := NewHandler(nil, encKey)
		h.encryptHasilLab(item, "")

		decTemplateID, err := crypto.Decrypt(item.DetailPK[0].IdTemplate, encKey)
		if err != nil || decTemplateID != "205" {
			t.Fatalf("expected encrypted IdTemplate to decrypt to 205, got %q (err=%v)", decTemplateID, err)
		}
	})

	t.Run("propagates error when fetchDetailPK fails", func(t *testing.T) {
		wantErr := errors.New("detail_periksa_lab read error")
		db := newStubLabDB(t, func(query string, _ []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "FROM detail_periksa_lab dpl") {
				return nil, nil, wantErr
			}
			return headerCols, [][]driver.Value{
				{
					"2026/09/04/000001", "MB010", "Kultur Darah Aerob", "MB", "Ranap",
					"2026-09-04", "14:30:00", "D001", "dr. Perujuk", "D002", "dr. PJ", "P001", "Analis B",
				},
			}, nil
		})

		repo := NewRepository(db)
		_, err := repo.DetailHasilLabMB(context.Background(), "2026/09/04/000001", "MB010", "2026-09-04", "14:30:00")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})
}

