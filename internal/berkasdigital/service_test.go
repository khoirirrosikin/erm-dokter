package berkasdigital_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	fetchMasterBerkasFn    func(ctx context.Context) ([]berkasdigital.MasterBerkasDigital, error)
	fetchBerkasByNoRawatFn func(ctx context.Context, noRawat string, kodeList []string) ([]berkasdigital.BerkasDigitalPerawatan, error)
}

func (m *mockRepository) FetchMasterBerkas(ctx context.Context) ([]berkasdigital.MasterBerkasDigital, error) {
	if m.fetchMasterBerkasFn != nil {
		return m.fetchMasterBerkasFn(ctx)
	}
	return nil, nil
}

func (m *mockRepository) FetchBerkasByNoRawat(ctx context.Context, noRawat string, kodeList []string) ([]berkasdigital.BerkasDigitalPerawatan, error) {
	if m.fetchBerkasByNoRawatFn != nil {
		return m.fetchBerkasByNoRawatFn(ctx, noRawat, kodeList)
	}
	return nil, nil
}

const testBaseURL = "http://berkasrawat.example.com/webapps/berkasrawat/"

func TestService_GetMasterBerkas(t *testing.T) {
	log := logger.New()
	mockRepo := &mockRepository{
		fetchMasterBerkasFn: func(ctx context.Context) ([]berkasdigital.MasterBerkasDigital, error) {
			return []berkasdigital.MasterBerkasDigital{
				{Kode: "001", Nama: "Berkas SEP"},
				{Kode: "015", Nama: "PATOLOGI ANATOMI"},
			}, nil
		},
	}

	svc := berkasdigital.NewService(mockRepo, testBaseURL, log)
	list, err := svc.GetMasterBerkas(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list))
	}
}

func TestService_GetBerkasByNoRawat(t *testing.T) {
	log := logger.New()
	mockRepo := &mockRepository{
		fetchBerkasByNoRawatFn: func(ctx context.Context, noRawat string, kodeList []string) ([]berkasdigital.BerkasDigitalPerawatan, error) {
			return []berkasdigital.BerkasDigitalPerawatan{
				{
					NoRawat:    "2026/04/22/000001",
					Kode:       "005",
					NamaBerkas: "HASIL LAB PK",
					LokasiFile: "pages/upload/pk.pdf",
				},
			}, nil
		},
	}

	svc := berkasdigital.NewService(mockRepo, testBaseURL, log)
	list, err := svc.GetBerkasByNoRawat(context.Background(), "2026/04/22/000001", []string{"005"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].NamaBerkas != "HASIL LAB PK" {
		t.Errorf("Unexpected result: %+v", list)
	}
}

func TestService_BuildFullURL_SecurityAndSpaces(t *testing.T) {
	log := logger.New()
	svc := berkasdigital.NewService(&mockRepository{}, testBaseURL, log)

	if got := svc.BuildFullURL("pages/upload/fnab.pdf"); got != "http://berkasrawat.example.com/webapps/berkasrawat/pages/upload/fnab.pdf" {
		t.Errorf("Unexpected normal URL: %s", got)
	}

	if got := svc.BuildFullURL("/pages/upload/Hasil PA (1).pdf"); got == "" {
		t.Errorf("Expected filename with spaces to build valid URL, got empty")
	}

	sameHostFull := "http://berkasrawat.example.com/webapps/berkasrawat/pages/upload/ok.pdf"
	if got := svc.BuildFullURL(sameHostFull); got != sameHostFull {
		t.Errorf("Expected same-host full URL preserved, got %s", got)
	}

	invalidCases := []string{
		"",
		"https://storage.rs.com/file.pdf",
		"http://169.254.169.254/latest/meta-data",
		"//evil.example.com/file.pdf",
		"../conf/config.php",
		"pages/upload/../../conf/config.php",
		"pages/upload/%2e%2e/config.php",
		"pages/upload/..%2fconfig.php",
		`pages\upload\fnab.pdf`,
		"pages/upload/fnab.pdf?cmd=whoami",
	}
	for _, tc := range invalidCases {
		if got := svc.BuildFullURL(tc); got != "" {
			t.Errorf("Expected invalid lokasi_file %q to be rejected, got %q", tc, got)
		}
	}
}

func TestService_GetBerkasStream_Success(t *testing.T) {
	dummyContent := []byte("%PDF-1.4 dummy pdf content")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(dummyContent)
	}))
	defer ts.Close()

	log := logger.New()
	svc := berkasdigital.NewService(&mockRepository{}, ts.URL, log)

	stream, err := svc.GetBerkasStream(context.Background(), ts.URL+"/test.pdf")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer stream.Body.Close()

	if stream.ContentType != "application/pdf" || stream.Disposition != "inline" {
		t.Errorf("Expected inline application/pdf, got contentType=%s disposition=%s", stream.ContentType, stream.Disposition)
	}

	bodyBytes, errRead := io.ReadAll(stream.Body)
	if errRead != nil {
		t.Fatalf("Failed to read stream body: %v", errRead)
	}
	if string(bodyBytes) != string(dummyContent) {
		t.Errorf("Unexpected body content: %s", string(bodyBytes))
	}
}

func TestService_GetBerkasStream_BlocksSSRFPathTraversalAndOpenRedirect(t *testing.T) {
	var internalHit atomic.Bool
	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalHit.Store(true)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("SECRET_INTERNAL_DATA"))
	}))
	defer internalServer.Close()

	storageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webapps/berkasrawat/redirect.pdf" {
			http.Redirect(w, r, internalServer.URL+"/secret", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-1.4 ok"))
	}))
	defer storageServer.Close()

	log := logger.New()
	svc := berkasdigital.NewService(&mockRepository{}, storageServer.URL+"/webapps/berkasrawat", log)

	ssrfTargets := []string{
		internalServer.URL + "/webapps/berkasrawat/file.pdf",
		storageServer.URL + "/otherapp/file.pdf",
		storageServer.URL + "/webapps/berkasrawat/../../etc/passwd",
		storageServer.URL + "/webapps/berkasrawat/%2e%2e/secret.php",
		storageServer.URL + "/webapps/berkasrawat/",
		storageServer.URL + "/webapps/berkasrawat/file.pdf?x=1",
	}
	for _, target := range ssrfTargets {
		_, err := svc.GetBerkasStream(context.Background(), target)
		var bizErr *apperror.BusinessError
		if !errors.As(err, &bizErr) {
			t.Errorf("Expected BusinessError for %q, got %v", target, err)
		}
	}

	_, err := svc.GetBerkasStream(context.Background(), storageServer.URL+"/webapps/berkasrawat/redirect.pdf")
	if err == nil {
		t.Fatalf("Expected redirect to internal server to be blocked")
	}
	if internalHit.Load() {
		t.Fatalf("SSRF Open Redirect reached internal server!")
	}
}

func TestService_GetBerkasStream_SanitizesXSSContentTypes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/xss.html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte("<script>alert(1)</script>"))
		case "/xss.svg":
			w.Header().Set("Content-Type", "image/svg+xml")
			_, _ = w.Write([]byte("<svg onload='alert(1)'/>"))
		}
	}))
	defer ts.Close()

	svc := berkasdigital.NewService(&mockRepository{}, ts.URL, logger.New())

	for _, p := range []string{"/xss.html", "/xss.svg"} {
		stream, err := svc.GetBerkasStream(context.Background(), ts.URL+p)
		if err != nil {
			t.Fatalf("Unexpected error for %s: %v", p, err)
		}
		_ = stream.Body.Close()
		if stream.ContentType != "application/octet-stream" || stream.Disposition != "attachment" {
			t.Errorf("Expected %s forced to attachment octet-stream, got ct=%s disp=%s", p, stream.ContentType, stream.Disposition)
		}
	}
}

func TestService_GetBerkasStream_RangeAndSlidingIdleTimeout(t *testing.T) {
	payload := bytes.Repeat([]byte("ABCDEFGH"), 16*1024)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/range.pdf":
			if r.Header.Get("Range") == "bytes=0-15" && r.Header.Get("If-Range") == `"v1"` {
				w.Header().Set("Content-Type", "application/pdf")
				w.Header().Set("Content-Range", "bytes 0-15/131072")
				w.Header().Set("Accept-Ranges", "bytes")
				w.WriteHeader(http.StatusPartialContent)
				_, _ = w.Write(payload[:16])
				return
			}
			w.WriteHeader(http.StatusOK)
		case "/slow-active.pdf":
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			for i := 0; i < 5; i++ {
				_, _ = w.Write([]byte("chunk-data-"))
				if flusher != nil {
					flusher.Flush()
				}
				time.Sleep(30 * time.Millisecond)
			}
		case "/stalled.pdf":
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			time.Sleep(220 * time.Millisecond)
			_, _ = w.Write([]byte("too-late"))
		}
	}))
	defer ts.Close()

	svc := berkasdigital.NewService(&mockRepository{}, ts.URL, logger.New(), berkasdigital.WithIdleTimeout(90*time.Millisecond))

	rangeCtx := berkasdigital.WithRangeHeaders(context.Background(), "bytes=0-15", `"v1"`)
	rangeStream, err := svc.GetBerkasStream(rangeCtx, ts.URL+"/range.pdf")
	if err != nil {
		t.Fatalf("Unexpected error on range request: %v", err)
	}
	rangeBytes, err := io.ReadAll(rangeStream.Body)
	_ = rangeStream.Body.Close()
	if err != nil {
		t.Fatalf("Failed reading range body: %v", err)
	}
	if rangeStream.StatusCode != http.StatusPartialContent || rangeStream.ContentRange != "bytes 0-15/131072" || len(rangeBytes) != 16 {
		t.Errorf("Unexpected range response: status=%d cr=%q len=%d", rangeStream.StatusCode, rangeStream.ContentRange, len(rangeBytes))
	}

	activeStream, err := svc.GetBerkasStream(context.Background(), ts.URL+"/slow-active.pdf")
	if err != nil {
		t.Fatalf("Unexpected error on active slow stream: %v", err)
	}
	activeBytes, err := io.ReadAll(activeStream.Body)
	_ = activeStream.Body.Close()
	if err != nil {
		t.Fatalf("Expected slow active stream to complete without timeout, got error: %v", err)
	}
	if string(activeBytes) != "chunk-data-chunk-data-chunk-data-chunk-data-chunk-data-" {
		t.Errorf("Unexpected active stream body: %q", string(activeBytes))
	}

	stalledStream, err := svc.GetBerkasStream(context.Background(), ts.URL+"/stalled.pdf")
	if err != nil {
		t.Fatalf("Unexpected initial error on stalled stream: %v", err)
	}
	_, err = io.ReadAll(stalledStream.Body)
	_ = stalledStream.Body.Close()
	if err == nil {
		t.Fatalf("Expected idle timeout error when upstream stalls without sending bytes")
	}
}

func TestService_GetBerkasStream_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	svc := berkasdigital.NewService(&mockRepository{}, ts.URL, logger.New())

	_, err := svc.GetBerkasStream(context.Background(), ts.URL+"/missing.pdf")
	if err == nil {
		t.Fatalf("Expected error for 404 response, got nil")
	}
}
