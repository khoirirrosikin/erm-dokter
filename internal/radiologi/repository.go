package radiologi

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

type Repository interface {
	DaftarHasilRadiologiKunjungan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, int, error)
	DaftarHasilRadiologiPasien(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, int, error)
	DetailHasilRadiologi(ctx context.Context, idHasil IdHasilRadiologi) (*HasilRadiologi, error)

	SimpanPermintaanRadiologi(ctx context.Context, noRawat string, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPermintaanRadiologiRequest, kodeTindakanList []string) (string, error)
	DaftarPermintaanRadiologiKunjungan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]DetailPermintaanRadiologi, error)
	DaftarPermintaanRadiologiPasien(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatPermintaanRadiologi) ([]DetailPermintaanRadiologi, int, error)
	DetailPermintaanRadiologi(ctx context.Context, noPermintaan string) (*DetailPermintaanRadiologi, error)
	UpdatePermintaanRadiologi(ctx context.Context, noPermintaan string, req SimpanPermintaanRadiologiRequest, kodeTindakanList []string) error
	HapusPermintaanRadiologi(ctx context.Context, noPermintaan string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) DaftarHasilRadiologiKunjungan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, int, error) {
	return r.queryRiwayatRadiologi(ctx, "pr.no_rawat = ?", noRawat, statusLanjut, filter)
}

func (r *repository) DaftarHasilRadiologiPasien(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, int, error) {
	return r.queryRiwayatRadiologi(ctx, "rp.no_rkm_medis = ?", noRM, statusLanjut, filter)
}

func (r *repository) queryRiwayatRadiologi(ctx context.Context, whereClause string, paramValue string, statusLanjut shared.StatusLanjut, filter FilterRiwayatRadiologi) ([]HasilRadiologi, int, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, whereClause)
	args = append(args, paramValue)

	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		conditions = append(conditions, "pr.status = 'Ralan'")
	case shared.StatusLanjutRawatInap:
		conditions = append(conditions, "pr.status = 'Ranap'")
	}

	if filter.Tanggal != "" {
		tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)
		if tglAwal != "" && tglAkhir != "" {
			conditions = append(conditions, "pr.tgl_periksa BETWEEN ? AND ?")
			args = append(args, tglAwal, tglAkhir)
		}
	}

	whereSQL := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM periksa_radiologi pr
		INNER JOIN reg_periksa rp ON rp.no_rawat = pr.no_rawat
		%s
	`, whereSQL)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]HasilRadiologi, 0), 0, nil
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			pr.no_rawat,
			pr.kd_jenis_prw AS kode_tindakan,
			COALESCE(jpr.nm_perawatan, pr.kd_jenis_prw) AS nama_tindakan,
			pr.status,
			DATE_FORMAT(pr.tgl_periksa, '%%Y-%%m-%%d') AS tanggal_periksa,
			pr.jam AS jam_periksa,
			pr.dokter_perujuk AS kd_dokter_perujuk,
			COALESCE(d_perujuk.nm_dokter, '-') AS nm_dokter_perujuk,
			pr.kd_dokter AS kd_dokter_radiologi,
			COALESCE(d_rad.nm_dokter, '-') AS nm_dokter_radiologi,
			pr.nip AS nip_petugas,
			COALESCE(p.nama, '-') AS nama_petugas,
			COALESCE(hr.hasil, '') AS hasil
		FROM periksa_radiologi pr
		INNER JOIN reg_periksa rp ON rp.no_rawat = pr.no_rawat
		LEFT JOIN jns_perawatan_radiologi jpr ON jpr.kd_jenis_prw = pr.kd_jenis_prw
		LEFT JOIN dokter d_perujuk ON d_perujuk.kd_dokter = pr.dokter_perujuk
		LEFT JOIN dokter d_rad ON d_rad.kd_dokter = pr.kd_dokter
		LEFT JOIN petugas p ON p.nip = pr.nip
		LEFT JOIN hasil_radiologi hr ON hr.no_rawat = pr.no_rawat AND hr.tgl_periksa = pr.tgl_periksa AND hr.jam = pr.jam
		%s
		ORDER BY pr.tgl_periksa DESC, pr.jam DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	argsWithPaging := append(args, filter.Limit, filter.Offset())
	rows, err := r.db.QueryContext(ctx, dataQuery, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []HasilRadiologi
	for rows.Next() {
		var item HasilRadiologi
		err = rows.Scan(
			&item.NoRawat,
			&item.KodeTindakan,
			&item.NamaTindakan,
			&item.Status,
			&item.TanggalPeriksa,
			&item.JamPeriksa,
			&item.DokterPerujuk.KodeDokter,
			&item.DokterPerujuk.NamaDokter,
			&item.DokterRadiologi.KodeDokter,
			&item.DokterRadiologi.NamaDokter,
			&item.Petugas.Nip,
			&item.Petugas.Nama,
			&item.Hasil,
		)
		if err != nil {
			return nil, 0, err
		}

		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	if err = r.attachGambarPACS(ctx, list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) DetailHasilRadiologi(ctx context.Context, idHasil IdHasilRadiologi) (*HasilRadiologi, error) {
	query := `
		SELECT 
			pr.no_rawat,
			pr.kd_jenis_prw AS kode_tindakan,
			COALESCE(jpr.nm_perawatan, pr.kd_jenis_prw) AS nama_tindakan,
			pr.status,
			DATE_FORMAT(pr.tgl_periksa, '%Y-%m-%d') AS tanggal_periksa,
			pr.jam AS jam_periksa,
			pr.dokter_perujuk AS kd_dokter_perujuk,
			COALESCE(d_perujuk.nm_dokter, '-') AS nm_dokter_perujuk,
			pr.kd_dokter AS kd_dokter_radiologi,
			COALESCE(d_rad.nm_dokter, '-') AS nm_dokter_radiologi,
			pr.nip AS nip_petugas,
			COALESCE(p.nama, '-') AS nama_petugas,
			COALESCE(hr.hasil, '') AS hasil
		FROM periksa_radiologi pr
		INNER JOIN reg_periksa rp ON rp.no_rawat = pr.no_rawat
		LEFT JOIN jns_perawatan_radiologi jpr ON jpr.kd_jenis_prw = pr.kd_jenis_prw
		LEFT JOIN dokter d_perujuk ON d_perujuk.kd_dokter = pr.dokter_perujuk
		LEFT JOIN dokter d_rad ON d_rad.kd_dokter = pr.kd_dokter
		LEFT JOIN petugas p ON p.nip = pr.nip
		LEFT JOIN hasil_radiologi hr ON hr.no_rawat = pr.no_rawat AND hr.tgl_periksa = pr.tgl_periksa AND hr.jam = pr.jam
		WHERE pr.no_rawat = ? AND pr.kd_jenis_prw = ? AND pr.tgl_periksa = ? AND pr.jam = ?
	`
	args := []any{idHasil.NoRawat, idHasil.KodeTindakan, idHasil.TanggalPeriksa, idHasil.JamPeriksa}

	var item HasilRadiologi
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&item.NoRawat,
		&item.KodeTindakan,
		&item.NamaTindakan,
		&item.Status,
		&item.TanggalPeriksa,
		&item.JamPeriksa,
		&item.DokterPerujuk.KodeDokter,
		&item.DokterPerujuk.NamaDokter,
		&item.DokterRadiologi.KodeDokter,
		&item.DokterRadiologi.NamaDokter,
		&item.Petugas.Nip,
		&item.Petugas.Nama,
		&item.Hasil,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	items := []HasilRadiologi{item}
	if err := r.attachGambarPACS(ctx, items); err != nil {
		return nil, err
	}

	return &items[0], nil
}

func (r *repository) attachGambarPACS(ctx context.Context, list []HasilRadiologi) error {
	if len(list) == 0 {
		return nil
	}

	var placeholders []string
	var args []any
	visited := make(map[string]bool)

	for _, item := range list {
		key := item.NoRawat + "|" + item.TanggalPeriksa + "|" + item.JamPeriksa
		if visited[key] {
			continue
		}
		visited[key] = true
		placeholders = append(placeholders, "(?, ?, ?)")
		args = append(args, item.NoRawat, item.TanggalPeriksa, item.JamPeriksa)
	}

	if len(placeholders) == 0 {
		return nil
	}

	inClause := strings.Join(placeholders, ", ")

	queryGambar := fmt.Sprintf(`
		SELECT 
			no_rawat,
			DATE_FORMAT(tgl_periksa, '%%Y-%%m-%%d') AS tanggal_periksa,
			jam AS jam_periksa,
			lokasi_gambar
		FROM gambar_radiologi_pacs
		WHERE (no_rawat, tgl_periksa, jam) IN (%s)
		ORDER BY lokasi_gambar ASC
	`, inClause)

	rowsGambar, err := r.db.QueryContext(ctx, queryGambar, args...)
	if err != nil {
		return err
	}
	defer rowsGambar.Close()

	mapGambar := make(map[string][]string)
	for rowsGambar.Next() {
		var noRawat, tanggalPeriksa, jamPeriksa, url string
		if err := rowsGambar.Scan(&noRawat, &tanggalPeriksa, &jamPeriksa, &url); err != nil {
			return err
		}
		key := noRawat + "|" + tanggalPeriksa + "|" + jamPeriksa
		mapGambar[key] = append(mapGambar[key], url)
	}
	if err := rowsGambar.Err(); err != nil {
		return err
	}

	for i := range list {
		key := list[i].NoRawat + "|" + list[i].TanggalPeriksa + "|" + list[i].JamPeriksa
		if urls, exists := mapGambar[key]; exists {
			list[i].GambarPACS = urls
		} else {
			list[i].GambarPACS = make([]string, 0)
		}
	}

	return nil
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "1062") || strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "primary")
}

func (r *repository) generateNoPermintaanRadiologi(ctx context.Context, tx *sql.Tx, tanggalPermintaan string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(tanggalPermintaan))
	prefixDate := time.Now().Format("20060102")
	if err == nil {
		prefixDate = parsedDate.Format("20060102")
	}
	prefix := "RAD" + prefixDate
	minOrder := prefix + "0000"
	maxOrder := prefix + "9999"

	query := `SELECT noorder FROM permintaan_radiologi WHERE noorder BETWEEN ? AND ? ORDER BY noorder DESC LIMIT 1 FOR UPDATE`
	var lastToday string
	err = tx.QueryRowContext(ctx, query, minOrder, maxOrder).Scan(&lastToday)
	if errors.Is(err, sql.ErrNoRows) {
		return prefix + "0001", nil
	}
	if err != nil {
		return "", err
	}

	if len(lastToday) != 15 || !strings.HasPrefix(lastToday, prefix) {
		return "RAD" + time.Now().Format("20060102150405"), nil
	}

	tail := lastToday[len(lastToday)-4:]
	next, err := strconv.Atoi(tail)
	if err != nil || next >= 9999 {
		return "RAD" + time.Now().Format("20060102150405"), nil
	}

	return fmt.Sprintf("%s%04d", prefix, next+1), nil
}

func (r *repository) SimpanPermintaanRadiologi(ctx context.Context, noRawat string, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanPermintaanRadiologiRequest, kodeTindakanList []string) (string, error) {
	var lastErr error
	status := strings.ToLower(string(statusLanjut))

	for attempt := 1; attempt <= 3; attempt++ {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return "", err
		}

		noPermintaan, err := r.generateNoPermintaanRadiologi(ctx, tx, req.TanggalPermintaan)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}

		insertHeaderQuery := `
			INSERT INTO permintaan_radiologi (
				noorder, no_rawat, tgl_permintaan, jam_permintaan, 
				tgl_sampel, jam_sampel, tgl_hasil, jam_hasil, 
				dokter_perujuk, status, informasi_tambahan, diagnosa_klinis
			) VALUES (
				?, ?, ?, ?, 
				'0000-00-00', '00:00:00', '0000-00-00', '00:00:00', 
				?, ?, ?, ?
			)
		`
		_, err = tx.ExecContext(ctx, insertHeaderQuery,
			noPermintaan,
			noRawat,
			req.TanggalPermintaan,
			req.JamPermintaan,
			kodeDokter,
			status,
			req.InformasiTambahan,
			req.DiagnosaKlinis,
		)
		if err != nil {
			_ = tx.Rollback()
			if isDuplicateKey(err) {
				lastErr = err
				continue
			}
			return "", err
		}

		stmtTindakan, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_pemeriksaan_radiologi (noorder, kd_jenis_prw, stts_bayar) VALUES (?, ?, 'Belum')`)
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
			_ = tx.Rollback()
			if isDuplicateKey(err) {
				lastErr = err
				continue
			}
			return "", err
		}

		return noPermintaan, nil
	}

	return "", fmt.Errorf("gagal membuat nomor order permintaan radiologi unik setelah 3 percobaan: %w", lastErr)
}

func (r *repository) DaftarPermintaanRadiologiKunjungan(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]DetailPermintaanRadiologi, error) {
	whereSQL := "WHERE pr.no_rawat = ?"
	args := []any{noRawat}

	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		whereSQL += " AND pr.status = 'ralan'"
	case shared.StatusLanjutRawatInap:
		whereSQL += " AND pr.status = 'ranap'"
	}

	query := fmt.Sprintf(`
		SELECT 
			pr.noorder AS no_permintaan,
			pr.no_rawat,
			DATE_FORMAT(pr.tgl_permintaan, '%%Y-%%m-%%d') AS tanggal_permintaan,
			pr.jam_permintaan,
			DATE_FORMAT(pr.tgl_sampel, '%%Y-%%m-%%d') AS tanggal_sampel,
			pr.jam_sampel,
			DATE_FORMAT(pr.tgl_hasil, '%%Y-%%m-%%d') AS tanggal_hasil,
			pr.jam_hasil,
			pr.dokter_perujuk AS kd_dokter_perujuk,
			COALESCE(d.nm_dokter, '-') AS nm_dokter_perujuk,
			CASE WHEN LOWER(pr.status) = 'ralan' THEN 'Ralan' ELSE 'Ranap' END AS status,
			pr.informasi_tambahan,
			pr.diagnosa_klinis
		FROM permintaan_radiologi pr
		LEFT JOIN dokter d ON d.kd_dokter = pr.dokter_perujuk
		%s
		ORDER BY pr.tgl_permintaan DESC, pr.jam_permintaan DESC, pr.noorder DESC
	`, whereSQL)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DetailPermintaanRadiologi
	var orderIDs []string
	orderIndexMap := make(map[string]int)

	for rows.Next() {
		var item DetailPermintaanRadiologi
		item.Pemeriksaan = make([]PemeriksaanRadiologiItem, 0)
		err := rows.Scan(
			&item.NoPermintaan,
			&item.NoRawat,
			&item.TanggalPermintaan,
			&item.JamPermintaan,
			&item.TanggalSampel,
			&item.JamSampel,
			&item.TanggalHasil,
			&item.JamHasil,
			&item.DokterPerujuk.KodeDokter,
			&item.DokterPerujuk.NamaDokter,
			&item.Status,
			&item.InformasiTambahan,
			&item.DiagnosaKlinis,
		)
		if err != nil {
			return nil, err
		}

		orderIndexMap[item.NoPermintaan] = len(result)
		orderIDs = append(orderIDs, item.NoPermintaan)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(orderIDs) == 0 {
		return make([]DetailPermintaanRadiologi, 0), nil
	}

	if err := r.attachItemPemeriksaan(ctx, orderIDs, result, orderIndexMap); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *repository) DaftarPermintaanRadiologiPasien(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatPermintaanRadiologi) ([]DetailPermintaanRadiologi, int, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, "rp.no_rkm_medis = ?")
	args = append(args, noRM)

	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		conditions = append(conditions, "pr.status = 'ralan'")
	case shared.StatusLanjutRawatInap:
		conditions = append(conditions, "pr.status = 'ranap'")
	}

	if filter.Tanggal != "" {
		tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)
		if tglAwal != "" && tglAkhir != "" {
			conditions = append(conditions, "pr.tgl_permintaan BETWEEN ? AND ?")
			args = append(args, tglAwal, tglAkhir)
		}
	}

	whereSQL := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT pr.noorder)
		FROM permintaan_radiologi pr
		INNER JOIN reg_periksa rp ON rp.no_rawat = pr.no_rawat
		%s
	`, whereSQL)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return make([]DetailPermintaanRadiologi, 0), 0, nil
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			pr.noorder AS no_permintaan,
			pr.no_rawat,
			DATE_FORMAT(pr.tgl_permintaan, '%%Y-%%m-%%d') AS tanggal_permintaan,
			pr.jam_permintaan,
			DATE_FORMAT(pr.tgl_sampel, '%%Y-%%m-%%d') AS tanggal_sampel,
			pr.jam_sampel,
			DATE_FORMAT(pr.tgl_hasil, '%%Y-%%m-%%d') AS tanggal_hasil,
			pr.jam_hasil,
			pr.dokter_perujuk AS kd_dokter_perujuk,
			COALESCE(d.nm_dokter, '-') AS nm_dokter_perujuk,
			CASE WHEN LOWER(pr.status) = 'ralan' THEN 'Ralan' ELSE 'Ranap' END AS status,
			pr.informasi_tambahan,
			pr.diagnosa_klinis
		FROM permintaan_radiologi pr
		INNER JOIN reg_periksa rp ON rp.no_rawat = pr.no_rawat
		LEFT JOIN dokter d ON d.kd_dokter = pr.dokter_perujuk
		%s
		ORDER BY pr.tgl_permintaan DESC, pr.jam_permintaan DESC, pr.noorder DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	argsWithPaging := append(args, filter.Limit, filter.Offset())
	rows, err := r.db.QueryContext(ctx, dataQuery, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []DetailPermintaanRadiologi
	var orderIDs []string
	orderIndexMap := make(map[string]int)

	for rows.Next() {
		var item DetailPermintaanRadiologi
		item.Pemeriksaan = make([]PemeriksaanRadiologiItem, 0)
		err := rows.Scan(
			&item.NoPermintaan,
			&item.NoRawat,
			&item.TanggalPermintaan,
			&item.JamPermintaan,
			&item.TanggalSampel,
			&item.JamSampel,
			&item.TanggalHasil,
			&item.JamHasil,
			&item.DokterPerujuk.KodeDokter,
			&item.DokterPerujuk.NamaDokter,
			&item.Status,
			&item.InformasiTambahan,
			&item.DiagnosaKlinis,
		)
		if err != nil {
			return nil, 0, err
		}

		orderIndexMap[item.NoPermintaan] = len(result)
		orderIDs = append(orderIDs, item.NoPermintaan)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(orderIDs) > 0 {
		if err := r.attachItemPemeriksaan(ctx, orderIDs, result, orderIndexMap); err != nil {
			return nil, 0, err
		}
	}

	return result, total, nil
}

func (r *repository) attachItemPemeriksaan(ctx context.Context, orderIDs []string, result []DetailPermintaanRadiologi, orderIndexMap map[string]int) error {
	placeholders := make([]string, len(orderIDs))
	argsPemeriksaan := make([]any, len(orderIDs))
	for i, id := range orderIDs {
		placeholders[i] = "?"
		argsPemeriksaan[i] = id
	}

	pemeriksaanQuery := fmt.Sprintf(`
		SELECT 
			ppr.noorder,
			ppr.kd_jenis_prw,
			COALESCE(jpr.nm_perawatan, ppr.kd_jenis_prw) AS nm_perawatan,
			COALESCE(ppr.stts_bayar, 'Belum') AS stts_bayar,
			COALESCE(jpr.total_byr, 0) AS biaya
		FROM permintaan_pemeriksaan_radiologi ppr
		LEFT JOIN jns_perawatan_radiologi jpr ON jpr.kd_jenis_prw = ppr.kd_jenis_prw
		WHERE ppr.noorder IN (%s)
		ORDER BY ppr.noorder, ppr.kd_jenis_prw
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, pemeriksaanQuery, argsPemeriksaan...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var noorder, kdJenisPrw, nmPerawatan, sttsBayar string
		var biaya float64
		if err := rows.Scan(&noorder, &kdJenisPrw, &nmPerawatan, &sttsBayar, &biaya); err != nil {
			return err
		}

		idx, ok := orderIndexMap[noorder]
		if ok {
			result[idx].Pemeriksaan = append(result[idx].Pemeriksaan, PemeriksaanRadiologiItem{
				KodeTindakan: kdJenisPrw,
				NamaTindakan: nmPerawatan,
				StatusBayar:  sttsBayar,
				Biaya:        biaya,
			})
		}
	}

	return rows.Err()
}

func (r *repository) DetailPermintaanRadiologi(ctx context.Context, noPermintaan string) (*DetailPermintaanRadiologi, error) {
	headerQuery := `
		SELECT 
			pr.noorder AS no_permintaan,
			pr.no_rawat,
			DATE_FORMAT(pr.tgl_permintaan, '%Y-%m-%d') AS tanggal_permintaan,
			pr.jam_permintaan,
			DATE_FORMAT(pr.tgl_sampel, '%Y-%m-%d') AS tanggal_sampel,
			pr.jam_sampel,
			DATE_FORMAT(pr.tgl_hasil, '%Y-%m-%d') AS tanggal_hasil,
			pr.jam_hasil,
			pr.dokter_perujuk AS kd_dokter_perujuk,
			COALESCE(d.nm_dokter, '-') AS nm_dokter_perujuk,
			CASE WHEN LOWER(pr.status) = 'ralan' THEN 'Ralan' ELSE 'Ranap' END AS status,
			pr.informasi_tambahan,
			pr.diagnosa_klinis
		FROM permintaan_radiologi pr
		LEFT JOIN dokter d ON d.kd_dokter = pr.dokter_perujuk
		WHERE pr.noorder = ?
	`

	var detail DetailPermintaanRadiologi
	err := r.db.QueryRowContext(ctx, headerQuery, noPermintaan).Scan(
		&detail.NoPermintaan,
		&detail.NoRawat,
		&detail.TanggalPermintaan,
		&detail.JamPermintaan,
		&detail.TanggalSampel,
		&detail.JamSampel,
		&detail.TanggalHasil,
		&detail.JamHasil,
		&detail.DokterPerujuk.KodeDokter,
		&detail.DokterPerujuk.NamaDokter,
		&detail.Status,
		&detail.InformasiTambahan,
		&detail.DiagnosaKlinis,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	pemeriksaanQuery := `
		SELECT 
			ppr.kd_jenis_prw,
			COALESCE(jpr.nm_perawatan, ppr.kd_jenis_prw) AS nm_perawatan,
			COALESCE(ppr.stts_bayar, 'Belum') AS stts_bayar,
			COALESCE(jpr.total_byr, 0) AS biaya
		FROM permintaan_pemeriksaan_radiologi ppr
		LEFT JOIN jns_perawatan_radiologi jpr ON jpr.kd_jenis_prw = ppr.kd_jenis_prw
		WHERE ppr.noorder = ?
		ORDER BY ppr.kd_jenis_prw
	`

	rows, err := r.db.QueryContext(ctx, pemeriksaanQuery, noPermintaan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	detail.Pemeriksaan = make([]PemeriksaanRadiologiItem, 0)
	for rows.Next() {
		var item PemeriksaanRadiologiItem
		if err := rows.Scan(&item.KodeTindakan, &item.NamaTindakan, &item.StatusBayar, &item.Biaya); err != nil {
			return nil, err
		}
		detail.Pemeriksaan = append(detail.Pemeriksaan, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &detail, nil
}

func (r *repository) UpdatePermintaanRadiologi(ctx context.Context, noPermintaan string, req SimpanPermintaanRadiologiRequest, kodeTindakanList []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateHeaderQuery := `
		UPDATE permintaan_radiologi 
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

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_radiologi WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	stmtTindakan, err := tx.PrepareContext(ctx, `INSERT INTO permintaan_pemeriksaan_radiologi (noorder, kd_jenis_prw, stts_bayar) VALUES (?, ?, 'Belum')`)
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

func (r *repository) HapusPermintaanRadiologi(ctx context.Context, noPermintaan string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_radiologi WHERE noorder = ?`, noPermintaan); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM permintaan_radiologi WHERE noorder = ?`, noPermintaan)
	if err != nil {
		return err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return err
	}

	return tx.Commit()
}
