package radiologi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Handler struct {
	service       Service
	encryptionKey string
}

func NewHandler(service Service, encryptionKey string) *Handler {
	return &Handler{
		service:       service,
		encryptionKey: encryptionKey,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.HandlerFunc) http.HandlerFunc, timeoutMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/v1/radiologi/hasil/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.DaftarHasilRadiologi)))
	mux.HandleFunc("GET /api/v1/radiologi/hasil/{status_lanjut}/pasien/{id_pasien}", authMiddleware(timeoutMiddleware(h.DaftarHasilRadiologiByRM)))
	mux.HandleFunc("GET /api/v1/radiologi/hasil/{id_radiologi}", authMiddleware(timeoutMiddleware(h.DetailHasilRadiologi)))

	mux.HandleFunc("POST /api/v1/radiologi/permintaan/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.SimpanPermintaanRadiologi)))
	mux.HandleFunc("GET /api/v1/radiologi/permintaan/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.DaftarPermintaanRadiologi)))
	mux.HandleFunc("GET /api/v1/radiologi/permintaan/{status_lanjut}/pasien/{id_pasien}", authMiddleware(timeoutMiddleware(h.DaftarPermintaanRadiologiByRM)))
	mux.HandleFunc("GET /api/v1/radiologi/permintaan/{id_permintaan}", authMiddleware(timeoutMiddleware(h.DetailPermintaanRadiologi)))
	mux.HandleFunc("PUT /api/v1/radiologi/permintaan/{status_lanjut}/{id_kunjungan}/{id_permintaan}", authMiddleware(timeoutMiddleware(h.UpdatePermintaanRadiologi)))
	mux.HandleFunc("DELETE /api/v1/radiologi/permintaan/{status_lanjut}/{id_kunjungan}/{id_permintaan}", authMiddleware(timeoutMiddleware(h.HapusPermintaanRadiologi)))
}

func (h *Handler) DaftarHasilRadiologi(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjutWithSemua(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := FilterRiwayatRadiologi{
		Tanggal: q.Get("tanggal"),
		Page:    page,
		Limit:   limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	list, meta, err := h.service.DaftarHasilRadiologi(r.Context(), noRawat, status, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptHasilRadiologiList(list, idKunjungan)

	response.SuccessWithMeta(w, "Berhasil mengambil riwayat hasil radiologi kunjungan", list, meta)
}

func (h *Handler) DaftarHasilRadiologiByRM(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idPasien := r.PathValue("id_pasien")

	status, ok := shared.ParseStatusLanjutWithSemua(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRM, err := crypto.Decrypt(idPasien, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID pasien tidak valid"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := FilterRiwayatRadiologi{
		Tanggal: q.Get("tanggal"),
		Page:    page,
		Limit:   limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	list, meta, err := h.service.DaftarHasilRadiologiByRM(r.Context(), noRM, status, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptHasilRadiologiList(list, "")

	response.SuccessWithMeta(w, "Berhasil mengambil riwayat hasil radiologi pasien", list, meta)
}

func (h *Handler) DetailHasilRadiologi(w http.ResponseWriter, r *http.Request) {
	idRadiologi := r.PathValue("id_radiologi")
	plainId, err := crypto.Decrypt(idRadiologi, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID hasil radiologi tidak valid"))
		return
	}

	idHasil, err := ParseIdHasilRadiologi(plainId)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format ID hasil radiologi tidak valid"))
		return
	}

	data, err := h.service.DetailHasilRadiologi(r.Context(), idHasil)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	data.Id = idRadiologi
	if encNoRawat, err := crypto.Encrypt(data.NoRawat, h.encryptionKey); err == nil {
		data.IdKunjungan = encNoRawat
	}

	response.Success(w, "Berhasil mengambil detail hasil radiologi", data)
}

func (h *Handler) SimpanPermintaanRadiologi(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawatURL, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req SimpanPermintaanRadiologiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data permintaan radiologi tidak valid"))
		return
	}

	if req.NoRawat != noRawatURL {
		apperror.HandleError(w, apperror.NewBusinessError("Nomor rawat permintaan radiologi tidak sesuai dengan ID kunjungan"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.decryptPermintaanPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	data, err := h.service.SimpanPermintaanRadiologi(r.Context(), kodeDokter, status, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoPermintaan, data.NoPermintaan)
	}
	h.encryptDetailPermintaanRadiologi(data)
	response.Created(w, "Berhasil menyimpan permintaan radiologi", data)
}

func (h *Handler) DaftarPermintaanRadiologi(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjutWithSemua(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	list, err := h.service.DaftarPermintaanRadiologi(r.Context(), noRawat, status)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range list {
		h.encryptDetailPermintaanRadiologi(&list[i])
	}

	response.Success(w, "Berhasil mengambil daftar permintaan radiologi", list)
}

func (h *Handler) DaftarPermintaanRadiologiByRM(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idPasien := r.PathValue("id_pasien")

	status, ok := shared.ParseStatusLanjutWithSemua(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRM, err := crypto.Decrypt(idPasien, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID pasien tidak valid"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := FilterRiwayatPermintaanRadiologi{
		Tanggal: q.Get("tanggal"),
		Page:    page,
		Limit:   limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	list, meta, err := h.service.DaftarPermintaanRadiologiByRM(r.Context(), noRM, status, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range list {
		h.encryptDetailPermintaanRadiologi(&list[i])
	}

	response.SuccessWithMeta(w, "Berhasil mengambil riwayat permintaan radiologi pasien", list, meta)
}

func (h *Handler) DetailPermintaanRadiologi(w http.ResponseWriter, r *http.Request) {
	idPermintaan := r.PathValue("id_permintaan")
	if idPermintaan == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan radiologi tidak valid"))
		return
	}

	noPermintaanURL, err := crypto.Decrypt(idPermintaan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan radiologi tidak valid"))
		return
	}

	data, err := h.service.DetailPermintaanRadiologi(r.Context(), noPermintaanURL)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if data != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoRawat, data.NoRawat)
	}
	h.encryptDetailPermintaanRadiologi(data)
	response.Success(w, "Berhasil mengambil detail permintaan radiologi", data)
}

func (h *Handler) UpdatePermintaanRadiologi(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	idPermintaan := r.PathValue("id_permintaan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawatURL, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	noPermintaanURL, err := crypto.Decrypt(idPermintaan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan radiologi tidak valid"))
		return
	}

	var req SimpanPermintaanRadiologiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data permintaan radiologi tidak valid"))
		return
	}

	if req.NoRawat != noRawatURL {
		apperror.HandleError(w, apperror.NewBusinessError("Nomor rawat permintaan radiologi tidak sesuai dengan ID kunjungan"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.decryptPermintaanPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	data, err := h.service.UpdatePermintaanRadiologi(r.Context(), kodeDokter, noRawatURL, noPermintaanURL, status, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptDetailPermintaanRadiologi(data)
	response.Success(w, "Berhasil memperbarui permintaan radiologi", data)
}

func (h *Handler) HapusPermintaanRadiologi(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	idPermintaan := r.PathValue("id_permintaan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawatURL, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	noPermintaanURL, err := crypto.Decrypt(idPermintaan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID permintaan radiologi tidak valid"))
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	err = h.service.HapusPermintaanRadiologi(r.Context(), kodeDokter, noRawatURL, noPermintaanURL, status)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil membatalkan permintaan radiologi", nil)
}

func (h *Handler) decryptPermintaanPayload(req *SimpanPermintaanRadiologiRequest) error {
	valErrs := make(apperror.ValidationError)
	seenTindakan := make(map[string]bool, len(req.Pemeriksaan))

	for i := range req.Pemeriksaan {
		plainTindakan, err := crypto.Decrypt(req.Pemeriksaan[i].IdTindakan, h.encryptionKey)
		if err != nil {
			valErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = "ID tindakan tidak valid"
			continue
		}
		if seenTindakan[plainTindakan] {
			valErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: Tindakan radiologi tidak boleh duplikat", i+1)
			continue
		}
		seenTindakan[plainTindakan] = true
		req.Pemeriksaan[i].KodeTindakan = plainTindakan
	}

	if len(valErrs) > 0 {
		return valErrs
	}
	return nil
}

func (h *Handler) encryptHasilRadiologiList(list []HasilRadiologi, defaultIdKunjungan string) {
	for i := range list {
		if encId, err := crypto.Encrypt(list[i].CompositeKey(), h.encryptionKey); err == nil {
			list[i].Id = encId
		}
		if defaultIdKunjungan != "" {
			list[i].IdKunjungan = defaultIdKunjungan
		} else if encNoRawat, err := crypto.Encrypt(list[i].NoRawat, h.encryptionKey); err == nil {
			list[i].IdKunjungan = encNoRawat
		}
	}
}

func (h *Handler) encryptDetailPermintaanRadiologi(detail *DetailPermintaanRadiologi) {
	if detail == nil {
		return
	}
	if encOrder, err := crypto.Encrypt(detail.NoPermintaan, h.encryptionKey); err == nil {
		detail.NoPermintaan = encOrder
	}
	for i := range detail.Pemeriksaan {
		if encTindakan, err := crypto.Encrypt(detail.Pemeriksaan[i].KodeTindakan, h.encryptionKey); err == nil {
			detail.Pemeriksaan[i].IdTindakan = encTindakan
		}
	}
}
