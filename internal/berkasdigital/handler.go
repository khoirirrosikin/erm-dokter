package berkasdigital

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/shared/apperror"
)

const (
	streamChunkSize        = 64 * 1024
	defaultWriteIdleWindow = 30 * time.Second
)

var streamBufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, streamChunkSize)
		return &buf
	},
}

type Handler struct {
	service          Service
	encryptionKey    string
	writeIdleTimeout time.Duration
}

func NewHandler(service Service, encryptionKey string) *Handler {
	return &Handler{
		service:          service,
		encryptionKey:    encryptionKey,
		writeIdleTimeout: defaultWriteIdleWindow,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.HandlerFunc) http.HandlerFunc, timeoutMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/v1/berkas-digital/{id_berkas}", authMiddleware(h.StreamBerkasDigital))
	mux.HandleFunc("GET /api/v1/master/berkas-digital", authMiddleware(timeoutMiddleware(h.MasterBerkasDigital)))
}

func (h *Handler) StreamBerkasDigital(w http.ResponseWriter, r *http.Request) {
	idBerkas := strings.TrimSpace(r.PathValue("id_berkas"))
	if idBerkas == "" {
		apperror.HandleError(w, apperror.NewBusinessError("ID berkas digital wajib diisi"))
		return
	}

	decrypted, err := crypto.Decrypt(idBerkas, h.encryptionKey)
	if err != nil {
		apperror.HandleError(w, apperror.NewBusinessError("Token berkas digital tidak valid"))
		return
	}

	targetURL, ok := ParseTokenPayload(decrypted)
	if !ok {
		apperror.HandleError(w, apperror.NewBusinessError("Token berkas digital tidak valid"))
		return
	}

	ctx := WithRangeHeaders(r.Context(), r.Header.Get("Range"), r.Header.Get("If-Range"))
	stream, err := h.service.GetBerkasStream(ctx, targetURL)
	if err != nil {
		apperror.HandleError(w, err)
		return
	}
	defer stream.Body.Close()

	if stream.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if cr := sanitizeHeaderValue(stream.ContentRange); cr != "" {
			w.Header().Set("Content-Range", cr)
		}
		if ar := sanitizeHeaderValue(stream.AcceptRanges); ar != "" {
			w.Header().Set("Accept-Ranges", ar)
		}
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	contentType, disposition, filename := ResolveSafeContentDisposition(stream.ContentType, stream.Filename)
	if stream.Disposition == "attachment" {
		disposition = "attachment"
		contentType = "application/octet-stream"
	}

	headers := w.Header()
	headers.Set("Content-Type", contentType)
	headers.Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", disposition, filename))
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("Content-Security-Policy", buildStreamCSP(contentType, disposition))

	if cl := sanitizeHeaderValue(stream.ContentLength); cl != "" {
		headers.Set("Content-Length", cl)
	}
	if ar := sanitizeHeaderValue(stream.AcceptRanges); ar != "" {
		headers.Set("Accept-Ranges", ar)
	}
	if cr := sanitizeHeaderValue(stream.ContentRange); cr != "" {
		headers.Set("Content-Range", cr)
	}
	if etag := sanitizeHeaderValue(stream.ETag); etag != "" {
		headers.Set("ETag", etag)
	}
	if lm := sanitizeHeaderValue(stream.LastModified); lm != "" {
		headers.Set("Last-Modified", lm)
	}

	statusCode := http.StatusOK
	if stream.StatusCode == http.StatusPartialContent {
		statusCode = http.StatusPartialContent
	}

	writeWindow := h.writeIdleTimeout
	if writeWindow <= 0 {
		writeWindow = defaultWriteIdleWindow
	}

	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Now().Add(writeWindow))
	w.WriteHeader(statusCode)

	bufPtr := streamBufPool.Get().(*[]byte)
	buf := *bufPtr
	defer streamBufPool.Put(bufPtr)

	for {
		n, readErr := stream.Body.Read(buf)
		if n > 0 {
			_ = rc.SetWriteDeadline(time.Now().Add(writeWindow))
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return
			}
		}
		if readErr != nil {
			break
		}
	}
	_ = rc.Flush()
}

func buildStreamCSP(contentType, disposition string) string {
	if disposition == "inline" && contentType == "application/pdf" {
		return "default-src 'none'; object-src 'self'; img-src 'self' data: blob:; style-src 'unsafe-inline'"
	}
	if disposition == "inline" {
		return "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; sandbox"
	}
	return "default-src 'none'; sandbox"
}

func (h *Handler) MasterBerkasDigital(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.GetMasterBerkas(r.Context())
	if err != nil {
		apperror.HandleError(w, err)
		return
	}

	response.Success(w, "Berhasil mengambil master berkas digital", list)
}
