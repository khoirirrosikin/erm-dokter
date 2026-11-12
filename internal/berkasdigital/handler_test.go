package berkasdigital_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/pkg/crypto"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/shared/apperror"
)

const testEncryptionKey = "secret-key-32-bytes-testing-12345"

type mockService struct {
	getMasterBerkasFn    func(ctx context.Context) ([]berkasdigital.MasterBerkasDigital, error)
	getBerkasByNoRawatFn func(ctx context.Context, noRawat string, kodeList []string) ([]berkasdigital.BerkasDigitalPerawatan, error)
	getBerkasStreamFn    func(ctx context.Context, targetURL string) (*berkasdigital.BerkasStream, error)
	buildFullURLFn       func(lokasiFile string) string
}

func (m *mockService) GetMasterBerkas(ctx context.Context) ([]berkasdigital.MasterBerkasDigital, error) {
	if m.getMasterBerkasFn != nil {
		return m.getMasterBerkasFn(ctx)
	}
	return nil, nil
}

func (m *mockService) GetBerkasByNoRawat(ctx context.Context, noRawat string, kodeList []string) ([]berkasdigital.BerkasDigitalPerawatan, error) {
	if m.getBerkasByNoRawatFn != nil {
		return m.getBerkasByNoRawatFn(ctx, noRawat, kodeList)
	}
	return nil, nil
}

func (m *mockService) GetBerkasStream(ctx context.Context, targetURL string) (*berkasdigital.BerkasStream, error) {
	if m.getBerkasStreamFn != nil {
		return m.getBerkasStreamFn(ctx, targetURL)
	}
	return nil, nil
}

func (m *mockService) BuildFullURL(lokasiFile string) string {
	if m.buildFullURLFn != nil {
		return m.buildFullURLFn(lokasiFile)
	}
	return ""
}

func TestHandler_MasterBerkasDigital_Success(t *testing.T) {
	mockSvc := &mockService{
		getMasterBerkasFn: func(ctx context.Context) ([]berkasdigital.MasterBerkasDigital, error) {
			return []berkasdigital.MasterBerkasDigital{
				{Kode: "001", Nama: "Berkas SEP"},
				{Kode: "015", Nama: "PATOLOGI ANATOMI"},
			}, nil
		},
	}

	handler := berkasdigital.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/master/berkas-digital", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Code    int                                 `json:"code"`
		Status  string                              `json:"status"`
		Message string                              `json:"message"`
		Data    []berkasdigital.MasterBerkasDigital `json:"data"`
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 items, got %d", len(resp.Data))
	}
}

func TestHandler_StreamBerkasDigital_InvalidAndCrossModuleTokens(t *testing.T) {
	handler := berkasdigital.NewHandler(&mockService{}, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	reqInvalid := httptest.NewRequest(http.MethodGet, "/api/v1/berkas-digital/invalid-token", nil)
	rrInvalid := httptest.NewRecorder()
	mux.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for invalid token, got %d", rrInvalid.Code)
	}

	noRawatToken, err := crypto.Encrypt("2026/04/22/000001", testEncryptionKey)
	if err != nil {
		t.Fatalf("Failed to encrypt token: %v", err)
	}
	reqCross := httptest.NewRequest(http.MethodGet, "/api/v1/berkas-digital/"+noRawatToken, nil)
	rrCross := httptest.NewRecorder()
	mux.ServeHTTP(rrCross, reqCross)
	if rrCross.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for cross-module no_rawat token, got %d", rrCross.Code)
	}
}

func TestHandler_StreamBerkasDigital_SuccessAndSecurityHeaders(t *testing.T) {
	targetURL := "http://berkasrawat.example.com/webapps/berkasrawat/pages/upload/fnab.pdf"
	encToken, err := crypto.Encrypt(berkasdigital.FormatTokenPayload(targetURL), testEncryptionKey)
	if err != nil {
		t.Fatalf("Failed to encrypt token: %v", err)
	}

	dummyContent := []byte("%PDF-1.4 dummy pdf content")
	mockSvc := &mockService{
		getBerkasStreamFn: func(ctx context.Context, url string) (*berkasdigital.BerkasStream, error) {
			if url != targetURL {
				return nil, apperror.NewNotFoundError("Berkas tidak ditemukan")
			}
			return &berkasdigital.BerkasStream{
				Body:          io.NopCloser(bytes.NewReader(dummyContent)),
				ContentType:   "application/pdf",
				ContentLength: "26",
				Filename:      "fnab.pdf",
			}, nil
		},
	}

	handler := berkasdigital.NewHandler(mockSvc, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/berkas-digital/"+encToken, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Expected Content-Type application/pdf, got %s", ct)
	}
	if nosniff := rr.Header().Get("X-Content-Type-Options"); nosniff != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options nosniff, got %s", nosniff)
	}
	csp := rr.Header().Get("Content-Security-Policy")
	if csp == "" || strings.Contains(csp, "sandbox") {
		t.Errorf("Expected PDF CSP without sandbox (for Chrome PDF viewer compatibility), got %q", csp)
	}
	if rr.Body.String() != string(dummyContent) {
		t.Errorf("Unexpected body content: %s", rr.Body.String())
	}
}

func TestHandler_StreamBerkasDigital_ForcesAttachmentOnHTMLAndSupportsLargeMultiChunkAndRange(t *testing.T) {
	largePDF := bytes.Repeat([]byte("%PDF-1.4-LARGE-BLOCK-"), 20*1024)

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/webapps/berkasrawat/evil.html":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<script>alert('xss')</script>"))
		case "/webapps/berkasrawat/big.pdf":
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Accept-Ranges", "bytes")
			if r.Header.Get("Range") == "bytes=0-1023" {
				w.Header().Set("Content-Range", "bytes 0-1023/"+strconv.Itoa(len(largePDF)))
				w.Header().Set("Content-Length", "1024")
				w.WriteHeader(http.StatusPartialContent)
				_, _ = w.Write(largePDF[:1024])
				return
			}
			w.Header().Set("Content-Length", strconv.Itoa(len(largePDF)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(largePDF)
		}
	}))
	defer storage.Close()

	svc := berkasdigital.NewService(&mockRepository{}, storage.URL+"/webapps/berkasrawat", logger.New())
	handler := berkasdigital.NewHandler(svc, testEncryptionKey)
	mux := http.NewServeMux()
	dummyMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	handler.RegisterRoutes(mux, dummyMiddleware, dummyMiddleware)

	htmlToken, _ := crypto.Encrypt(berkasdigital.FormatTokenPayload(storage.URL+"/webapps/berkasrawat/evil.html"), testEncryptionKey)
	reqHTML := httptest.NewRequest(http.MethodGet, "/api/v1/berkas-digital/"+htmlToken, nil)
	rrHTML := httptest.NewRecorder()
	mux.ServeHTTP(rrHTML, reqHTML)

	if rrHTML.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rrHTML.Code)
	}
	if ct := rrHTML.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("Expected forced application/octet-stream, got %s", ct)
	}
	if disp := rrHTML.Header().Get("Content-Disposition"); !strings.HasPrefix(disp, "attachment;") {
		t.Errorf("Expected forced attachment disposition, got %s", disp)
	}
	if csp := rrHTML.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "sandbox") {
		t.Errorf("Expected sandbox CSP on attachment response, got %s", csp)
	}

	pdfToken, _ := crypto.Encrypt(berkasdigital.FormatTokenPayload(storage.URL+"/webapps/berkasrawat/big.pdf"), testEncryptionKey)

	reqFull := httptest.NewRequest(http.MethodGet, "/api/v1/berkas-digital/"+pdfToken, nil)
	rrFull := httptest.NewRecorder()
	mux.ServeHTTP(rrFull, reqFull)
	if rrFull.Code != http.StatusOK || rrFull.Body.Len() != len(largePDF) {
		t.Fatalf("Expected full multi-chunk stream len %d, got code=%d len=%d", len(largePDF), rrFull.Code, rrFull.Body.Len())
	}

	reqRange := httptest.NewRequest(http.MethodGet, "/api/v1/berkas-digital/"+pdfToken, nil)
	reqRange.Header.Set("Range", "bytes=0-1023")
	rrRange := httptest.NewRecorder()
	mux.ServeHTTP(rrRange, reqRange)
	if rrRange.Code != http.StatusPartialContent {
		t.Fatalf("Expected 206 Partial Content, got %d", rrRange.Code)
	}
	if rrRange.Body.Len() != 1024 || rrRange.Header().Get("Content-Range") == "" {
		t.Errorf("Expected 1024 partial bytes and Content-Range header, got len=%d cr=%q", rrRange.Body.Len(), rrRange.Header().Get("Content-Range"))
	}
}
