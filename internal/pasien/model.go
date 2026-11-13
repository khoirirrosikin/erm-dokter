package pasien

import (
	"strings"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
	"erm-dokter/internal/shared/formatter"
)

type Pasien struct {
	NoRM         string `json:"no_rkm_medis"`
	Nama         string `json:"nm_pasien"`
	NIK          string `json:"no_ktp"`
	NoBPJS       string `json:"no_peserta"`
	JenisKelamin string `json:"jk"`
	TanggalLahir string `json:"tgl_lahir"`
	TempatLahir  string `json:"tmp_lahir"`
	Alamat       string `json:"alamat"`
	NoTelp       string `json:"no_tlp"`
}

func (p *Pasien) FormatNoRekamMedis() string {
	return formatter.FormatNoRekamMedis(p.NoRM)
}

func (p *Pasien) HitungUmurLengkap() formatter.Umur {
	return formatter.HitungUmur(p.TanggalLahir)
}

func (p *Pasien) FormatUmur() string {
	return formatter.FormatUmur(p.TanggalLahir)
}

func (p *Pasien) FormatJenisKelamin() string {
	return formatter.FormatJenisKelamin(p.JenisKelamin)
}

type RiwayatRujukanInternal struct {
	KodePoli   string `json:"kode_poli"`
	NamaPoli   string `json:"nama_poli"`
	KodeDokter string `json:"kode_dokter"`
	NamaDokter string `json:"nama_dokter"`
}

type RiwayatKamarInap struct {
	KodeKamar     string `json:"kode_kamar"`
	NamaBangsal   string `json:"nama_bangsal"`
	Kelas         string `json:"kelas"`
	TanggalMasuk  string `json:"tanggal_masuk"`
	JamMasuk      string `json:"jam_masuk"`
	TanggalKeluar string `json:"tanggal_keluar"`
	JamKeluar     string `json:"jam_keluar"`
	StatusPulang  string `json:"status_pulang"`
	LamaInap      int    `json:"lama_inap"`
	DiagnosaAwal  string `json:"diagnosa_awal"`
	DiagnosaAkhir string `json:"diagnosa_akhir"`
}

type RiwayatRawatInap struct {
	DPJP  []string           `json:"dpjp"`
	Kamar []RiwayatKamarInap `json:"kamar"`
}

type RiwayatKunjungan struct {
	IdKunjungan       string                   `json:"id_kunjungan"`
	IdPasien          string                   `json:"id_pasien"`
	NoRawat           string                   `json:"no_rawat"`
	NoRegistrasi      string                   `json:"no_registrasi"`
	TanggalRegistrasi string                   `json:"tanggal_registrasi"`
	JamRegistrasi     string                   `json:"jam_registrasi"`
	NoRekamMedis      string                   `json:"no_rekam_medis"`
	NamaPasien        string                   `json:"nama_pasien"`
	StatusLanjut      shared.StatusLanjut      `json:"status_lanjut"`
	StatusBayar       string                   `json:"status_bayar"`
	KodePenjamin      string                   `json:"kode_penjamin"`
	NamaPenjamin      string                   `json:"nama_penjamin"`
	KodeDokter        string                   `json:"kode_dokter"`
	NamaDokter        string                   `json:"nama_dokter"`
	KodePoli          string                   `json:"kode_poli,omitempty"`
	NamaPoli          string                   `json:"nama_poli,omitempty"`
	RujukanInternal   []RiwayatRujukanInternal `json:"rujukan_internal,omitempty"`
	StatusPeriksa     string                   `json:"status_periksa"`
	KamarInap         *RiwayatRawatInap        `json:"kamar_inap,omitempty"`
}

type FilterRiwayatKunjungan struct {
	Tanggal string `json:"tanggal"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

func (f *FilterRiwayatKunjungan) Sanitize() {
	f.Tanggal = strings.TrimSpace(f.Tanggal)

	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 3
	}
	if f.Limit > 50 {
		f.Limit = 50
	}
}

func (f *FilterRiwayatKunjungan) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if f.Tanggal != "" {
		shared.ValidasiRentangTanggal(f.Tanggal, errs)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (f *FilterRiwayatKunjungan) Offset() int {
	return (f.Page - 1) * f.Limit
}
