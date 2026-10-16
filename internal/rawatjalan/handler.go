package rawatjalan

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Handler struct {
	rawatJalanService Service
	encryptionKey     string
}

func NewHandler(service Service, encryptionKey string) *Handler {
	return &Handler{
		rawatJalanService: service,
		encryptionKey:     encryptionKey,
	}
}

func (h *Handler) RegisterRoutes(
	mux *http.ServeMux,
	authMiddleware func(http.HandlerFunc) http.HandlerFunc,
	serviceAuthMiddleware func(http.HandlerFunc) http.HandlerFunc,
	timeoutMiddleware func(http.HandlerFunc) http.HandlerFunc,
) {
	mux.HandleFunc("GET /api/v1/rawat-jalan/status-pemeriksaan", authMiddleware(timeoutMiddleware(h.DaftarStatusPemeriksaan)))
	mux.HandleFunc("GET /api/v1/rawat-jalan/status-lanjut", authMiddleware(timeoutMiddleware(h.DaftarStatusLanjut)))
	mux.HandleFunc("GET /api/v1/rawat-jalan/status-bayar", authMiddleware(timeoutMiddleware(h.DaftarStatusBayar)))
	mux.HandleFunc("GET /api/v1/rawat-jalan/jenis-antrean", authMiddleware(timeoutMiddleware(h.DaftarJenisAntrean)))

	mux.HandleFunc("GET /api/v1/rawat-jalan/antrean", serviceAuthMiddleware(timeoutMiddleware(h.DaftarAntreanDokter)))
	mux.HandleFunc("GET /api/v1/rawat-jalan/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.DetailKunjungan)))
	mux.HandleFunc("POST /api/v1/rawat-jalan/{id_kunjungan}/panggil", serviceAuthMiddleware(timeoutMiddleware(h.PanggilAntrean)))
}

func (h *Handler) DaftarAntreanDokter(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := FilterAntreanDokter{
		Tanggal:           q.Get("tanggal"),
		Penjamin:          q.Get("penjamin"),
		StatusPemeriksaan: StatusPemeriksaan(q.Get("status_pemeriksaan")),
		JenisAntrean:      JenisAntrean(q.Get("jenis_antrean")),
		StatusLanjut:      shared.StatusLanjut(q.Get("status_lanjut")),
		Keyword:           q.Get("keyword"),
		OrderBy:           q.Get("order_by"),
		SortOrder:         q.Get("sort_order"),
		Page:              page,
		Limit:             limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context(), true)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	daftarAntrean, meta, err := h.rawatJalanService.DaftarAntreanDokter(r.Context(), kodeDokter, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range daftarAntrean {
		kunjungan := &daftarAntrean[i]
		if encryptedId, err := crypto.Encrypt(kunjungan.NoRawat, h.encryptionKey); err == nil {
			kunjungan.Id = encryptedId
		}
		if encryptedIdPasien, err := crypto.Encrypt(kunjungan.NoRekamMedis, h.encryptionKey); err == nil {
			kunjungan.IdPasien = encryptedIdPasien
		}
		kunjungan.NoRekamMedis = kunjungan.FormatNoRekamMedis()
	}

	response.SuccessWithMeta(w, "Berhasil mengambil daftar antrean dokter", daftarAntrean, meta)
}

func (h *Handler) DetailKunjungan(w http.ResponseWriter, r *http.Request) {
	idKunjungan := r.PathValue("id_kunjungan")
	if idKunjungan == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan wajib diisi"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	kunjungan, err := h.rawatJalanService.DetailKunjungan(r.Context(), noRawat, kodeDokter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if kunjungan == nil {
		apperror.HandleError(w, apperror.NewNotFoundError("Detail kunjungan pasien tidak ditemukan"))
		return
	}
	kunjungan.Id = idKunjungan
	if encryptedIdPasien, err := crypto.Encrypt(kunjungan.NoRekamMedis, h.encryptionKey); err == nil {
		kunjungan.IdPasien = encryptedIdPasien
	}
	kunjungan.NoRekamMedis = kunjungan.FormatNoRekamMedis()

	response.Success(w, "Berhasil mengambil detail kunjungan pasien", kunjungan)
}

func (h *Handler) DaftarStatusPemeriksaan(w http.ResponseWriter, r *http.Request) {
	data := h.rawatJalanService.DaftarStatusPemeriksaan(r.Context())
	response.Success(w, "Berhasil mengambil referensi status pemeriksaan", data)
}

func (h *Handler) DaftarStatusLanjut(w http.ResponseWriter, r *http.Request) {
	data := h.rawatJalanService.DaftarStatusLanjut(r.Context())
	response.Success(w, "Berhasil mengambil referensi status lanjut", data)
}

func (h *Handler) DaftarStatusBayar(w http.ResponseWriter, r *http.Request) {
	data := h.rawatJalanService.DaftarStatusBayar(r.Context())
	response.Success(w, "Berhasil mengambil referensi status bayar", data)
}

func (h *Handler) DaftarJenisAntrean(w http.ResponseWriter, r *http.Request) {
	data := h.rawatJalanService.DaftarJenisAntrean(r.Context())
	response.Success(w, "Berhasil mengambil referensi jenis antrean", data)
}

func (h *Handler) PanggilAntrean(w http.ResponseWriter, r *http.Request) {
	idKunjungan := strings.TrimSpace(r.PathValue("id_kunjungan"))
	if idKunjungan == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan wajib diisi"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req PanggilAntreanRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context(), true)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	hasil, err := h.rawatJalanService.PanggilAntrean(r.Context(), noRawat, req.KodePoli, kodeDokter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil memanggil antrean pasien ke display", hasil)
}

