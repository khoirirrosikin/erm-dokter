package laboratorium

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"erm-dokter/internal/pkg/database"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/formatter"
)

func (r *repository) DaftarHasilLabPA(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, int, error) {
	return r.queryRiwayatLabPA(ctx, "pl.no_rawat = ?", noRawat, statusLanjut, filter)
}

func (r *repository) DaftarHasilLabPAByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, int, error) {
	return r.queryRiwayatLabPA(ctx, "rp.no_rkm_medis = ?", noRM, statusLanjut, filter)
}

func (r *repository) queryRiwayatLabPA(ctx context.Context, whereClause string, paramValue string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, int, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, whereClause)
	args = append(args, paramValue)

	conditions = append(conditions, "pl.kategori = 'PA'")

	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		conditions = append(conditions, "pl.status = 'Ralan'")
	case shared.StatusLanjutRawatInap:
		conditions = append(conditions, "pl.status = 'Ranap'")
	}

	if filter.Tanggal != "" {
		tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)
		if tglAwal != "" && tglAkhir != "" {
			conditions = append(conditions, "pl.tgl_periksa BETWEEN ? AND ?")
			args = append(args, tglAwal, tglAkhir)
		}
	}

	whereSQL := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM periksa_lab pl
		INNER JOIN reg_periksa rp ON rp.no_rawat = pl.no_rawat
		%s
	`, whereSQL)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]HasilLaboratorium, 0), 0, nil
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			pl.no_rawat,
			pl.kd_jenis_prw AS kode_tindakan,
			COALESCE(jpl.nm_perawatan, pl.kd_jenis_prw) AS nama_tindakan,
			pl.kategori,
			pl.status,
			COALESCE(DATE_FORMAT(pl.tgl_periksa, '%%Y-%%m-%%d'), '') AS tgl_periksa,
			pl.jam AS jam_periksa,
			pl.dokter_perujuk AS kd_dokter_perujuk,
			COALESCE(d_perujuk.nm_dokter, '-') AS nm_dokter_perujuk,
			pl.kd_dokter AS kd_dokter_pj,
			COALESCE(d_pj.nm_dokter, '-') AS nm_dokter_pj,
			pl.nip AS nip_petugas,
			COALESCE(p.nama, '-') AS nama_petugas
		FROM periksa_lab pl
		INNER JOIN reg_periksa rp ON rp.no_rawat = pl.no_rawat
		LEFT JOIN jns_perawatan_lab jpl ON jpl.kd_jenis_prw = pl.kd_jenis_prw
		LEFT JOIN dokter d_perujuk ON d_perujuk.kd_dokter = pl.dokter_perujuk
		LEFT JOIN dokter d_pj ON d_pj.kd_dokter = pl.kd_dokter
		LEFT JOIN petugas p ON p.nip = pl.nip
		%s
		ORDER BY pl.tgl_periksa DESC, pl.jam DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	argsWithPaging := append(args, filter.Limit, filter.Offset())
	rows, err := r.db.QueryContext(ctx, dataQuery, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []HasilLaboratorium
	for rows.Next() {
		var item HasilLaboratorium
		err = rows.Scan(
			&item.NoRawat,
			&item.KodeTindakan,
			&item.NamaTindakan,
			&item.Kategori,
			&item.Status,
			&item.TanggalPeriksa,
			&item.JamPeriksa,
			&item.DokterPerujuk.KodeDokter,
			&item.DokterPerujuk.NamaDokter,
			&item.DokterPJ.KodeDokter,
			&item.DokterPJ.NamaDokter,
			&item.Petugas.Nip,
			&item.Petugas.Nama,
		)
		if err != nil {
			return nil, 0, err
		}

		detailPA, errPA := r.fetchDetailPA(ctx, item.NoRawat, item.KodeTindakan, item.TanggalPeriksa, item.JamPeriksa)
		if errPA != nil {
			return nil, 0, errPA
		}
		item.DetailPA = detailPA

		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) DetailHasilLabPA(ctx context.Context, noRawat string, kodeTindakan string, tanggalPeriksa string, jamPeriksa string) (*HasilLaboratorium, error) {
	query := `
		SELECT 
			pl.no_rawat,
			pl.kd_jenis_prw AS kode_tindakan,
			COALESCE(jpl.nm_perawatan, pl.kd_jenis_prw) AS nama_tindakan,
			pl.kategori,
			pl.status,
			COALESCE(DATE_FORMAT(pl.tgl_periksa, '%Y-%m-%d'), '') AS tgl_periksa,
			pl.jam AS jam_periksa,
			pl.dokter_perujuk AS kd_dokter_perujuk,
			COALESCE(d_perujuk.nm_dokter, '-') AS nm_dokter_perujuk,
			pl.kd_dokter AS kd_dokter_pj,
			COALESCE(d_pj.nm_dokter, '-') AS nm_dokter_pj,
			pl.nip AS nip_petugas,
			COALESCE(p.nama, '-') AS nama_petugas
		FROM periksa_lab pl
		INNER JOIN reg_periksa rp ON rp.no_rawat = pl.no_rawat
		LEFT JOIN jns_perawatan_lab jpl ON jpl.kd_jenis_prw = pl.kd_jenis_prw
		LEFT JOIN dokter d_perujuk ON d_perujuk.kd_dokter = pl.dokter_perujuk
		LEFT JOIN dokter d_pj ON d_pj.kd_dokter = pl.kd_dokter
		LEFT JOIN petugas p ON p.nip = pl.nip
		WHERE pl.no_rawat = ? AND pl.kd_jenis_prw = ? AND pl.tgl_periksa = ? AND pl.jam = ? AND pl.kategori = 'PA'
	`

	var item HasilLaboratorium
	err := r.db.QueryRowContext(ctx, query, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa).Scan(
		&item.NoRawat,
		&item.KodeTindakan,
		&item.NamaTindakan,
		&item.Kategori,
		&item.Status,
		&item.TanggalPeriksa,
		&item.JamPeriksa,
		&item.DokterPerujuk.KodeDokter,
		&item.DokterPerujuk.NamaDokter,
		&item.DokterPJ.KodeDokter,
		&item.DokterPJ.NamaDokter,
		&item.Petugas.Nip,
		&item.Petugas.Nama,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	detailPA, errPA := r.fetchDetailPA(ctx, item.NoRawat, item.KodeTindakan, item.TanggalPeriksa, item.JamPeriksa)
	if errPA != nil {
		return nil, errPA
	}
	item.DetailPA = detailPA

	return &item, nil
}

func (r *repository) fetchDetailPA(ctx context.Context, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa string) (*HasilLabPA, error) {
	query := `
		SELECT 
			COALESCE(diagnosa_klinik, '') AS diagnosa_klinik,
			COALESCE(makroskopik, '') AS makroskopis,
			COALESCE(mikroskopik, '') AS mikroskopis,
			COALESCE(kesimpulan, '') AS kesimpulan,
			COALESCE(kesan, '') AS kesan
		FROM detail_periksa_labpa
		WHERE no_rawat = ? AND kd_jenis_prw = ? AND tgl_periksa = ? AND jam = ?
		LIMIT 1
	`

	var pa HasilLabPA
	err := r.db.QueryRowContext(ctx, query, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa).Scan(
		&pa.DiagnosaKlinik,
		&pa.Makroskopis,
		&pa.Mikroskopis,
		&pa.Kesimpulan,
		&pa.Kesan,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &HasilLabPA{
				DiagnosaKlinik: "",
				Makroskopis:    "",
				Mikroskopis:    "",
				Kesimpulan:     "",
				Kesan:          "",
			}, nil
		}
		return nil, err
	}

	return &pa, nil
}

func toNullString(val string) any {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func toZeroDateIfEmpty(val string) string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return "0000-00-00"
	}
	return trimmed
}

func (r *repository) generateNoPermintaanPA(ctx context.Context, tx *sql.Tx, tanggalPermintaan string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(tanggalPermintaan))
	prefixDate := time.Now().Format("20060102")
	if err == nil {
		prefixDate = parsedDate.Format("20060102")
	}
	prefix := "PA" + prefixDate
	minOrder := prefix + "0000"
	maxOrder := prefix + "9999"

	query := `SELECT noorder FROM permintaan_labpa WHERE noorder BETWEEN ? AND ? ORDER BY noorder DESC LIMIT 1 FOR UPDATE`
	var lastToday string
	err = tx.QueryRowContext(ctx, query, minOrder, maxOrder).Scan(&lastToday)
	if errors.Is(err, sql.ErrNoRows) {
		return prefix + "0001", nil
	}
	if err != nil {
		return "", err
	}

	if len(lastToday) != 14 || !strings.HasPrefix(lastToday, prefix) {
		return "PA" + time.Now().Format("20060102150405"), nil
	}

	tail := lastToday[len(lastToday)-4:]
	next, err := strconv.Atoi(tail)
	if err != nil || next >= 9999 {
		return "PA" + time.Now().Format("20060102150405"), nil
	}

	return fmt.Sprintf("%s%04d", prefix, next+1), nil
}

func (r *repository) SimpanPermintaanLabPA(ctx context.Context, noRawat string, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPARequest, kodeTindakanList []string) (string, error) {
	var lastErr error
	status := strings.ToLower(string(statusLanjut))

	for attempt := 1; attempt <= 3; attempt++ {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return "", err
		}

		noPermintaan, err := r.generateNoPermintaanPA(ctx, tx, req.TanggalPermintaan)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}

		insertHeaderQuery := `
			INSERT INTO permintaan_labpa (
				noorder, no_rawat, tgl_permintaan, jam_permintaan, 
				tgl_sampel, jam_sampel, tgl_hasil, jam_hasil, 
				dokter_perujuk, status, informasi_tambahan, diagnosa_klinis,
				pengambilan_bahan, diperoleh_dengan, lokasi_jaringan, diawetkan_dengan,
				pernah_dilakukan_di, tanggal_pa_sebelumnya, nomor_pa_sebelumnya, diagnosa_pa_sebelumnya
			) VALUES (
				?, ?, ?, ?, 
				'0000-00-00', '00:00:00', '0000-00-00', '00:00:00', 
				?, ?, ?, ?,
				?, ?, ?, ?,
				?, ?, ?, ?
			)
		`
		tglPengambilanBahan := strings.TrimSpace(req.PengambilanBahan)
		if tglPengambilanBahan == "" {
			tglPengambilanBahan = strings.TrimSpace(req.TanggalPermintaan)
		}

		_, err = tx.ExecContext(ctx, insertHeaderQuery,
			noPermintaan, noRawat, req.TanggalPermintaan, req.JamPermintaan,
			kodeDokter, status, req.InformasiTambahan, req.DiagnosaKlinis,
			toNullString(tglPengambilanBahan),
			toNullString(req.DiperolehDengan),
			toNullString(req.LokasiJaringan),
			toNullString(req.DiawetkanDengan),
			toNullString(req.PernahDilakukanDi),
			toZeroDateIfEmpty(req.TanggalPASebelumnya),
			toNullString(req.NomorPASebelumnya),
			toNullString(req.DiagnosaPASebelumnya),
		)
		if err != nil {
			_ = tx.Rollback()
			if isDuplicateKey(err) {
				lastErr = err
				continue
			}
			return "", err
		}

		stmtTindakan, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_pemeriksaan_labpa (noorder, kd_jenis_prw, stts_bayar) VALUES (?, ?, 'Belum')`)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
		defer stmtTindakan.Close()

		for _, kodeTindakan := range kodeTindakanList {
			if _, err := stmtTindakan.ExecContext(ctx, noPermintaan, kodeTindakan); err != nil {
				_ = tx.Rollback()
				return "", err
			}
		}

		if err := tx.Commit(); err != nil {
			return "", err
		}

		return noPermintaan, nil
	}

	if lastErr != nil {
		return "", fmt.Errorf("Gagal membuat nomor permintaan laboratorium PA karena bentrok nomor urut, silakan coba kirim ulang: %w", lastErr)
	}
	return "", errors.New("Gagal membuat nomor permintaan laboratorium PA setelah 3 percobaan")
}

func (r *repository) DaftarPermintaanLabPA(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]PermintaanLabPA, error) {
	statusCondition := ""
	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		statusCondition = " AND pl.status = 'ralan'"
	case shared.StatusLanjutRawatInap:
		statusCondition = " AND pl.status = 'ranap'"
	}

	query := fmt.Sprintf(`
		SELECT 
			pl.noorder AS no_permintaan,
			pl.no_rawat,
			DATE_FORMAT(pl.tgl_permintaan, '%%Y-%%m-%%d') AS tgl_permintaan,
			pl.jam_permintaan,
			DATE_FORMAT(pl.tgl_sampel, '%%Y-%%m-%%d') AS tgl_sampel,
			pl.jam_sampel,
			DATE_FORMAT(pl.tgl_hasil, '%%Y-%%m-%%d') AS tgl_hasil,
			pl.jam_hasil,
			pl.dokter_perujuk,
			COALESCE(d.nm_dokter, '-') AS nama_dokter_perujuk,
			pl.status,
			pl.informasi_tambahan,
			pl.diagnosa_klinis,
			COALESCE(DATE_FORMAT(pl.pengambilan_bahan, '%%Y-%%m-%%d'), '') AS pengambilan_bahan,
			COALESCE(pl.diperoleh_dengan, '') AS diperoleh_dengan,
			COALESCE(pl.lokasi_jaringan, '') AS lokasi_jaringan,
			COALESCE(pl.diawetkan_dengan, '') AS diawetkan_dengan,
			COALESCE(pl.pernah_dilakukan_di, '') AS pernah_dilakukan_di,
			COALESCE(DATE_FORMAT(pl.tanggal_pa_sebelumnya, '%%Y-%%m-%%d'), '') AS tanggal_pa_sebelumnya,
			COALESCE(pl.nomor_pa_sebelumnya, '') AS nomor_pa_sebelumnya,
			COALESCE(pl.diagnosa_pa_sebelumnya, '') AS diagnosa_pa_sebelumnya
		FROM permintaan_labpa pl
		LEFT JOIN dokter d ON d.kd_dokter = pl.dokter_perujuk
		WHERE pl.no_rawat = ?%s
		ORDER BY pl.tgl_permintaan DESC, pl.jam_permintaan DESC
	`, statusCondition)
	rows, err := r.db.QueryContext(ctx, query, noRawat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PermintaanLabPA, 0)
	for rows.Next() {
		var item PermintaanLabPA
		err := rows.Scan(
			&item.NoPermintaan,
			&item.NoRawat,
			&item.TanggalPermintaan,
			&item.JamPermintaan,
			&item.TanggalSampel,
			&item.JamSampel,
			&item.TanggalHasil,
			&item.JamHasil,
			&item.KodeDokterPerujuk,
			&item.NamaDokterPerujuk,
			&item.Status,
			&item.InformasiTambahan,
			&item.DiagnosaKlinis,
			&item.PengambilanBahan,
			&item.DiperolehDengan,
			&item.LokasiJaringan,
			&item.DiawetkanDengan,
			&item.PernahDilakukanDi,
			&item.TanggalPASebelumnya,
			&item.NomorPASebelumnya,
			&item.DiagnosaPASebelumnya,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) DaftarPermintaanLabPAByRM(ctx context.Context, noRkmMedis string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]PermintaanLabPA, int, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, "rp.no_rkm_medis = ?")
	args = append(args, noRkmMedis)

	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		conditions = append(conditions, "pl.status = 'ralan'")
	case shared.StatusLanjutRawatInap:
		conditions = append(conditions, "pl.status = 'ranap'")
	}

	if filter.Tanggal != "" {
		tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)
		if tglAwal != "" && tglAkhir != "" {
			conditions = append(conditions, "pl.tgl_permintaan BETWEEN ? AND ?")
			args = append(args, tglAwal, tglAkhir)
		}
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT pl.noorder)
		FROM permintaan_labpa pl
		INNER JOIN reg_periksa rp ON rp.no_rawat = pl.no_rawat
		WHERE %s
	`, whereClause)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return make([]PermintaanLabPA, 0), 0, nil
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			pl.noorder AS no_permintaan,
			pl.no_rawat,
			DATE_FORMAT(pl.tgl_permintaan, '%%Y-%%m-%%d') AS tgl_permintaan,
			pl.jam_permintaan,
			DATE_FORMAT(pl.tgl_sampel, '%%Y-%%m-%%d') AS tgl_sampel,
			pl.jam_sampel,
			DATE_FORMAT(pl.tgl_hasil, '%%Y-%%m-%%d') AS tgl_hasil,
			pl.jam_hasil,
			pl.dokter_perujuk,
			COALESCE(d.nm_dokter, '-') AS nama_dokter_perujuk,
			pl.status,
			pl.informasi_tambahan,
			pl.diagnosa_klinis,
			COALESCE(DATE_FORMAT(pl.pengambilan_bahan, '%%Y-%%m-%%d'), '') AS pengambilan_bahan,
			COALESCE(pl.diperoleh_dengan, '') AS diperoleh_dengan,
			COALESCE(pl.lokasi_jaringan, '') AS lokasi_jaringan,
			COALESCE(pl.diawetkan_dengan, '') AS diawetkan_dengan,
			COALESCE(pl.pernah_dilakukan_di, '') AS pernah_dilakukan_di,
			COALESCE(DATE_FORMAT(pl.tanggal_pa_sebelumnya, '%%Y-%%m-%%d'), '') AS tanggal_pa_sebelumnya,
			COALESCE(pl.nomor_pa_sebelumnya, '') AS nomor_pa_sebelumnya,
			COALESCE(pl.diagnosa_pa_sebelumnya, '') AS diagnosa_pa_sebelumnya
		FROM permintaan_labpa pl
		INNER JOIN reg_periksa rp ON rp.no_rawat = pl.no_rawat
		LEFT JOIN dokter d ON d.kd_dokter = pl.dokter_perujuk
		WHERE %s
		ORDER BY pl.tgl_permintaan DESC, pl.jam_permintaan DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	selectArgs := append(args, filter.Limit, filter.Offset())
	rows, err := r.db.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []PermintaanLabPA
	for rows.Next() {
		var item PermintaanLabPA
		err := rows.Scan(
			&item.NoPermintaan,
			&item.NoRawat,
			&item.TanggalPermintaan,
			&item.JamPermintaan,
			&item.TanggalSampel,
			&item.JamSampel,
			&item.TanggalHasil,
			&item.JamHasil,
			&item.KodeDokterPerujuk,
			&item.NamaDokterPerujuk,
			&item.Status,
			&item.InformasiTambahan,
			&item.DiagnosaKlinis,
			&item.PengambilanBahan,
			&item.DiperolehDengan,
			&item.LokasiJaringan,
			&item.DiawetkanDengan,
			&item.PernahDilakukanDi,
			&item.TanggalPASebelumnya,
			&item.NomorPASebelumnya,
			&item.DiagnosaPASebelumnya,
		)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) DetailPermintaanLabPA(ctx context.Context, noPermintaan string) (*DetailPermintaanLabPA, error) {
	headerQuery := `
		SELECT 
			pl.noorder AS no_permintaan,
			pl.no_rawat,
			DATE_FORMAT(pl.tgl_permintaan, '%Y-%m-%d') AS tgl_permintaan,
			pl.jam_permintaan,
			DATE_FORMAT(pl.tgl_sampel, '%Y-%m-%d') AS tgl_sampel,
			pl.jam_sampel,
			DATE_FORMAT(pl.tgl_hasil, '%Y-%m-%d') AS tgl_hasil,
			pl.jam_hasil,
			pl.dokter_perujuk,
			COALESCE(d.nm_dokter, '-') AS nama_dokter_perujuk,
			pl.status,
			pl.informasi_tambahan,
			pl.diagnosa_klinis,
			COALESCE(DATE_FORMAT(pl.pengambilan_bahan, '%Y-%m-%d'), '') AS pengambilan_bahan,
			COALESCE(pl.diperoleh_dengan, '') AS diperoleh_dengan,
			COALESCE(pl.lokasi_jaringan, '') AS lokasi_jaringan,
			COALESCE(pl.diawetkan_dengan, '') AS diawetkan_dengan,
			COALESCE(pl.pernah_dilakukan_di, '') AS pernah_dilakukan_di,
			COALESCE(DATE_FORMAT(pl.tanggal_pa_sebelumnya, '%Y-%m-%d'), '') AS tanggal_pa_sebelumnya,
			COALESCE(pl.nomor_pa_sebelumnya, '') AS nomor_pa_sebelumnya,
			COALESCE(pl.diagnosa_pa_sebelumnya, '') AS diagnosa_pa_sebelumnya
		FROM permintaan_labpa pl
		LEFT JOIN dokter d ON d.kd_dokter = pl.dokter_perujuk
		WHERE pl.noorder = ?
	`
	var detail DetailPermintaanLabPA
	err := r.db.QueryRowContext(ctx, headerQuery, noPermintaan).Scan(
		&detail.NoPermintaan,
		&detail.NoRawat,
		&detail.TanggalPermintaan,
		&detail.JamPermintaan,
		&detail.TanggalSampel,
		&detail.JamSampel,
		&detail.TanggalHasil,
		&detail.JamHasil,
		&detail.KodeDokterPerujuk,
		&detail.NamaDokterPerujuk,
		&detail.Status,
		&detail.InformasiTambahan,
		&detail.DiagnosaKlinis,
		&detail.PengambilanBahan,
		&detail.DiperolehDengan,
		&detail.LokasiJaringan,
		&detail.DiawetkanDengan,
		&detail.PernahDilakukanDi,
		&detail.TanggalPASebelumnya,
		&detail.NomorPASebelumnya,
		&detail.DiagnosaPASebelumnya,
	)
	if err != nil {
		return nil, err
	}

	detailQuery := `
		SELECT 
			ppl.kd_jenis_prw AS kode_tindakan,
			COALESCE(jpl.nm_perawatan, ppl.kd_jenis_prw) AS nama_tindakan,
			COALESCE(jpl.total_byr, 0) AS biaya,
			COALESCE(ppl.stts_bayar, 'Belum') AS status_bayar
		FROM permintaan_pemeriksaan_labpa ppl
		LEFT JOIN jns_perawatan_lab jpl ON jpl.kd_jenis_prw = ppl.kd_jenis_prw
		WHERE ppl.noorder = ?
		ORDER BY ppl.kd_jenis_prw ASC
	`
	rows, err := r.db.QueryContext(ctx, detailQuery, noPermintaan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	detail.Pemeriksaan = make([]PemeriksaanLabPAItem, 0)
	for rows.Next() {
		var item PemeriksaanLabPAItem
		err := rows.Scan(
			&item.KodeTindakan,
			&item.NamaTindakan,
			&item.Biaya,
			&item.StatusBayar,
		)
		if err != nil {
			return nil, err
		}
		detail.Pemeriksaan = append(detail.Pemeriksaan, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &detail, nil
}

func (r *repository) UpdatePermintaanLabPA(ctx context.Context, noPermintaan string, req SimpanPermintaanLabPARequest, kodeTindakanList []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tglPengambilanBahan := strings.TrimSpace(req.PengambilanBahan)
	if tglPengambilanBahan == "" {
		tglPengambilanBahan = strings.TrimSpace(req.TanggalPermintaan)
	}

	updateHeaderQuery := `
		UPDATE permintaan_labpa
		SET tgl_permintaan = ?, jam_permintaan = ?, informasi_tambahan = ?, diagnosa_klinis = ?,
		    pengambilan_bahan = ?, diperoleh_dengan = ?, lokasi_jaringan = ?, diawetkan_dengan = ?,
		    pernah_dilakukan_di = ?, tanggal_pa_sebelumnya = ?, nomor_pa_sebelumnya = ?, diagnosa_pa_sebelumnya = ?
		WHERE noorder = ?
	`
	res, err := tx.ExecContext(ctx, updateHeaderQuery,
		req.TanggalPermintaan, req.JamPermintaan, req.InformasiTambahan, req.DiagnosaKlinis,
		toNullString(tglPengambilanBahan),
		toNullString(req.DiperolehDengan),
		toNullString(req.LokasiJaringan),
		toNullString(req.DiawetkanDengan),
		toNullString(req.PernahDilakukanDi),
		toZeroDateIfEmpty(req.TanggalPASebelumnya),
		toNullString(req.NomorPASebelumnya),
		toNullString(req.DiagnosaPASebelumnya),
		noPermintaan,
	)
	if err != nil {
		return err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_labpa WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	stmtTindakan, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_pemeriksaan_labpa (noorder, kd_jenis_prw, stts_bayar) VALUES (?, ?, 'Belum')`)
	if err != nil {
		return err
	}
	defer stmtTindakan.Close()

	for _, kodeTindakan := range kodeTindakanList {
		if _, err := stmtTindakan.ExecContext(ctx, noPermintaan, kodeTindakan); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) HapusPermintaanLabPA(ctx context.Context, noPermintaan string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_labpa WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM permintaan_labpa WHERE noorder = ?`, noPermintaan)
	if err != nil {
		return err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return err
	}

	return tx.Commit()
}
