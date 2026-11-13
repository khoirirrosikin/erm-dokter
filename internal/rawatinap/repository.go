package rawatinap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/formatter"
)

type Repository interface {
	CekStatusKamarInap(ctx context.Context, noRawat string) (isKamarAktif bool, hasRecordKamar bool, err error)
	GetWaktuKeluarTerakhir(ctx context.Context, noRawat string) (tanggalKeluar string, jamKeluar string, err error)
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
				WHEN (stts_pulang = '-' OR stts_pulang IS NULL)
				 AND (tgl_keluar IS NULL OR tgl_keluar < '1000-01-01')
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

func (r *repository) GetWaktuKeluarTerakhir(ctx context.Context, noRawat string) (string, string, error) {
	query := `
		SELECT
			COALESCE(DATE_FORMAT(ki.tgl_keluar, '%Y-%m-%d'), '') AS tgl_keluar,
			COALESCE(ki.jam_keluar, '') AS jam_keluar
		FROM kamar_inap ki
		WHERE ki.no_rawat = ?
		  AND ki.tgl_keluar >= '1000-01-01'
		ORDER BY
			(COALESCE(ki.stts_pulang, '-') <> 'Pindah Kamar') DESC,
			ki.tgl_keluar DESC,
			ki.jam_keluar DESC,
			ki.tgl_masuk DESC,
			ki.jam_masuk DESC
		LIMIT 1
	`
	var tglKeluar, jamKeluar string
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&tglKeluar, &jamKeluar)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("gagal query waktu keluar terakhir kamar inap: %w", err)
	}

	tglKeluar = strings.TrimSpace(tglKeluar)
	if tglKeluar == "0000-00-00" {
		return "", "", nil
	}
	return tglKeluar, strings.TrimSpace(jamKeluar), nil
}

func (r *repository) GetKelasRawat(ctx context.Context, noRawat string) (string, error) {
	query := `
		SELECT COALESCE(k.kelas, '')
		FROM kamar_inap ki
		LEFT JOIN kamar k ON ki.kd_kamar = k.kd_kamar
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
	sqlKamarInapAktif = "(ki.stts_pulang = '-' OR ki.stts_pulang IS NULL) AND (ki.tgl_keluar IS NULL OR ki.tgl_keluar < '1000-01-01')"
	sqlKamarInapKRS   = "ki.stts_pulang NOT IN ('-', 'Pindah Kamar') AND ki.tgl_keluar >= '1000-01-01'"
)

const selectPasienRawatInap = `
	SELECT 
		ki.no_rawat,
		COALESCE(rp.no_reg, '') AS no_registrasi,
		COALESCE(DATE_FORMAT(rp.tgl_registrasi, '%Y-%m-%d'), '') AS tanggal_registrasi,
		COALESCE(rp.jam_reg, '') AS jam_registrasi,
		COALESCE(rp.no_rkm_medis, '') AS no_rkm_medis,
		COALESCE(p.nm_pasien, '') AS nama_pasien,
		COALESCE(p.jk, '') AS jenis_kelamin,
		COALESCE(DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d'), '') AS tanggal_lahir,
		COALESCE(p.alamat, rp.almt_pj, '-') AS alamat,
		COALESCE(ki.kd_kamar, '') AS kode_kamar,
		COALESCE(k.kelas, '') AS kelas_kamar,
		COALESCE(b.kd_bangsal, '') AS kode_bangsal,
		COALESCE(b.nm_bangsal, '') AS nama_bangsal,
		COALESCE(DATE_FORMAT(ki.tgl_masuk, '%Y-%m-%d'), '') AS tanggal_masuk,
		COALESCE(ki.jam_masuk, '') AS jam_masuk,
		DATE_FORMAT(ki.tgl_keluar, '%Y-%m-%d') AS tanggal_keluar,
		ki.jam_keluar,
		COALESCE(ki.stts_pulang, '-') AS status_pulang,
		CAST(ROUND(COALESCE(ki.lama, 0)) AS SIGNED) AS lama_inap,
		COALESCE(rp.kd_pj, '') AS kode_penjamin,
		COALESCE(pj.png_jawab, '') AS nama_penjamin,
		COALESCE(rp.status_bayar, '') AS status_bayar,
		COALESCE(rp.kd_dokter, '') AS kode_dokter_rawat_jalan,
		COALESCE(d_reg.nm_dokter, '') AS dokter_rawat_jalan,
		COALESCE(ki.diagnosa_awal, '-') AS diagnosa_awal,
		ki.diagnosa_akhir,
		COALESCE(p.gol_darah, '-') AS golongan_darah,
		COALESCE(p.agama, '-') AS agama,
		COALESCE(p.no_tlp, '-') AS no_telepon,
		COALESCE(p.no_peserta, '-') AS no_peserta,
		COALESCE(p.no_ktp, '-') AS no_ktp,
		COALESCE(rp.p_jawab, '-') AS penanggung_jawab,
		COALESCE(rp.hubunganpj, '-') AS hubungan_penanggung_jawab,
		COALESCE(rp.almt_pj, '-') AS alamat_penanggung_jawab
`

const lookupJoinsKamarInap = `
	LEFT JOIN reg_periksa rp ON ki.no_rawat = rp.no_rawat
	LEFT JOIN pasien p ON rp.no_rkm_medis = p.no_rkm_medis
	LEFT JOIN kamar k ON ki.kd_kamar = k.kd_kamar
	LEFT JOIN bangsal b ON k.kd_bangsal = b.kd_bangsal
	LEFT JOIN penjab pj ON rp.kd_pj = pj.kd_pj
	LEFT JOIN dokter d_reg ON rp.kd_dokter = d_reg.kd_dokter
`

const fromAndJoinKamarInap = `
	FROM kamar_inap ki
` + lookupJoinsKamarInap

type ranapJoinFlags struct {
	regPeriksa bool
	pasien     bool
	kamar      bool
	bangsal    bool
	penjab     bool
}

func (f ranapJoinFlags) toSQL() string {
	if f.pasien || f.penjab {
		f.regPeriksa = true
	}
	if f.bangsal {
		f.kamar = true
	}

	var b strings.Builder
	if f.regPeriksa {
		b.WriteString("\n\tLEFT JOIN reg_periksa rp ON ki.no_rawat = rp.no_rawat")
	}
	if f.pasien {
		b.WriteString("\n\tLEFT JOIN pasien p ON rp.no_rkm_medis = p.no_rkm_medis")
	}
	if f.kamar {
		b.WriteString("\n\tLEFT JOIN kamar k ON ki.kd_kamar = k.kd_kamar")
	}
	if f.bangsal {
		b.WriteString("\n\tLEFT JOIN bangsal b ON k.kd_bangsal = b.kd_bangsal")
	}
	if f.penjab {
		b.WriteString("\n\tLEFT JOIN penjab pj ON rp.kd_pj = pj.kd_pj")
	}
	return b.String()
}

func filterJoinFlags(kodeDokterLogin string, filter FilterPasienRawatInap) ranapJoinFlags {
	var flags ranapJoinFlags
	if filter.Bangsal != "" {
		flags.bangsal = true
	}
	if filter.Kelas != "" {
		flags.kamar = true
	}
	if filter.Penjamin != "" {
		flags.regPeriksa = true
	}
	if filter.ScopeDPJP == ScopeDPJPDokter && kodeDokterLogin != "" {
		flags.regPeriksa = true
	}
	if filter.Keyword != "" {
		flags.pasien = true
	}
	return flags
}

func buildFilterJoins(kodeDokterLogin string, filter FilterPasienRawatInap) string {
	return filterJoinFlags(kodeDokterLogin, filter).toSQL()
}

func buildPageKeyJoins(kodeDokterLogin string, filter FilterPasienRawatInap) string {
	flags := filterJoinFlags(kodeDokterLogin, filter)
	switch filter.OrderBy {
	case OrderByNamaPasien:
		flags.pasien = true
	case OrderByKelas:
		flags.kamar = true
	case OrderByBangsal:
		flags.bangsal = true
	case OrderByPenjamin:
		flags.penjab = true
	}
	return flags.toSQL()
}

func buildStatusFilter(filter FilterPasienRawatInap, tglAwal, tglAkhir string) (string, []any) {
	hasRentangTanggal := tglAwal != "" && tglAkhir != ""

	if filter.StatusKunjungan == StatusKunjunganTanggalMRS {
		var parts []string
		var args []any

		if filter.StatusPulang == StatusPulangBelumPulang {
			parts = append(parts, sqlKamarInapAktif)
		} else if filter.StatusPulang != "" {
			parts = append(parts, "ki.stts_pulang = ?")
			args = append(args, string(filter.StatusPulang))
		}

		if hasRentangTanggal {
			parts = append(parts, "ki.tgl_masuk BETWEEN ? AND ?")
			args = append(args, tglAwal, tglAkhir)
		}

		return strings.Join(parts, " AND "), args
	}

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
		return fmt.Sprintf("ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir)
	case OrderByWaktuKRS:
		return fmt.Sprintf("ki.tgl_keluar %s, ki.jam_keluar %s, ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir, dir, dir)
	case OrderByNamaPasien:
		return fmt.Sprintf("p.nm_pasien %s, ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir, dir)
	case OrderByKamar:
		return fmt.Sprintf("ki.kd_kamar %s, ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir, dir)
	case OrderByKelas:
		return fmt.Sprintf("k.kelas %s, ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir, dir)
	case OrderByBangsal:
		return fmt.Sprintf("b.nm_bangsal %s, ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir, dir)
	case OrderByPenjamin:
		return fmt.Sprintf("pj.png_jawab %s, ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir, dir)
	case OrderByStatusPulang:
		return fmt.Sprintf("ki.stts_pulang %s, ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir, dir)
	default:
		return fmt.Sprintf("ki.tgl_masuk %s, ki.jam_masuk %s, ki.no_rawat %s", dir, dir, dir)
	}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanKunjunganInap(s scanner) (*KunjunganRawatInap, error) {
	var (
		item          KunjunganRawatInap
		tglKeluar     sql.NullString
		jamKeluar     sql.NullString
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

	item.DPJP = make([]string, 0)
	item.Umur = item.FormatUmur()
	return &item, nil
}

func (r *repository) hydrateDPJP(ctx context.Context, items []KunjunganRawatInap) error {
	if len(items) == 0 {
		return nil
	}

	uniqueNoRawats := make([]string, 0, len(items))
	seenNoRawat := make(map[string]struct{}, len(items))
	for i := range items {
		if items[i].DPJP == nil {
			items[i].DPJP = make([]string, 0)
		}
		nr := items[i].NoRawat
		if nr == "" {
			continue
		}
		if _, exists := seenNoRawat[nr]; !exists {
			seenNoRawat[nr] = struct{}{}
			uniqueNoRawats = append(uniqueNoRawats, nr)
		}
	}
	if len(uniqueNoRawats) == 0 {
		return nil
	}

	placeholders := shared.CreateInPlaceholders(len(uniqueNoRawats))
	args := make([]any, len(uniqueNoRawats))
	for i, nr := range uniqueNoRawats {
		args[i] = nr
	}

	query := fmt.Sprintf(`
		SELECT
			dr.no_rawat,
			COALESCE(d.nm_dokter, '') AS nm_dokter
		FROM dpjp_ranap dr
		INNER JOIN dokter d ON dr.kd_dokter = d.kd_dokter
		WHERE dr.no_rawat IN (%s)
	`, placeholders)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("gagal query dpjp rawat inap: %w", err)
	}
	defer rows.Close()

	dpjpSetByNoRawat := make(map[string]map[string]struct{}, len(uniqueNoRawats))
	for rows.Next() {
		var noRawat, nmDokter string
		if err := rows.Scan(&noRawat, &nmDokter); err != nil {
			return fmt.Errorf("gagal scan dpjp rawat inap: %w", err)
		}
		clean := strings.TrimSpace(nmDokter)
		if clean == "" {
			continue
		}
		set, ok := dpjpSetByNoRawat[noRawat]
		if !ok {
			set = make(map[string]struct{}, 2)
			dpjpSetByNoRawat[noRawat] = set
		}
		set[clean] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterasi dpjp rawat inap: %w", err)
	}

	sortedByNoRawat := make(map[string][]string, len(dpjpSetByNoRawat))
	for nr, set := range dpjpSetByNoRawat {
		names := make([]string, 0, len(set))
		for name := range set {
			names = append(names, name)
		}
		sort.Strings(names)
		sortedByNoRawat[nr] = names
	}

	for i := range items {
		if names, ok := sortedByNoRawat[items[i].NoRawat]; ok {
			copied := make([]string, len(names))
			copy(copied, names)
			items[i].DPJP = copied
		}
	}

	return nil
}

func (r *repository) countPasienRawatInap(ctx context.Context, filterJoins string, whereClause string, args []any) (int, error) {
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM kamar_inap ki%s
		%s
	`, filterJoins, whereClause)

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
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset()
	if offset < 0 {
		offset = 0
	}

	conditions, whereArgs := r.buildFilterConditions(kodeDokterLogin, filter)

	var whereClause string
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	pageKeyJoins := buildPageKeyJoins(kodeDokterLogin, filter)
	orderClause := buildOrderClause(filter.OrderBy, filter.SortOrder)

	dataQuery := fmt.Sprintf(`
		%s
		FROM (
			SELECT ki.no_rawat, ki.tgl_masuk, ki.jam_masuk
			FROM kamar_inap ki%s
			%s
			ORDER BY %s
			LIMIT ? OFFSET ?
		) AS page_keys
		INNER JOIN kamar_inap ki
			ON ki.no_rawat = page_keys.no_rawat
		   AND ki.tgl_masuk = page_keys.tgl_masuk
		   AND ki.jam_masuk = page_keys.jam_masuk
		%s
		ORDER BY %s
	`, selectPasienRawatInap, pageKeyJoins, whereClause, orderClause, lookupJoinsKamarInap, orderClause)

	finalArgs := make([]any, 0, len(whereArgs)+2)
	finalArgs = append(finalArgs, whereArgs...)
	finalArgs = append(finalArgs, limit, offset)

	list, err := r.scanPasienRows(ctx, dataQuery, finalArgs, limit)
	if err != nil {
		return nil, 0, err
	}

	if len(list) > 0 {
		if err := r.hydrateDPJP(ctx, list); err != nil {
			return nil, 0, err
		}
	}

	if offset == 0 && len(list) < limit {
		return list, len(list), nil
	}

	filterJoins := buildFilterJoins(kodeDokterLogin, filter)
	total, err := r.countPasienRawatInap(ctx, filterJoins, whereClause, whereArgs)
	if err != nil {
		return nil, 0, err
	}
	if len(list) > 0 && offset+len(list) > total {
		total = offset + len(list)
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

	slice := []KunjunganRawatInap{*item}
	if err := r.hydrateDPJP(ctx, slice); err != nil {
		return nil, err
	}
	item.DPJP = slice[0].DPJP

	return item, nil
}
