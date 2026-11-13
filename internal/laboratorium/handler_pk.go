package laboratorium

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (h *Handler) SimpanPermintaanLabPK(w http.ResponseWriter, r *http.Request) {
	statusLanjut, ok := shared.ParseStatusLanjut(r.PathValue("status_lanjut"))
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	idKunjungan := strings.TrimSpace(r.PathValue("id_kunjungan"))
	noRawatURL, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req SimpanPermintaanLabPKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data permintaan laboratorium tidak valid"))
		return
	}

	if req.NoRawat != noRawatURL {
		apperror.HandleError(w, apperror.NewBusinessError("Nomor rawat permintaan laboratorium tidak sesuai dengan ID kunjungan"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.decryptTindakanLabPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	data, err := h.service.SimpanPermintaanLabPK(r.Context(), kodeDokter, statusLanjut, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoPermintaan, data.NoPermintaan)
		h.encryptDetailPermintaanLabPK(data, idKunjungan)
	}

	response.Created(w, "Berhasil mengirim permintaan laboratorium PK", data)
}

func (h *Handler) DaftarPermintaanLabPK(w http.ResponseWriter, r *http.Request) {
	statusLanjut, ok := shared.ParseStatusLanjutWithSemua(r.PathValue("status_lanjut"))
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	idKunjungan := strings.TrimSpace(r.PathValue("id_kunjungan"))
	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	data, err := h.service.GetDaftarPermintaanLabPK(r.Context(), noRawat, statusLanjut)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptPermintaanLabPKList(data, idKunjungan)
	response.Success(w, "Berhasil mengambil daftar permintaan laboratorium PK", data)
}

func (h *Handler) DaftarPermintaanLabPKByRM(w http.ResponseWriter, r *http.Request) {
	statusLanjut, ok := shared.ParseStatusLanjutWithSemua(r.PathValue("status_lanjut"))
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	idPasien := strings.TrimSpace(r.PathValue("id_pasien"))
	noRM, err := crypto.Decrypt(idPasien, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID pasien tidak valid"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := FilterRiwayatLab{
		Tanggal: q.Get("tanggal"),
		Page:    page,
		Limit:   limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	data, meta, err := h.service.GetRiwayatPermintaanLabPKByRM(r.Context(), noRM, statusLanjut, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptPermintaanLabPKList(data, "")
	response.SuccessWithMeta(w, "Berhasil mengambil riwayat permintaan laboratorium PK pasien", data, meta)
}

func (h *Handler) DetailPermintaanLabPK(w http.ResponseWriter, r *http.Request) {
	idPermintaan := strings.TrimSpace(r.PathValue("id_permintaan"))
	if idPermintaan == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan laboratorium wajib diisi"))
		return
	}

	noPermintaan, err := crypto.Decrypt(idPermintaan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan laboratorium tidak valid"))
		return
	}

	data, err := h.service.GetDetailPermintaanLabPK(r.Context(), noPermintaan)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoRawat, data.NoRawat)
		h.encryptDetailPermintaanLabPK(data, "")
	}

	response.Success(w, "Berhasil mengambil detail permintaan laboratorium PK", data)
}

func (h *Handler) UpdatePermintaanLabPK(w http.ResponseWriter, r *http.Request) {
	statusLanjut, ok := shared.ParseStatusLanjut(r.PathValue("status_lanjut"))
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	idKunjungan := strings.TrimSpace(r.PathValue("id_kunjungan"))
	noRawatURL, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	idPermintaan := strings.TrimSpace(r.PathValue("id_permintaan"))
	noPermintaanURL, err := crypto.Decrypt(idPermintaan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan laboratorium tidak valid"))
		return
	}

	var req SimpanPermintaanLabPKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data permintaan laboratorium tidak valid"))
		return
	}

	if req.NoRawat != noRawatURL {
		apperror.HandleError(w, apperror.NewBusinessError("Nomor rawat permintaan laboratorium tidak sesuai dengan ID kunjungan"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.decryptTindakanLabPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	data, err := h.service.UpdatePermintaanLabPK(r.Context(), kodeDokter, noRawatURL, noPermintaanURL, statusLanjut, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoPermintaan, data.NoPermintaan)
		h.encryptDetailPermintaanLabPK(data, idKunjungan)
	}

	response.Success(w, "Berhasil memperbarui permintaan laboratorium PK", data)
}

func (h *Handler) HapusPermintaanLabPK(w http.ResponseWriter, r *http.Request) {
	statusLanjut, ok := shared.ParseStatusLanjut(r.PathValue("status_lanjut"))
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	idKunjungan := strings.TrimSpace(r.PathValue("id_kunjungan"))
	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	idPermintaan := strings.TrimSpace(r.PathValue("id_permintaan"))
	noPermintaan, err := crypto.Decrypt(idPermintaan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan laboratorium tidak valid"))
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if err := h.service.HapusPermintaanLabPK(r.Context(), noRawat, noPermintaan, statusLanjut, kodeDokter); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil membatalkan permintaan laboratorium PK", nil)
}

func (h *Handler) decryptTindakanLabPayload(req *SimpanPermintaanLabPKRequest) error {
	valErrs := make(apperror.ValidationError)
	seenTindakan := make(map[string]bool, len(req.Pemeriksaan))

	for i := range req.Pemeriksaan {
		kodeTindakan, err := crypto.Decrypt(req.Pemeriksaan[i].IdTindakan, h.encryptionKey)
		if err != nil {
			valErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = "ID tindakan tidak valid"
			continue
		}
		if seenTindakan[kodeTindakan] {
			valErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: Tindakan laboratorium tidak boleh duplikat", i+1)
			continue
		}
		seenTindakan[kodeTindakan] = true
		req.Pemeriksaan[i].KodeTindakan = kodeTindakan

		seenTemplate := make(map[int]bool, len(req.Pemeriksaan[i].IdTemplate))
		for j, encIdTemplate := range req.Pemeriksaan[i].IdTemplate {
			idTemplateStr, err := crypto.Decrypt(encIdTemplate, h.encryptionKey)
			if err != nil {
				valErrs[fmt.Sprintf("pemeriksaan[%d].id_template[%d]", i, j)] = "ID template tidak valid"
				continue
			}
			idTemplateInt, err := strconv.Atoi(idTemplateStr)
			if err != nil {
				valErrs[fmt.Sprintf("pemeriksaan[%d].id_template[%d]", i, j)] = "Format ID template tidak valid"
				continue
			}
			if seenTemplate[idTemplateInt] {
				valErrs[fmt.Sprintf("pemeriksaan[%d].id_template[%d]", i, j)] = fmt.Sprintf("Pemeriksaan ke-%d parameter ke-%d: Template pengujian tidak boleh duplikat", i+1, j+1)
				continue
			}
			seenTemplate[idTemplateInt] = true
			req.Pemeriksaan[i].KodeTemplate = append(req.Pemeriksaan[i].KodeTemplate, idTemplateInt)
		}
	}

	if len(valErrs) > 0 {
		return valErrs
	}
	return nil
}

func (h *Handler) encryptPermintaanLabPK(item *PermintaanLabPK, defaultIdKunjungan string) {
	if item == nil {
		return
	}
	h.encryptPermintaanLabHeader(&item.PermintaanLabHeader, defaultIdKunjungan)
}

func (h *Handler) encryptPermintaanLabPKList(items []PermintaanLabPK, defaultIdKunjungan string) {
	for i := range items {
		h.encryptPermintaanLabPK(&items[i], defaultIdKunjungan)
	}
}

func (h *Handler) encryptDetailPermintaanLabPK(detail *DetailPermintaanLabPK, defaultIdKunjungan string) {
	if detail == nil {
		return
	}
	h.encryptPermintaanLabPK(&detail.PermintaanLabPK, defaultIdKunjungan)
	for i := range detail.Pemeriksaan {
		detail.Pemeriksaan[i].IdTindakan, _ = crypto.Encrypt(detail.Pemeriksaan[i].KodeTindakan, h.encryptionKey)
		for j := range detail.Pemeriksaan[i].DetailTemplate {
			if detail.Pemeriksaan[i].DetailTemplate[j].IdTemplate != "" {
				detail.Pemeriksaan[i].DetailTemplate[j].IdTemplate, _ = crypto.Encrypt(detail.Pemeriksaan[i].DetailTemplate[j].IdTemplate, h.encryptionKey)
			}
		}
	}
}
