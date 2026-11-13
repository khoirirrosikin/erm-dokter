package diagnosa

import (
	"errors"
	"fmt"
	"strings"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type IdDiagnosa struct {
	NoRawat string
	Kode    string
	Status  shared.StatusLanjut
}

func (id IdDiagnosa) CompositeKey() string {
	return fmt.Sprintf("%s~%s~%s", id.NoRawat, id.Kode, id.Status)
}

func ParseIdDiagnosa(decryptedKey string) (IdDiagnosa, error) {
	parts := strings.Split(decryptedKey, "~")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return IdDiagnosa{}, errors.New("ID diagnosa tidak valid")
	}
	status, ok := shared.ParseStatusLanjut(parts[2])
	if !ok {
		return IdDiagnosa{}, errors.New("Status ID diagnosa tidak valid")
	}
	return IdDiagnosa{
		NoRawat: parts[0],
		Kode:    parts[1],
		Status:  status,
	}, nil
}

type IdProsedur struct {
	NoRawat string
	Kode    string
	Status  shared.StatusLanjut
}

func (id IdProsedur) CompositeKey() string {
	return fmt.Sprintf("%s~%s~%s", id.NoRawat, id.Kode, id.Status)
}

func ParseIdProsedur(decryptedKey string) (IdProsedur, error) {
	parts := strings.Split(decryptedKey, "~")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return IdProsedur{}, errors.New("ID prosedur tidak valid")
	}
	status, ok := shared.ParseStatusLanjut(parts[2])
	if !ok {
		return IdProsedur{}, errors.New("Status ID prosedur tidak valid")
	}
	return IdProsedur{
		NoRawat: parts[0],
		Kode:    parts[1],
		Status:  status,
	}, nil
}

type DiagnosaPasien struct {
	Id             string              `json:"id"`
	IdKunjungan    string              `json:"id_kunjungan"`
	NoRawat        string              `json:"-"`
	Kode           string              `json:"kode"`
	Nama           string              `json:"nama"`
	Status         shared.StatusLanjut `json:"status"`
	Prioritas      int                 `json:"prioritas"`
	StatusPenyakit string              `json:"status_penyakit"`
}

func (d *DiagnosaPasien) ToIdDiagnosa() IdDiagnosa {
	return IdDiagnosa{
		NoRawat: d.NoRawat,
		Kode:    d.Kode,
		Status:  d.Status,
	}
}

func (d *DiagnosaPasien) CompositeKey() string {
	return d.ToIdDiagnosa().CompositeKey()
}

type ProsedurPasien struct {
	Id          string              `json:"id"`
	IdKunjungan string              `json:"id_kunjungan"`
	NoRawat     string              `json:"-"`
	Kode        string              `json:"kode"`
	Nama        string              `json:"nama"`
	Status      shared.StatusLanjut `json:"status"`
	Prioritas   int                 `json:"prioritas"`
}

func (p *ProsedurPasien) ToIdProsedur() IdProsedur {
	return IdProsedur{
		NoRawat: p.NoRawat,
		Kode:    p.Kode,
		Status:  p.Status,
	}
}

func (p *ProsedurPasien) CompositeKey() string {
	return p.ToIdProsedur().CompositeKey()
}

type DaftarDiagnosaProsedurResponse struct {
	Diagnosa []DiagnosaPasien `json:"diagnosa"`
	Prosedur []ProsedurPasien `json:"prosedur"`
}

type RiwayatPasienResponse struct {
	Diagnosa []DiagnosaPasien `json:"diagnosa"`
	Prosedur []ProsedurPasien `json:"prosedur"`
}

type TambahDiagnosaRequest struct {
	Kode           string `json:"kode"`
	Prioritas      *int   `json:"prioritas,omitempty"`
	StatusPenyakit string `json:"status_penyakit,omitempty"`
}

func (r *TambahDiagnosaRequest) Sanitize() {
	r.Kode = strings.TrimSpace(r.Kode)
	r.StatusPenyakit = strings.TrimSpace(r.StatusPenyakit)
}

func (r *TambahDiagnosaRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if r.Kode == "" {
		errs["kode"] = "Kode ICD-10 wajib diisi"
	}

	if r.Prioritas != nil {
		if *r.Prioritas < 1 {
			errs["prioritas"] = "Prioritas minimal 1"
		} else if *r.Prioritas > 127 {
			errs["prioritas"] = "Prioritas maksimal 127"
		}
	}

	if r.StatusPenyakit != "" && r.StatusPenyakit != "Baru" && r.StatusPenyakit != "Lama" {
		errs["status_penyakit"] = "Status penyakit harus 'Baru' atau 'Lama'"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateDiagnosaRequest struct {
	Prioritas      *int   `json:"prioritas,omitempty"`
	StatusPenyakit string `json:"status_penyakit,omitempty"`
}

func (r *UpdateDiagnosaRequest) Sanitize() {
	r.StatusPenyakit = strings.TrimSpace(r.StatusPenyakit)
}

func (r *UpdateDiagnosaRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if r.Prioritas == nil && r.StatusPenyakit == "" {
		errs["payload"] = "Minimal salah satu dari prioritas atau status_penyakit harus diisi"
	}

	if r.Prioritas != nil {
		if *r.Prioritas < 1 {
			errs["prioritas"] = "Prioritas minimal 1"
		} else if *r.Prioritas > 127 {
			errs["prioritas"] = "Prioritas maksimal 127"
		}
	}

	if r.StatusPenyakit != "" && r.StatusPenyakit != "Baru" && r.StatusPenyakit != "Lama" {
		errs["status_penyakit"] = "Status penyakit harus 'Baru' atau 'Lama'"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type TambahProsedurRequest struct {
	Kode      string `json:"kode"`
	Prioritas *int   `json:"prioritas,omitempty"`
}

func (r *TambahProsedurRequest) Sanitize() {
	r.Kode = strings.TrimSpace(r.Kode)
}

func (r *TambahProsedurRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if r.Kode == "" {
		errs["kode"] = "Kode ICD-9 wajib diisi"
	}

	if r.Prioritas != nil {
		if *r.Prioritas < 1 {
			errs["prioritas"] = "Prioritas minimal 1"
		} else if *r.Prioritas > 127 {
			errs["prioritas"] = "Prioritas maksimal 127"
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateProsedurRequest struct {
	Prioritas *int `json:"prioritas"`
}

func (r *UpdateProsedurRequest) Sanitize() {
}

func (r *UpdateProsedurRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if r.Prioritas == nil {
		errs["prioritas"] = "Prioritas wajib diisi"
	} else if *r.Prioritas < 1 {
		errs["prioritas"] = "Prioritas minimal 1"
	} else if *r.Prioritas > 127 {
		errs["prioritas"] = "Prioritas maksimal 127"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type ReorderItemRequest struct {
	Id        string `json:"id,omitempty"`
	Kode      string `json:"kode,omitempty"`
	Prioritas int    `json:"prioritas"`
}

type ReorderRequest struct {
	Items []ReorderItemRequest `json:"items"`
}

func (r *ReorderRequest) Sanitize() {
	for i := range r.Items {
		r.Items[i].Id = strings.TrimSpace(r.Items[i].Id)
		r.Items[i].Kode = strings.TrimSpace(r.Items[i].Kode)
	}
}

func (r *ReorderRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	if len(r.Items) == 0 {
		errs["items"] = "Daftar urutan items tidak boleh kosong"
		return errs
	}

	seenKey := make(map[string]bool)
	seenPrioritas := make(map[int]bool)
	for i, item := range r.Items {
		key := item.Kode
		if key == "" {
			key = item.Id
		}
		if key == "" {
			errs[fmt.Sprintf("items[%d]", i)] = "Kode atau ID item wajib diisi"
		}
		if item.Prioritas < 1 {
			errs[fmt.Sprintf("items[%d].prioritas", i)] = "Prioritas minimal 1"
		} else if item.Prioritas > 127 {
			errs[fmt.Sprintf("items[%d].prioritas", i)] = "Prioritas maksimal 127"
		} else if seenPrioritas[item.Prioritas] {
			errs[fmt.Sprintf("items[%d].prioritas", i)] = fmt.Sprintf("Prioritas %d duplikat dalam urutan", item.Prioritas)
		} else {
			seenPrioritas[item.Prioritas] = true
		}
		if key != "" && seenKey[key] {
			errs[fmt.Sprintf("items[%d]", i)] = fmt.Sprintf("Item '%s' duplikat dalam urutan", key)
		}
		if key != "" {
			seenKey[key] = true
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type SimulasiEklaimRequest struct {
	Diagnosa []string `json:"diagnosa,omitempty"`
	Prosedur []string `json:"prosedur,omitempty"`
}

func (r *SimulasiEklaimRequest) Sanitize() {
	var cleanDiagnosa []string
	for _, d := range r.Diagnosa {
		if trimmed := strings.TrimSpace(d); trimmed != "" {
			cleanDiagnosa = append(cleanDiagnosa, trimmed)
		}
	}
	r.Diagnosa = cleanDiagnosa

	var cleanProsedur []string
	for _, p := range r.Prosedur {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			cleanProsedur = append(cleanProsedur, trimmed)
		}
	}
	r.Prosedur = cleanProsedur
}

func (r *SimulasiEklaimRequest) Validate() apperror.ValidationError {
	errs := make(apperror.ValidationError)

	for i, d := range r.Diagnosa {
		if strings.TrimSpace(d) == "" {
			errs[fmt.Sprintf("diagnosa[%d]", i)] = fmt.Sprintf("Diagnosa ke-%d: Kode ICD-10 wajib diisi", i+1)
		}
	}

	for i, p := range r.Prosedur {
		if strings.TrimSpace(p) == "" {
			errs[fmt.Sprintf("prosedur[%d]", i)] = fmt.Sprintf("Prosedur ke-%d: Kode ICD-9 wajib diisi", i+1)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type SpecialCMGResponse struct {
	Kode      string `json:"kode"`
	Deskripsi string `json:"deskripsi"`
	Tarif     int64  `json:"tarif"`
	Tipe      string `json:"tipe"`
}

type SimulasiEklaimResponse struct {
	KodeCBG       string               `json:"kode_cbg"`
	DeskripsiCBG  string               `json:"deskripsi_cbg"`
	Tarif         int64                `json:"tarif"`
	BaseTarif     int64                `json:"base_tarif"`
	Kelas         string               `json:"kelas"`
	JenisRawat    string               `json:"jenis_rawat"`
	SeverityLevel string               `json:"severity_level"`
	SpecialCMG    []SpecialCMGResponse `json:"special_cmg,omitempty"`
}
