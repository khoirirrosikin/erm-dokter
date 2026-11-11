package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/pkg/database"
)

const selectPenilaianMedisRalanKandungan = `
	SELECT 
		pmrk.no_rawat,
		COALESCE(DATE_FORMAT(pmrk.tanggal, '%Y-%m-%d %H:%i:%s'), '') AS tanggal_penilaian,
		pmrk.kd_dokter,
		COALESCE(d.nm_dokter, '') AS nama_dokter,
		pmrk.anamnesis,
		COALESCE(pmrk.hubungan, '') AS hubungan,
		COALESCE(pmrk.keluhan_utama, '') AS keluhan_utama,
		COALESCE(pmrk.rps, '') AS rps,
		COALESCE(pmrk.rpd, '') AS rpd,
		COALESCE(pmrk.rpk, '') AS rpk,
		COALESCE(pmrk.rpo, '') AS rpo,
		COALESCE(pmrk.alergi, '') AS alergi,
		pmrk.keadaan,
		COALESCE(pmrk.gcs, '') AS gcs,
		pmrk.kesadaran,
		COALESCE(pmrk.td, '') AS td,
		COALESCE(pmrk.nadi, '') AS nadi,
		COALESCE(pmrk.rr, '') AS rr,
		COALESCE(pmrk.suhu, '') AS suhu,
		COALESCE(pmrk.spo, '') AS spo,
		COALESCE(pmrk.bb, '') AS bb,
		COALESCE(pmrk.tb, '') AS tb,
		pmrk.kepala,
		pmrk.mata,
		pmrk.gigi,
		pmrk.tht,
		pmrk.thoraks,
		pmrk.abdomen,
		pmrk.genital,
		pmrk.ekstremitas,
		pmrk.kulit,
		COALESCE(pmrk.ket_fisik, '') AS ket_fisik,
		COALESCE(pmrk.tfu, '') AS tfu,
		COALESCE(pmrk.tbj, '') AS tbj,
		COALESCE(pmrk.his, '') AS his,
		pmrk.kontraksi,
		COALESCE(pmrk.djj, '') AS djj,
		COALESCE(pmrk.inspeksi, '') AS inspeksi,
		COALESCE(pmrk.inspekulo, '') AS inspekulo,
		COALESCE(pmrk.vt, '') AS vt,
		COALESCE(pmrk.rt, '') AS rt,
		COALESCE(pmrk.ultra, '') AS ultra,
		COALESCE(pmrk.kardio, '') AS kardio,
		COALESCE(pmrk.lab, '') AS lab,
		COALESCE(pmrk.diagnosis, '') AS diagnosis,
		COALESCE(pmrk.tata, '') AS tata,
		COALESCE(pmrk.konsul, '') AS konsul
	FROM penilaian_medis_ralan_kandungan pmrk
	INNER JOIN dokter d ON d.kd_dokter = pmrk.kd_dokter
`

func scanPenilaianMedisRalanKandungan(scanner interface{ Scan(dest ...any) error }) (*PenilaianMedisRalanKandungan, error) {
	var item PenilaianMedisRalanKandungan
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
		&item.Abdomen,
		&item.Genital,
		&item.Ekstremitas,
		&item.Kulit,
		&item.KeteranganFisik,
		&item.TinggiFundusUteri,
		&item.TaksiranBeratJanin,
		&item.His,
		&item.Kontraksi,
		&item.DenyutJantungJanin,
		&item.Inspeksi,
		&item.Inspekulo,
		&item.VaginalToucher,
		&item.RectalToucher,
		&item.Ultrasonografi,
		&item.Kardiotokografi,
		&item.Laboratorium,
		&item.Diagnosis,
		&item.TataLaksana,
		&item.KonsultasiRujukan,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) DetailPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRalanKandungan, error) {
	query := selectPenilaianMedisRalanKandungan + ` WHERE pmrk.no_rawat = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, noRawat)

	item, err := scanPenilaianMedisRalanKandungan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return item, nil
}

func (r *repository) RiwayatPenilaianMedisRalanKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalanKandungan, error) {
	query := selectPenilaianMedisRalanKandungan + ` INNER JOIN reg_periksa rp ON rp.no_rawat = pmrk.no_rawat WHERE rp.no_rkm_medis = ? ORDER BY pmrk.tanggal DESC`
	rows, err := r.db.QueryContext(ctx, query, noRM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PenilaianMedisRalanKandungan, 0)
	for rows.Next() {
		item, err := scanPenilaianMedisRalanKandungan(rows)
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

func (r *repository) CekPenilaianMedisRalanKandunganAda(ctx context.Context, noRawat string) (bool, error) {
	query := `SELECT COUNT(1) FROM penilaian_medis_ralan_kandungan WHERE no_rawat = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const insertPenilaianMedisRalanKandungan = `
	INSERT INTO penilaian_medis_ralan_kandungan (
		no_rawat, tanggal, kd_dokter, anamnesis, hubungan, keluhan_utama,
		rps, rpd, rpk, rpo, alergi, keadaan, gcs, kesadaran,
		td, nadi, rr, suhu, spo, bb, tb,
		kepala, mata, gigi, tht, thoraks, abdomen, genital, ekstremitas, kulit,
		ket_fisik, tfu, tbj, his, kontraksi, djj,
		inspeksi, inspekulo, vt, rt, ultra, kardio, lab,
		diagnosis, tata, konsul
	) VALUES (
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?
	)
`

func (r *repository) SimpanPenilaianMedisRalanKandungan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRalanKandunganRequest) error {
	_, err := r.db.ExecContext(
		ctx,
		insertPenilaianMedisRalanKandungan,
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
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.Kulit,
		req.KeteranganFisik,
		req.TinggiFundusUteri,
		req.TaksiranBeratJanin,
		req.His,
		req.Kontraksi,
		req.DenyutJantungJanin,
		req.Inspeksi,
		req.Inspekulo,
		req.VaginalToucher,
		req.RectalToucher,
		req.Ultrasonografi,
		req.Kardiotokografi,
		req.Laboratorium,
		req.Diagnosis,
		req.TataLaksana,
		req.KonsultasiRujukan,
	)
	if err != nil {
		return err
	}
	return nil
}

const updatePenilaianMedisRalanKandungan = `
	UPDATE penilaian_medis_ralan_kandungan SET
		tanggal = ?, anamnesis = ?, hubungan = ?, keluhan_utama = ?,
		rps = ?, rpd = ?, rpk = ?, rpo = ?, alergi = ?, keadaan = ?, gcs = ?, kesadaran = ?,
		td = ?, nadi = ?, rr = ?, suhu = ?, spo = ?, bb = ?, tb = ?,
		kepala = ?, mata = ?, gigi = ?, tht = ?, thoraks = ?, abdomen = ?, genital = ?, ekstremitas = ?, kulit = ?,
		ket_fisik = ?, tfu = ?, tbj = ?, his = ?, kontraksi = ?, djj = ?,
		inspeksi = ?, inspekulo = ?, vt = ?, rt = ?, ultra = ?, kardio = ?, lab = ?,
		diagnosis = ?, tata = ?, konsul = ?
	WHERE no_rawat = ?
`

func (r *repository) UpdatePenilaianMedisRalanKandungan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRalanKandunganRequest) error {
	res, err := r.db.ExecContext(
		ctx,
		updatePenilaianMedisRalanKandungan,
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
		req.Abdomen,
		req.Genital,
		req.Ekstremitas,
		req.Kulit,
		req.KeteranganFisik,
		req.TinggiFundusUteri,
		req.TaksiranBeratJanin,
		req.His,
		req.Kontraksi,
		req.DenyutJantungJanin,
		req.Inspeksi,
		req.Inspekulo,
		req.VaginalToucher,
		req.RectalToucher,
		req.Ultrasonografi,
		req.Kardiotokografi,
		req.Laboratorium,
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

func (r *repository) HapusPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) error {
	query := `DELETE FROM penilaian_medis_ralan_kandungan WHERE no_rawat = ?`
	res, err := r.db.ExecContext(ctx, query, noRawat)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}
