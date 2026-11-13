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

func (r *repository) DaftarHasilLabPK(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, int, error) {
	return r.queryRiwayatLabPK(ctx, "pl.no_rawat = ?", noRawat, statusLanjut, filter)
}

func (r *repository) DaftarHasilLabPKByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, int, error) {
	return r.queryRiwayatLabPK(ctx, "rp.no_rkm_medis = ?", noRM, statusLanjut, filter)
}

func (r *repository) queryRiwayatLabPK(ctx context.Context, whereClause string, paramValue string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]HasilLaboratorium, int, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, whereClause)
	args = append(args, paramValue)

	conditions = append(conditions, "pl.kategori = 'PK'")

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

	list := make([]HasilLaboratorium, 0)
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

		item.DetailPK = make([]ItemHasilLabPK, 0)
		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	if err = rows.Close(); err != nil {
		return nil, 0, err
	}

	if err = r.hydrateDetailPKList(ctx, list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) DetailHasilLabPK(ctx context.Context, noRawat string, kodeTindakan string, tanggalPeriksa string, jamPeriksa string) (*HasilLaboratorium, error) {
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
		WHERE pl.no_rawat = ? AND pl.kd_jenis_prw = ? AND pl.tgl_periksa = ? AND pl.jam = ? AND pl.kategori = 'PK'
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

	detailPK, err := r.fetchDetailPK(ctx, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa)
	if err != nil {
		return nil, err
	}
	item.DetailPK = detailPK

	return &item, nil
}

func (r *repository) fetchDetailPK(ctx context.Context, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa string) ([]ItemHasilLabPK, error) {
	query := `
		SELECT 
			COALESCE(dpl.id_template, 0) AS id_template,
			COALESCE(tl.Pemeriksaan, '') AS nama_pemeriksaan,
			COALESCE(dpl.nilai, '') AS nilai,
			COALESCE(tl.satuan, '') AS satuan,
			COALESCE(dpl.nilai_rujukan, '') AS nilai_rujukan,
			COALESCE(dpl.keterangan, '') AS keterangan
		FROM detail_periksa_lab dpl
		LEFT JOIN template_laboratorium tl ON tl.id_template = dpl.id_template
		WHERE dpl.no_rawat = ? AND dpl.kd_jenis_prw = ? AND dpl.tgl_periksa = ? AND dpl.jam = ?
		ORDER BY tl.urut ASC, dpl.id_template ASC
	`

	rows, err := r.db.QueryContext(ctx, query, noRawat, kodeTindakan, tanggalPeriksa, jamPeriksa)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	details := make([]ItemHasilLabPK, 0)
	for rows.Next() {
		var idTemplate int
		var d ItemHasilLabPK
		err = rows.Scan(
			&idTemplate,
			&d.NamaPemeriksaan,
			&d.Nilai,
			&d.Satuan,
			&d.NilaiRujukan,
			&d.Keterangan,
		)
		if err != nil {
			return nil, err
		}
		if idTemplate > 0 {
			d.IdTemplate = strconv.Itoa(idTemplate)
		}
		details = append(details, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return details, nil
}

type periksaLabKey struct {
	noRawat        string
	kodeTindakan   string
	tanggalPeriksa string
	jamPeriksa     string
}

func (r *repository) hydrateDetailPKList(ctx context.Context, list []HasilLaboratorium) error {
	if len(list) == 0 {
		return nil
	}

	indexByKey := make(map[periksaLabKey][]int, len(list))
	clauses := make([]string, 0, len(list))
	args := make([]any, 0, len(list)*4)

	for i := range list {
		list[i].DetailPK = make([]ItemHasilLabPK, 0)
		key := periksaLabKey{
			noRawat:        list[i].NoRawat,
			kodeTindakan:   list[i].KodeTindakan,
			tanggalPeriksa: list[i].TanggalPeriksa,
			jamPeriksa:     list[i].JamPeriksa,
		}
		if _, exists := indexByKey[key]; !exists {
			clauses = append(clauses, "(dpl.no_rawat = ? AND dpl.kd_jenis_prw = ? AND dpl.tgl_periksa = ? AND dpl.jam = ?)")
			args = append(args, key.noRawat, key.kodeTindakan, key.tanggalPeriksa, key.jamPeriksa)
		}
		indexByKey[key] = append(indexByKey[key], i)
	}

	query := fmt.Sprintf(`
		SELECT 
			dpl.no_rawat,
			dpl.kd_jenis_prw AS kode_tindakan,
			COALESCE(DATE_FORMAT(dpl.tgl_periksa, '%%Y-%%m-%%d'), '') AS tgl_periksa,
			dpl.jam AS jam_periksa,
			COALESCE(dpl.id_template, 0) AS id_template,
			COALESCE(tl.Pemeriksaan, '') AS nama_pemeriksaan,
			COALESCE(dpl.nilai, '') AS nilai,
			COALESCE(tl.satuan, '') AS satuan,
			COALESCE(dpl.nilai_rujukan, '') AS nilai_rujukan,
			COALESCE(dpl.keterangan, '') AS keterangan
		FROM detail_periksa_lab dpl
		LEFT JOIN template_laboratorium tl ON tl.id_template = dpl.id_template
		WHERE %s
		ORDER BY tl.urut ASC, dpl.id_template ASC
	`, strings.Join(clauses, " OR "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key periksaLabKey
		var idTemplate int
		var d ItemHasilLabPK
		err = rows.Scan(
			&key.noRawat,
			&key.kodeTindakan,
			&key.tanggalPeriksa,
			&key.jamPeriksa,
			&idTemplate,
			&d.NamaPemeriksaan,
			&d.Nilai,
			&d.Satuan,
			&d.NilaiRujukan,
			&d.Keterangan,
		)
		if err != nil {
			return err
		}
		if idTemplate > 0 {
			d.IdTemplate = strconv.Itoa(idTemplate)
		}
		for _, idx := range indexByKey[key] {
			list[idx].DetailPK = append(list[idx].DetailPK, d)
		}
	}

	return rows.Err()
}

func (r *repository) SimpanPermintaanLabPK(ctx context.Context, noRawat string, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPKRequest, kodeTindakanList []string, templateMap map[string][]int) (string, error) {
	status := strings.ToLower(string(statusLanjut))
	return database.RunOrderTxWithRetry(ctx, database.DefaultOrderMaxAttempts, func(attempt int) (string, bool, error) {
		return r.simpanPermintaanLabPKOnce(ctx, attempt, noRawat, kodeDokter, status, req, kodeTindakanList, templateMap)
	})
}

func (r *repository) simpanPermintaanLabPKOnce(
	ctx context.Context,
	attempt int,
	noRawat string,
	kodeDokter string,
	status string,
	req SimpanPermintaanLabPKRequest,
	kodeTindakanList []string,
	templateMap map[string][]int,
) (string, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = tx.Rollback() }()

	noPermintaan, err := database.GenerateDailyOrderNumber(ctx, tx, database.OrderSpecLabPK, req.TanggalPermintaan, attempt, time.Now())
	if err != nil {
		return "", database.IsDeadlockOrLockTimeout(err), err
	}

	insertHeaderQuery := `
		INSERT INTO permintaan_lab (
			noorder, no_rawat, tgl_permintaan, jam_permintaan, 
			tgl_sampel, jam_sampel, tgl_hasil, jam_hasil, 
			dokter_perujuk, status, informasi_tambahan, diagnosa_klinis
		) VALUES (?, ?, ?, ?, '0000-00-00', '00:00:00', '0000-00-00', '00:00:00', ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, insertHeaderQuery,
		noPermintaan, noRawat, req.TanggalPermintaan, req.JamPermintaan,
		kodeDokter, status, req.InformasiTambahan, req.DiagnosaKlinis,
	)
	if err != nil {
		return "", database.IsRetryableOrderError(err), err
	}

	if len(kodeTindakanList) > 0 {
		stmtTindakan, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_pemeriksaan_lab (noorder, kd_jenis_prw, stts_bayar) VALUES (?, ?, 'Belum')`)
		if err != nil {
			return "", database.IsDeadlockOrLockTimeout(err), err
		}
		defer stmtTindakan.Close()

		stmtDetail, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_detail_permintaan_lab (noorder, kd_jenis_prw, id_template, stts_bayar) VALUES (?, ?, ?, 'Belum')`)
		if err != nil {
			return "", database.IsDeadlockOrLockTimeout(err), err
		}
		defer stmtDetail.Close()

		for _, kodeTindakan := range kodeTindakanList {
			if _, err := stmtTindakan.ExecContext(ctx, noPermintaan, kodeTindakan); err != nil {
				return "", database.IsDeadlockOrLockTimeout(err), err
			}

			if templates, ok := templateMap[kodeTindakan]; ok && len(templates) > 0 {
				for _, idTemplate := range templates {
					if _, err := stmtDetail.ExecContext(ctx, noPermintaan, kodeTindakan, idTemplate); err != nil {
						return "", database.IsDeadlockOrLockTimeout(err), err
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", database.IsRetryableOrderError(err), err
	}

	return noPermintaan, false, nil
}

func (r *repository) DaftarPermintaanLabPK(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]PermintaanLabPK, error) {
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
			pl.diagnosa_klinis
		FROM permintaan_lab pl
		LEFT JOIN dokter d ON d.kd_dokter = pl.dokter_perujuk
		WHERE pl.no_rawat = ?%s
		ORDER BY pl.tgl_permintaan DESC, pl.jam_permintaan DESC
	`, statusCondition)
	rows, err := r.db.QueryContext(ctx, query, noRawat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PermintaanLabPK, 0)
	for rows.Next() {
		var item PermintaanLabPK
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

func (r *repository) DaftarPermintaanLabPKByRM(ctx context.Context, noRkmMedis string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]PermintaanLabPK, int, error) {
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
		FROM permintaan_lab pl
		INNER JOIN reg_periksa rp ON rp.no_rawat = pl.no_rawat
		WHERE %s
	`, whereClause)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return make([]PermintaanLabPK, 0), 0, nil
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
			pl.diagnosa_klinis
		FROM permintaan_lab pl
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

	var list []PermintaanLabPK
	for rows.Next() {
		var item PermintaanLabPK
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

func (r *repository) DetailPermintaanLabPK(ctx context.Context, noPermintaan string) (*DetailPermintaanLabPK, error) {
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
			pl.diagnosa_klinis
		FROM permintaan_lab pl
		LEFT JOIN dokter d ON d.kd_dokter = pl.dokter_perujuk
		WHERE pl.noorder = ?
	`
	var header PermintaanLabPK
	err := r.db.QueryRowContext(ctx, headerQuery, noPermintaan).Scan(
		&header.NoPermintaan,
		&header.NoRawat,
		&header.TanggalPermintaan,
		&header.JamPermintaan,
		&header.TanggalSampel,
		&header.JamSampel,
		&header.TanggalHasil,
		&header.JamHasil,
		&header.KodeDokterPerujuk,
		&header.NamaDokterPerujuk,
		&header.Status,
		&header.InformasiTambahan,
		&header.DiagnosaKlinis,
	)
	if err != nil {
		return nil, err
	}

	detailQuery := `
		SELECT 
			ppl.kd_jenis_prw AS kode_tindakan,
			COALESCE(jpl.nm_perawatan, ppl.kd_jenis_prw) AS nama_tindakan,
			COALESCE(ppl.stts_bayar, 'Belum') AS status_bayar,
			COALESCE(pdpl.id_template, 0) AS id_template,
			COALESCE(tl.Pemeriksaan, '') AS nama_pemeriksaan,
			COALESCE(tl.satuan, '') AS satuan,
			COALESCE(tl.nilai_rujukan_ld, '') AS nilai_rujukan,
			COALESCE(pdpl.stts_bayar, '') AS status_bayar_template
		FROM permintaan_pemeriksaan_lab ppl
		LEFT JOIN jns_perawatan_lab jpl ON jpl.kd_jenis_prw = ppl.kd_jenis_prw
		LEFT JOIN permintaan_detail_permintaan_lab pdpl ON pdpl.noorder = ppl.noorder AND pdpl.kd_jenis_prw = ppl.kd_jenis_prw
		LEFT JOIN template_laboratorium tl ON tl.id_template = pdpl.id_template
		WHERE ppl.noorder = ?
		ORDER BY ppl.kd_jenis_prw ASC, tl.urut ASC, pdpl.id_template ASC
	`
	rows, err := r.db.QueryContext(ctx, detailQuery, noPermintaan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pemeriksaanMap := make(map[string]*PemeriksaanLabPKItem)
	var pemeriksaanOrder []string

	for rows.Next() {
		var kodeTindakan, namaTindakan, statusBayar string
		var idTemplate int
		var namaPemeriksaan, satuan, nilaiRujukan, statusBayarTemplate string

		err := rows.Scan(
			&kodeTindakan,
			&namaTindakan,
			&statusBayar,
			&idTemplate,
			&namaPemeriksaan,
			&satuan,
			&nilaiRujukan,
			&statusBayarTemplate,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := pemeriksaanMap[kodeTindakan]; !exists {
			pemeriksaanMap[kodeTindakan] = &PemeriksaanLabPKItem{
				KodeTindakan:   kodeTindakan,
				NamaTindakan:   namaTindakan,
				StatusBayar:    statusBayar,
				DetailTemplate: make([]DetailTemplateLabPKItem, 0),
			}
			pemeriksaanOrder = append(pemeriksaanOrder, kodeTindakan)
		}

		if idTemplate > 0 {
			pemeriksaanMap[kodeTindakan].DetailTemplate = append(pemeriksaanMap[kodeTindakan].DetailTemplate, DetailTemplateLabPKItem{
				IdTemplate:      strconv.Itoa(idTemplate),
				NamaPemeriksaan: namaPemeriksaan,
				Satuan:          satuan,
				NilaiRujukan:    nilaiRujukan,
				StatusBayar:     statusBayarTemplate,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	pemeriksaanList := make([]PemeriksaanLabPKItem, 0, len(pemeriksaanOrder))
	for _, kodeTindakan := range pemeriksaanOrder {
		pemeriksaanList = append(pemeriksaanList, *pemeriksaanMap[kodeTindakan])
	}

	return &DetailPermintaanLabPK{
		PermintaanLabPK: header,
		Pemeriksaan:     pemeriksaanList,
	}, nil
}

func (r *repository) UpdatePermintaanLabPK(ctx context.Context, noPermintaan string, req SimpanPermintaanLabPKRequest, kodeTindakanList []string, templateMap map[string][]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateHeaderQuery := `
		UPDATE permintaan_lab 
		SET tgl_permintaan = ?, jam_permintaan = ?, informasi_tambahan = ?, diagnosa_klinis = ?
		WHERE noorder = ?
	`
	res, err := tx.ExecContext(ctx, updateHeaderQuery, req.TanggalPermintaan, req.JamPermintaan, req.InformasiTambahan, req.DiagnosaKlinis, noPermintaan)
	if err != nil {
		return err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_detail_permintaan_lab WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_lab WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	stmtTindakan, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_pemeriksaan_lab (noorder, kd_jenis_prw, stts_bayar) VALUES (?, ?, 'Belum')`)
	if err != nil {
		return err
	}
	defer stmtTindakan.Close()

	stmtDetail, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_detail_permintaan_lab (noorder, kd_jenis_prw, id_template, stts_bayar) VALUES (?, ?, ?, 'Belum')`)
	if err != nil {
		return err
	}
	defer stmtDetail.Close()

	for _, kodeTindakan := range kodeTindakanList {
		if _, err := stmtTindakan.ExecContext(ctx, noPermintaan, kodeTindakan); err != nil {
			return err
		}

		if templates, ok := templateMap[kodeTindakan]; ok && len(templates) > 0 {
			for _, idTemplate := range templates {
				if _, err := stmtDetail.ExecContext(ctx, noPermintaan, kodeTindakan, idTemplate); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

func (r *repository) HapusPermintaanLabPK(ctx context.Context, noPermintaan string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_detail_permintaan_lab WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_lab WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM permintaan_lab WHERE noorder = ?`, noPermintaan)
	if err != nil {
		return err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return err
	}

	return tx.Commit()
}
