package rawatinap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/formatter"
)

type Repository interface {
	CekStatusKamarInap(ctx context.Context, noRawat string) (isKamarAktif bool, hasRecordKamar bool, err error)
	GetKelasRawat(ctx context.Context, noRawat string) (string, error)
	DaftarPasienRawatInap(ctx context.Context, kodeDokterLogin string, filter FilterPasienRawatInap) ([]KunjunganRawatInap, int, error)
	DetailPasienRawatInap(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*KunjunganRawatInap, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CekStatusKamarInap(ctx context.Context, noRawat string) (bool, bool, error) {
	query := `
		SELECT
			COUNT(*) > 0 AS has_record,
			COALESCE(SUM(CASE
				WHEN COALESCE(stts_pulang, '-') = '-'
				 AND (tgl_keluar IS NULL OR COALESCE(DATE_FORMAT(tgl_keluar, '%Y-%m-%d'), '0000-00-00') = '0000-00-00')
				THEN 1 ELSE 0 END), 0) > 0 AS is_aktif
		FROM kamar_inap
		WHERE no_rawat = ?
	`
	var hasRecord, isAktif bool
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&hasRecord, &isAktif)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}

	return isAktif, hasRecord, nil
}

func (r *repository) GetKelasRawat(ctx context.Context, noRawat string) (string, error) {
	query := `
		SELECT COALESCE(k.kelas, '')
		FROM kamar_inap ki
		INNER JOIN kamar k ON ki.kd_kamar = k.kd_kamar
		WHERE ki.no_rawat = ?
		ORDER BY (COALESCE(ki.stts_pulang, '-') = '-') DESC, ki.tgl_masuk DESC, ki.jam_masuk DESC
		LIMIT 1
	`
	var kelas string
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&kelas)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("gagal query kelas rawat inap: %w", err)
	}
	return strings.TrimSpace(kelas), nil
}

const (
	sqlKamarInapAktif = "COALESCE(ki.stts_pulang, '-') = '-' AND (ki.tgl_keluar IS NULL OR COALESCE(DATE_FORMAT(ki.tgl_keluar, '%Y-%m-%d'), '0000-00-00') = '0000-00-00')"
	sqlKamarInapKRS   = "COALESCE(ki.stts_pulang, '-') <> '-' AND ki.stts_pulang <> 'Pindah Kamar' AND ki.tgl_keluar IS NOT NULL AND COALESCE(DATE_FORMAT(ki.tgl_keluar, '%Y-%m-%d'), '0000-00-00') <> '0000-00-00'"
)

const selectPasienRawatInap = `
	SELECT 
		ki.no_rawat,
		rp.no_reg AS no_registrasi,
		DATE_FORMAT(rp.tgl_registrasi, '%Y-%m-%d') AS tanggal_registrasi,
		rp.jam_reg AS jam_registrasi,
		rp.no_rkm_medis,
		p.nm_pasien AS nama_pasien,
		p.jk AS jenis_kelamin,
		DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d') AS tanggal_lahir,
		COALESCE(p.alamat, rp.almt_pj, '-') AS alamat,
		ki.kd_kamar AS kode_kamar,
		k.kelas AS kelas_kamar,
		b.kd_bangsal AS kode_bangsal,
		b.nm_bangsal AS nama_bangsal,
		DATE_FORMAT(ki.tgl_masuk, '%Y-%m-%d') AS tanggal_masuk,
		ki.jam_masuk,
		DATE_FORMAT(ki.tgl_keluar, '%Y-%m-%d') AS tanggal_keluar,
		ki.jam_keluar,
		ki.stts_pulang AS status_pulang,
		ki.lama AS lama_inap,
		rp.kd_pj AS kode_penjamin,
		pj.png_jawab AS nama_penjamin,
		rp.status_bayar,
		rp.kd_dokter AS kode_dokter_rawat_jalan,
		d_reg.nm_dokter AS dokter_rawat_jalan,
		COALESCE((
			SELECT GROUP_CONCAT(DISTINCT d_dpjp.nm_dokter ORDER BY d_dpjp.nm_dokter SEPARATOR ';;;')
			FROM dpjp_ranap dr
			INNER JOIN dokter d_dpjp ON dr.kd_dokter = d_dpjp.kd_dokter
			WHERE dr.no_rawat = ki.no_rawat
		), '') AS dpjp,
		COALESCE(ki.diagnosa_awal, '-') AS diagnosa_awal,
		COALESCE(ki.diagnosa_akhir, '-') AS diagnosa_akhir,
		COALESCE(p.gol_darah, '-') AS golongan_darah,
		COALESCE(p.agama, '-') AS agama,
		COALESCE(p.no_tlp, '-') AS no_telepon,
		COALESCE(p.no_peserta, '-') AS no_peserta,
		COALESCE(p.no_ktp, '-') AS no_ktp,
		COALESCE(rp.p_jawab, '-') AS penanggung_jawab,
		COALESCE(rp.hubunganpj, '-') AS hubungan_penanggung_jawab,
		COALESCE(rp.almt_pj, '-') AS alamat_penanggung_jawab
`

const fromAndJoinKamarInap = `
	FROM kamar_inap ki
	INNER JOIN reg_periksa rp ON ki.no_rawat = rp.no_rawat
	INNER JOIN pasien p ON rp.no_rkm_medis = p.no_rkm_medis
	INNER JOIN kamar k ON ki.kd_kamar = k.kd_kamar
	INNER JOIN bangsal b ON k.kd_bangsal = b.kd_bangsal
	INNER JOIN penjab pj ON rp.kd_pj = pj.kd_pj
	INNER JOIN dokter d_reg ON rp.kd_dokter = d_reg.kd_dokter
`

func buildStatusFilter(filter FilterPasienRawatInap, tglAwal, tglAkhir string) (string, []any) {
	hasRentangTanggal := tglAwal != "" && tglAkhir != ""

	if filter.StatusPulang == StatusPulangBelumPulang {
		return sqlKamarInapAktif, nil
	}

	if filter.StatusPulang != "" && hasRentangTanggal {
		return "ki.stts_pulang = ? AND ki.tgl_keluar BETWEEN ? AND ?", []any{string(filter.StatusPulang), tglAwal, tglAkhir}
	}

	if filter.StatusPulang != "" {
		return "ki.stts_pulang = ?", []any{string(filter.StatusPulang)}
	}

	if filter.StatusKunjungan == StatusKunjunganBelumKRS {
		return sqlKamarInapAktif, nil
	}

	if filter.StatusKunjungan == StatusKunjunganTanggalMRS && hasRentangTanggal {
		return "ki.tgl_masuk BETWEEN ? AND ?", []any{tglAwal, tglAkhir}
	}

	if filter.StatusKunjungan == StatusKunjunganTanggalKRS && hasRentangTanggal {
		return sqlKamarInapKRS + " AND ki.tgl_keluar BETWEEN ? AND ?", []any{tglAwal, tglAkhir}
	}

	if filter.StatusKunjungan == StatusKunjunganTanggalKRS {
		return sqlKamarInapKRS, nil
	}

	return "", nil
}

func buildDokterFilter(kodeDokterLogin string, filter FilterPasienRawatInap) (string, []any) {
	if filter.ScopeDPJP == ScopeDPJPDokter && kodeDokterLogin != "" {
		return "(EXISTS (SELECT 1 FROM dpjp_ranap dr WHERE dr.no_rawat = ki.no_rawat AND dr.kd_dokter = ?) OR (NOT EXISTS (SELECT 1 FROM dpjp_ranap dr_sub WHERE dr_sub.no_rawat = ki.no_rawat) AND rp.kd_dokter = ?))", []any{kodeDokterLogin, kodeDokterLogin}
	}

	return "", nil
}

func (r *repository) buildFilterConditions(kodeDokterLogin string, filter FilterPasienRawatInap) ([]string, []any) {
	var conditions []string
	var args []any

	tglAwal, tglAkhir := formatter.ParseRentangTanggal(filter.Tanggal)

	statusCond, statusArgs := buildStatusFilter(filter, tglAwal, tglAkhir)
	if statusCond != "" {
		conditions = append(conditions, statusCond)
		args = append(args, statusArgs...)
	}

	if filter.Bangsal != "" {
		conditions = append(conditions, "b.kd_bangsal = ?")
		args = append(args, filter.Bangsal)
	}

	if filter.Kelas != "" {
		conditions = append(conditions, "k.kelas = ?")
		args = append(args, filter.Kelas)
	}

	if filter.Penjamin != "" {
		conditions = append(conditions, "rp.kd_pj = ?")
		args = append(args, filter.Penjamin)
	}

	dokterCond, dokterArgs := buildDokterFilter(kodeDokterLogin, filter)
	if dokterCond != "" {
		conditions = append(conditions, dokterCond)
		args = append(args, dokterArgs...)
	}

	if filter.Keyword != "" {
		keywordPattern := "%" + filter.Keyword + "%"
		conditions = append(conditions, "(p.nm_pasien LIKE ? OR rp.no_rkm_medis LIKE ? OR p.alamat LIKE ? OR rp.almt_pj LIKE ? OR ki.no_rawat LIKE ?)")
		args = append(args, keywordPattern, keywordPattern, keywordPattern, keywordPattern, keywordPattern)
	}

	return conditions, args
}

func buildOrderClause(orderBy OrderByRanap, sortOrder string) string {
	dir := "DESC"
	if sortOrder == "ASC" {
		dir = "ASC"
	}

	switch orderBy {
	case OrderByWaktuMRS:
		return fmt.Sprintf("ki.tgl_masuk %s, ki.jam_masuk %s", dir, dir)
	case OrderByWaktuKRS:
		return fmt.Sprintf("ki.tgl_keluar %s, ki.jam_keluar %s", dir, dir)
	case OrderByNamaPasien:
		return fmt.Sprintf("p.nm_pasien %s", dir)
	case OrderByKamar:
		return fmt.Sprintf("ki.kd_kamar %s", dir)
	case OrderByKelas:
		return fmt.Sprintf("k.kelas %s", dir)
	case OrderByBangsal:
		return fmt.Sprintf("b.nm_bangsal %s", dir)
	case OrderByPenjamin:
		return fmt.Sprintf("pj.png_jawab %s", dir)
	case OrderByStatusPulang:
		return fmt.Sprintf("ki.stts_pulang %s", dir)
	default:
		return fmt.Sprintf("ki.tgl_masuk %s, ki.jam_masuk %s", dir, dir)
	}
}

func parseDPJPList(dpjpStr sql.NullString) []string {
	if !dpjpStr.Valid || strings.TrimSpace(dpjpStr.String) == "" {
		return make([]string, 0)
	}

	parts := strings.Split(dpjpStr.String, ";;;")
	list := make([]string, 0, len(parts))
	for _, p := range parts {
		clean := strings.TrimSpace(p)
		if clean != "" {
			list = append(list, clean)
		}
	}
	return list
}

type scanner interface {
	Scan(dest ...any) error
}

func scanKunjunganInap(s scanner) (*KunjunganRawatInap, error) {
	var (
		item          KunjunganRawatInap
		tglKeluar     sql.NullString
		jamKeluar     sql.NullString
		dpjpStr       sql.NullString
		diagnosaAkhir sql.NullString
	)

	err := s.Scan(
		&item.NoRawat,
		&item.NoRegistrasi,
		&item.TanggalRegistrasi,
		&item.JamRegistrasi,
		&item.NoRekamMedis,
		&item.NamaPasien,
		&item.JenisKelamin,
		&item.TanggalLahir,
		&item.Alamat,
		&item.KodeKamar,
		&item.KelasKamar,
		&item.KodeBangsal,
		&item.NamaBangsal,
		&item.TanggalMasuk,
		&item.JamMasuk,
		&tglKeluar,
		&jamKeluar,
		&item.StatusPulang,
		&item.LamaInap,
		&item.KodePenjamin,
		&item.NamaPenjamin,
		&item.StatusBayar,
		&item.KodeDokterRawatJalan,
		&item.DokterRawatJalan,
		&dpjpStr,
		&item.DiagnosaAwal,
		&diagnosaAkhir,
		&item.GolonganDarah,
		&item.Agama,
		&item.NoTelepon,
		&item.NoPeserta,
		&item.NoKTP,
		&item.PenanggungJawab,
		&item.HubunganPenanggungJawab,
		&item.AlamatPenanggungJawab,
	)
	if err != nil {
		return nil, err
	}

	item.StatusLanjut = shared.StatusLanjutRawatInap
	if tglKeluar.Valid && tglKeluar.String != "0000-00-00" {
		item.TanggalKeluar = tglKeluar.String
	}
	if jamKeluar.Valid && jamKeluar.String != "00:00:00" {
		item.JamKeluar = jamKeluar.String
	}
	if diagnosaAkhir.Valid && diagnosaAkhir.String != "-" && diagnosaAkhir.String != "" {
		item.DiagnosaAkhir = diagnosaAkhir.String
	}

	item.DPJP = parseDPJPList(dpjpStr)
	item.Umur = item.FormatUmur()
	return &item, nil
}

func (r *repository) countPasienRawatInap(ctx context.Context, whereClause string, args []any) (int, error) {
	countQuery := fmt.Sprintf(`
		SELECT COUNT(ki.no_rawat) 
		%s 
		%s
	`, fromAndJoinKamarInap, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("gagal menghitung total pasien rawat inap: %w", err)
	}
	return total, nil
}

func (r *repository) scanPasienRows(ctx context.Context, query string, args []any, limit int) ([]KunjunganRawatInap, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal query daftar pasien rawat inap: %w", err)
	}
	defer rows.Close()

	result := make([]KunjunganRawatInap, 0, limit)
	for rows.Next() {
		item, err := scanKunjunganInap(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal scan data pasien rawat inap: %w", err)
		}
		result = append(result, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterasi data pasien rawat inap: %w", err)
	}

	return result, nil
}

func (r *repository) DaftarPasienRawatInap(ctx context.Context, kodeDokterLogin string, filter FilterPasienRawatInap) ([]KunjunganRawatInap, int, error) {
	conditions, whereArgs := r.buildFilterConditions(kodeDokterLogin, filter)

	var whereClause string
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	total, err := r.countPasienRawatInap(ctx, whereClause, whereArgs)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return make([]KunjunganRawatInap, 0), 0, nil
	}

	orderClause := buildOrderClause(filter.OrderBy, filter.SortOrder)
	dataQuery := fmt.Sprintf(`
		%s
		%s
		%s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, selectPasienRawatInap, fromAndJoinKamarInap, whereClause, orderClause)

	finalArgs := make([]any, 0, len(whereArgs)+2)
	finalArgs = append(finalArgs, whereArgs...)
	finalArgs = append(finalArgs, filter.Limit, filter.Offset())

	list, err := r.scanPasienRows(ctx, dataQuery, finalArgs, filter.Limit)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) DetailPasienRawatInap(ctx context.Context, noRawat string, tglMasuk string, jamMasuk string) (*KunjunganRawatInap, error) {
	dataQuery := fmt.Sprintf(`
		%s
		%s
		WHERE ki.no_rawat = ? AND ki.tgl_masuk = ? AND ki.jam_masuk = ?
		LIMIT 1
	`, selectPasienRawatInap, fromAndJoinKamarInap)

	row := r.db.QueryRowContext(ctx, dataQuery, noRawat, tglMasuk, jamMasuk)
	item, err := scanKunjunganInap(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("gagal query detail pasien rawat inap: %w", err)
	}

	return item, nil
}
