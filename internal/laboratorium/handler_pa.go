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

func (h *Handler) SimpanPermintaanLabPA(w http.ResponseWriter, r *http.Request) {
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

	var req SimpanPermintaanLabPARequest
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

	if err := h.decryptTindakanLabPAPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	data, err := h.service.SimpanPermintaanLabPA(r.Context(), kodeDokter, statusLanjut, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoPermintaan, data.NoPermintaan)
		h.encryptDetailPermintaanLabPA(data, idKunjungan)
	}

	response.Created(w, "Berhasil mengirim permintaan laboratorium PA", data)
}

func (h *Handler) DaftarPermintaanLabPA(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.service.GetDaftarPermintaanLabPA(r.Context(), noRawat, statusLanjut)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptPermintaanLabPAList(data, idKunjungan)
	response.Success(w, "Berhasil mengambil daftar permintaan laboratorium PA", data)
}

func (h *Handler) DaftarPermintaanLabPAByRM(w http.ResponseWriter, r *http.Request) {
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

	data, meta, err := h.service.GetRiwayatPermintaanLabPAByRM(r.Context(), noRM, statusLanjut, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptPermintaanLabPAList(data, "")
	response.SuccessWithMeta(w, "Berhasil mengambil riwayat permintaan laboratorium PA pasien", data, meta)
}

func (h *Handler) DetailPermintaanLabPA(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.service.GetDetailPermintaanLabPA(r.Context(), noPermintaan)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoRawat, data.NoRawat)
		h.encryptDetailPermintaanLabPA(data, "")
	}

	response.Success(w, "Berhasil mengambil detail permintaan laboratorium PA", data)
}

func (h *Handler) UpdatePermintaanLabPA(w http.ResponseWriter, r *http.Request) {
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

	var req SimpanPermintaanLabPARequest
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

	if err := h.decryptTindakanLabPAPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	data, err := h.service.UpdatePermintaanLabPA(r.Context(), kodeDokter, noRawatURL, noPermintaanURL, statusLanjut, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoPermintaan, data.NoPermintaan)
		h.encryptDetailPermintaanLabPA(data, idKunjungan)
	}

	response.Success(w, "Berhasil memperbarui permintaan laboratorium PA", data)
}

func (h *Handler) HapusPermintaanLabPA(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.HapusPermintaanLabPA(r.Context(), noRawat, noPermintaan, statusLanjut, kodeDokter); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil membatalkan permintaan laboratorium PA", nil)
}

func (h *Handler) decryptTindakanLabPAPayload(req *SimpanPermintaanLabPARequest) error {
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
	}

	if len(valErrs) > 0 {
		return valErrs
	}
	return nil
}

func (h *Handler) encryptPermintaanLabPA(item *PermintaanLabPA, defaultIdKunjungan string) {
	if item == nil {
		return
	}
	h.encryptPermintaanLabHeader(&item.PermintaanLabHeader, defaultIdKunjungan)
}

func (h *Handler) encryptPermintaanLabPAList(items []PermintaanLabPA, defaultIdKunjungan string) {
	for i := range items {
		h.encryptPermintaanLabPA(&items[i], defaultIdKunjungan)
	}
}

func (h *Handler) encryptDetailPermintaanLabPA(detail *DetailPermintaanLabPA, defaultIdKunjungan string) {
	if detail == nil {
		return
	}
	h.encryptPermintaanLabPA(&detail.PermintaanLabPA, defaultIdKunjungan)
	for i := range detail.Pemeriksaan {
		detail.Pemeriksaan[i].IdTindakan, _ = crypto.Encrypt(detail.Pemeriksaan[i].KodeTindakan, h.encryptionKey)
	}
}
