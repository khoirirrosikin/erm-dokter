package rujukaninternal

import (
	"encoding/json"
	"net/http"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/response"
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
	mux.HandleFunc("GET /api/v1/rujukan-internal/opsi-poli", authMiddleware(timeoutMiddleware(h.DaftarOpsiPoliDokter)))
	mux.HandleFunc("GET /api/v1/rujukan-internal/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.DaftarRujukanInternal)))
	mux.HandleFunc("POST /api/v1/rujukan-internal/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.SimpanRujukanInternal)))
	mux.HandleFunc("DELETE /api/v1/rujukan-internal/{id_kunjungan}/{id_rujukan}", authMiddleware(timeoutMiddleware(h.HapusRujukanInternal)))
}

func (h *Handler) DaftarOpsiPoliDokter(w http.ResponseWriter, r *http.Request) {
	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	keyword := r.URL.Query().Get("keyword")
	opsi, err := h.service.DaftarOpsiPoliDokter(r.Context(), kodeDokter, keyword)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range opsi {
		if encrypted, err := crypto.Encrypt(opsi[i].CompositeKey(), h.encryptionKey); err == nil {
			opsi[i].Id = encrypted
		}
	}

	response.Success(w, "Berhasil mengambil opsi dokter dan poliklinik rujukan", opsi)
}

func (h *Handler) DaftarRujukanInternal(w http.ResponseWriter, r *http.Request) {
	idKunjungan := r.PathValue("id_kunjungan")
	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	daftar, err := h.service.DaftarRujukanInternal(r.Context(), noRawat)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range daftar {
		if encrypted, err := crypto.Encrypt(daftar[i].CompositeKey(), h.encryptionKey); err == nil {
			daftar[i].Id = encrypted
		}
		daftar[i].IdKunjungan = idKunjungan
	}

	response.Success(w, "Berhasil mengambil daftar rujukan internal", daftar)
}

func (h *Handler) SimpanRujukanInternal(w http.ResponseWriter, r *http.Request) {
	idKunjungan := r.PathValue("id_kunjungan")
	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req SimpanRujukanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data rujukan tidak valid"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	decryptedTujuan, err := crypto.Decrypt(req.IdTujuan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID poli rujukan tidak valid"))
		return
	}

	target, err := ParseIdOpsiPoliDokter(decryptedTujuan)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError(err.Error()))
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	hasil, err := h.service.SimpanRujukanInternal(r.Context(), kodeDokter, noRawat, target.KodePoli, target.KodeDokter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if hasil != nil {
		if encrypted, err := crypto.Encrypt(hasil.CompositeKey(), h.encryptionKey); err == nil {
			hasil.Id = encrypted
		}
		hasil.IdKunjungan = idKunjungan
	}

	response.Created(w, "Berhasil menyimpan rujukan internal", hasil)
}

func (h *Handler) HapusRujukanInternal(w http.ResponseWriter, r *http.Request) {
	idKunjungan := r.PathValue("id_kunjungan")
	idRujukan := r.PathValue("id_rujukan")

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	decryptedId, err := crypto.Decrypt(idRujukan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID rujukan tidak valid"))
		return
	}

	parsedId, err := ParseIdRujukanInternal(decryptedId)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError(err.Error()))
		return
	}

	if parsedId.NoRawat != noRawat {
		apperror.HandleError(w, apperror.NewForbiddenError("ID rujukan tidak sesuai dengan ID kunjungan"))
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if err := h.service.HapusRujukanInternal(r.Context(), kodeDokter, noRawat, parsedId.KodePoli, parsedId.KodeDokter); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil membatalkan rujukan internal", nil)
}
