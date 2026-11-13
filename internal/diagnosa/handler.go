package diagnosa

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) RegisterRoutes(
	mux *http.ServeMux,
	authMiddleware func(http.HandlerFunc) http.HandlerFunc,
	timeoutMiddleware func(http.HandlerFunc) http.HandlerFunc,
) {
	mux.HandleFunc("GET /api/v1/diagnosa/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.DaftarDiagnosaProsedur)))
	mux.HandleFunc("GET /api/v1/diagnosa/{status_lanjut}/pasien/{id_pasien}", authMiddleware(timeoutMiddleware(h.RiwayatPasien)))
	mux.HandleFunc("POST /api/v1/diagnosa/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.TambahDiagnosa)))
	mux.HandleFunc("PUT /api/v1/diagnosa/{status_lanjut}/{id_kunjungan}/reorder", authMiddleware(timeoutMiddleware(h.ReorderDiagnosa)))
	mux.HandleFunc("PUT /api/v1/diagnosa/{status_lanjut}/{id_kunjungan}/{id}", authMiddleware(timeoutMiddleware(h.UpdateDiagnosa)))
	mux.HandleFunc("DELETE /api/v1/diagnosa/{status_lanjut}/{id_kunjungan}/{id}", authMiddleware(timeoutMiddleware(h.HapusDiagnosa)))
	mux.HandleFunc("POST /api/v1/diagnosa/{status_lanjut}/{id_kunjungan}/simulasi-eklaim", authMiddleware(timeoutMiddleware(h.SimulasiEklaim)))

	mux.HandleFunc("POST /api/v1/prosedur/{status_lanjut}/{id_kunjungan}", authMiddleware(timeoutMiddleware(h.TambahProsedur)))
	mux.HandleFunc("PUT /api/v1/prosedur/{status_lanjut}/{id_kunjungan}/reorder", authMiddleware(timeoutMiddleware(h.ReorderProsedur)))
	mux.HandleFunc("PUT /api/v1/prosedur/{status_lanjut}/{id_kunjungan}/{id}", authMiddleware(timeoutMiddleware(h.UpdateProsedur)))
	mux.HandleFunc("DELETE /api/v1/prosedur/{status_lanjut}/{id_kunjungan}/{id}", authMiddleware(timeoutMiddleware(h.HapusProsedur)))
}

func (h *Handler) DaftarDiagnosaProsedur(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.service.DaftarDiagnosaProsedur(r.Context(), noRawat, status)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range data.Diagnosa {
		item := &data.Diagnosa[i]
		if encId, err := crypto.Encrypt(item.CompositeKey(), h.encryptionKey); err == nil {
			item.Id = encId
		}
		item.IdKunjungan = idKunjungan
	}

	for i := range data.Prosedur {
		item := &data.Prosedur[i]
		if encId, err := crypto.Encrypt(item.CompositeKey(), h.encryptionKey); err == nil {
			item.Id = encId
		}
		item.IdKunjungan = idKunjungan
	}

	response.Success(w, "Berhasil mengambil daftar diagnosa dan prosedur", data)
}

func (h *Handler) RiwayatPasien(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.service.RiwayatPasien(r.Context(), noRekamMedis, status)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	for i := range data.Diagnosa {
		item := &data.Diagnosa[i]
		if encId, err := crypto.Encrypt(item.CompositeKey(), h.encryptionKey); err == nil {
			item.Id = encId
		}
		if encRawat, err := crypto.Encrypt(item.NoRawat, h.encryptionKey); err == nil {
			item.IdKunjungan = encRawat
		}
	}

	for i := range data.Prosedur {
		item := &data.Prosedur[i]
		if encId, err := crypto.Encrypt(item.CompositeKey(), h.encryptionKey); err == nil {
			item.Id = encId
		}
		if encRawat, err := crypto.Encrypt(item.NoRawat, h.encryptionKey); err == nil {
			item.IdKunjungan = encRawat
		}
	}

	response.Success(w, "Berhasil mengambil riwayat diagnosa dan prosedur pasien", data)
}

func (h *Handler) TambahDiagnosa(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req TambahDiagnosaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data diagnosa tidak valid"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	diagnosa, err := h.service.TambahDiagnosa(r.Context(), noRawat, status, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if diagnosa != nil {
		if encId, err := crypto.Encrypt(diagnosa.CompositeKey(), h.encryptionKey); err == nil {
			diagnosa.Id = encId
		}
		diagnosa.IdKunjungan = idKunjungan
	}

	response.Created(w, "Berhasil menambahkan diagnosa", diagnosa)
}

func (h *Handler) UpdateDiagnosa(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	encryptedId := r.PathValue("id")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	decryptedId, err := crypto.Decrypt(encryptedId, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID diagnosa tidak valid"))
		return
	}

	idDiagnosa, err := ParseIdDiagnosa(decryptedId)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError(err.Error()))
		return
	}

	if idDiagnosa.NoRawat != noRawat {
		apperror.HandleError(w, apperror.NewBusinessError("ID diagnosa tidak sesuai dengan ID kunjungan"))
		return
	}

	if idDiagnosa.Status != status {
		apperror.HandleError(w, apperror.NewBusinessError("Status ID diagnosa tidak sesuai dengan status lanjut kunjungan"))
		return
	}

	var req UpdateDiagnosaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data diagnosa tidak valid"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.service.UpdateDiagnosa(r.Context(), idDiagnosa, req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil memperbarui data diagnosa", nil)
}

func (h *Handler) HapusDiagnosa(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	encryptedId := r.PathValue("id")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	decryptedId, err := crypto.Decrypt(encryptedId, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID diagnosa tidak valid"))
		return
	}

	idDiagnosa, err := ParseIdDiagnosa(decryptedId)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError(err.Error()))
		return
	}

	if idDiagnosa.NoRawat != noRawat {
		apperror.HandleError(w, apperror.NewBusinessError("ID diagnosa tidak sesuai dengan ID kunjungan"))
		return
	}

	if idDiagnosa.Status != status {
		apperror.HandleError(w, apperror.NewBusinessError("Status ID diagnosa tidak sesuai dengan status lanjut kunjungan"))
		return
	}

	if err := h.service.HapusDiagnosa(r.Context(), idDiagnosa); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil menghapus data diagnosa", nil)
}

func (h *Handler) resolveReorderItems(items []ReorderItemRequest, noRawat string, status shared.StatusLanjut, isProsedur bool) error {
	label := "diagnosa"
	if isProsedur {
		label = "prosedur"
	}

	seenKode := make(map[string]bool, len(items))
	for i := range items {
		if items[i].Id != "" {
			decryptedId, err := crypto.Decrypt(items[i].Id, h.encryptionKey)
			if err != nil {
				return apperror.NewBusinessError("ID " + label + " tidak valid")
			}

			var parsedNoRawat, parsedKode string
			var parsedStatus shared.StatusLanjut
			if isProsedur {
				idProsedur, err := ParseIdProsedur(decryptedId)
				if err != nil {
					return apperror.NewBusinessError(err.Error())
				}
				parsedNoRawat, parsedKode, parsedStatus = idProsedur.NoRawat, idProsedur.Kode, idProsedur.Status
			} else {
				idDiagnosa, err := ParseIdDiagnosa(decryptedId)
				if err != nil {
					return apperror.NewBusinessError(err.Error())
				}
				parsedNoRawat, parsedKode, parsedStatus = idDiagnosa.NoRawat, idDiagnosa.Kode, idDiagnosa.Status
			}

			if parsedNoRawat != noRawat {
				return apperror.NewBusinessError("ID " + label + " tidak sesuai dengan ID kunjungan")
			}
			if parsedStatus != status {
				return apperror.NewBusinessError("Status ID " + label + " tidak sesuai dengan status lanjut kunjungan")
			}
			if items[i].Kode != "" && items[i].Kode != parsedKode {
				return apperror.ValidationError{
					"items": "Kode tidak sesuai dengan ID " + label + " pada urutan item",
				}
			}
			items[i].Kode = parsedKode
		}

		if seenKode[items[i].Kode] {
			return apperror.ValidationError{
				"items": "Item '" + items[i].Kode + "' duplikat dalam urutan",
			}
		}
		seenKode[items[i].Kode] = true
	}

	return nil
}

func (h *Handler) ReorderDiagnosa(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data urutan diagnosa tidak valid"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.resolveReorderItems(req.Items, noRawat, status, false); err != nil {
		apperror.HandleError(w, err)
		return
	}

	if err := h.service.ReorderDiagnosa(r.Context(), noRawat, status, req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil mengurutkan data diagnosa", nil)
}

func (h *Handler) TambahProsedur(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req TambahProsedurRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data prosedur tidak valid"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	prosedur, err := h.service.TambahProsedur(r.Context(), noRawat, status, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	if prosedur != nil {
		if encId, err := crypto.Encrypt(prosedur.CompositeKey(), h.encryptionKey); err == nil {
			prosedur.Id = encId
		}
		prosedur.IdKunjungan = idKunjungan
	}

	response.Created(w, "Berhasil menambahkan prosedur", prosedur)
}

func (h *Handler) UpdateProsedur(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	encryptedId := r.PathValue("id")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	decryptedId, err := crypto.Decrypt(encryptedId, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID prosedur tidak valid"))
		return
	}

	idProsedur, err := ParseIdProsedur(decryptedId)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError(err.Error()))
		return
	}

	if idProsedur.NoRawat != noRawat {
		apperror.HandleError(w, apperror.NewBusinessError("ID prosedur tidak sesuai dengan ID kunjungan"))
		return
	}

	if idProsedur.Status != status {
		apperror.HandleError(w, apperror.NewBusinessError("Status ID prosedur tidak sesuai dengan status lanjut kunjungan"))
		return
	}

	var req UpdateProsedurRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data prosedur tidak valid"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.service.UpdateProsedur(r.Context(), idProsedur, req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil memperbarui data prosedur", nil)
}

func (h *Handler) HapusProsedur(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")
	encryptedId := r.PathValue("id")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	decryptedId, err := crypto.Decrypt(encryptedId, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID prosedur tidak valid"))
		return
	}

	idProsedur, err := ParseIdProsedur(decryptedId)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError(err.Error()))
		return
	}

	if idProsedur.NoRawat != noRawat {
		apperror.HandleError(w, apperror.NewBusinessError("ID prosedur tidak sesuai dengan ID kunjungan"))
		return
	}

	if idProsedur.Status != status {
		apperror.HandleError(w, apperror.NewBusinessError("Status ID prosedur tidak sesuai dengan status lanjut kunjungan"))
		return
	}

	if err := h.service.HapusProsedur(r.Context(), idProsedur); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil menghapus data prosedur", nil)
}

func (h *Handler) ReorderProsedur(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Format data urutan prosedur tidak valid"))
		return
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	if err := h.resolveReorderItems(req.Items, noRawat, status, true); err != nil {
		apperror.HandleError(w, err)
		return
	}

	if err := h.service.ReorderProsedur(r.Context(), noRawat, status, req); err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil mengurutkan data prosedur", nil)
}

func (h *Handler) SimulasiEklaim(w http.ResponseWriter, r *http.Request) {
	statusLanjut := r.PathValue("status_lanjut")
	idKunjungan := r.PathValue("id_kunjungan")

	status, ok := shared.ParseStatusLanjut(statusLanjut)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Status lanjut tidak valid"))
		return
	}

	noRawat, err := crypto.Decrypt(idKunjungan, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("ID kunjungan tidak valid"))
		return
	}

	var req SimulasiEklaimRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apperror.HandleError(w, apperror.NewBusinessError("Format data simulasi tidak valid"))
			return
		}
	}

	req.Sanitize()
	if errs := req.Validate(); errs != nil {
		apperror.HandleError(w, errs)
		return
	}

	result, err := h.service.SimulasiEklaim(r.Context(), noRawat, status, req)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Simulasi biaya E-Klaim berhasil", result)
}
