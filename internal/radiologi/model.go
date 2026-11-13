package radiologi

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type IdHasilRadiologi struct {
	NoRawat        string
	KodeTindakan   string
	TanggalPeriksa string
	JamPeriksa     string
}

func (id IdHasilRadiologi) CompositeKey() string {
	return fmt.Sprintf("%s~%s~%s~%s", id.NoRawat, id.KodeTindakan, id.TanggalPeriksa, id.JamPeriksa)
}

func ParseIdHasilRadiologi(decryptedKey string) (IdHasilRadiologi, error) {
	parts := strings.Split(decryptedKey, "~")
	if len(parts) != 4 {
		return IdHasilRadiologi{}, errors.New("format ID hasil radiologi tidak valid")
	}
	return IdHasilRadiologi{
		NoRawat:        parts[0],
		KodeTindakan:   parts[1],
		TanggalPeriksa: parts[2],
		JamPeriksa:     parts[3],
	}, nil
}

type DokterInfo struct {
	KodeDokter string `json:"kode_dokter"`
	NamaDokter string `json:"nama_dokter"`
}

type PetugasInfo struct {
	Nip  string `json:"nip"`
	Nama string `json:"nama"`
}

type HasilRadiologi struct {
	Id              string      `json:"id"`
	IdKunjungan     string      `json:"id_kunjungan"`
	NoRawat         string      `json:"no_rawat"`
	KodeTindakan    string      `json:"kode_tindakan"`
	NamaTindakan    string      `json:"nama_tindakan"`
	Status          string      `json:"status"`
	TanggalPeriksa  string      `json:"tanggal_periksa"`
	JamPeriksa      string      `json:"jam_periksa"`
	DokterPerujuk   DokterInfo  `json:"dokter_perujuk"`
	DokterRadiologi DokterInfo  `json:"dokter_radiologi"`
	Petugas         PetugasInfo `json:"petugas"`
	Hasil           string      `json:"hasil"`
	GambarPACS      []string    `json:"gambar_pacs"`
}

func (h *HasilRadiologi) ToIdHasilRadiologi() IdHasilRadiologi {
	return IdHasilRadiologi{
		NoRawat:        h.NoRawat,
		KodeTindakan:   h.KodeTindakan,
		TanggalPeriksa: h.TanggalPeriksa,
		JamPeriksa:     h.JamPeriksa,
	}
}

func (h *HasilRadiologi) CompositeKey() string {
	return h.ToIdHasilRadiologi().CompositeKey()
}

type FilterRiwayatRadiologi struct {
	Tanggal string `json:"tanggal,omitempty"`
	Page    int    `json:"page,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

func (f *FilterRiwayatRadiologi) Sanitize() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 5
	} else if f.Limit > 100 {
		f.Limit = 100
	}
	f.Tanggal = strings.TrimSpace(f.Tanggal)
}

func (f FilterRiwayatRadiologi) Offset() int {
	return (f.Page - 1) * f.Limit
}

func (f FilterRiwayatRadiologi) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if f.Tanggal != "" {
		shared.ValidasiRentangTanggal(f.Tanggal, errs)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type PermintaanRadiologiHeader struct {
	NoPermintaan      string     `json:"id_permintaan"`
	NoRawat           string     `json:"-"`
	TanggalPermintaan string     `json:"tanggal_permintaan"`
	JamPermintaan     string     `json:"jam_permintaan"`
	TanggalSampel     string     `json:"tanggal_sampel"`
	JamSampel         string     `json:"jam_sampel"`
	TanggalHasil      string     `json:"tanggal_hasil"`
	JamHasil          string     `json:"jam_hasil"`
	DokterPerujuk     DokterInfo `json:"dokter_perujuk"`
	Status            string     `json:"status"`
	InformasiTambahan string     `json:"informasi_tambahan"`
	DiagnosaKlinis    string     `json:"diagnosa_klinis"`
}

type PemeriksaanRadiologiItem struct {
	IdTindakan   string  `json:"id_tindakan"`
	KodeTindakan string  `json:"kode_tindakan"`
	NamaTindakan string  `json:"nama_tindakan"`
	Biaya        float64 `json:"biaya,omitempty"`
	StatusBayar  string  `json:"status_bayar"`
}

type DetailPermintaanRadiologi struct {
	PermintaanRadiologiHeader
	Pemeriksaan []PemeriksaanRadiologiItem `json:"pemeriksaan"`
}

type ItemPemeriksaanRadiologiRequest struct {
	IdTindakan   string `json:"id_tindakan"`
	KodeTindakan string `json:"-"`
}

type SimpanPermintaanRadiologiRequest struct {
	NoRawat           string                            `json:"no_rawat"`
	TanggalPermintaan string                            `json:"tanggal_permintaan"`
	JamPermintaan     string                            `json:"jam_permintaan"`
	InformasiTambahan string                            `json:"informasi_tambahan"`
	DiagnosaKlinis    string                            `json:"diagnosa_klinis"`
	Pemeriksaan       []ItemPemeriksaanRadiologiRequest `json:"pemeriksaan"`
}

func (req *SimpanPermintaanRadiologiRequest) Sanitize() {
	req.NoRawat = strings.TrimSpace(req.NoRawat)
	req.TanggalPermintaan = strings.TrimSpace(req.TanggalPermintaan)
	req.JamPermintaan = strings.TrimSpace(req.JamPermintaan)
	if len(req.JamPermintaan) == 5 {
		req.JamPermintaan += ":00"
	}

	req.InformasiTambahan = strings.TrimSpace(req.InformasiTambahan)
	if req.InformasiTambahan == "" {
		req.InformasiTambahan = "-"
	}

	req.DiagnosaKlinis = strings.TrimSpace(req.DiagnosaKlinis)
	if req.DiagnosaKlinis == "" {
		req.DiagnosaKlinis = "-"
	}

	for i := range req.Pemeriksaan {
		req.Pemeriksaan[i].IdTindakan = strings.TrimSpace(req.Pemeriksaan[i].IdTindakan)
	}
}

func (req SimpanPermintaanRadiologiRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if req.NoRawat == "" {
		errs["no_rawat"] = "Nomor rawat wajib diisi"
	}

	if req.TanggalPermintaan == "" {
		errs["tanggal_permintaan"] = "Tanggal permintaan wajib diisi"
	} else if _, err := time.Parse("2006-01-02", req.TanggalPermintaan); err != nil {
		errs["tanggal_permintaan"] = "Format tanggal permintaan harus YYYY-MM-DD"
	}

	if req.JamPermintaan == "" {
		errs["jam_permintaan"] = "Jam permintaan wajib diisi"
	} else if _, err := time.Parse("15:04:05", req.JamPermintaan); err != nil {
		errs["jam_permintaan"] = "Format jam permintaan harus HH:mm:ss"
	}

	if req.TanggalPermintaan != "" && req.JamPermintaan != "" {
		waktuReq, err := time.ParseInLocation("2006-01-02 15:04:05", req.TanggalPermintaan+" "+req.JamPermintaan, time.Local)
		if err == nil && waktuReq.After(time.Now().Add(5*time.Minute)) {
			errs["tanggal_permintaan"] = "Waktu permintaan tidak boleh di masa depan"
		}
	}

	if len(req.InformasiTambahan) > 60 {
		errs["informasi_tambahan"] = "Informasi tambahan maksimal 60 karakter"
	}

	if len(req.DiagnosaKlinis) > 80 {
		errs["diagnosa_klinis"] = "Diagnosa klinis maksimal 80 karakter"
	}

	if len(req.Pemeriksaan) == 0 {
		errs["pemeriksaan"] = "Pemeriksaan radiologi minimal harus memilih 1 tindakan"
	}

	seenTindakan := make(map[string]bool, len(req.Pemeriksaan))
	for i, item := range req.Pemeriksaan {
		if item.IdTindakan == "" {
			errs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: ID tindakan wajib diisi", i+1)
			continue
		}
		if seenTindakan[item.IdTindakan] {
			errs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: Tindakan radiologi tidak boleh duplikat", i+1)
			continue
		}
		seenTindakan[item.IdTindakan] = true
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type FilterRiwayatPermintaanRadiologi struct {
	Tanggal string
	Page    int
	Limit   int
}

func (f *FilterRiwayatPermintaanRadiologi) Sanitize() {
	f.Tanggal = strings.TrimSpace(f.Tanggal)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 5
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
}

func (f FilterRiwayatPermintaanRadiologi) Offset() int {
	return (f.Page - 1) * f.Limit
}

func (f FilterRiwayatPermintaanRadiologi) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if f.Tanggal != "" {
		parts := strings.Split(f.Tanggal, ",")
		if len(parts) == 1 {
			if _, err := time.Parse("2006-01-02", strings.TrimSpace(parts[0])); err != nil {
				errs["tanggal"] = "Format tanggal harus YYYY-MM-DD atau YYYY-MM-DD,YYYY-MM-DD"
			}
		} else if len(parts) == 2 {
			tglAwal, err1 := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
			tglAkhir, err2 := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
			if err1 != nil || err2 != nil {
				errs["tanggal"] = "Format rentang tanggal harus YYYY-MM-DD,YYYY-MM-DD"
			} else if tglAwal.After(tglAkhir) {
				errs["tanggal"] = "Tanggal awal tidak boleh lebih besar dari tanggal akhir"
			}
		} else {
			errs["tanggal"] = "Filter tanggal maksimal terdiri dari 2 rentang tanggal"
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
