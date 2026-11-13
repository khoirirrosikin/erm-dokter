package pasien

import (
	"net/http"
	"strconv"
	"strings"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/shared/apperror"
	"erm-dokter/internal/shared/formatter"
)

type Handler struct {
	pasienService Service
	encryptionKey string
}

func NewHandler(service Service, encryptionKey string) *Handler {
	return &Handler{
		pasienService: service,
		encryptionKey: encryptionKey,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.HandlerFunc) http.HandlerFunc, timeoutMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/v1/pasien/{id_pasien}/riwayat-kunjungan", authMiddleware(timeoutMiddleware(h.RiwayatKunjungan)))
}

func (h *Handler) RiwayatKunjungan(w http.ResponseWriter, r *http.Request) {
	if _, err := middleware.GetKodeDokter(r.Context()); err != nil {
		apperror.HandleError(w, err)
		return
	}

	idPasien := strings.TrimSpace(r.PathValue("id_pasien"))
	if idPasien == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID pasien wajib diisi"))
		return
	}

	cleanNoRM, err := crypto.Decrypt(idPasien, h.encryptionKey)
	if err != nil || strings.TrimSpace(cleanNoRM) == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID pasien tidak valid"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := FilterRiwayatKunjungan{
		Tanggal: q.Get("tanggal"),
		Page:    page,
		Limit:   limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	daftar, meta, err := h.pasienService.RiwayatKunjunganPasien(r.Context(), cleanNoRM, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range daftar {
		item := &daftar[i]
		if encIdKunjungan, err := crypto.Encrypt(item.NoRawat, h.encryptionKey); err == nil {
			item.IdKunjungan = encIdKunjungan
		}
		if encIdPasien, err := crypto.Encrypt(item.NoRekamMedis, h.encryptionKey); err == nil {
			item.IdPasien = encIdPasien
		}
		item.NoRekamMedis = formatter.FormatNoRekamMedis(item.NoRekamMedis)
	}

	response.SuccessWithMeta(w, "Berhasil mengambil riwayat kunjungan pasien", daftar, meta)
}
