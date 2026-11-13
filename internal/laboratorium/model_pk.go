package laboratorium

import (
	"fmt"
	"strings"

	"erm-dokter/internal/shared/apperror"
)

type ItemHasilLabPK struct {
	IdTemplate      string `json:"id_template,omitempty"`
	NamaPemeriksaan string `json:"nama_pemeriksaan"`
	Nilai           string `json:"nilai"`
	Satuan          string `json:"satuan"`
	NilaiRujukan    string `json:"nilai_rujukan"`
	Keterangan      string `json:"keterangan"`
}

type PermintaanLabPK struct {
	PermintaanLabHeader
}

type DetailTemplateLabPKItem struct {
	IdTemplate      string `json:"id_template"`
	NamaPemeriksaan string `json:"nama_pemeriksaan"`
	Satuan          string `json:"satuan"`
	NilaiRujukan    string `json:"nilai_rujukan"`
	StatusBayar     string `json:"status_bayar"`
}

type PemeriksaanLabPKItem struct {
	IdTindakan     string                    `json:"id_tindakan"`
	KodeTindakan   string                    `json:"kode_tindakan"`
	NamaTindakan   string                    `json:"nama_tindakan"`
	StatusBayar    string                    `json:"status_bayar"`
	DetailTemplate []DetailTemplateLabPKItem `json:"detail_template,omitempty"`
}

type DetailPermintaanLabPK struct {
	PermintaanLabPK
	Pemeriksaan []PemeriksaanLabPKItem `json:"pemeriksaan"`
}

type ItemPemeriksaanLabPKRequest struct {
	IdTindakan   string   `json:"id_tindakan"`
	KodeTindakan string   `json:"-"`
	IdTemplate   []string `json:"id_template,omitempty"`
	KodeTemplate []int    `json:"-"`
}

type SimpanPermintaanLabPKRequest struct {
	PermintaanLabHeaderRequest
	Pemeriksaan []ItemPemeriksaanLabPKRequest `json:"pemeriksaan"`
}

func (req *SimpanPermintaanLabPKRequest) Sanitize() {
	req.PermintaanLabHeaderRequest.Sanitize()

	for i := range req.Pemeriksaan {
		req.Pemeriksaan[i].IdTindakan = strings.TrimSpace(req.Pemeriksaan[i].IdTindakan)
		for j := range req.Pemeriksaan[i].IdTemplate {
			req.Pemeriksaan[i].IdTemplate[j] = strings.TrimSpace(req.Pemeriksaan[i].IdTemplate[j])
		}
	}
}

func (req SimpanPermintaanLabPKRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	req.PermintaanLabHeaderRequest.Validate(errs)

	if len(req.Pemeriksaan) == 0 {
		errs["pemeriksaan"] = "Pemeriksaan laboratorium minimal harus memilih 1 tindakan"
	}

	seenTindakan := make(map[string]bool, len(req.Pemeriksaan))
	for i, item := range req.Pemeriksaan {
		prefix := fmt.Sprintf("pemeriksaan[%d]", i)
		if item.IdTindakan == "" {
			errs[prefix+".id_tindakan"] = fmt.Sprintf("Pemeriksaan ke-%d: ID tindakan wajib diisi", i+1)
		} else if seenTindakan[item.IdTindakan] {
			errs[prefix+".id_tindakan"] = fmt.Sprintf("Pemeriksaan ke-%d: Tindakan laboratorium tidak boleh duplikat", i+1)
		} else {
			seenTindakan[item.IdTindakan] = true
		}
		seenTemplate := make(map[string]bool, len(item.IdTemplate))
		for j, idTemplate := range item.IdTemplate {
			if idTemplate == "" {
				errs[fmt.Sprintf("%s.id_template[%d]", prefix, j)] = fmt.Sprintf("Pemeriksaan ke-%d parameter ke-%d: ID template pengujian wajib diisi", i+1, j+1)
			} else if seenTemplate[idTemplate] {
				errs[fmt.Sprintf("%s.id_template[%d]", prefix, j)] = fmt.Sprintf("Pemeriksaan ke-%d parameter ke-%d: Template pengujian tidak boleh duplikat", i+1, j+1)
			} else {
				seenTemplate[idTemplate] = true
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
