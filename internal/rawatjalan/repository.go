package rawatjalan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

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

const selectKunjunganBukanRujukan = `
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
		COALESCE(r.almt_pj, '-') AS alamat_penanggung_jawab
	FROM reg_periksa r
	INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis
	INNER JOIN poliklinik pol ON r.kd_poli = pol.kd_poli
	INNER JOIN dokter d ON r.kd_dokter = d.kd_dokter
	INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj
`

const selectKunjunganRujukan = `
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
		COALESCE(r.almt_pj, '-') AS alamat_penanggung_jawab
	FROM rujukan_internal_poli rip
	INNER JOIN reg_periksa r ON rip.no_rawat = r.no_rawat
	INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis
	INNER JOIN poliklinik pol ON r.kd_poli = pol.kd_poli
	INNER JOIN dokter d ON r.kd_dokter = d.kd_dokter
	INNER JOIN penjab pj ON r.kd_pj = pj.kd_pj
	INNER JOIN poliklinik pol_rip ON rip.kd_poli = pol_rip.kd_poli
	INNER JOIN dokter d_rip ON rip.kd_dokter = d_rip.kd_dokter
`

var orderByMapping = map[string]func(string) string{
	"waktu_registrasi":   func(dir string) string { return fmt.Sprintf("r.tgl_registrasi %s, r.jam_reg %s", dir, dir) },
	"nama_pasien":        func(dir string) string { return fmt.Sprintf("p.nm_pasien %s", dir) },
	"penjamin":           func(dir string) string { return fmt.Sprintf("pj.png_jawab %s", dir) },
	"status_pemeriksaan": func(dir string) string { return fmt.Sprintf("r.stts %s", dir) },
	"status_lanjut":      func(dir string) string { return fmt.Sprintf("r.status_lanjut %s", dir) },
	"status_bayar":       func(dir string) string { return fmt.Sprintf("r.status_bayar %s", dir) },
	"jenis_antrean":      func(dir string) string { return fmt.Sprintf("r.tgl_registrasi %s, r.jam_reg %s", dir, dir) },
}

func sortAntrean(data []KunjunganRawatJalan, orderBy string, sortOrder string) {
	asc := strings.ToUpper(sortOrder) == "ASC"
	sort.SliceStable(data, func(i, j int) bool {
		a, b := data[i], data[j]
		var result int
		switch orderBy {
		case "nama_pasien":
			result = strings.Compare(a.NamaPasien, b.NamaPasien)
		case "penjamin":
			result = strings.Compare(a.NamaPenjamin, b.NamaPenjamin)
		case "status_pemeriksaan":
			result = strings.Compare(string(a.StatusPemeriksaan), string(b.StatusPemeriksaan))
		case "status_lanjut":
			result = strings.Compare(string(a.StatusLanjut), string(b.StatusLanjut))
		case "status_bayar":
			result = strings.Compare(string(a.StatusBayar), string(b.StatusBayar))
		case "jenis_antrean":
			result = strings.Compare(string(a.JenisAntrean), string(b.JenisAntrean))
		default:
			result = strings.Compare(a.TanggalRegistrasi, b.TanggalRegistrasi)
			if result == 0 {
				result = strings.Compare(a.JamRegistrasi, b.JamRegistrasi)
			}
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

func buildBranchConditions(dokterCol string, kodeDokter string, filter FilterAntreanDokter) (string, []any) {
	var conditions []string
	var args []any

	if kodeDokter != "" {
		conditions = append(conditions, dokterCol+" = ?")
		args = append(args, kodeDokter)
	}

	if filter.Tanggal != "" {
		tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)
		if tglAwal != "" && tglAkhir != "" {
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
	where, args := buildBranchConditions("r.kd_dokter", kodeDokter, filter)
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
	where, args := buildBranchConditions("rip.kd_dokter", kodeDokter, filter)
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
		where, args := buildBranchConditions("r.kd_dokter", kodeDokter, filter)
		return selectKunjunganBukanRujukan + where, args

	case JenisAntreanRujukan:
		where, args := buildBranchConditions("rip.kd_dokter", kodeDokter, filter)
		return selectKunjunganRujukan + where, args

	default:
		where1, args1 := buildBranchConditions("r.kd_dokter", kodeDokter, filter)
		where2, args2 := buildBranchConditions("rip.kd_dokter", kodeDokter, filter)

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

func (r *repository) DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter FilterAntreanDokter) ([]KunjunganRawatJalan, int, error) {
	innerBuilder, exists := orderByMapping[filter.OrderBy]
	if !exists {
		innerBuilder = orderByMapping["waktu_registrasi"]
	}
	innerOrder := " ORDER BY " + innerBuilder(filter.SortOrder)

	switch filter.JenisAntrean {
	case JenisAntreanTidakRujukan:
		totalData, err := r.countBukanRujukan(ctx, kodeDokter, filter)
		if err != nil {
			return nil, 0, fmt.Errorf("gagal menghitung total antrean: %w", err)
		}
		if totalData == 0 {
			return make([]KunjunganRawatJalan, 0), 0, nil
		}
		where, args := buildBranchConditions("r.kd_dokter", kodeDokter, filter)
		query := selectKunjunganBukanRujukan + where + innerOrder + " LIMIT ? OFFSET ?"
		args = append(args, filter.Limit, filter.Offset())
		daftarAntrean, err := r.scanAntreanRows(ctx, query, args)
		if err != nil {
			return nil, 0, err
		}
		return daftarAntrean, totalData, nil

	case JenisAntreanRujukan:
		totalData, err := r.countRujukan(ctx, kodeDokter, filter)
		if err != nil {
			return nil, 0, fmt.Errorf("gagal menghitung total antrean: %w", err)
		}
		if totalData == 0 {
			return make([]KunjunganRawatJalan, 0), 0, nil
		}
		where, args := buildBranchConditions("rip.kd_dokter", kodeDokter, filter)
		query := selectKunjunganRujukan + where + innerOrder + " LIMIT ? OFFSET ?"
		args = append(args, filter.Limit, filter.Offset())
		daftarAntrean, err := r.scanAntreanRows(ctx, query, args)
		if err != nil {
			return nil, 0, err
		}
		return daftarAntrean, totalData, nil

	default:
		if kodeDokter == "" {
			totalData, err := r.countBukanRujukan(ctx, "", filter)
			if err != nil {
				return nil, 0, fmt.Errorf("gagal menghitung total antrean: %w", err)
			}
			if totalData == 0 {
				return make([]KunjunganRawatJalan, 0), 0, nil
			}

			where, args := buildBranchConditions("r.kd_dokter", "", filter)
			pageQuery := "SELECT r.no_rawat FROM reg_periksa r"
			if filter.Keyword != "" {
				pageQuery += " INNER JOIN pasien p ON r.no_rkm_medis = p.no_rkm_medis"
			}
			pageQuery += where + innerOrder + " LIMIT ? OFFSET ?"
			pageArgs := append(append([]any(nil), args...), filter.Limit, filter.Offset())

			idRows, err := r.db.QueryContext(ctx, pageQuery, pageArgs...)
			if err != nil {
				return nil, 0, fmt.Errorf("gagal query halaman antrean: %w", err)
			}
			defer idRows.Close()

			var pageNoRawats []string
			for idRows.Next() {
				var nr string
				if err := idRows.Scan(&nr); err == nil {
					pageNoRawats = append(pageNoRawats, nr)
				}
			}
			if err := idRows.Err(); err != nil {
				return nil, 0, fmt.Errorf("gagal membaca baris halaman antrean: %w", err)
			}

			if len(pageNoRawats) == 0 {
				return make([]KunjunganRawatJalan, 0), totalData, nil
			}

			placeholders := make([]string, len(pageNoRawats))
			idArgs := make([]any, len(pageNoRawats))
			for i, nr := range pageNoRawats {
				placeholders[i] = "?"
				idArgs[i] = nr
			}
			detailQuery := selectKunjunganBukanRujukan + fmt.Sprintf(" WHERE r.no_rawat IN (%s) ", strings.Join(placeholders, ", ")) + innerOrder
			daftarAntrean, err := r.scanAntreanRows(ctx, detailQuery, idArgs)
			if err != nil {
				return nil, 0, err
			}

			return daftarAntrean, totalData, nil
		}

		countNonRujukan, err := r.countBukanRujukan(ctx, kodeDokter, filter)
		if err != nil {
			return nil, 0, fmt.Errorf("gagal menghitung antrean bukan rujukan: %w", err)
		}

		countRujukan, err := r.countRujukan(ctx, kodeDokter, filter)
		if err != nil {
			return nil, 0, fmt.Errorf("gagal menghitung antrean rujukan: %w", err)
		}

		totalData := countNonRujukan + countRujukan
		if totalData == 0 {
			return make([]KunjunganRawatJalan, 0), 0, nil
		}

		if countRujukan == 0 {
			where, args := buildBranchConditions("r.kd_dokter", kodeDokter, filter)
			query := selectKunjunganBukanRujukan + where + innerOrder + " LIMIT ? OFFSET ?"
			args = append(args, filter.Limit, filter.Offset())
			daftarAntrean, err := r.scanAntreanRows(ctx, query, args)
			if err != nil {
				return nil, 0, err
			}
			return daftarAntrean, totalData, nil
		}

		if countNonRujukan == 0 {
			where, args := buildBranchConditions("rip.kd_dokter", kodeDokter, filter)
			query := selectKunjunganRujukan + where + innerOrder + " LIMIT ? OFFSET ?"
			args = append(args, filter.Limit, filter.Offset())
			daftarAntrean, err := r.scanAntreanRows(ctx, query, args)
			if err != nil {
				return nil, 0, err
			}
			return daftarAntrean, totalData, nil
		}

		innerLimit := filter.Offset() + filter.Limit
		innerLimitClause := fmt.Sprintf(" LIMIT %d", innerLimit)

		where1, args1 := buildBranchConditions("r.kd_dokter", kodeDokter, filter)
		query1 := selectKunjunganBukanRujukan + where1 + innerOrder + innerLimitClause

		where2, args2 := buildBranchConditions("rip.kd_dokter", kodeDokter, filter)
		query2 := selectKunjunganRujukan + where2 + innerOrder + innerLimitClause

		branch1, err := r.scanAntreanRows(ctx, query1, args1)
		if err != nil {
			return nil, 0, err
		}

		branch2, err := r.scanAntreanRows(ctx, query2, args2)
		if err != nil {
			return nil, 0, err
		}

		merged := append(branch1, branch2...)
		sortAntrean(merged, filter.OrderBy, filter.SortOrder)

		start := filter.Offset()
		end := start + filter.Limit
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
	_, err := r.db.ExecContext(ctx, query, string(status), noRawat)
	if err != nil {
		return fmt.Errorf("gagal update status pemeriksaan kunjungan: %w", err)
	}
	return nil
}

