package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/pkg/database"
)

const selectPenilaianMedisRanapKandungan = `
	SELECT 
		pmrnk.no_rawat,
		COALESCE(DATE_FORMAT(pmrnk.tanggal, '%Y-%m-%d %H:%i:%s'), '') AS tanggal_penilaian,
		pmrnk.kd_dokter,
		COALESCE(d.nm_dokter, '') AS nama_dokter,
		pmrnk.anamnesis,
		COALESCE(pmrnk.hubungan, '') AS hubungan,
		COALESCE(pmrnk.keluhan_utama, '') AS keluhan_utama,
		COALESCE(pmrnk.rps, '') AS rps,
		COALESCE(pmrnk.rpd, '') AS rpd,
		COALESCE(pmrnk.rpk, '') AS rpk,
		COALESCE(pmrnk.rpo, '') AS rpo,
		COALESCE(pmrnk.alergi, '') AS alergi,
		pmrnk.keadaan,
		COALESCE(pmrnk.gcs, '') AS gcs,
		pmrnk.kesadaran,
		COALESCE(pmrnk.td, '') AS td,
		COALESCE(pmrnk.nadi, '') AS nadi,
		COALESCE(pmrnk.rr, '') AS rr,
		COALESCE(pmrnk.suhu, '') AS suhu,
		COALESCE(pmrnk.spo, '') AS spo,
		COALESCE(pmrnk.bb, '') AS bb,
		COALESCE(pmrnk.tb, '') AS tb,
		pmrnk.kepala,
		pmrnk.mata,
		pmrnk.gigi,
		pmrnk.tht,
		pmrnk.thoraks,
		pmrnk.jantung,
		pmrnk.paru,
		pmrnk.abdomen,
		pmrnk.genital,
		pmrnk.ekstremitas,
		pmrnk.kulit,
		COALESCE(pmrnk.ket_fisik, '') AS ket_fisik,
		COALESCE(pmrnk.tfu, '') AS tfu,
		COALESCE(pmrnk.tbj, '') AS tbj,
		COALESCE(pmrnk.his, '') AS his,
		pmrnk.kontraksi,
		COALESCE(pmrnk.djj, '') AS djj,
		COALESCE(pmrnk.inspeksi, '') AS inspeksi,
		COALESCE(pmrnk.inspekulo, '') AS inspekulo,
		COALESCE(pmrnk.vt, '') AS vt,
		COALESCE(pmrnk.rt, '') AS rt,
		COALESCE(pmrnk.ultra, '') AS ultra,
		COALESCE(pmrnk.kardio, '') AS kardio,
		COALESCE(pmrnk.lab, '') AS lab,
		COALESCE(pmrnk.diagnosis, '') AS diagnosis,
		COALESCE(pmrnk.tata, '') AS tata,
		COALESCE(pmrnk.edukasi, '') AS edukasi
	FROM penilaian_medis_ranap_kandungan pmrnk
	INNER JOIN dokter d ON d.kd_dokter = pmrnk.kd_dokter
`

func scanPenilaianMedisRanapKandungan(scanner interface{ Scan(dest ...any) error }) (*PenilaianMedisRanapKandungan, error) {
	var item PenilaianMedisRanapKandungan
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
		&item.Edukasi,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) DetailPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRanapKandungan, error) {
	query := selectPenilaianMedisRanapKandungan + ` WHERE pmrnk.no_rawat = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, noRawat)

	item, err := scanPenilaianMedisRanapKandungan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return item, nil
}

func (r *repository) RiwayatPenilaianMedisRanapKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanapKandungan, error) {
	query := selectPenilaianMedisRanapKandungan + ` INNER JOIN reg_periksa rp ON rp.no_rawat = pmrnk.no_rawat WHERE rp.no_rkm_medis = ? ORDER BY pmrnk.tanggal DESC`
	rows, err := r.db.QueryContext(ctx, query, noRM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]PenilaianMedisRanapKandungan, 0)
	for rows.Next() {
		item, err := scanPenilaianMedisRanapKandungan(rows)
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

func (r *repository) CekPenilaianMedisRanapKandunganAda(ctx context.Context, noRawat string) (bool, error) {
	query := `SELECT COUNT(1) FROM penilaian_medis_ranap_kandungan WHERE no_rawat = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const insertPenilaianMedisRanapKandungan = `
	INSERT INTO penilaian_medis_ranap_kandungan (
		no_rawat, tanggal, kd_dokter, anamnesis, hubungan, keluhan_utama,
		rps, rpd, rpk, rpo, alergi, keadaan, gcs, kesadaran,
		td, nadi, rr, suhu, spo, bb, tb,
		kepala, mata, gigi, tht, thoraks, jantung, paru, abdomen, genital, ekstremitas, kulit,
		ket_fisik, tfu, tbj, his, kontraksi, djj,
		inspeksi, inspekulo, vt, rt, ultra, kardio, lab,
		diagnosis, tata, edukasi
	) VALUES (
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?,
		?, ?, ?
	)
`

func (r *repository) SimpanPenilaianMedisRanapKandungan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRanapKandunganRequest) error {
	_, err := r.db.ExecContext(
		ctx,
		insertPenilaianMedisRanapKandungan,
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
		req.Edukasi,
	)
	if err != nil {
		return err
	}
	return nil
}

const updatePenilaianMedisRanapKandungan = `
	UPDATE penilaian_medis_ranap_kandungan SET
		tanggal = ?, anamnesis = ?, hubungan = ?, keluhan_utama = ?,
		rps = ?, rpd = ?, rpk = ?, rpo = ?, alergi = ?, keadaan = ?, gcs = ?, kesadaran = ?,
		td = ?, nadi = ?, rr = ?, suhu = ?, spo = ?, bb = ?, tb = ?,
		kepala = ?, mata = ?, gigi = ?, tht = ?, thoraks = ?, jantung = ?, paru = ?, abdomen = ?, genital = ?, ekstremitas = ?, kulit = ?,
		ket_fisik = ?, tfu = ?, tbj = ?, his = ?, kontraksi = ?, djj = ?,
		inspeksi = ?, inspekulo = ?, vt = ?, rt = ?, ultra = ?, kardio = ?, lab = ?,
		diagnosis = ?, tata = ?, edukasi = ?
	WHERE no_rawat = ?
`

func (r *repository) UpdatePenilaianMedisRanapKandungan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRanapKandunganRequest) error {
	res, err := r.db.ExecContext(
		ctx,
		updatePenilaianMedisRanapKandungan,
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
		req.Edukasi,
		noRawat,
	)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}

func (r *repository) HapusPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) error {
	query := `DELETE FROM penilaian_medis_ranap_kandungan WHERE no_rawat = ?`
	res, err := r.db.ExecContext(ctx, query, noRawat)
	if err != nil {
		return err
	}

	return database.CheckRowsAffected(res)
}
