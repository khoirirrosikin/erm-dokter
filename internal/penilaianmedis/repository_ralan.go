package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/pkg/database"
)

const selectPenilaianMedisRalan = `
	SELECT 
		pmr.no_rawat,
		COALESCE(DATE_FORMAT(pmr.tanggal, '%Y-%m-%d %H:%i:%s'), '') AS tanggal_penilaian,
		pmr.kd_dokter,
		COALESCE(d.nm_dokter, '') AS nama_dokter,
		pmr.anamnesis,
		COALESCE(pmr.hubungan, '') AS hubungan,
		COALESCE(pmr.keluhan_utama, '') AS keluhan_utama,
		COALESCE(pmr.rps, '') AS rps,
		COALESCE(pmr.rpd, '') AS rpd,
		COALESCE(pmr.rpk, '') AS rpk,
		COALESCE(pmr.rpo, '') AS rpo,
		COALESCE(pmr.alergi, '') AS alergi,
		pmr.keadaan,
		COALESCE(pmr.gcs, '') AS gcs,
		pmr.kesadaran,
		COALESCE(pmr.td, '') AS td,
		COALESCE(pmr.nadi, '') AS nadi,
		COALESCE(pmr.rr, '') AS rr,
		COALESCE(pmr.suhu, '') AS suhu,
		COALESCE(pmr.spo, '') AS spo,
		COALESCE(pmr.bb, '') AS bb,
		COALESCE(pmr.tb, '') AS tb,
		pmr.kepala,
		pmr.gigi,
		pmr.tht,
		pmr.thoraks,
		pmr.abdomen,
		pmr.genital,
		pmr.ekstremitas,
		pmr.kulit,
		COALESCE(pmr.ket_fisik, '') AS ket_fisik,
		COALESCE(pmr.ket_lokalis, '') AS ket_lokalis,
		COALESCE(pmr.penunjang, '') AS penunjang,
		COALESCE(pmr.diagnosis, '') AS diagnosis,
		COALESCE(pmr.tata, '') AS tata,
		COALESCE(pmr.konsulrujuk, '') AS konsulrujuk
	FROM penilaian_medis_ralan pmr
	INNER JOIN dokter d ON d.kd_dokter = pmr.kd_dokter
`

func scanPenilaianMedisRalan(scanner interface{ Scan(dest ...any) error }) (*PenilaianMedisRalan, error) {
	var item PenilaianMedisRalan
	err := scanner.Scan(
		&item.NoRawat,
		&item.TanggalPenilaian,
		&item.KodeDokter,
		&item.NamaDokter,
		&item.Anamnesis,
		&item.Hubungan,
		&item.KeluhanUtama,
		&item.RiwayatPenyakitSekarang,
		&item.RiwayatPenyakitDahulu,
		&item.RiwayatPenyakitKeluarga,
		&item.RiwayatPengobatan,
		&item.Alergi,
		&item.Keadaan,
		&item.Gcs,
		&item.Kesadaran,
		&item.Tensi,
		&item.Nadi,
		&item.Respirasi,
		&item.SuhuTubuh,
		&item.SpO2,
		&item.BeratBadan,
		&item.TinggiBadan,
		&item.Kepala,
		&item.Gigi,
		&item.TelingaHidungTenggorok,
		&item.Thoraks,
		&item.Abdomen,
		&item.Genital,
		&item.Ekstremitas,
		&item.Kulit,
		&item.KeteranganFisik,
		&item.KeteranganLokalis,
		&item.PemeriksaanPenunjang,
		&item.Diagnosis,
		&item.TataLaksana,
		&item.KonsultasiRujukan,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) DetailPenilaianMedisRalan(ctx context.Context, noRawat string) (*PenilaianMedisRalan, error) {
	query := selectPenilaianMedisRalan + ` WHERE pmr.no_rawat = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, noRawat)

	item, err := scanPenilaianMedisRalan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return item, nil
}

func (r *repository) RiwayatPenilaianMedisRalanByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalan, error) {
	query := selectPenilaianMedisRalan + ` INNER JOIN reg_periksa rp ON rp.no_rawat = pmr.no_rawat WHERE rp.no_rkm_medis = ? ORDER BY pmr.tanggal DESC`
	rows, err := r.db.QueryContext(ctx, query, noRM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PenilaianMedisRalan, 0)
	for rows.Next() {
		item, err := scanPenilaianMedisRalan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) CekPenilaianMedisRalanAda(ctx context.Context, noRawat string) (bool, error) {
	query := `SELECT COUNT(1) FROM penilaian_medis_ralan WHERE no_rawat = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const insertPenilaianMedisRalan = `
	INSERT INTO penilaian_medis_ralan (
		no_rawat, tanggal, kd_dokter, anamnesis, hubungan, keluhan_utama,
		rps, rpd, rpk, rpo, alergi, keadaan, gcs, kesadaran,
		td, nadi, rr, suhu, spo, bb, tb,
		kepala, gigi, tht, thoraks, abdomen, genital, ekstremitas, kulit,
		ket_fisik, ket_lokalis, penunjang, diagnosis, tata, konsulrujuk
	) VALUES (
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?
	)
`

func (r *repository) SimpanPenilaianMedisRalan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRalanRequest) error {
	_, err := r.db.ExecContext(
		ctx,
		insertPenilaianMedisRalan,
		noRawat,
		req.TanggalPenilaian,
		kodeDokter,
		req.Anamnesis,
		req.Hubungan,
		req.KeluhanUtama,
		req.RiwayatPenyakitSekarang,
		req.RiwayatPenyakitDahulu,
		req.RiwayatPenyakitKeluarga,
		req.RiwayatPengobatan,
		req.Alergi,
		req.Keadaan,
		req.Gcs,
		req.Kesadaran,
		req.Tensi,
		req.Nadi,
		req.Respirasi,
		req.SuhuTubuh,
		req.SpO2,
		req.BeratBadan,
		req.TinggiBadan,
		req.Kepala,
		req.Gigi,
		req.TelingaHidungTenggorok,
		req.Thoraks,
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.Kulit,
		req.KeteranganFisik,
		req.KeteranganLokalis,
		req.PemeriksaanPenunjang,
		req.Diagnosis,
		req.TataLaksana,
		req.KonsultasiRujukan,
	)
	if err != nil {
		return err
	}
	return nil
}

const updatePenilaianMedisRalan = `
	UPDATE penilaian_medis_ralan SET
		tanggal = ?, anamnesis = ?, hubungan = ?, keluhan_utama = ?,
		rps = ?, rpd = ?, rpk = ?, rpo = ?, alergi = ?, keadaan = ?, gcs = ?, kesadaran = ?,
		td = ?, nadi = ?, rr = ?, suhu = ?, spo = ?, bb = ?, tb = ?,
		kepala = ?, gigi = ?, tht = ?, thoraks = ?, abdomen = ?, genital = ?, ekstremitas = ?, kulit = ?,
		ket_fisik = ?, ket_lokalis = ?, penunjang = ?, diagnosis = ?, tata = ?, konsulrujuk = ?
	WHERE no_rawat = ?
`

func (r *repository) UpdatePenilaianMedisRalan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRalanRequest) error {
	res, err := r.db.ExecContext(
		ctx,
		updatePenilaianMedisRalan,
		req.TanggalPenilaian,
		req.Anamnesis,
		req.Hubungan,
		req.KeluhanUtama,
		req.RiwayatPenyakitSekarang,
		req.RiwayatPenyakitDahulu,
		req.RiwayatPenyakitKeluarga,
		req.RiwayatPengobatan,
		req.Alergi,
		req.Keadaan,
		req.Gcs,
		req.Kesadaran,
		req.Tensi,
		req.Nadi,
		req.Respirasi,
		req.SuhuTubuh,
		req.SpO2,
		req.BeratBadan,
		req.TinggiBadan,
		req.Kepala,
		req.Gigi,
		req.TelingaHidungTenggorok,
		req.Thoraks,
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.Kulit,
		req.KeteranganFisik,
		req.KeteranganLokalis,
		req.PemeriksaanPenunjang,
		req.Diagnosis,
		req.TataLaksana,
		req.KonsultasiRujukan,
		noRawat,
	)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}

func (r *repository) HapusPenilaianMedisRalan(ctx context.Context, noRawat string) error {
	query := `DELETE FROM penilaian_medis_ralan WHERE no_rawat = ?`
	res, err := r.db.ExecContext(ctx, query, noRawat)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}
