package penilaianmedis

import (
	"context"

	"erm-dokter/internal/pkg/database"
)

const selectPenilaianMedisRanap = `
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
		pmr.tht,
		pmr.thoraks,
		pmr.jantung,
		pmr.paru,
		pmr.abdomen,
		pmr.genital,
		pmr.ekstremitas,
		pmr.kulit,
		COALESCE(pmr.ket_fisik, '') AS ket_fisik,
		COALESCE(pmr.ket_lokalis, '') AS ket_lokalis,
		COALESCE(pmr.lab, '') AS lab,
		COALESCE(pmr.rad, '') AS rad,
		COALESCE(pmr.penunjang, '') AS penunjang,
		COALESCE(pmr.diagnosis, '') AS diagnosis,
		COALESCE(pmr.tata, '') AS tata,
		COALESCE(pmr.edukasi, '') AS edukasi
	FROM penilaian_medis_ranap pmr
	INNER JOIN dokter d ON d.kd_dokter = pmr.kd_dokter
`

func scanPenilaianMedisRanap(scanner interface{ Scan(dest ...any) error }) (*PenilaianMedisRanap, error) {
	var item PenilaianMedisRanap
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
		&item.TelingaHidungTenggorok,
		&item.Thoraks,
		&item.Jantung,
		&item.Paru,
		&item.Abdomen,
		&item.Genital,
		&item.Ekstremitas,
		&item.Kulit,
		&item.KeteranganFisik,
		&item.KeteranganLokalis,
		&item.Laboratorium,
		&item.Radiologi,
		&item.PemeriksaanPenunjang,
		&item.Diagnosis,
		&item.TataLaksana,
		&item.Edukasi,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) DetailPenilaianMedisRanap(ctx context.Context, noRawat string) (*PenilaianMedisRanap, error) {
	query := selectPenilaianMedisRanap + " WHERE pmr.no_rawat = ?"
	row := r.db.QueryRowContext(ctx, query, noRawat)
	return scanPenilaianMedisRanap(row)
}

func (r *repository) RiwayatPenilaianMedisRanapByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanap, error) {
	query := selectPenilaianMedisRanap + `
		INNER JOIN reg_periksa rp ON rp.no_rawat = pmr.no_rawat
		WHERE rp.no_rkm_medis = ?
		ORDER BY pmr.tanggal DESC
	`

	rows, err := r.db.QueryContext(ctx, query, noRM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PenilaianMedisRanap, 0)
	for rows.Next() {
		item, err := scanPenilaianMedisRanap(rows)
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

func (r *repository) CekPenilaianMedisRanapAda(ctx context.Context, noRawat string) (bool, error) {
	query := `SELECT COUNT(1) FROM penilaian_medis_ranap WHERE no_rawat = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const insertPenilaianMedisRanap = `
	INSERT INTO penilaian_medis_ranap (
		no_rawat, tanggal, kd_dokter, anamnesis, hubungan, keluhan_utama,
		rps, rpd, rpk, rpo, alergi, keadaan, gcs, kesadaran,
		td, nadi, rr, suhu, spo, bb, tb,
		kepala, mata, gigi, tht, thoraks, jantung, paru, abdomen, genital, ekstremitas, kulit,
		ket_fisik, ket_lokalis, lab, rad, penunjang, diagnosis, tata, edukasi
	) VALUES (
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?
	)
`

func (r *repository) SimpanPenilaianMedisRanap(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRanapRequest) error {
	_, err := r.db.ExecContext(
		ctx,
		insertPenilaianMedisRanap,
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
		req.TelingaHidungTenggorok,
		req.Thoraks,
		req.Jantung,
		req.Paru,
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.Kulit,
		req.KeteranganFisik,
		req.KeteranganLokalis,
		req.Laboratorium,
		req.Radiologi,
		req.PemeriksaanPenunjang,
		req.Diagnosis,
		req.TataLaksana,
		req.Edukasi,
	)
	if err != nil {
		return err
	}
	return nil
}

const updatePenilaianMedisRanap = `
	UPDATE penilaian_medis_ranap SET
		tanggal = ?, anamnesis = ?, hubungan = ?, keluhan_utama = ?,
		rps = ?, rpd = ?, rpk = ?, rpo = ?, alergi = ?, keadaan = ?, gcs = ?, kesadaran = ?,
		td = ?, nadi = ?, rr = ?, suhu = ?, spo = ?, bb = ?, tb = ?,
		kepala = ?, mata = ?, gigi = ?, tht = ?, thoraks = ?, jantung = ?, paru = ?, abdomen = ?, genital = ?, ekstremitas = ?, kulit = ?,
		ket_fisik = ?, ket_lokalis = ?, lab = ?, rad = ?, penunjang = ?, diagnosis = ?, tata = ?, edukasi = ?
	WHERE no_rawat = ?
`

func (r *repository) UpdatePenilaianMedisRanap(ctx context.Context, noRawat string, req UpdatePenilaianMedisRanapRequest) error {
	res, err := r.db.ExecContext(
		ctx,
		updatePenilaianMedisRanap,
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
		req.TelingaHidungTenggorok,
		req.Thoraks,
		req.Jantung,
		req.Paru,
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.Kulit,
		req.KeteranganFisik,
		req.KeteranganLokalis,
		req.Laboratorium,
		req.Radiologi,
		req.PemeriksaanPenunjang,
		req.Diagnosis,
		req.TataLaksana,
		req.Edukasi,
		noRawat,
	)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}

func (r *repository) HapusPenilaianMedisRanap(ctx context.Context, noRawat string) error {
	query := `DELETE FROM penilaian_medis_ranap WHERE no_rawat = ?`
	res, err := r.db.ExecContext(ctx, query, noRawat)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}
