package resep

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
	DaftarResep(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, int, error)
	DaftarResepByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, int, error)
	DetailResep(ctx context.Context, noResep string) (*Resep, error)
	SimpanResep(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanResepRequest) (*Resep, error)
	DaftarAturanPakai(ctx context.Context, keyword string) ([]AturanPakai, error)
	DaftarMetodeRacik(ctx context.Context) ([]MetodeRacik, error)
	CekKeberadaanMetodeRacik(ctx context.Context, listKodeRacik []string) (map[string]bool, error)
	HapusResep(ctx context.Context, noResep string) error
	UpdateResep(ctx context.Context, noResep string, req SimpanResepRequest) (*Resep, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) DaftarResep(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, int, error) {
	return r.queryResep(ctx, "ro.no_rawat = ?", noRawat, statusLanjut, filter)
}

func (r *repository) DaftarResepByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, int, error) {
	return r.queryResep(ctx, "rp.no_rkm_medis = ?", noRM, statusLanjut, filter)
}

func (r *repository) queryResep(ctx context.Context, whereClause string, paramValue string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, int, error) {
	tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)
	useTglFilter := tglAwal != "" && tglAkhir != ""

	statusCondition := ""
	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		statusCondition = " AND ro.status = 'ralan'"
	case shared.StatusLanjutRawatInap:
		statusCondition = " AND ro.status = 'ranap'"
	case "Semua", "":

	default:
		return nil, 0, fmt.Errorf("repository: status lanjut '%s' tidak didukung", statusLanjut)
	}

	tanggalCondition := ""
	if useTglFilter {
		tanggalCondition = " AND ro.tgl_peresepan BETWEEN ? AND ?"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM resep_obat ro
		INNER JOIN reg_periksa rp ON rp.no_rawat = ro.no_rawat
		INNER JOIN dokter d ON d.kd_dokter = ro.kd_dokter
		WHERE %s AND ro.tgl_peresepan != '0000-00-00'%s%s`, whereClause, statusCondition, tanggalCondition)

	countArgs := []any{paramValue}
	if useTglFilter {
		countArgs = append(countArgs, tglAwal, tglAkhir)
	}

	var totalData int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalData); err != nil {
		return nil, 0, err
	}

	if totalData == 0 {
		return make([]Resep, 0), 0, nil
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			ro.no_resep,
			ro.no_rawat,
			DATE_FORMAT(ro.tgl_peresepan, '%%Y-%%m-%%d') AS tanggal_peresepan,
			ro.jam_peresepan,
			ro.status,
			ro.kd_dokter,
			COALESCE(d.nm_dokter, '') AS nama_dokter
		FROM resep_obat ro
		INNER JOIN reg_periksa rp ON rp.no_rawat = ro.no_rawat
		INNER JOIN dokter d ON d.kd_dokter = ro.kd_dokter
		WHERE %s AND ro.tgl_peresepan != '0000-00-00'%s%s
		ORDER BY ro.tgl_peresepan DESC, ro.jam_peresepan DESC
		LIMIT ? OFFSET ?`, whereClause, statusCondition, tanggalCondition)

	selectArgs := []any{paramValue}
	if useTglFilter {
		selectArgs = append(selectArgs, tglAwal, tglAkhir)
	}
	selectArgs = append(selectArgs, filter.Limit, filter.Offset())

	rows, err := r.db.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var listResep []Resep
	var listNoResep []string
	resepIndexMap := make(map[string]int)

	for rows.Next() {
		var res Resep
		if err := rows.Scan(
			&res.NoResep,
			&res.NoRawat,
			&res.TanggalPeresepan,
			&res.JamPeresepan,
			&res.Status,
			&res.KodeDokter,
			&res.NamaDokter,
		); err != nil {
			return nil, 0, err
		}

		res.ResepDokter = make([]ResepDokter, 0)
		res.ResepDokterRacikan = make([]ResepDokterRacikan, 0)

		resepIndexMap[res.NoResep] = len(listResep)
		listResep = append(listResep, res)
		listNoResep = append(listNoResep, res.NoResep)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(listNoResep) == 0 {
		return listResep, totalData, nil
	}

	if err := r.loadResepDokter(ctx, listNoResep, listResep, resepIndexMap); err != nil {
		return nil, 0, err
	}

	if err := r.loadResepRacikan(ctx, listNoResep, listResep, resepIndexMap); err != nil {
		return nil, 0, err
	}

	return listResep, totalData, nil
}

func (r *repository) loadResepDokter(ctx context.Context, listNoResep []string, listResep []Resep, resepIndexMap map[string]int) error {
	if len(listNoResep) == 0 {
		return nil
	}
	inPlaceholders := shared.CreateInPlaceholders(len(listNoResep))
	args := make([]any, len(listNoResep))
	for i, nr := range listNoResep {
		args[i] = nr
	}

	query := fmt.Sprintf(`
		SELECT 
			rd.no_resep,
			rd.kode_brng AS kode_obat,
			COALESCE(dtb.nama_brng, '') AS nama_obat,
			rd.jml,
			COALESCE(dtb.kode_sat, '') AS satuan,
			rd.aturan_pakai
		FROM resep_dokter rd
		LEFT JOIN databarang dtb ON rd.kode_brng = dtb.kode_brng
		WHERE rd.no_resep IN (%s)`, inPlaceholders)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rd ResepDokter
		if err := rows.Scan(
			&rd.NoResep,
			&rd.KodeObat,
			&rd.NamaObat,
			&rd.Jumlah,
			&rd.Satuan,
			&rd.AturanPakai,
		); err != nil {
			return err
		}

		if idx, exists := resepIndexMap[rd.NoResep]; exists {
			listResep[idx].ResepDokter = append(listResep[idx].ResepDokter, rd)
		}
	}

	return rows.Err()
}

func (r *repository) loadResepRacikan(ctx context.Context, listNoResep []string, listResep []Resep, resepIndexMap map[string]int) error {
	if len(listNoResep) == 0 {
		return nil
	}
	inPlaceholders := shared.CreateInPlaceholders(len(listNoResep))
	args := make([]any, len(listNoResep))
	for i, nr := range listNoResep {
		args[i] = nr
	}

	queryRacik := fmt.Sprintf(`
		SELECT 
			rdr.no_resep,
			rdr.no_racik,
			rdr.nama_racik,
			rdr.kd_racik,
			COALESCE(mr.nm_racik, '') AS metode_racik,
			rdr.jml_dr,
			rdr.aturan_pakai,
			rdr.keterangan
		FROM resep_dokter_racikan rdr
		LEFT JOIN metode_racik mr ON rdr.kd_racik = mr.kd_racik
		WHERE rdr.no_resep IN (%s)
		ORDER BY rdr.no_resep ASC, CAST(rdr.no_racik AS UNSIGNED) ASC, rdr.no_racik ASC`, inPlaceholders)

	rows, err := r.db.QueryContext(ctx, queryRacik, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	type racikanPos struct {
		resepIdx int
		racikIdx int
	}
	racikanMap := make(map[string]racikanPos)

	for rows.Next() {
		var rdr ResepDokterRacikan
		if err := rows.Scan(
			&rdr.NoResep,
			&rdr.NoRacik,
			&rdr.NamaRacik,
			&rdr.KodeMetodeRacik,
			&rdr.NamaMetodeRacik,
			&rdr.JumlahRacikan,
			&rdr.AturanPakai,
			&rdr.Keterangan,
		); err != nil {
			return err
		}
		rdr.DetailRacikan = make([]ResepDokterRacikanDetail, 0)

		if idx, exists := resepIndexMap[rdr.NoResep]; exists {
			listResep[idx].ResepDokterRacikan = append(listResep[idx].ResepDokterRacikan, rdr)
			lastIdx := len(listResep[idx].ResepDokterRacikan) - 1
			key := rdr.NoResep + "\x00" + rdr.NoRacik
			racikanMap[key] = racikanPos{
				resepIdx: idx,
				racikIdx: lastIdx,
			}
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(racikanMap) == 0 {
		return nil
	}

	queryDetail := fmt.Sprintf(`
		SELECT 
			rdrd.no_resep,
			rdrd.no_racik,
			rdrd.kode_brng,
			COALESCE(dtb.nama_brng, '') AS nama_obat,
			rdrd.kandungan,
			rdrd.jml,
			COALESCE(dtb.kode_sat, '') AS satuan
		FROM resep_dokter_racikan_detail rdrd
		LEFT JOIN databarang dtb ON rdrd.kode_brng = dtb.kode_brng
		WHERE rdrd.no_resep IN (%s)
		ORDER BY rdrd.no_resep ASC, CAST(rdrd.no_racik AS UNSIGNED) ASC, rdrd.no_racik ASC, rdrd.kode_brng ASC`, inPlaceholders)

	detailRows, err := r.db.QueryContext(ctx, queryDetail, args...)
	if err != nil {
		return err
	}
	defer detailRows.Close()

	for detailRows.Next() {
		var d ResepDokterRacikanDetail
		if err := detailRows.Scan(
			&d.NoResep,
			&d.NoRacik,
			&d.KodeObat,
			&d.NamaObat,
			&d.Kandungan,
			&d.Jumlah,
			&d.Satuan,
		); err != nil {
			return err
		}

		key := d.NoResep + "\x00" + d.NoRacik
		if pos, exists := racikanMap[key]; exists {
			racik := &listResep[pos.resepIdx].ResepDokterRacikan[pos.racikIdx]
			racik.DetailRacikan = append(racik.DetailRacikan, d)
		}
	}

	return detailRows.Err()
}

func (r *repository) DaftarAturanPakai(ctx context.Context, keyword string) ([]AturanPakai, error) {
	cleanKeyword := strings.ReplaceAll(strings.TrimSpace(keyword), " ", "")
	query := "SELECT aturan FROM master_aturan_pakai"
	var args []any

	if cleanKeyword != "" {
		query += " WHERE REPLACE(aturan, ' ', '') LIKE ?"
		args = append(args, "%"+cleanKeyword+"%")
	}

	query += " ORDER BY aturan ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]AturanPakai, 0)
	for rows.Next() {
		var ap AturanPakai
		if err := rows.Scan(&ap.AturanPakai); err != nil {
			return nil, err
		}
		list = append(list, ap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarMetodeRacik(ctx context.Context) ([]MetodeRacik, error) {
	query := "SELECT kd_racik AS kode, nm_racik AS nama FROM metode_racik ORDER BY nm_racik ASC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]MetodeRacik, 0)
	for rows.Next() {
		var mr MetodeRacik
		if err := rows.Scan(&mr.Kode, &mr.Nama); err != nil {
			return nil, err
		}
		list = append(list, mr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) CekKeberadaanMetodeRacik(ctx context.Context, listKodeRacik []string) (map[string]bool, error) {
	uniqueCodes := make([]string, 0, len(listKodeRacik))
	seen := make(map[string]bool)
	for _, code := range listKodeRacik {
		trimmed := strings.TrimSpace(code)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			uniqueCodes = append(uniqueCodes, trimmed)
		}
	}

	if len(uniqueCodes) == 0 {
		return map[string]bool{}, nil
	}

	inPlaceholders := shared.CreateInPlaceholders(len(uniqueCodes))
	args := make([]any, len(uniqueCodes))
	for i, c := range uniqueCodes {
		args[i] = c
	}

	query := fmt.Sprintf("SELECT kd_racik FROM metode_racik WHERE kd_racik IN (%s)", inPlaceholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	foundMap := make(map[string]bool)
	for rows.Next() {
		var kode string
		if err := rows.Scan(&kode); err != nil {
			return nil, err
		}
		foundMap[kode] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return foundMap, nil
}

func (r *repository) DetailResep(ctx context.Context, noResep string) (*Resep, error) {
	query := `
		SELECT 
			ro.no_resep,
			ro.no_rawat,
			DATE_FORMAT(ro.tgl_peresepan, '%Y-%m-%d') AS tanggal_peresepan,
			ro.jam_peresepan,
			DATE_FORMAT(ro.tgl_perawatan, '%Y-%m-%d') AS tanggal_perawatan,
			ro.jam AS jam_perawatan,
			DATE_FORMAT(ro.tgl_penyerahan, '%Y-%m-%d') AS tanggal_penyerahan,
			ro.jam_penyerahan,
			ro.status,
			ro.kd_dokter,
			COALESCE(d.nm_dokter, '') AS nama_dokter
		FROM resep_obat ro
		INNER JOIN dokter d ON d.kd_dokter = ro.kd_dokter
		WHERE ro.no_resep = ?`

	var res Resep
	err := r.db.QueryRowContext(ctx, query, noResep).Scan(
		&res.NoResep,
		&res.NoRawat,
		&res.TanggalPeresepan,
		&res.JamPeresepan,
		&res.TanggalPerawatan,
		&res.JamPerawatan,
		&res.TanggalPenyerahan,
		&res.JamPenyerahan,
		&res.Status,
		&res.KodeDokter,
		&res.NamaDokter,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	res.ResepDokter = make([]ResepDokter, 0)
	res.ResepDokterRacikan = make([]ResepDokterRacikan, 0)

	listResep := []Resep{res}
	resepIndexMap := map[string]int{res.NoResep: 0}

	if err := r.loadResepDokter(ctx, []string{res.NoResep}, listResep, resepIndexMap); err != nil {
		return nil, err
	}
	if err := r.loadResepRacikan(ctx, []string{res.NoResep}, listResep, resepIndexMap); err != nil {
		return nil, err
	}

	return &listResep[0], nil
}

func (r *repository) HapusResep(ctx context.Context, noResep string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM resep_dokter_racikan_detail WHERE no_resep = ?", noResep); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM resep_dokter_racikan WHERE no_resep = ?", noResep); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM resep_dokter WHERE no_resep = ?", noResep); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, "DELETE FROM resep_obat WHERE no_resep = ?", noResep)
	if err != nil {
		return err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) UpdateResep(ctx context.Context, noResep string, req SimpanResepRequest) (*Resep, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	updateQuery := `UPDATE resep_obat SET tgl_peresepan = ?, jam_peresepan = ? WHERE no_resep = ?`
	res, err := tx.ExecContext(ctx, updateQuery, req.TanggalPeresepan, req.JamPeresepan, noResep)
	if err != nil {
		return nil, err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM resep_dokter_racikan_detail WHERE no_resep = ?", noResep); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM resep_dokter_racikan WHERE no_resep = ?", noResep); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM resep_dokter WHERE no_resep = ?", noResep); err != nil {
		return nil, err
	}

	if err := r.insertResepDokter(ctx, tx, noResep, req.ResepDokter); err != nil {
		return nil, err
	}
	if err := r.insertResepRacikan(ctx, tx, noResep, req.ResepRacikan); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.DetailResep(ctx, noResep)
}

func (r *repository) generateNoResep(ctx context.Context, tx *sql.Tx, tglPeresepan string) (string, error) {

	parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(tglPeresepan))
	prefix := time.Now().Format("20060102")
	if err == nil {
		prefix = parsedDate.Format("20060102")
	}

	minOrder := prefix + "0000"
	maxOrder := prefix + "9999"

	query := `SELECT no_resep FROM resep_obat WHERE no_resep BETWEEN ? AND ? ORDER BY no_resep DESC LIMIT 1 FOR UPDATE`
	var lastToday string
	err = tx.QueryRowContext(ctx, query, minOrder, maxOrder).Scan(&lastToday)
	if errors.Is(err, sql.ErrNoRows) {
		return prefix + "0001", nil
	}
	if err != nil {
		return "", err
	}

	if len(lastToday) != 12 || !strings.HasPrefix(lastToday, prefix) {
		return time.Now().Format("20060102150405"), nil
	}

	tail := lastToday[len(lastToday)-4:]
	next, err := strconv.Atoi(tail)
	if err != nil || next >= 9999 {
		return time.Now().Format("20060102150405"), nil
	}

	return fmt.Sprintf("%s%04d", prefix, next+1), nil
}

func (r *repository) SimpanResep(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanResepRequest) (*Resep, error) {
	var dbStatus string
	switch statusLanjut {
	case shared.StatusLanjutRawatJalan:
		dbStatus = "ralan"
	case shared.StatusLanjutRawatInap:
		dbStatus = "ranap"
	default:
		return nil, fmt.Errorf("repository: status lanjut '%s' tidak didukung", statusLanjut)
	}

	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}

		noResep, err := r.generateNoResep(ctx, tx, req.TanggalPeresepan)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if err := r.insertHeader(ctx, tx, noResep, req, kodeDokter, dbStatus); err != nil {
			_ = tx.Rollback()
			if isDuplicateKey(err) {
				lastErr = err
				continue
			}
			return nil, err
		}

		if err := r.insertResepDokter(ctx, tx, noResep, req.ResepDokter); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if err := r.insertResepRacikan(ctx, tx, noResep, req.ResepRacikan); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			if isDuplicateKey(err) {
				lastErr = err
				continue
			}
			return nil, err
		}

		return r.DetailResep(ctx, noResep)
	}

	return nil, fmt.Errorf("gagal mendapatkan nomor resep unik setelah 3 kali percobaan: %w", lastErr)
}

func (r *repository) insertHeader(ctx context.Context, tx *sql.Tx, noResep string, req SimpanResepRequest, kodeDokter string, dbStatus string) error {
	query := `
		INSERT INTO resep_obat (
			no_resep, tgl_perawatan, jam, no_rawat, kd_dokter,
			tgl_peresepan, jam_peresepan, status, tgl_penyerahan, jam_penyerahan
		) VALUES (?, '0000-00-00', '00:00:00', ?, ?, ?, ?, ?, '0000-00-00', '00:00:00')`

	_, err := tx.ExecContext(ctx, query, noResep, req.NoRawat, kodeDokter, req.TanggalPeresepan, req.JamPeresepan, dbStatus)
	return err
}

func (r *repository) insertResepDokter(ctx context.Context, tx *sql.Tx, noResep string, items []ResepDokterInput) error {
	if len(items) == 0 {
		return nil
	}

	query := `INSERT INTO resep_dokter (no_resep, kode_brng, jml, aturan_pakai) VALUES (?, ?, ?, ?)`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rd := range items {
		if _, err := stmt.ExecContext(ctx, noResep, rd.KodeObat, rd.Jumlah, rd.AturanPakai); err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) insertResepRacikan(ctx context.Context, tx *sql.Tx, noResep string, items []ResepRacikanInput) error {
	if len(items) == 0 {
		return nil
	}

	queryRacik := `INSERT INTO resep_dokter_racikan (no_resep, no_racik, nama_racik, kd_racik, jml_dr, aturan_pakai, keterangan) VALUES (?, ?, ?, ?, ?, ?, ?)`
	stmtRacik, err := tx.PrepareContext(ctx, queryRacik)
	if err != nil {
		return err
	}
	defer stmtRacik.Close()

	queryDetail := `INSERT INTO resep_dokter_racikan_detail (no_resep, no_racik, kode_brng, p1, p2, kandungan, jml) VALUES (?, ?, ?, 1, 1, ?, ?)`
	stmtDetail, err := tx.PrepareContext(ctx, queryDetail)
	if err != nil {
		return err
	}
	defer stmtDetail.Close()

	for i, rr := range items {
		noRacik := i + 1
		if _, err := stmtRacik.ExecContext(ctx, noResep, noRacik, rr.NamaRacik, rr.KodeRacik, rr.JumlahRacikan, rr.AturanPakai, rr.Keterangan); err != nil {
			return err
		}

		for _, d := range rr.Detail {
			if _, err := stmtDetail.ExecContext(ctx, noResep, noRacik, d.KodeObat, d.Kandungan, d.Jumlah); err != nil {
				return err
			}
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
