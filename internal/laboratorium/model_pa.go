package laboratorium

import (
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/shared/apperror"
)

type HasilLabPA struct {
	DiagnosaKlinik string `json:"diagnosa_klinik"`
	Makroskopis    string `json:"makroskopis"`
	Mikroskopis    string `json:"mikroskopis"`
	Kesimpulan     string `json:"kesimpulan"`
	Kesan          string `json:"kesan"`
}

type DataKlinisJaringanPA struct {
	PengambilanBahan     string `json:"pengambilan_bahan,omitempty"`
	DiperolehDengan      string `json:"diperoleh_dengan,omitempty"`
	LokasiJaringan       string `json:"lokasi_jaringan,omitempty"`
	DiawetkanDengan      string `json:"diawetkan_dengan,omitempty"`
	PernahDilakukanDi    string `json:"pernah_dilakukan_di,omitempty"`
	TanggalPASebelumnya  string `json:"tanggal_pa_sebelumnya,omitempty"`
	NomorPASebelumnya    string `json:"nomor_pa_sebelumnya,omitempty"`
	DiagnosaPASebelumnya string `json:"diagnosa_pa_sebelumnya,omitempty"`
}

func (d *DataKlinisJaringanPA) Sanitize(tanggalPermintaan string) {
	d.PengambilanBahan = strings.TrimSpace(d.PengambilanBahan)
	if d.PengambilanBahan == "" {
		d.PengambilanBahan = strings.TrimSpace(tanggalPermintaan)
	}
	d.DiperolehDengan = strings.TrimSpace(d.DiperolehDengan)
	d.LokasiJaringan = strings.TrimSpace(d.LokasiJaringan)
	d.DiawetkanDengan = strings.TrimSpace(d.DiawetkanDengan)
	d.PernahDilakukanDi = strings.TrimSpace(d.PernahDilakukanDi)
	d.TanggalPASebelumnya = strings.TrimSpace(d.TanggalPASebelumnya)
	if d.TanggalPASebelumnya == "" {
		d.TanggalPASebelumnya = "0000-00-00"
	}
	d.NomorPASebelumnya = strings.TrimSpace(d.NomorPASebelumnya)
	d.DiagnosaPASebelumnya = strings.TrimSpace(d.DiagnosaPASebelumnya)
}

func (d DataKlinisJaringanPA) Validate(errs apperror.ValidationError) {
	if len(d.DiperolehDengan) > 40 {
		errs["diperoleh_dengan"] = "Keterangan diperoleh dengan maksimal 40 karakter"
	}
	if len(d.LokasiJaringan) > 40 {
		errs["lokasi_jaringan"] = "Lokasi pengambilan jaringan maksimal 40 karakter"
	}
	if len(d.DiawetkanDengan) > 40 {
		errs["diawetkan_dengan"] = "Keterangan diawetkan dengan maksimal 40 karakter"
	}
	if len(d.PernahDilakukanDi) > 100 {
		errs["pernah_dilakukan_di"] = "Riwayat faskes PA sebelumnya maksimal 100 karakter"
	}
	if len(d.NomorPASebelumnya) > 20 {
		errs["nomor_pa_sebelumnya"] = "Nomor PA sebelumnya maksimal 20 karakter"
	}
	if len(d.DiagnosaPASebelumnya) > 100 {
		errs["diagnosa_pa_sebelumnya"] = "Diagnosa PA sebelumnya maksimal 100 karakter"
	}

	if d.PengambilanBahan != "" {
		tglBahan, errBahan := time.Parse("2006-01-02", d.PengambilanBahan)
		if errBahan != nil {
			errs["pengambilan_bahan"] = "Format tanggal pengambilan bahan harus YYYY-MM-DD"
		} else {
			now := time.Now()
			todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
			if tglBahan.After(todayEnd) {
				errs["pengambilan_bahan"] = "Tanggal pengambilan bahan tidak boleh melebihi hari ini"
			}
		}
	}

	if d.TanggalPASebelumnya != "" && d.TanggalPASebelumnya != "0000-00-00" {
		tglPASblm, errPASblm := time.Parse("2006-01-02", d.TanggalPASebelumnya)
		if errPASblm != nil {
			errs["tanggal_pa_sebelumnya"] = "Format tanggal PA sebelumnya harus YYYY-MM-DD"
		} else {
			now := time.Now()
			todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
			if tglPASblm.After(todayEnd) {
				errs["tanggal_pa_sebelumnya"] = "Tanggal PA sebelumnya tidak boleh melebihi hari ini"
			}
		}
	}
}

type PermintaanLabPA struct {
	PermintaanLabHeader
	DataKlinisJaringanPA
}

type PemeriksaanLabPAItem struct {
	IdTindakan   string  `json:"id_tindakan"`
	KodeTindakan string  `json:"kode_tindakan"`
	NamaTindakan string  `json:"nama_tindakan"`
	Biaya        float64 `json:"biaya,omitempty"`
	StatusBayar  string  `json:"status_bayar"`
}

type DetailPermintaanLabPA struct {
	PermintaanLabPA
	Pemeriksaan []PemeriksaanLabPAItem `json:"pemeriksaan"`
}

type ItemPemeriksaanLabPARequest struct {
	IdTindakan   string `json:"id_tindakan"`
	KodeTindakan string `json:"-"`
}

type SimpanPermintaanLabPARequest struct {
	PermintaanLabHeaderRequest
	DataKlinisJaringanPA
	Pemeriksaan []ItemPemeriksaanLabPARequest `json:"pemeriksaan"`
}

func (req *SimpanPermintaanLabPARequest) Sanitize() {
	req.PermintaanLabHeaderRequest.Sanitize()
	req.DataKlinisJaringanPA.Sanitize(req.TanggalPermintaan)

	for i := range req.Pemeriksaan {
		req.Pemeriksaan[i].IdTindakan = strings.TrimSpace(req.Pemeriksaan[i].IdTindakan)
	}
}

func (req SimpanPermintaanLabPARequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	req.PermintaanLabHeaderRequest.Validate(errs)
	req.DataKlinisJaringanPA.Validate(errs)

	if len(req.Pemeriksaan) == 0 {
		errs["pemeriksaan"] = "Pemeriksaan laboratorium minimal harus memilih 1 tindakan"
	}

	seenTindakan := make(map[string]bool, len(req.Pemeriksaan))
	for i, item := range req.Pemeriksaan {
		if item.IdTindakan == "" {
			errs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: ID tindakan wajib diisi", i+1)
		} else if seenTindakan[item.IdTindakan] {
			errs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: Tindakan laboratorium tidak boleh duplikat", i+1)
		} else {
			seenTindakan[item.IdTindakan] = true
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
