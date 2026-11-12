package rawatjalan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"erm-dokter/internal/pkg/database"
	"erm-dokter/internal/shared/formatter"
)

type Repository interface {
	DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter FilterAntreanDokter) ([]KunjunganRawatJalan, int, error)
	DetailKunjungan(ctx context.Context, noRawat string, kodeDokter string) (*KunjunganRawatJalan, error)
	RiwayatKunjunganPasien(ctx context.Context, noRM string) ([]KunjunganRawatJalan, error)
	GetWaktuRegistrasi(ctx context.Context, noRawat string) (tanggal string, jam string, exists bool, err error)
	GetInfoRegistrasi(ctx context.Context, noRawat string) (*InfoRegistrasiPasien, error)
	GetKodePoliByNoRawat(ctx context.Context, noRawat string, kodeDokter string) (string, error)
	CekAksesKunjungan(ctx context.Context, noRawat string, kodeDokter string) (exists bool, authorized bool, err error)
	UpdateStatusPemeriksaan(ctx context.Context, noRawat string, status StatusPemeriksaan) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

const globalNoRawatWindowMonths = 3

const colsKunjunganBukanRujukan = `
	SELECT 
		r.no_rawat,
		r.no_reg AS no_registrasi,
		DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d') AS tanggal_registrasi,
		r.jam_reg AS jam_registrasi,
		r.no_rkm_medis AS no_rekam_medis,
		p.nm_pasien AS nama_pasien,
		p.jk AS jenis_kelamin,
		DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d') AS tanggal_lahir,
		r.almt_pj AS alamat,
		r.kd_poli AS kode_poli_asal,
		pol.nm_poli AS nama_poli_asal,
		r.kd_dokter AS kode_dokter_asal,
		d.nm_dokter AS nama_dokter_asal,
		'' AS kode_poli_rujukan,
		'' AS nama_poli_rujukan,
		'' AS kode_dokter_rujukan,
		'' AS nama_dokter_rujukan,
		r.kd_pj AS kode_penjamin,
		pj.png_jawab AS nama_penjamin,
		r.stts AS status_pemeriksaan,
		r.status_lanjut,
		r.status_bayar,
		'Bukan Rujukan' AS jenis_antrean,
		COALESCE(p.gol_darah, '-') AS golongan_darah,
		COALESCE(p.agama, '-') AS agama,
		COALESCE(p.no_tlp, '-') AS no_telepon,
		COALESCE(p.no_peserta, '-') AS no_peserta,
		COALESCE(p.no_ktp, '-') AS no_ktp,
		COALESCE(r.p_jawab, '-') AS penanggung_jawab,
		COALESCE(r.hubunganpj, '-') AS hubungan_penanggung_jawab,
		COALESCE(r.almt_pj, '-') AS alamat_penanggung_jawab`

const selectKunjunganBukanRujukan = colsKunjunganBukanRujukan + `
	FROM reg_periksa r
	INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis
	INNER JOIN poliklinik pol ON r.kd_poli = pol.kd_poli
	INNER JOIN dokter d ON r.kd_dokter = d.kd_dokter
	INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj
`

const colsKunjunganRujukan = `
	SELECT 
		r.no_rawat,
		r.no_reg AS no_registrasi,
		DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d') AS tanggal_registrasi,
		r.jam_reg AS jam_registrasi,
		r.no_rkm_medis AS no_rekam_medis,
		p.nm_pasien AS nama_pasien,
		p.jk AS jenis_kelamin,
		DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d') AS tanggal_lahir,
		r.almt_pj AS alamat,
		r.kd_poli AS kode_poli_asal,
		pol.nm_poli AS nama_poli_asal,
		r.kd_dokter AS kode_dokter_asal,
		d.nm_dokter AS nama_dokter_asal,
		rip.kd_poli AS kode_poli_rujukan,
		pol_rip.nm_poli AS nama_poli_rujukan,
		rip.kd_dokter AS kode_dokter_rujukan,
		d_rip.nm_dokter AS nama_dokter_rujukan,
		r.kd_pj AS kode_penjamin,
		pj.png_jawab AS nama_penjamin,
		r.stts AS status_pemeriksaan,
		r.status_lanjut,
		r.status_bayar,
		'Rujukan' AS jenis_antrean,
		COALESCE(p.gol_darah, '-') AS golongan_darah,
		COALESCE(p.agama, '-') AS agama,
		COALESCE(p.no_tlp, '-') AS no_telepon,
		COALESCE(p.no_peserta, '-') AS no_peserta,
		COALESCE(p.no_ktp, '-') AS no_ktp,
		COALESCE(r.p_jawab, '-') AS penanggung_jawab,
		COALESCE(r.hubunganpj, '-') AS hubungan_penanggung_jawab,
		COALESCE(r.almt_pj, '-') AS alamat_penanggung_jawab`

const selectKunjunganRujukan = colsKunjunganRujukan + `
	FROM rujukan_internal_poli rip
	INNER JOIN reg_periksa r ON rip.no_rawat = r.no_rawat
	INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis
	INNER JOIN poliklinik pol ON r.kd_poli = pol.kd_poli
	INNER JOIN dokter d ON r.kd_dokter = d.kd_dokter
	INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj
	INNER JOIN poliklinik pol_rip ON rip.kd_poli = pol_rip.kd_poli
	INNER JOIN dokter d_rip ON rip.kd_dokter = d_rip.kd_dokter
`

func normalizeAntreanSort(orderBy, sortOrder string) (string, string) {
	col := strings.ToLower(strings.TrimSpace(orderBy))
	if col == "" {
		col = "waktu_registrasi"
	}
	dir := "ASC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "DESC") {
		dir = "DESC"
	}
	return col, dir
}

func buildAntreanOrderClause(orderBy, sortOrder string, isRujukan bool) string {
	tieTail := ""
	if isRujukan {
		tieTail = ", rip.kd_dokter ASC"
	}

	switch orderBy {
	case "nama_pasien":
		return fmt.Sprintf("p.nm_pasien %s, r.tgl_registrasi ASC, r.jam_reg ASC, r.no_rawat ASC%s", sortOrder, tieTail)
	case "penjamin":
		return fmt.Sprintf("pj.png_jawab %s, r.tgl_registrasi ASC, r.jam_reg ASC, r.no_rawat ASC%s", sortOrder, tieTail)
	case "status_pemeriksaan":
		return fmt.Sprintf("r.stts %s, r.tgl_registrasi ASC, r.jam_reg ASC, r.no_rawat ASC%s", sortOrder, tieTail)
	case "status_lanjut":
		return fmt.Sprintf("r.status_lanjut %s, r.tgl_registrasi ASC, r.jam_reg ASC, r.no_rawat ASC%s", sortOrder, tieTail)
	case "status_bayar":
		return fmt.Sprintf("r.status_bayar %s, r.tgl_registrasi ASC, r.jam_reg ASC, r.no_rawat ASC%s", sortOrder, tieTail)
	case "jenis_antrean":
		return fmt.Sprintf("r.tgl_registrasi ASC, r.jam_reg ASC, r.no_rawat ASC%s", tieTail)
	default:
		return fmt.Sprintf("r.tgl_registrasi %s, r.jam_reg %s, r.no_rawat %s%s", sortOrder, sortOrder, sortOrder, tieTail)
	}
}

func compareStringsFold(a, b string) int {
	if cmp := strings.Compare(strings.ToLower(a), strings.ToLower(b)); cmp != 0 {
		return cmp
	}
	return strings.Compare(a, b)
}

func compareTieBreaker(a, b KunjunganRawatJalan) int {
	if cmp := strings.Compare(a.TanggalRegistrasi, b.TanggalRegistrasi); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(a.JamRegistrasi, b.JamRegistrasi); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(a.NoRawat, b.NoRawat); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(string(a.JenisAntrean), string(b.JenisAntrean)); cmp != 0 {
		return cmp
	}
	return strings.Compare(a.KodeDokterRujukan, b.KodeDokterRujukan)
}

func sortAntrean(data []KunjunganRawatJalan, orderBy string, sortOrder string) {
	col, dir := normalizeAntreanSort(orderBy, sortOrder)
	asc := dir == "ASC"
	sort.SliceStable(data, func(i, j int) bool {
		a, b := data[i], data[j]
		var result int
		switch col {
		case "nama_pasien":
			result = compareStringsFold(a.NamaPasien, b.NamaPasien)
		case "penjamin":
			result = compareStringsFold(a.NamaPenjamin, b.NamaPenjamin)
		case "status_pemeriksaan":
			result = compareStringsFold(string(a.StatusPemeriksaan), string(b.StatusPemeriksaan))
		case "status_lanjut":
			result = compareStringsFold(string(a.StatusLanjut), string(b.StatusLanjut))
		case "status_bayar":
			result = compareStringsFold(string(a.StatusBayar), string(b.StatusBayar))
		case "jenis_antrean":
			result = compareStringsFold(string(a.JenisAntrean), string(b.JenisAntrean))
		default:
			result = strings.Compare(a.TanggalRegistrasi, b.TanggalRegistrasi)
			if result == 0 {
				result = strings.Compare(a.JamRegistrasi, b.JamRegistrasi)
			}
			if result == 0 {
				result = strings.Compare(a.NoRawat, b.NoRawat)
			}
		}
		if result == 0 {
			return compareTieBreaker(a, b) < 0
		}
		if asc {
			return result < 0
		}
		return result > 0
	})
}

type scanner interface {
	Scan(dest ...any) error
}

func scanKunjungan(s scanner) (*KunjunganRawatJalan, error) {
	var k KunjunganRawatJalan
	err := s.Scan(
		&k.NoRawat,
		&k.NoRegistrasi,
		&k.TanggalRegistrasi,
		&k.JamRegistrasi,
		&k.NoRekamMedis,
		&k.NamaPasien,
		&k.JenisKelamin,
		&k.TanggalLahir,
		&k.Alamat,
		&k.KodePoliAsal,
		&k.NamaPoliAsal,
		&k.KodeDokterAsal,
		&k.NamaDokterAsal,
		&k.KodePoliRujukan,
		&k.NamaPoliRujukan,
		&k.KodeDokterRujukan,
		&k.NamaDokterRujukan,
		&k.KodePenjamin,
		&k.NamaPenjamin,
		&k.StatusPemeriksaan,
		&k.StatusLanjut,
		&k.StatusBayar,
		&k.JenisAntrean,
		&k.GolonganDarah,
		&k.Agama,
		&k.NoTelepon,
		&k.NoPeserta,
		&k.NoKTP,
		&k.PenanggungJawab,
		&k.HubunganPenanggungJawab,
		&k.AlamatPenanggungJawab,
	)
	if err != nil {
		return nil, err
	}
	k.JenisKelamin = k.FormatJenisKelamin()
	k.Umur = k.FormatUmur()
	return &k, nil
}

func buildNoRawatBounds(tglAwal, tglAkhir string) (string, string, bool) {
	start, err1 := time.Parse("2006-01-02", tglAwal)
	end, err2 := time.Parse("2006-01-02", tglAkhir)
	if err1 != nil || err2 != nil || start.After(end) {
		return "", "", false
	}
	lower := start.AddDate(0, -globalNoRawatWindowMonths, 0).Format("2006/01/02") + "/"
	upper := end.AddDate(0, globalNoRawatWindowMonths, 1).Format("2006/01/02") + "/"
	return lower, upper, true
}

func buildBranchConditions(noRawatCol string, dokterCol string, kodeDokter string, filter FilterAntreanDokter) (string, []any) {
	var conditions []string
	var args []any

	if kodeDokter != "" {
		conditions = append(conditions, dokterCol+" = ?")
		args = append(args, kodeDokter)
	}

	if filter.Tanggal != "" {
		tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)
		if tglAwal != "" && tglAkhir != "" {
			if kodeDokter == "" && noRawatCol != "" {
				if lower, upper, ok := buildNoRawatBounds(tglAwal, tglAkhir); ok {
					conditions = append(
						conditions,
						noRawatCol+" >= CONVERT(? USING latin1)",
						noRawatCol+" < CONVERT(? USING latin1)",
					)
					args = append(args, lower, upper)
				}
			}
			conditions = append(conditions, "r.tgl_registrasi BETWEEN ? AND ?")
			args = append(args, tglAwal, tglAkhir)
		}
	}

	if filter.Penjamin != "" {
		conditions = append(conditions, "r.kd_pj = ?")
		args = append(args, filter.Penjamin)
	}

	if filter.StatusPemeriksaan != "" {
		conditions = append(conditions, "r.stts = ?")
		args = append(args, filter.StatusPemeriksaan)
	}

	if filter.StatusLanjut != "" {
		conditions = append(conditions, "r.status_lanjut = ?")
		args = append(args, filter.StatusLanjut)
	}

	if filter.Keyword != "" {
		conditions = append(conditions, "(p.nm_pasien LIKE ? OR r.no_rkm_medis LIKE ? OR r.no_rawat LIKE ?)")
		keywordPattern := "%" + filter.Keyword + "%"
		args = append(args, keywordPattern, keywordPattern, keywordPattern)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *repository) countBukanRujukan(ctx context.Context, kodeDokter string, filter FilterAntreanDokter) (int, error) {
	from := "FROM reg_periksa r"
	if filter.Keyword != "" {
		from += " INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis"
	}
	where, args := buildBranchConditions("r.no_rawat", "r.kd_dokter", kodeDokter, filter)
	query := "SELECT COUNT(*) " + from + where
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) countRujukan(ctx context.Context, kodeDokter string, filter FilterAntreanDokter) (int, error) {
	from := "FROM rujukan_internal_poli rip INNER JOIN reg_periksa r ON rip.no_rawat = r.no_rawat"
	if filter.Keyword != "" {
		from += " INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis"
	}
	where, args := buildBranchConditions("rip.no_rawat", "rip.kd_dokter", kodeDokter, filter)
	query := "SELECT COUNT(*) " + from + where
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func buildBaseQuery(kodeDokter string, filter FilterAntreanDokter) (string, []any) {
	switch filter.JenisAntrean {
	case JenisAntreanTidakRujukan:
		where, args := buildBranchConditions("", "r.kd_dokter", kodeDokter, filter)
		return selectKunjunganBukanRujukan + where, args

	case JenisAntreanRujukan:
		where, args := buildBranchConditions("", "rip.kd_dokter", kodeDokter, filter)
		return selectKunjunganRujukan + where, args

	default:
		where1, args1 := buildBranchConditions("", "r.kd_dokter", kodeDokter, filter)
		where2, args2 := buildBranchConditions("", "rip.kd_dokter", kodeDokter, filter)

		query1 := selectKunjunganBukanRujukan + where1
		query2 := selectKunjunganRujukan + where2

		combinedQuery := fmt.Sprintf("%s UNION ALL %s", query1, query2)
		combinedArgs := append(args1, args2...)
		return combinedQuery, combinedArgs
	}
}

func (r *repository) scanAntreanRows(ctx context.Context, query string, args []any) ([]KunjunganRawatJalan, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal query data antrean: %w", err)
	}
	defer rows.Close()

	result := make([]KunjunganRawatJalan, 0)
	for rows.Next() {
		kunjungan, err := scanKunjungan(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal scan data antrean: %w", err)
		}
		result = append(result, *kunjungan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error saat iterasi data antrean: %w", err)
	}

	return result, nil
}

func (r *repository) queryBukanRujukanPage(
	ctx context.Context,
	kodeDokter string,
	filter FilterAntreanDokter,
	orderBy string,
	sortOrder string,
	limit int,
	offset int,
) ([]KunjunganRawatJalan, error) {
	innerFrom := "FROM reg_periksa r"
	if filter.Keyword != "" || orderBy == "nama_pasien" {
		innerFrom += " INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis"
	}
	if orderBy == "penjamin" {
		innerFrom += " INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj"
	}

	where, args := buildBranchConditions("r.no_rawat", "r.kd_dokter", kodeDokter, filter)
	orderClause := buildAntreanOrderClause(orderBy, sortOrder, false)

	query := fmt.Sprintf(`%s
	FROM (
		SELECT r.no_rawat
		%s%s
		ORDER BY %s
		LIMIT ? OFFSET ?
	) AS page_r
	INNER JOIN reg_periksa r ON page_r.no_rawat = r.no_rawat
	INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis
	INNER JOIN poliklinik pol ON r.kd_poli = pol.kd_poli
	INNER JOIN dokter d ON r.kd_dokter = d.kd_dokter
	INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj
	ORDER BY %s`, colsKunjunganBukanRujukan, innerFrom, where, orderClause, orderClause)

	pageArgs := append(append([]any(nil), args...), limit, offset)
	return r.scanAntreanRows(ctx, query, pageArgs)
}

func (r *repository) queryRujukanPage(
	ctx context.Context,
	kodeDokter string,
	filter FilterAntreanDokter,
	orderBy string,
	sortOrder string,
	limit int,
	offset int,
) ([]KunjunganRawatJalan, error) {
	innerFrom := "FROM rujukan_internal_poli rip INNER JOIN reg_periksa r ON rip.no_rawat = r.no_rawat"
	if filter.Keyword != "" || orderBy == "nama_pasien" {
		innerFrom += " INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis"
	}
	if orderBy == "penjamin" {
		innerFrom += " INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj"
	}

	where, args := buildBranchConditions("rip.no_rawat", "rip.kd_dokter", kodeDokter, filter)
	orderClause := buildAntreanOrderClause(orderBy, sortOrder, true)

	query := fmt.Sprintf(`%s
	FROM (
		SELECT rip.no_rawat, rip.kd_dokter
		%s%s
		ORDER BY %s
		LIMIT ? OFFSET ?
	) AS page_rip
	INNER JOIN rujukan_internal_poli rip ON page_rip.no_rawat = rip.no_rawat AND page_rip.kd_dokter = rip.kd_dokter
	INNER JOIN reg_periksa r ON rip.no_rawat = r.no_rawat
	INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis
	INNER JOIN poliklinik pol ON r.kd_poli = pol.kd_poli
	INNER JOIN dokter d ON r.kd_dokter = d.kd_dokter
	INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj
	INNER JOIN poliklinik pol_rip ON rip.kd_poli = pol_rip.kd_poli
	INNER JOIN dokter d_rip ON rip.kd_dokter = d_rip.kd_dokter
	ORDER BY %s`, colsKunjunganRujukan, innerFrom, where, orderClause, orderClause)

	pageArgs := append(append([]any(nil), args...), limit, offset)
	return r.scanAntreanRows(ctx, query, pageArgs)
}

func (r *repository) DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter FilterAntreanDokter) ([]KunjunganRawatJalan, int, error) {
	orderBy, sortOrder := normalizeAntreanSort(filter.OrderBy, filter.SortOrder)
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset()
	if offset < 0 {
		offset = 0
	}

	switch filter.JenisAntrean {
	case JenisAntreanTidakRujukan:
		var (
			totalData     int
			daftarAntrean []KunjunganRawatJalan
			countErr      error
			dataErr       error
			wg            sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			totalData, countErr = r.countBukanRujukan(ctx, kodeDokter, filter)
		}()
		go func() {
			defer wg.Done()
			daftarAntrean, dataErr = r.queryBukanRujukanPage(ctx, kodeDokter, filter, orderBy, sortOrder, limit, offset)
		}()
		wg.Wait()

		if countErr != nil {
			return nil, 0, fmt.Errorf("gagal menghitung total antrean: %w", countErr)
		}
		if dataErr != nil {
			return nil, 0, dataErr
		}
		if totalData == 0 {
			return make([]KunjunganRawatJalan, 0), 0, nil
		}
		return daftarAntrean, totalData, nil

	case JenisAntreanRujukan:
		var (
			totalData     int
			daftarAntrean []KunjunganRawatJalan
			countErr      error
			dataErr       error
			wg            sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			totalData, countErr = r.countRujukan(ctx, kodeDokter, filter)
		}()
		go func() {
			defer wg.Done()
			daftarAntrean, dataErr = r.queryRujukanPage(ctx, kodeDokter, filter, orderBy, sortOrder, limit, offset)
		}()
		wg.Wait()

		if countErr != nil {
			return nil, 0, fmt.Errorf("gagal menghitung total antrean: %w", countErr)
		}
		if dataErr != nil {
			return nil, 0, dataErr
		}
		if totalData == 0 {
			return make([]KunjunganRawatJalan, 0), 0, nil
		}
		return daftarAntrean, totalData, nil

	default:
		var (
			countNonRujukan int
			countRujukan    int
			errNonRujukan   error
			errRujukan      error
			countWG         sync.WaitGroup
		)
		countWG.Add(2)
		go func() {
			defer countWG.Done()
			countNonRujukan, errNonRujukan = r.countBukanRujukan(ctx, kodeDokter, filter)
		}()
		go func() {
			defer countWG.Done()
			countRujukan, errRujukan = r.countRujukan(ctx, kodeDokter, filter)
		}()
		countWG.Wait()

		if errNonRujukan != nil {
			return nil, 0, fmt.Errorf("gagal menghitung antrean bukan rujukan: %w", errNonRujukan)
		}
		if errRujukan != nil {
			return nil, 0, fmt.Errorf("gagal menghitung antrean rujukan: %w", errRujukan)
		}

		totalData := countNonRujukan + countRujukan
		if totalData == 0 {
			return make([]KunjunganRawatJalan, 0), 0, nil
		}

		if countRujukan == 0 {
			daftarAntrean, err := r.queryBukanRujukanPage(ctx, kodeDokter, filter, orderBy, sortOrder, limit, offset)
			if err != nil {
				return nil, 0, err
			}
			return daftarAntrean, totalData, nil
		}

		if countNonRujukan == 0 {
			daftarAntrean, err := r.queryRujukanPage(ctx, kodeDokter, filter, orderBy, sortOrder, limit, offset)
			if err != nil {
				return nil, 0, err
			}
			return daftarAntrean, totalData, nil
		}

		innerLimit := offset + limit
		var (
			branch1 []KunjunganRawatJalan
			branch2 []KunjunganRawatJalan
			err1    error
			err2    error
			dataWG  sync.WaitGroup
		)
		dataWG.Add(2)
		go func() {
			defer dataWG.Done()
			branch1, err1 = r.queryBukanRujukanPage(ctx, kodeDokter, filter, orderBy, sortOrder, innerLimit, 0)
		}()
		go func() {
			defer dataWG.Done()
			branch2, err2 = r.queryRujukanPage(ctx, kodeDokter, filter, orderBy, sortOrder, innerLimit, 0)
		}()
		dataWG.Wait()

		if err1 != nil {
			return nil, 0, err1
		}
		if err2 != nil {
			return nil, 0, err2
		}

		merged := append(branch1, branch2...)
		sortAntrean(merged, orderBy, sortOrder)

		start := offset
		end := start + limit
		if start >= len(merged) {
			return make([]KunjunganRawatJalan, 0), totalData, nil
		}
		if end > len(merged) {
			end = len(merged)
		}

		return merged[start:end], totalData, nil
	}
}

func (r *repository) DetailKunjungan(ctx context.Context, noRawat string, kodeDokter string) (*KunjunganRawatJalan, error) {
	baseQuery, baseArgs := buildBaseQuery(kodeDokter, FilterAntreanDokter{})

	query := fmt.Sprintf(`
		SELECT * 
		FROM (%s) AS t 
		WHERE t.status_pemeriksaan <> 'Batal' AND t.no_rawat = ? 
		LIMIT 1
	`, baseQuery)

	args := append(baseArgs, noRawat)
	row := r.db.QueryRowContext(ctx, query, args...)
	kunjungan, err := scanKunjungan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("gagal query detail kunjungan: %w", err)
	}
	return kunjungan, nil
}

func (r *repository) RiwayatKunjunganPasien(ctx context.Context, noRekamMedis string) ([]KunjunganRawatJalan, error) {
	baseQuery, baseArgs := buildBaseQuery("", FilterAntreanDokter{})

	query := fmt.Sprintf(`
		SELECT * 
		FROM (%s) AS t 
		WHERE t.status_pemeriksaan <> 'Batal' AND t.no_rekam_medis = ? 
		ORDER BY t.tanggal_registrasi DESC, t.jam_registrasi DESC
	`, baseQuery)

	args := append(baseArgs, noRekamMedis)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal query riwayat kunjungan pasien: %w", err)
	}
	defer rows.Close()

	listKunjungan := make([]KunjunganRawatJalan, 0)
	for rows.Next() {
		kunjungan, err := scanKunjungan(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal scan riwayat kunjungan: %w", err)
		}
		listKunjungan = append(listKunjungan, *kunjungan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterasi riwayat kunjungan: %w", err)
	}

	return listKunjungan, nil
}

func (r *repository) GetWaktuRegistrasi(ctx context.Context, noRawat string) (string, string, bool, error) {
	query := `
		SELECT 
			DATE_FORMAT(tgl_registrasi, '%Y-%m-%d') AS tgl_registrasi,
			jam_reg
		FROM reg_periksa
		WHERE no_rawat = ?
	`
	var tanggalRegistrasi, jamRegistrasi string
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&tanggalRegistrasi, &jamRegistrasi)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, fmt.Errorf("gagal query waktu registrasi: %w", err)
	}
	return tanggalRegistrasi, jamRegistrasi, true, nil
}

func (r *repository) GetInfoRegistrasi(ctx context.Context, noRawat string) (*InfoRegistrasiPasien, error) {
	query := `
		SELECT 
			DATE_FORMAT(tgl_registrasi, '%Y-%m-%d') AS tgl_registrasi,
			jam_reg,
			kd_pj,
			status_bayar
		FROM reg_periksa
		WHERE no_rawat = ?
	`
	var info InfoRegistrasiPasien
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(
		&info.TanggalRegistrasi,
		&info.JamRegistrasi,
		&info.KodePenjamin,
		&info.StatusBayar,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("gagal query info registrasi: %w", err)
	}
	return &info, nil
}

func (r *repository) GetKodePoliByNoRawat(ctx context.Context, noRawat string, kodeDokter string) (string, error) {
	if strings.TrimSpace(kodeDokter) != "" {
		var kdPoliRujukan string
		err := r.db.QueryRowContext(ctx, "SELECT kd_poli FROM rujukan_internal_poli WHERE no_rawat = ? AND kd_dokter = ? LIMIT 1", noRawat, kodeDokter).Scan(&kdPoliRujukan)
		if err == nil && strings.TrimSpace(kdPoliRujukan) != "" {
			return strings.TrimSpace(kdPoliRujukan), nil
		}
	}

	var kdPoli string
	err := r.db.QueryRowContext(ctx, "SELECT kd_poli FROM reg_periksa WHERE no_rawat = ? LIMIT 1", noRawat).Scan(&kdPoli)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("gagal query kode poli: %w", err)
	}
	return strings.TrimSpace(kdPoli), nil
}

func (r *repository) CekAksesKunjungan(ctx context.Context, noRawat string, kodeDokter string) (bool, bool, error) {
	var kdDokter string
	err := r.db.QueryRowContext(ctx, "SELECT kd_dokter FROM reg_periksa WHERE no_rawat = ? LIMIT 1", noRawat).Scan(&kdDokter)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("gagal query kunjungan reg_periksa: %w", err)
	}

	cleanKodeDokter := strings.TrimSpace(kodeDokter)
	if cleanKodeDokter == "" || strings.TrimSpace(kdDokter) == cleanKodeDokter {
		return true, true, nil
	}

	var countRujukan int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rujukan_internal_poli WHERE no_rawat = ? AND kd_dokter = ?", noRawat, cleanKodeDokter).Scan(&countRujukan)
	if err != nil {
		return true, false, fmt.Errorf("gagal query rujukan internal: %w", err)
	}

	return true, countRujukan > 0, nil
}

func (r *repository) UpdateStatusPemeriksaan(ctx context.Context, noRawat string, status StatusPemeriksaan) error {
	query := "UPDATE reg_periksa SET stts = ? WHERE no_rawat = ?"
	res, err := r.db.ExecContext(ctx, query, string(status), noRawat)
	if err != nil {
		return fmt.Errorf("gagal update status pemeriksaan kunjungan: %w", err)
	}
	return database.CheckRowsAffected(res)
}

