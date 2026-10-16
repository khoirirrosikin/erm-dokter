package rawatjalan

import (
	"strings"
	"time"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
	"erm-dokter/internal/shared/formatter"
)

type KunjunganRawatJalan struct {
	Id                      string              `json:"id"`
	IdPasien                string              `json:"id_pasien"`
	NoRawat                 string              `json:"no_rawat"`
	NoRegistrasi            string              `json:"no_registrasi"`
	TanggalRegistrasi       string              `json:"tanggal_registrasi"`
	JamRegistrasi           string              `json:"jam_registrasi"`
	NoRekamMedis            string              `json:"no_rekam_medis"`
	NamaPasien              string              `json:"nama_pasien"`
	JenisKelamin            string              `json:"jenis_kelamin"`
	TanggalLahir            string              `json:"tanggal_lahir"`
	Umur                    string              `json:"umur"`
	Alamat                  string              `json:"alamat"`
	KodePoliAsal            string              `json:"kode_poli_asal"`
	NamaPoliAsal            string              `json:"nama_poli_asal"`
	KodeDokterAsal          string              `json:"kode_dokter_asal"`
	NamaDokterAsal          string              `json:"nama_dokter_asal"`
	KodePoliRujukan         string              `json:"kode_poli_rujukan,omitempty"`
	NamaPoliRujukan         string              `json:"nama_poli_rujukan,omitempty"`
	KodeDokterRujukan       string              `json:"kode_dokter_rujukan,omitempty"`
	NamaDokterRujukan       string              `json:"nama_dokter_rujukan,omitempty"`
	KodePenjamin            string              `json:"kode_penjamin"`
	NamaPenjamin            string              `json:"nama_penjamin"`
	StatusPemeriksaan       StatusPemeriksaan   `json:"status_pemeriksaan"`
	StatusLanjut            shared.StatusLanjut `json:"status_lanjut"`
	StatusBayar             StatusBayar         `json:"status_bayar"`
	JenisAntrean            JenisAntrean        `json:"jenis_antrean"`
	GolonganDarah           string              `json:"golongan_darah"`
	Agama                   string              `json:"agama"`
	NoTelepon               string              `json:"no_telepon"`
	NoPeserta               string              `json:"no_peserta"`
	NoKTP                   string              `json:"no_ktp"`
	PenanggungJawab         string              `json:"penanggung_jawab"`
	HubunganPenanggungJawab string              `json:"hubungan_penanggung_jawab"`
	AlamatPenanggungJawab   string              `json:"alamat_penanggung_jawab"`
}

func (k *KunjunganRawatJalan) FormatJenisKelamin() string {
	return formatter.FormatJenisKelamin(k.JenisKelamin)
}

func (k *KunjunganRawatJalan) FormatUmur() string {
	return formatter.FormatUmur(k.TanggalLahir)
}

func (k *KunjunganRawatJalan) FormatNoRekamMedis() string {
	return formatter.FormatNoRekamMedis(k.NoRekamMedis)
}

type InfoRegistrasiPasien struct {
	TanggalRegistrasi string
	JamRegistrasi     string
	KodePenjamin      string
	StatusBayar       string
}

type OpsiReferensi struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type FilterAntreanDokter struct {
	Tanggal           string              `json:"tanggal,omitempty"`
	Penjamin          string              `json:"penjamin,omitempty"`
	StatusPemeriksaan StatusPemeriksaan   `json:"status_pemeriksaan,omitempty"`
	JenisAntrean      JenisAntrean        `json:"jenis_antrean,omitempty"`
	StatusLanjut      shared.StatusLanjut `json:"status_lanjut,omitempty"`
	Keyword           string              `json:"keyword,omitempty"`
	OrderBy           string              `json:"order_by,omitempty"`
	SortOrder         string              `json:"sort_order,omitempty"`
	Page              int                 `json:"page,omitempty"`
	Limit             int                 `json:"limit,omitempty"`
}

func (f *FilterAntreanDokter) Sanitize() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 20
	} else if f.Limit > 100 {
		f.Limit = 100
	}
	if f.OrderBy == "" {
		f.OrderBy = "waktu_registrasi"
	}
	if f.SortOrder == "" {
		f.SortOrder = "ASC"
	} else {
		f.SortOrder = strings.ToUpper(strings.TrimSpace(f.SortOrder))
	}
	if strings.TrimSpace(f.Tanggal) == "" {
		today := time.Now().Format("2006-01-02")
		f.Tanggal = today + "," + today
	}
}

func (f FilterAntreanDokter) Offset() int {
	return (f.Page - 1) * f.Limit
}

func (f FilterAntreanDokter) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)
	if !shared.SortOrder(f.SortOrder).IsValid() {
		errs["sort_order"] = "Jenis pengurutan tidak valid"
	}
	if !OrderBy(f.OrderBy).IsValid() {
		errs["order_by"] = "Jenis pengurutan tidak valid"
	}
	if f.JenisAntrean != "" && !JenisAntrean(f.JenisAntrean).IsValid() {
		errs["jenis_antrean"] = "Jenis antrean tidak valid"
	}
	if f.StatusPemeriksaan != "" && !StatusPemeriksaan(f.StatusPemeriksaan).IsValid() {
		errs["status_pemeriksaan"] = "Status pemeriksaan tidak valid"
	}
	if f.StatusLanjut != "" && !shared.StatusLanjut(f.StatusLanjut).IsValid() {
		errs["status_lanjut"] = "Status lanjut tidak valid"
	}
	if f.Tanggal != "" {
		shared.ValidasiRentangTanggal(f.Tanggal, errs)
	}
	keyword := strings.TrimSpace(f.Keyword)
	if len(keyword) > 0 && len(keyword) < 3 {
		errs["keyword"] = "Kata kunci pencarian minimal 3 karakter"
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type PanggilAntreanRequest struct {
	KodePoli string `json:"kd_poli,omitempty"`
}

type HasilPanggilAntrean struct {
	NoRawat  string `json:"no_rawat"`
	KodePoli string `json:"kd_poli"`
}

