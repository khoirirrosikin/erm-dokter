package resumepasien

import (
	"strings"
	"time"

	"erm-dokter/internal/shared/apperror"
)

type InfoKamarInap struct {
	TanggalMasuk  string `json:"tanggal_masuk,omitempty"`
	JamMasuk      string `json:"jam_masuk,omitempty"`
	TanggalKeluar string `json:"tanggal_keluar,omitempty"`
	JamKeluar     string `json:"jam_keluar,omitempty"`
	Kamar         string `json:"kamar,omitempty"`
}

type DataResumePasienRanap struct {
	DiagnosaAwal                 string        `json:"diagnosa_awal"`
	AlasanRawat                  string        `json:"alasan_rawat"`
	KeluhanUtama                 string        `json:"keluhan_utama"`
	PemeriksaanFisik             string        `json:"pemeriksaan_fisik"`
	JalannyaPenyakit             string        `json:"jalannya_penyakit"`
	HasilPemeriksaanRadiologi    string        `json:"hasil_pemeriksaan_radiologi"`
	HasilPemeriksaanLaboratorium string        `json:"hasil_pemeriksaan_laboratorium"`
	TindakanAtauOperasi          string        `json:"tindakan_atau_operasi"`
	ObatSelamaPerawatan          string        `json:"obat_selama_perawatan"`
	DiagnosaUtama                string        `json:"diagnosa_utama"`
	KdDiagnosaUtama              string        `json:"kd_diagnosa_utama,omitempty"`
	DiagnosaSekunder             string        `json:"diagnosa_sekunder"`
	KdDiagnosaSekunder           string        `json:"kd_diagnosa_sekunder,omitempty"`
	DiagnosaSekunder2            string        `json:"diagnosa_sekunder2"`
	KdDiagnosaSekunder2          string        `json:"kd_diagnosa_sekunder2,omitempty"`
	DiagnosaSekunder3            string        `json:"diagnosa_sekunder3"`
	KdDiagnosaSekunder3          string        `json:"kd_diagnosa_sekunder3,omitempty"`
	DiagnosaSekunder4            string        `json:"diagnosa_sekunder4"`
	KdDiagnosaSekunder4          string        `json:"kd_diagnosa_sekunder4,omitempty"`
	ProsedurUtama                string        `json:"prosedur_utama"`
	KdProsedurUtama              string        `json:"kd_prosedur_utama,omitempty"`
	ProsedurSekunder             string        `json:"prosedur_sekunder"`
	KdProsedurSekunder           string        `json:"kd_prosedur_sekunder,omitempty"`
	ProsedurSekunder2            string        `json:"prosedur_sekunder2"`
	KdProsedurSekunder2          string        `json:"kd_prosedur_sekunder2,omitempty"`
	ProsedurSekunder3            string        `json:"prosedur_sekunder3"`
	KdProsedurSekunder3          string        `json:"kd_prosedur_sekunder3,omitempty"`
	Alergi                       string        `json:"alergi"`
	Diet                         string        `json:"diet"`
	HasilLaboratoriumPending     string        `json:"hasil_laboratorium_pending"`
	InstruksiAtauEdukasi         string        `json:"instruksi_atau_edukasi"`
	CaraKeluar                   CaraKeluar    `json:"cara_keluar"`
	KeteranganKeluar             string        `json:"keterangan_keluar"`
	KeadaanPulang                KeadaanPulang `json:"keadaan_pulang"`
	KeteranganKeadaanPulang      string        `json:"keterangan_keadaan_pulang"`
	Dilanjutkan                  Dilanjutkan   `json:"dilanjutkan"`
	KeteranganDilanjutkan        string        `json:"keterangan_dilanjutkan"`
	WaktuKontrol                 string        `json:"waktu_kontrol"`
	ObatPulang                   string        `json:"obat_pulang"`
}

func (d *DataResumePasienRanap) Sanitize() {
	d.DiagnosaAwal = strings.TrimSpace(d.DiagnosaAwal)
	d.AlasanRawat = strings.TrimSpace(d.AlasanRawat)
	d.KeluhanUtama = strings.TrimSpace(d.KeluhanUtama)
	d.PemeriksaanFisik = strings.TrimSpace(d.PemeriksaanFisik)
	d.JalannyaPenyakit = strings.TrimSpace(d.JalannyaPenyakit)
	d.HasilPemeriksaanRadiologi = strings.TrimSpace(d.HasilPemeriksaanRadiologi)
	d.HasilPemeriksaanLaboratorium = strings.TrimSpace(d.HasilPemeriksaanLaboratorium)
	d.TindakanAtauOperasi = strings.TrimSpace(d.TindakanAtauOperasi)
	d.ObatSelamaPerawatan = strings.TrimSpace(d.ObatSelamaPerawatan)
	d.DiagnosaUtama = strings.TrimSpace(d.DiagnosaUtama)
	d.KdDiagnosaUtama = strings.TrimSpace(d.KdDiagnosaUtama)
	d.DiagnosaSekunder = strings.TrimSpace(d.DiagnosaSekunder)
	d.KdDiagnosaSekunder = strings.TrimSpace(d.KdDiagnosaSekunder)
	d.DiagnosaSekunder2 = strings.TrimSpace(d.DiagnosaSekunder2)
	d.KdDiagnosaSekunder2 = strings.TrimSpace(d.KdDiagnosaSekunder2)
	d.DiagnosaSekunder3 = strings.TrimSpace(d.DiagnosaSekunder3)
	d.KdDiagnosaSekunder3 = strings.TrimSpace(d.KdDiagnosaSekunder3)
	d.DiagnosaSekunder4 = strings.TrimSpace(d.DiagnosaSekunder4)
	d.KdDiagnosaSekunder4 = strings.TrimSpace(d.KdDiagnosaSekunder4)
	d.ProsedurUtama = strings.TrimSpace(d.ProsedurUtama)
	d.KdProsedurUtama = strings.TrimSpace(d.KdProsedurUtama)
	d.ProsedurSekunder = strings.TrimSpace(d.ProsedurSekunder)
	d.KdProsedurSekunder = strings.TrimSpace(d.KdProsedurSekunder)
	d.ProsedurSekunder2 = strings.TrimSpace(d.ProsedurSekunder2)
	d.KdProsedurSekunder2 = strings.TrimSpace(d.KdProsedurSekunder2)
	d.ProsedurSekunder3 = strings.TrimSpace(d.ProsedurSekunder3)
	d.KdProsedurSekunder3 = strings.TrimSpace(d.KdProsedurSekunder3)
	d.Alergi = strings.TrimSpace(d.Alergi)
	d.Diet = strings.TrimSpace(d.Diet)
	d.HasilLaboratoriumPending = strings.TrimSpace(d.HasilLaboratoriumPending)
	d.InstruksiAtauEdukasi = strings.TrimSpace(d.InstruksiAtauEdukasi)
	d.CaraKeluar = CaraKeluar(strings.TrimSpace(string(d.CaraKeluar)))
	d.KeteranganKeluar = strings.TrimSpace(d.KeteranganKeluar)
	d.KeadaanPulang = KeadaanPulang(strings.TrimSpace(string(d.KeadaanPulang)))
	d.KeteranganKeadaanPulang = strings.TrimSpace(d.KeteranganKeadaanPulang)
	d.Dilanjutkan = Dilanjutkan(strings.TrimSpace(string(d.Dilanjutkan)))
	d.KeteranganDilanjutkan = strings.TrimSpace(d.KeteranganDilanjutkan)

	d.WaktuKontrol = strings.TrimSpace(d.WaktuKontrol)
	if strings.HasPrefix(d.WaktuKontrol, "0000-00-00") {
		d.WaktuKontrol = ""
	}
	if len(d.WaktuKontrol) == 16 {
		d.WaktuKontrol += ":00"
	}

	if d.KeadaanPulang == KeadaanPulangMeninggal {
		d.WaktuKontrol = "0000-00-00 00:00:00"
	} else if d.KeadaanPulang == KeadaanPulangRujuk && d.WaktuKontrol == "" {
		d.WaktuKontrol = "0000-00-00 00:00:00"
	}
	d.ObatPulang = strings.TrimSpace(d.ObatPulang)
}

func (d *DataResumePasienRanap) Validate(errs apperror.ValidationError) {
	if d.DiagnosaAwal == "" {
		errs["diagnosa_awal"] = "Diagnosa awal masuk wajib diisi"
	} else if len(d.DiagnosaAwal) > 500 {
		errs["diagnosa_awal"] = "Diagnosa awal masuk maksimal 500 karakter"
	}

	if d.AlasanRawat == "" {
		errs["alasan_rawat"] = "Alasan masuk dirawat wajib diisi"
	} else if len(d.AlasanRawat) > 100 {
		errs["alasan_rawat"] = "Alasan masuk dirawat maksimal 100 karakter"
	}

	if d.KeluhanUtama == "" {
		errs["keluhan_utama"] = "Keluhan utama wajib diisi"
	}

	if d.DiagnosaUtama == "" {
		errs["diagnosa_utama"] = "Diagnosa utama wajib diisi"
	} else if len(d.DiagnosaUtama) > 150 {
		errs["diagnosa_utama"] = "Diagnosa utama maksimal 150 karakter"
	}

	if len(d.KdDiagnosaUtama) > 10 {
		errs["kd_diagnosa_utama"] = "Kode diagnosa utama maksimal 10 karakter"
	}
	if len(d.DiagnosaSekunder) > 150 {
		errs["diagnosa_sekunder"] = "Diagnosa sekunder 1 maksimal 150 karakter"
	}
	if len(d.KdDiagnosaSekunder) > 10 {
		errs["kd_diagnosa_sekunder"] = "Kode diagnosa sekunder 1 maksimal 10 karakter"
	}
	if len(d.DiagnosaSekunder2) > 150 {
		errs["diagnosa_sekunder2"] = "Diagnosa sekunder 2 maksimal 150 karakter"
	}
	if len(d.KdDiagnosaSekunder2) > 10 {
		errs["kd_diagnosa_sekunder2"] = "Kode diagnosa sekunder 2 maksimal 10 karakter"
	}
	if len(d.DiagnosaSekunder3) > 150 {
		errs["diagnosa_sekunder3"] = "Diagnosa sekunder 3 maksimal 150 karakter"
	}
	if len(d.KdDiagnosaSekunder3) > 10 {
		errs["kd_diagnosa_sekunder3"] = "Kode diagnosa sekunder 3 maksimal 10 karakter"
	}
	if len(d.DiagnosaSekunder4) > 150 {
		errs["diagnosa_sekunder4"] = "Diagnosa sekunder 4 maksimal 150 karakter"
	}
	if len(d.KdDiagnosaSekunder4) > 10 {
		errs["kd_diagnosa_sekunder4"] = "Kode diagnosa sekunder 4 maksimal 10 karakter"
	}

	if len(d.ProsedurUtama) > 150 {
		errs["prosedur_utama"] = "Prosedur utama maksimal 150 karakter"
	}
	if len(d.KdProsedurUtama) > 8 {
		errs["kd_prosedur_utama"] = "Kode prosedur utama maksimal 8 karakter"
	}
	if len(d.ProsedurSekunder) > 150 {
		errs["prosedur_sekunder"] = "Prosedur sekunder 1 maksimal 150 karakter"
	}
	if len(d.KdProsedurSekunder) > 8 {
		errs["kd_prosedur_sekunder"] = "Kode prosedur sekunder 1 maksimal 8 karakter"
	}
	if len(d.ProsedurSekunder2) > 150 {
		errs["prosedur_sekunder2"] = "Prosedur sekunder 2 maksimal 150 karakter"
	}
	if len(d.KdProsedurSekunder2) > 8 {
		errs["kd_prosedur_sekunder2"] = "Kode prosedur sekunder 2 maksimal 8 karakter"
	}
	if len(d.ProsedurSekunder3) > 150 {
		errs["prosedur_sekunder3"] = "Prosedur sekunder 3 maksimal 150 karakter"
	}
	if len(d.KdProsedurSekunder3) > 8 {
		errs["kd_prosedur_sekunder3"] = "Kode prosedur sekunder 3 maksimal 8 karakter"
	}

	if len(d.Alergi) > 100 {
		errs["alergi"] = "Alergi maksimal 100 karakter"
	}

	if d.CaraKeluar == "" {
		errs["cara_keluar"] = "Cara keluar wajib diisi"
	} else if !d.CaraKeluar.IsValid() {
		errs["cara_keluar"] = "Cara keluar tidak valid"
	}
	if len(d.KeteranganKeluar) > 50 {
		errs["keterangan_keluar"] = "Keterangan cara keluar maksimal 50 karakter"
	}

	if d.KeadaanPulang == "" {
		errs["keadaan_pulang"] = "Keadaan pulang wajib diisi"
	} else if !d.KeadaanPulang.IsValidRanap() {
		errs["keadaan_pulang"] = "Keadaan pulang tidak valid"
	}
	if len(d.KeteranganKeadaanPulang) > 50 {
		errs["keterangan_keadaan_pulang"] = "Keterangan keadaan pulang maksimal 50 karakter"
	}

	if d.Dilanjutkan == "" {
		errs["dilanjutkan"] = "Status dilanjutkan wajib diisi"
	} else if !d.Dilanjutkan.IsValid() {
		errs["dilanjutkan"] = "Status dilanjutkan tidak valid"
	}
	if len(d.KeteranganDilanjutkan) > 50 {
		errs["keterangan_dilanjutkan"] = "Keterangan status dilanjutkan maksimal 50 karakter"
	}

	if d.KeadaanPulang != KeadaanPulangMeninggal && d.KeadaanPulang != KeadaanPulangRujuk {
		if d.WaktuKontrol == "" || d.WaktuKontrol == "0000-00-00 00:00:00" {
			errs["waktu_kontrol"] = "Waktu kontrol wajib diisi"
		}
	}

	if d.WaktuKontrol != "" && d.WaktuKontrol != "0000-00-00 00:00:00" {
		waktu, err := time.ParseInLocation("2006-01-02 15:04:05", d.WaktuKontrol, time.Local)
		if err != nil {
			errs["waktu_kontrol"] = "Format tanggal & jam kontrol harus YYYY-MM-DD HH:mm:ss"
		} else if waktu.Hour() >= 14 {
			errs["waktu_kontrol"] = "Jam kontrol harus kurang dari jam 14:00"
		}
	}
}

type ResumePasienRanap struct {
	IdKunjungan string `json:"id_kunjungan"`
	NoRawat     string `json:"-"`
	KodeDokter  string `json:"kode_dokter"`
	NamaDokter  string `json:"nama_dokter"`
	InfoKamarInap
	DataResumePasienRanap
}

type SimpanResumePasienRanapRequest struct {
	NoRawat string `json:"no_rawat"`
	DataResumePasienRanap
}

func (req *SimpanResumePasienRanapRequest) Sanitize() {
	req.NoRawat = strings.TrimSpace(req.NoRawat)
	req.DataResumePasienRanap.Sanitize()
}

func (req *SimpanResumePasienRanapRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)
	if req.NoRawat == "" {
		errs["no_rawat"] = "Nomor rawat wajib diisi"
	}
	req.DataResumePasienRanap.Validate(errs)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateResumePasienRanapRequest struct {
	DataResumePasienRanap
}

func (req *UpdateResumePasienRanapRequest) Sanitize() {
	req.DataResumePasienRanap.Sanitize()
}

func (req *UpdateResumePasienRanapRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)
	req.DataResumePasienRanap.Validate(errs)
	if len(errs) > 0 {
		return errs
	}
	return nil
}
