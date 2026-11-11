package resumepasien

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/pkg/database"
)

const selectResumePasienRanap = `
	SELECT 
		rpr.no_rawat,
		rpr.kd_dokter,
		COALESCE(d.nm_dokter, '') AS nama_dokter,
		COALESCE(rpr.diagnosa_awal, '') AS diagnosa_awal,
		COALESCE(rpr.alasan, '') AS alasan,
		COALESCE(rpr.keluhan_utama, '') AS keluhan_utama,
		COALESCE(rpr.pemeriksaan_fisik, '') AS pemeriksaan_fisik,
		COALESCE(rpr.jalannya_penyakit, '') AS jalannya_penyakit,
		COALESCE(rpr.pemeriksaan_penunjang, '') AS pemeriksaan_penunjang,
		COALESCE(rpr.hasil_laborat, '') AS hasil_laborat,
		COALESCE(rpr.tindakan_dan_operasi, '') AS tindakan_dan_operasi,
		COALESCE(rpr.obat_di_rs, '') AS obat_di_rs,
		COALESCE(rpr.diagnosa_utama, '') AS diagnosa_utama,
		COALESCE(rpr.kd_diagnosa_utama, '') AS kd_diagnosa_utama,
		COALESCE(rpr.diagnosa_sekunder, '') AS diagnosa_sekunder,
		COALESCE(rpr.kd_diagnosa_sekunder, '') AS kd_diagnosa_sekunder,
		COALESCE(rpr.diagnosa_sekunder2, '') AS diagnosa_sekunder2,
		COALESCE(rpr.kd_diagnosa_sekunder2, '') AS kd_diagnosa_sekunder2,
		COALESCE(rpr.diagnosa_sekunder3, '') AS diagnosa_sekunder3,
		COALESCE(rpr.kd_diagnosa_sekunder3, '') AS kd_diagnosa_sekunder3,
		COALESCE(rpr.diagnosa_sekunder4, '') AS diagnosa_sekunder4,
		COALESCE(rpr.kd_diagnosa_sekunder4, '') AS kd_diagnosa_sekunder4,
		COALESCE(rpr.prosedur_utama, '') AS prosedur_utama,
		COALESCE(rpr.kd_prosedur_utama, '') AS kd_prosedur_utama,
		COALESCE(rpr.prosedur_sekunder, '') AS prosedur_sekunder,
		COALESCE(rpr.kd_prosedur_sekunder, '') AS kd_prosedur_sekunder,
		COALESCE(rpr.prosedur_sekunder2, '') AS prosedur_sekunder2,
		COALESCE(rpr.kd_prosedur_sekunder2, '') AS kd_prosedur_sekunder2,
		COALESCE(rpr.prosedur_sekunder3, '') AS prosedur_sekunder3,
		COALESCE(rpr.kd_prosedur_sekunder3, '') AS kd_prosedur_sekunder3,
		COALESCE(rpr.alergi, '') AS alergi,
		COALESCE(rpr.diet, '') AS diet,
		COALESCE(rpr.lab_belum, '') AS lab_belum,
		COALESCE(rpr.edukasi, '') AS edukasi,
		rpr.cara_keluar,
		COALESCE(rpr.ket_keluar, '') AS ket_keluar,
		rpr.keadaan,
		COALESCE(rpr.ket_keadaan, '') AS ket_keadaan,
		rpr.dilanjutkan,
		COALESCE(rpr.ket_dilanjutkan, '') AS ket_dilanjutkan,
		COALESCE(DATE_FORMAT(rpr.kontrol, '%Y-%m-%d %H:%i:%s'), '') AS kontrol,
		COALESCE(rpr.obat_pulang, '') AS obat_pulang,
		COALESCE(ki.tgl_masuk, '') AS tgl_masuk,
		COALESCE(ki.jam_masuk, '') AS jam_masuk,
		COALESCE(ki.tgl_keluar, '') AS tgl_keluar,
		COALESCE(ki.jam_keluar, '') AS jam_keluar
	FROM resume_pasien_ranap rpr
	INNER JOIN dokter d ON d.kd_dokter = rpr.kd_dokter
	LEFT JOIN (
		SELECT 
			no_rawat,
			MIN(tgl_masuk) AS tgl_masuk,
			MIN(jam_masuk) AS jam_masuk,
			MAX(tgl_keluar) AS tgl_keluar,
			MAX(jam_keluar) AS jam_keluar
		FROM kamar_inap
		GROUP BY no_rawat
	) ki ON ki.no_rawat = rpr.no_rawat
`

func scanResumePasienRanap(scanner interface{ Scan(dest ...any) error }) (*ResumePasienRanap, error) {
	var item ResumePasienRanap
	var (
		caraKeluar  string
		keadaan     string
		dilanjutkan string
		tglKeluar   string
		jamKeluar   string
	)

	err := scanner.Scan(
		&item.NoRawat,
		&item.KodeDokter,
		&item.NamaDokter,
		&item.DiagnosaAwal,
		&item.AlasanRawat,
		&item.KeluhanUtama,
		&item.PemeriksaanFisik,
		&item.JalannyaPenyakit,
		&item.HasilPemeriksaanRadiologi,
		&item.HasilPemeriksaanLaboratorium,
		&item.TindakanAtauOperasi,
		&item.ObatSelamaPerawatan,
		&item.DiagnosaUtama,
		&item.KdDiagnosaUtama,
		&item.DiagnosaSekunder,
		&item.KdDiagnosaSekunder,
		&item.DiagnosaSekunder2,
		&item.KdDiagnosaSekunder2,
		&item.DiagnosaSekunder3,
		&item.KdDiagnosaSekunder3,
		&item.DiagnosaSekunder4,
		&item.KdDiagnosaSekunder4,
		&item.ProsedurUtama,
		&item.KdProsedurUtama,
		&item.ProsedurSekunder,
		&item.KdProsedurSekunder,
		&item.ProsedurSekunder2,
		&item.KdProsedurSekunder2,
		&item.ProsedurSekunder3,
		&item.KdProsedurSekunder3,
		&item.Alergi,
		&item.Diet,
		&item.HasilLaboratoriumPending,
		&item.InstruksiAtauEdukasi,
		&caraKeluar,
		&item.KeteranganKeluar,
		&keadaan,
		&item.KeteranganKeadaanPulang,
		&dilanjutkan,
		&item.KeteranganDilanjutkan,
		&item.WaktuKontrol,
		&item.ObatPulang,
		&item.TanggalMasuk,
		&item.JamMasuk,
		&tglKeluar,
		&jamKeluar,
	)
	if err != nil {
		return nil, err
	}

	item.CaraKeluar = CaraKeluar(caraKeluar)
	item.KeadaanPulang = KeadaanPulang(keadaan)
	item.Dilanjutkan = Dilanjutkan(dilanjutkan)

	if tglKeluar != "0000-00-00" {
		item.TanggalKeluar = tglKeluar
	}
	if jamKeluar != "00:00:00" {
		item.JamKeluar = jamKeluar
	}

	return &item, nil
}

func (r *repository) DetailResumePasienRanap(ctx context.Context, noRawat string) (*ResumePasienRanap, error) {
	query := selectResumePasienRanap + " WHERE rpr.no_rawat = ?"
	row := r.db.QueryRowContext(ctx, query, noRawat)
	item, err := scanResumePasienRanap(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return item, nil
}

func (r *repository) RiwayatResumePasienRanapByNoRM(ctx context.Context, noRM string) ([]ResumePasienRanap, error) {
	query := selectResumePasienRanap + `
		INNER JOIN reg_periksa rp ON rp.no_rawat = rpr.no_rawat
		WHERE rp.no_rkm_medis = ?
		ORDER BY rp.tgl_registrasi DESC, rp.jam_reg DESC
	`
	rows, err := r.db.QueryContext(ctx, query, noRM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ResumePasienRanap
	for rows.Next() {
		item, err := scanResumePasienRanap(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []ResumePasienRanap{}
	}
	return list, nil
}

func (r *repository) CekResumePasienRanapAda(ctx context.Context, noRawat string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM resume_pasien_ranap WHERE no_rawat = ?)"
	var exists bool
	err := r.db.QueryRowContext(ctx, query, noRawat).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *repository) SimpanResumePasienRanap(ctx context.Context, noRawat, kodeDokter string, req SimpanResumePasienRanapRequest) (*ResumePasienRanap, error) {
	query := `
		INSERT INTO resume_pasien_ranap (
			no_rawat, kd_dokter, diagnosa_awal, alasan, keluhan_utama, pemeriksaan_fisik,
			jalannya_penyakit, pemeriksaan_penunjang, hasil_laborat, tindakan_dan_operasi,
			obat_di_rs, diagnosa_utama, kd_diagnosa_utama, diagnosa_sekunder, kd_diagnosa_sekunder,
			diagnosa_sekunder2, kd_diagnosa_sekunder2, diagnosa_sekunder3, kd_diagnosa_sekunder3,
			diagnosa_sekunder4, kd_diagnosa_sekunder4, prosedur_utama, kd_prosedur_utama,
			prosedur_sekunder, kd_prosedur_sekunder, prosedur_sekunder2, kd_prosedur_sekunder2,
			prosedur_sekunder3, kd_prosedur_sekunder3, alergi, diet, lab_belum, edukasi,
			cara_keluar, ket_keluar, keadaan, ket_keadaan, dilanjutkan, ket_dilanjutkan,
			kontrol, obat_pulang
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?
		)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		noRawat,
		kodeDokter,
		req.DiagnosaAwal,
		req.AlasanRawat,
		req.KeluhanUtama,
		req.PemeriksaanFisik,
		req.JalannyaPenyakit,
		req.HasilPemeriksaanRadiologi,
		req.HasilPemeriksaanLaboratorium,
		req.TindakanAtauOperasi,
		req.ObatSelamaPerawatan,
		req.DiagnosaUtama,
		req.KdDiagnosaUtama,
		req.DiagnosaSekunder,
		req.KdDiagnosaSekunder,
		req.DiagnosaSekunder2,
		req.KdDiagnosaSekunder2,
		req.DiagnosaSekunder3,
		req.KdDiagnosaSekunder3,
		req.DiagnosaSekunder4,
		req.KdDiagnosaSekunder4,
		req.ProsedurUtama,
		req.KdProsedurUtama,
		req.ProsedurSekunder,
		req.KdProsedurSekunder,
		req.ProsedurSekunder2,
		req.KdProsedurSekunder2,
		req.ProsedurSekunder3,
		req.KdProsedurSekunder3,
		req.Alergi,
		req.Diet,
		req.HasilLaboratoriumPending,
		req.InstruksiAtauEdukasi,
		string(req.CaraKeluar),
		req.KeteranganKeluar,
		string(req.KeadaanPulang),
		req.KeteranganKeadaanPulang,
		string(req.Dilanjutkan),
		req.KeteranganDilanjutkan,
		req.WaktuKontrol,
		req.ObatPulang,
	)
	if err != nil {
		return nil, err
	}

	return r.DetailResumePasienRanap(ctx, noRawat)
}

func (r *repository) UpdateResumePasienRanap(ctx context.Context, noRawat string, req UpdateResumePasienRanapRequest) (*ResumePasienRanap, error) {
	query := `
		UPDATE resume_pasien_ranap SET
			diagnosa_awal = ?, alasan = ?, keluhan_utama = ?, pemeriksaan_fisik = ?,
			jalannya_penyakit = ?, pemeriksaan_penunjang = ?, hasil_laborat = ?, tindakan_dan_operasi = ?,
			obat_di_rs = ?, diagnosa_utama = ?, kd_diagnosa_utama = ?, diagnosa_sekunder = ?, kd_diagnosa_sekunder = ?,
			diagnosa_sekunder2 = ?, kd_diagnosa_sekunder2 = ?, diagnosa_sekunder3 = ?, kd_diagnosa_sekunder3 = ?,
			diagnosa_sekunder4 = ?, kd_diagnosa_sekunder4 = ?, prosedur_utama = ?, kd_prosedur_utama = ?,
			prosedur_sekunder = ?, kd_prosedur_sekunder = ?, prosedur_sekunder2 = ?, kd_prosedur_sekunder2 = ?,
			prosedur_sekunder3 = ?, kd_prosedur_sekunder3 = ?, alergi = ?, diet = ?, lab_belum = ?, edukasi = ?,
			cara_keluar = ?, ket_keluar = ?, keadaan = ?, ket_keadaan = ?, dilanjutkan = ?, ket_dilanjutkan = ?,
			kontrol = ?, obat_pulang = ?
		WHERE no_rawat = ?
	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		req.DiagnosaAwal,
		req.AlasanRawat,
		req.KeluhanUtama,
		req.PemeriksaanFisik,
		req.JalannyaPenyakit,
		req.HasilPemeriksaanRadiologi,
		req.HasilPemeriksaanLaboratorium,
		req.TindakanAtauOperasi,
		req.ObatSelamaPerawatan,
		req.DiagnosaUtama,
		req.KdDiagnosaUtama,
		req.DiagnosaSekunder,
		req.KdDiagnosaSekunder,
		req.DiagnosaSekunder2,
		req.KdDiagnosaSekunder2,
		req.DiagnosaSekunder3,
		req.KdDiagnosaSekunder3,
		req.DiagnosaSekunder4,
		req.KdDiagnosaSekunder4,
		req.ProsedurUtama,
		req.KdProsedurUtama,
		req.ProsedurSekunder,
		req.KdProsedurSekunder,
		req.ProsedurSekunder2,
		req.KdProsedurSekunder2,
		req.ProsedurSekunder3,
		req.KdProsedurSekunder3,
		req.Alergi,
		req.Diet,
		req.HasilLaboratoriumPending,
		req.InstruksiAtauEdukasi,
		string(req.CaraKeluar),
		req.KeteranganKeluar,
		string(req.KeadaanPulang),
		req.KeteranganKeadaanPulang,
		string(req.Dilanjutkan),
		req.KeteranganDilanjutkan,
		req.WaktuKontrol,
		req.ObatPulang,
		noRawat,
	)
	if err != nil {
		return nil, err
	}
	if err := database.CheckRowsAffected(res); err != nil {
		return nil, err
	}

	return r.DetailResumePasienRanap(ctx, noRawat)
}

func (r *repository) HapusResumePasienRanap(ctx context.Context, noRawat string) error {
	query := "DELETE FROM resume_pasien_ranap WHERE no_rawat = ?"
	res, err := r.db.ExecContext(ctx, query, noRawat)
	if err != nil {
		return err
	}
	return database.CheckRowsAffected(res)
}
