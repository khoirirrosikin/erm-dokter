package resep

import (
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Resep struct {
	Id                 string               `json:"id"`
	IdKunjungan        string               `json:"id_kunjungan"`
	NoResep            string               `json:"no_resep"`
	NoRawat            string               `json:"no_rawat"`
	TanggalPeresepan   string               `json:"tanggal_peresepan"`
	JamPeresepan       string               `json:"jam_peresepan"`
	TanggalPerawatan   string               `json:"tanggal_perawatan,omitempty"`
	JamPerawatan       string               `json:"jam_perawatan,omitempty"`
	TanggalPenyerahan  string               `json:"tanggal_penyerahan,omitempty"`
	JamPenyerahan      string               `json:"jam_penyerahan,omitempty"`
	Status             string               `json:"status"`
	KodeDokter         string               `json:"kode_dokter"`
	NamaDokter         string               `json:"nama_dokter"`
	ResepDokter        []ResepDokter        `json:"resep_dokter"`
	ResepDokterRacikan []ResepDokterRacikan `json:"resep_dokter_racikan"`
}

type ItemObatResep struct {
	IdObat   string  `json:"id_obat"`
	KodeObat string  `json:"kode_obat"`
	NamaObat string  `json:"nama_obat"`
	Jumlah   float64 `json:"jumlah"`
	Satuan   string  `json:"satuan"`
}

type ResepDokter struct {
	NoResep string `json:"-"`
	ItemObatResep
	AturanPakai string `json:"aturan_pakai"`
}

type ResepDokterRacikan struct {
	NoResep         string                     `json:"-"`
	NoRacik         string                     `json:"no_racik"`
	NamaRacik       string                     `json:"nama_racik"`
	KodeMetodeRacik string                     `json:"kode_racik"`
	NamaMetodeRacik string                     `json:"metode_racik"`
	JumlahRacikan   int                        `json:"jumlah_racikan"`
	AturanPakai     string                     `json:"aturan_pakai"`
	Keterangan      string                     `json:"keterangan"`
	DetailRacikan   []ResepDokterRacikanDetail `json:"detail_racikan"`
}

type ResepDokterRacikanDetail struct {
	NoResep string `json:"-"`
	NoRacik string `json:"-"`
	ItemObatResep
	Kandungan string `json:"kandungan"`
}

type FilterDaftarResep struct {
	Tanggal string `json:"tanggal,omitempty"`
	Page    int    `json:"page,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

func (f *FilterDaftarResep) Sanitize() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 5
	} else if f.Limit > 100 {
		f.Limit = 100
	}
}

func (f FilterDaftarResep) Offset() int {
	return (f.Page - 1) * f.Limit
}

func (f FilterDaftarResep) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)
	if f.Tanggal != "" {
		shared.ValidasiRentangTanggal(f.Tanggal, errs)
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type AturanPakai struct {
	AturanPakai string `json:"aturan_pakai"`
}

type MetodeRacik struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type ItemObatInput struct {
	IdObat   string  `json:"id_obat"`
	KodeObat string  `json:"-"`
	Jumlah   float64 `json:"jumlah"`
}

func (i *ItemObatInput) Sanitize() {
	i.IdObat = strings.TrimSpace(i.IdObat)
}

func (i ItemObatInput) Validate(prefix, label string, errs apperror.ValidationError) {
	if i.IdObat == "" {
		errs[prefix+".id_obat"] = fmt.Sprintf("%s: ID obat wajib diisi", label)
	}
	if i.Jumlah <= 0 {
		errs[prefix+".jumlah"] = fmt.Sprintf("%s: Jumlah obat harus lebih dari 0", label)
	}
}

type SimpanResepRequest struct {
	NoRawat          string              `json:"no_rawat"`
	TanggalPeresepan string              `json:"tanggal_peresepan"`
	JamPeresepan     string              `json:"jam_peresepan"`
	ResepDokter      []ResepDokterInput  `json:"resep_dokter"`
	ResepRacikan     []ResepRacikanInput `json:"resep_racikan"`
}

type ResepDokterInput struct {
	ItemObatInput
	AturanPakai string `json:"aturan_pakai"`
}

type ResepRacikanInput struct {
	NamaRacik     string                    `json:"nama_racik"`
	KodeRacik     string                    `json:"kode_racik"`
	JumlahRacikan int                       `json:"jumlah_racikan"`
	AturanPakai   string                    `json:"aturan_pakai"`
	Keterangan    string                    `json:"keterangan"`
	Detail        []ResepRacikanDetailInput `json:"detail"`
}

type ResepRacikanDetailInput struct {
	ItemObatInput
	Kandungan string `json:"kandungan"`
}

func (r *SimpanResepRequest) Sanitize() {
	r.NoRawat = strings.TrimSpace(r.NoRawat)
	r.TanggalPeresepan = strings.TrimSpace(r.TanggalPeresepan)
	r.JamPeresepan = strings.TrimSpace(r.JamPeresepan)
	if r.TanggalPeresepan == "" {
		r.TanggalPeresepan = time.Now().Format("2006-01-02")
	}
	if r.JamPeresepan == "" {
		r.JamPeresepan = time.Now().Format("15:04:05")
	}

	for i := range r.ResepDokter {
		r.ResepDokter[i].ItemObatInput.Sanitize()
		r.ResepDokter[i].AturanPakai = strings.TrimSpace(r.ResepDokter[i].AturanPakai)
	}

	for i := range r.ResepRacikan {
		r.ResepRacikan[i].NamaRacik = strings.TrimSpace(r.ResepRacikan[i].NamaRacik)
		r.ResepRacikan[i].KodeRacik = strings.TrimSpace(r.ResepRacikan[i].KodeRacik)
		r.ResepRacikan[i].AturanPakai = strings.TrimSpace(r.ResepRacikan[i].AturanPakai)
		r.ResepRacikan[i].Keterangan = strings.TrimSpace(r.ResepRacikan[i].Keterangan)

		for j := range r.ResepRacikan[i].Detail {
			r.ResepRacikan[i].Detail[j].ItemObatInput.Sanitize()
			r.ResepRacikan[i].Detail[j].Kandungan = strings.TrimSpace(r.ResepRacikan[i].Detail[j].Kandungan)
		}
	}
}

func (r *SimpanResepRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if r.NoRawat == "" {
		errs["no_rawat"] = "Nomor rawat wajib diisi"
	}

	tgl, errTgl := time.Parse("2006-01-02", r.TanggalPeresepan)
	if errTgl != nil {
		errs["tanggal_peresepan"] = "Format tanggal peresepan harus YYYY-MM-DD"
	}

	jam, errJam := time.Parse("15:04:05", r.JamPeresepan)
	if errJam != nil {
		errs["jam_peresepan"] = "Format jam peresepan harus HH:mm:ss"
	}

	if errTgl == nil && errJam == nil {
		waktuPeresepan := time.Date(
			tgl.Year(), tgl.Month(), tgl.Day(),
			jam.Hour(), jam.Minute(), jam.Second(), 0,
			time.Local,
		)
		if waktuPeresepan.After(time.Now()) {
			errs["tanggal_peresepan"] = "Waktu peresepan tidak boleh melebihi waktu saat ini"
		}
	}

	if len(r.ResepDokter) == 0 && len(r.ResepRacikan) == 0 {
		errs["resep"] = "Resep obat harus memiliki minimal 1 obat non-racikan atau obat racikan"
	}

	seenObatDokter := make(map[string]bool, len(r.ResepDokter))
	for i, rd := range r.ResepDokter {
		prefix := fmt.Sprintf("resep_dokter[%d]", i)
		label := fmt.Sprintf("Obat ke-%d", i+1)
		rd.ItemObatInput.Validate(prefix, label, errs)
		if rd.IdObat != "" {
			if seenObatDokter[rd.IdObat] {
				errs[prefix+".id_obat"] = fmt.Sprintf("%s: Obat tidak boleh duplikat dalam daftar non-racikan", label)
			} else {
				seenObatDokter[rd.IdObat] = true
			}
		}
		if rd.AturanPakai == "" {
			errs[prefix+".aturan_pakai"] = fmt.Sprintf("%s: Aturan pakai wajib diisi", label)
		}
	}

	for i, rr := range r.ResepRacikan {
		prefix := fmt.Sprintf("resep_racikan[%d]", i)
		if rr.NamaRacik == "" {
			errs[prefix+".nama_racik"] = fmt.Sprintf("Racikan ke-%d: Nama racikan wajib diisi", i+1)
		}
		if rr.KodeRacik == "" {
			errs[prefix+".kode_racik"] = fmt.Sprintf("Racikan ke-%d: Metode/kode racik wajib diisi", i+1)
		}
		if rr.JumlahRacikan <= 0 {
			errs[prefix+".jumlah_racikan"] = fmt.Sprintf("Racikan ke-%d: Jumlah racikan harus lebih dari 0", i+1)
		}
		if rr.AturanPakai == "" {
			errs[prefix+".aturan_pakai"] = fmt.Sprintf("Racikan ke-%d: Aturan pakai wajib diisi", i+1)
		}
		if len(rr.Detail) == 0 {
			errs[prefix+".detail"] = fmt.Sprintf("Racikan ke-%d: Komposisi obat racikan minimal harus memiliki 1 bahan", i+1)
			continue
		}

		seenBahanRacikan := make(map[string]bool, len(rr.Detail))
		for j, d := range rr.Detail {
			detailPrefix := fmt.Sprintf("%s.detail[%d]", prefix, j)
			detailLabel := fmt.Sprintf("Racikan ke-%d bahan ke-%d", i+1, j+1)
			d.ItemObatInput.Validate(detailPrefix, detailLabel, errs)
			if d.IdObat != "" {
				if seenBahanRacikan[d.IdObat] {
					errs[detailPrefix+".id_obat"] = fmt.Sprintf("%s: Bahan obat tidak boleh duplikat dalam satu racikan", detailLabel)
				} else {
					seenBahanRacikan[d.IdObat] = true
				}
			}
		}

	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
