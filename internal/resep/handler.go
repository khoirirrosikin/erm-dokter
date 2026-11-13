package resep

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
	resepService  Service
	encryptionKey string
}

func NewHandler(service Service, encryptionKey string) *Handler {
	return &Handler{
		resepService:  service,
		encryptionKey: encryptionKey,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.HandlerFunc) http.HandlerFunc, timeoutMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/v1/resep/aturan-pakai", authMiddleware(timeoutMiddleware(h.DaftarAturanPakai)))
	mux.HandleFunc("GET /api/v1/resep/metode-racik", authMiddleware(timeoutMiddleware(h.DaftarMetodeRacik)))
	mux.HandleFunc("GET /api/v1/resep/{status_lanjut}/pasien/{id_pasien}", authMiddleware(timeoutMiddleware(h.DaftarResepByRM)))
	mux.HandleFunc("GET /api/v1/resep/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.DaftarResep)))
	mux.HandleFunc("GET /api/v1/resep/{id_resep}", authMiddleware(timeoutMiddleware(h.DetailResep)))
	mux.HandleFunc("POST /api/v1/resep/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.SimpanResep)))
	mux.HandleFunc("PUT /api/v1/resep/{status_lanjut}/{id_kunjungan}/{id_resep}", authMiddleware(timeoutMiddleware(h.UpdateResep)))
	mux.HandleFunc("DELETE /api/v1/resep/{status_lanjut}/{id_kunjungan}/{id_resep}", authMiddleware(timeoutMiddleware(h.HapusResep)))
}

func (h *Handler) DaftarAturanPakai(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	data, err := h.resepService.DaftarAturanPakai(r.Context(), keyword)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil mengambil daftar aturan pakai", data)
}

func (h *Handler) DaftarMetodeRacik(w http.ResponseWriter, r *http.Request) {
	data, err := h.resepService.DaftarMetodeRacik(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil mengambil daftar metode racik", data)
}

func (h *Handler) DaftarResep(w http.ResponseWriter, r *http.Request) {
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

	filter := FilterDaftarResep{
		Tanggal: q.Get("tanggal"),
		Page:    page,
		Limit:   limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	daftarResep, meta, err := h.resepService.DaftarResep(r.Context(), noRawat, status, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptResepResponse(daftarResep, idKunjungan)

	response.SuccessWithMeta(w, "Berhasil mengambil daftar resep", daftarResep, meta)
}

func (h *Handler) DaftarResepByRM(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idPasien := r.PathValue("id_pasien")

	status, ok := shared.ParseStatusLanjutWithSemua(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRekamMedis, err := crypto.Decrypt(idPasien, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID pasien tidak valid"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := FilterDaftarResep{
		Tanggal: q.Get("tanggal"),
		Page:    page,
		Limit:   limit,
	}

	filter.Sanitize()
	if errs := filter.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	daftarResep, meta, err := h.resepService.DaftarResepByRM(r.Context(), noRekamMedis, status, filter)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	h.encryptResepResponse(daftarResep, "")

	response.SuccessWithMeta(w, "Berhasil mengambil riwayat resep pasien", daftarResep, meta)
}

func (h *Handler) DetailResep(w http.ResponseWriter, r *http.Request) {
	idResep := r.PathValue("id_resep")
	if idResep == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID resep tidak valid"))
		return
	}

	noResep, err := crypto.Decrypt(idResep, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID resep tidak valid"))
		return
	}

	resep, err := h.resepService.DetailResep(r.Context(), noResep)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if resep != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoRawat, resep.NoRawat)
	}

	listResep := []Resep{*resep}
	h.encryptResepResponse(listResep, "")
	*resep = listResep[0]

	response.Success(w, "Berhasil mengambil detail resep obat", resep)
}

func (h *Handler) SimpanResep(w http.ResponseWriter, r *http.Request) {

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

	var req SimpanResepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data resep tidak valid"))
		return
	}

	if req.NoRawat != noRawatURL {
		apperror.HandleError(w, apperror.NewBusinessError("Nomor rawat resep tidak sesuai dengan ID kunjungan"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.decryptObatPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	resep, err := h.resepService.SimpanResep(r.Context(), kodeDokter, status, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if resep != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoResep, resep.NoResep)
		listResep := []Resep{*resep}
		h.encryptResepResponse(listResep, idKunjungan)
		*resep = listResep[0]
	}

	response.Created(w, "Berhasil menyimpan resep obat", resep)
}

func (h *Handler) UpdateResep(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	idResep := r.PathValue("id_resep")

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

	noResep, err := crypto.Decrypt(idResep, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID resep tidak valid"))
		return
	}

	var req SimpanResepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data resep tidak valid"))
		return
	}

	if req.NoRawat != noRawatURL {
		apperror.HandleError(w, apperror.NewBusinessError("Nomor rawat resep tidak sesuai dengan ID kunjungan"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.decryptObatPayload(&req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	resep, err := h.resepService.UpdateResep(r.Context(), kodeDokter, noRawatURL, noResep, status, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if resep != nil {
		middleware.SetLogRef(w, middleware.HeaderLogNoResep, resep.NoResep)
		listResep := []Resep{*resep}
		h.encryptResepResponse(listResep, idKunjungan)
		*resep = listResep[0]
	}

	response.Success(w, "Berhasil memperbarui resep obat", resep)
}

func (h *Handler) HapusResep(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	idResep := r.PathValue("id_resep")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("status lanjut tidak valid (Ralan/Ranap)"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	noResep, err := crypto.Decrypt(idResep, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID resep tidak valid"))
		return
	}

	kodeDokter, err := middleware.GetKodeDokter(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if err := h.resepService.HapusResep(r.Context(), kodeDokter, noRawat, noResep, status); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil menghapus resep obat", nil)
}

func (h *Handler) decryptObatPayload(req *SimpanResepRequest) error {
	valErrs := make(apperror.ValidationError)

	seenDokter := make(map[string]bool, len(req.ResepDokter))
	for i := range req.ResepDokter {
		kodeObat, err := crypto.Decrypt(req.ResepDokter[i].IdObat, h.encryptionKey)
		if err != nil {
			valErrs[fmt.Sprintf("resep_dokter[%d].id_obat", i)] = "ID obat tidak valid"
			continue
		}
		if seenDokter[kodeObat] {
			valErrs[fmt.Sprintf("resep_dokter[%d].id_obat", i)] = fmt.Sprintf("Obat ke-%d: Obat tidak boleh duplikat dalam daftar non-racikan", i+1)
			continue
		}
		seenDokter[kodeObat] = true
		req.ResepDokter[i].KodeObat = kodeObat
	}

	for i := range req.ResepRacikan {
		seenDetail := make(map[string]bool, len(req.ResepRacikan[i].Detail))
		for j := range req.ResepRacikan[i].Detail {
			kodeObat, err := crypto.Decrypt(req.ResepRacikan[i].Detail[j].IdObat, h.encryptionKey)
			if err != nil {
				valErrs[fmt.Sprintf("resep_racikan[%d].detail[%d].id_obat", i, j)] = "ID obat tidak valid"
				continue
			}
			if seenDetail[kodeObat] {
				valErrs[fmt.Sprintf("resep_racikan[%d].detail[%d].id_obat", i, j)] = fmt.Sprintf("Racikan ke-%d bahan ke-%d: Bahan obat tidak boleh duplikat dalam satu racikan", i+1, j+1)
				continue
			}
			seenDetail[kodeObat] = true
			req.ResepRacikan[i].Detail[j].KodeObat = kodeObat
		}
	}

	if len(valErrs) > 0 {
		return valErrs
	}
	return nil
}

func (h *Handler) encryptResepResponse(daftarResep []Resep, defaultIdKunjungan string) {

	for i := range daftarResep {
		res := &daftarResep[i]

		if encryptedId, err := crypto.Encrypt(res.NoResep, h.encryptionKey); err == nil {
			res.Id = encryptedId
		}

		if defaultIdKunjungan != "" {
			res.IdKunjungan = defaultIdKunjungan
		} else if encryptedKunjungan, err := crypto.Encrypt(res.NoRawat, h.encryptionKey); err == nil {
			res.IdKunjungan = encryptedKunjungan
		}

		for j := range res.ResepDokter {
			rd := &res.ResepDokter[j]
			if encryptedObat, err := crypto.Encrypt(rd.KodeObat, h.encryptionKey); err == nil {
				rd.IdObat = encryptedObat
			}
		}

		for j := range res.ResepDokterRacikan {
			rdr := &res.ResepDokterRacikan[j]
			for k := range rdr.DetailRacikan {
				detail := &rdr.DetailRacikan[k]
				if encryptedObat, err := crypto.Encrypt(detail.KodeObat, h.encryptionKey); err == nil {
					detail.IdObat = encryptedObat
				}
			}
		}
	}
}
