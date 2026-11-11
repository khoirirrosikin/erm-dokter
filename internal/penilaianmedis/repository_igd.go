package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/pkg/database"
)

const selectPenilaianMedisIGD = `
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
		pmr.mata,
		pmr.gigi,
		pmr.leher,
		pmr.thoraks,
		pmr.abdomen,
		pmr.genital,
		pmr.ekstremitas,
		COALESCE(pmr.ket_fisik, '') AS ket_fisik,
		COALESCE(pmr.ket_lokalis, '') AS ket_lokalis,
		COALESCE(pmr.ekg, '') AS ekg,
		COALESCE(pmr.rad, '') AS rad,
		COALESCE(pmr.lab, '') AS lab,
		COALESCE(pmr.diagnosis, '') AS diagnosis,
		COALESCE(pmr.tata, '') AS tata
	FROM penilaian_medis_igd pmr
	INNER JOIN dokter d ON d.kd_dokter = pmr.kd_dokter
`

func scanPenilaianMedisIGD(scanner interface{ Scan(dest ...any) error }) (*PenilaianMedisIGD, error) {
	var item PenilaianMedisIGD
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
		&item.Mata,
		&item.Gigi,
		&item.Leher,
		&item.Thoraks,
		&item.Abdomen,
		&item.Genital,
		&item.Ekstremitas,
		&item.KeteranganFisik,
		&item.KeteranganLokalis,
		&item.EKG,
		&item.Radiologi,
		&item.Laboratorium,
		&item.Diagnosis,
		&item.TataLaksana,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) DetailPenilaianMedisIGD(ctx context.Context, noRawat string) (*PenilaianMedisIGD, error) {
	query := selectPenilaianMedisIGD + ` WHERE pmr.no_rawat = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, noRawat)

	item, err := scanPenilaianMedisIGD(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return item, nil
}

func (r *repository) RiwayatPenilaianMedisIGDByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisIGD, error) {
	query := selectPenilaianMedisIGD + ` INNER JOIN reg_periksa rp ON rp.no_rawat = pmr.no_rawat WHERE rp.no_rkm_medis = ? ORDER BY pmr.tanggal DESC`
	rows, err := r.db.QueryContext(ctx, query, noRM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PenilaianMedisIGD, 0)
	for rows.Next() {
		item, err := scanPenilaianMedisIGD(rows)
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

func (r *repository) CekPenilaianMedisIGDAda(ctx context.Context, noRawat string) (bool, error) {
	query := `SELECT COUNT(1) FROM penilaian_medis_igd WHERE no_rawat = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const insertPenilaianMedisIGD = `
	INSERT INTO penilaian_medis_igd (
		no_rawat, tanggal, kd_dokter, anamnesis, hubungan, keluhan_utama,
		rps, rpd, rpk, rpo, alergi, keadaan, gcs, kesadaran,
		td, nadi, rr, suhu, spo, bb, tb,
		kepala, mata, gigi, leher, thoraks, abdomen, genital, ekstremitas,
		ket_fisik, ket_lokalis, ekg, rad, lab, diagnosis, tata
	) VALUES (
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?
	)
`

func (r *repository) SimpanPenilaianMedisIGD(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisIGDRequest) error {
	_, err := r.db.ExecContext(
		ctx,
		insertPenilaianMedisIGD,
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
		req.Mata,
		req.Gigi,
		req.Leher,
		req.Thoraks,
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.KeteranganFisik,
		req.KeteranganLokalis,
		req.EKG,
		req.Radiologi,
		req.Laboratorium,
		req.Diagnosis,
		req.TataLaksana,
	)
	if err != nil {
		return err
	}
	return nil
}

const updatePenilaianMedisIGD = `
	UPDATE penilaian_medis_igd SET
		tanggal = ?, anamnesis = ?, hubungan = ?, keluhan_utama = ?,
		rps = ?, rpd = ?, rpk = ?, rpo = ?, alergi = ?, keadaan = ?, gcs = ?, kesadaran = ?,
		td = ?, nadi = ?, rr = ?, suhu = ?, spo = ?, bb = ?, tb = ?,
		kepala = ?, mata = ?, gigi = ?, leher = ?, thoraks = ?, abdomen = ?, genital = ?, ekstremitas = ?,
		ket_fisik = ?, ket_lokalis = ?, ekg = ?, rad = ?, lab = ?, diagnosis = ?, tata = ?
	WHERE no_rawat = ?
`

func (r *repository) UpdatePenilaianMedisIGD(ctx context.Context, noRawat string, req UpdatePenilaianMedisIGDRequest) error {
	res, err := r.db.ExecContext(
		ctx,
		updatePenilaianMedisIGD,
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
		req.Mata,
		req.Gigi,
		req.Leher,
		req.Thoraks,
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.KeteranganFisik,
		req.KeteranganLokalis,
		req.EKG,
		req.Radiologi,
		req.Laboratorium,
		req.Diagnosis,
		req.TataLaksana,
		noRawat,
	)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}

func (r *repository) HapusPenilaianMedisIGD(ctx context.Context, noRawat string) error {
	query := `DELETE FROM penilaian_medis_igd WHERE no_rawat = ?`
	res, err := r.db.ExecContext(ctx, query, noRawat)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}
