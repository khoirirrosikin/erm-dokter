package berkasdigital

import (
	"context"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/shared/apperror"
)

const (
	tokenPrefix              = "berkas:"
	defaultStreamIdleTimeout = 30 * time.Second
	maxStreamRedirects       = 3
)

type rangeContextKey struct{}

type rangeHeaders struct {
	rangeVal   string
	ifRangeVal string
}

func WithRangeHeaders(ctx context.Context, rangeHeader, ifRangeHeader string) context.Context {
	cleanRange := sanitizeHeaderValue(rangeHeader)
	if cleanRange == "" || !strings.HasPrefix(strings.ToLower(cleanRange), "bytes=") {
		return ctx
	}

	cleanIfRange := sanitizeHeaderValue(ifRangeHeader)
	if len(cleanIfRange) > 256 {
		cleanIfRange = ""
	}

	return context.WithValue(ctx, rangeContextKey{}, rangeHeaders{
		rangeVal:   cleanRange,
		ifRangeVal: cleanIfRange,
	})
}

func FormatTokenPayload(targetURL string) string {
	targetURL = strings.TrimSpace(targetURL)
	if targetURL == "" {
		return ""
	}
	if strings.HasPrefix(targetURL, tokenPrefix) {
		return targetURL
	}
	return tokenPrefix + targetURL
}

func ParseTokenPayload(decrypted string) (string, bool) {
	decrypted = strings.TrimSpace(decrypted)
	if strings.HasPrefix(decrypted, tokenPrefix) {
		rawURL := strings.TrimSpace(strings.TrimPrefix(decrypted, tokenPrefix))
		return rawURL, rawURL != ""
	}
	if strings.HasPrefix(decrypted, "http://") || strings.HasPrefix(decrypted, "https://") {
		return decrypted, true
	}
	return "", false
}

type ServiceOption func(*service)

func WithIdleTimeout(d time.Duration) ServiceOption {
	return func(s *service) {
		if d > 0 {
			s.idleTimeout = d
		}
	}
}

type Service interface {
	GetMasterBerkas(ctx context.Context) ([]MasterBerkasDigital, error)
	GetBerkasByNoRawat(ctx context.Context, noRawat string, kodeList []string) ([]BerkasDigitalPerawatan, error)
	GetBerkasStream(ctx context.Context, targetURL string) (*BerkasStream, error)
	BuildFullURL(lokasiFile string) string
}

type service struct {
	repo             Repository
	urlBerkasDigital string
	baseURL          *url.URL
	basePrefix       string
	idleTimeout      time.Duration
	log              *logger.Logger
	httpClient       *http.Client
}

func NewService(repo Repository, urlBerkasDigital string, log *logger.Logger, opts ...ServiceOption) Service {
	parsedBase, basePrefix, normalizedBase := parseBaseURL(urlBerkasDigital)

	s := &service{
		repo:             repo,
		urlBerkasDigital: normalizedBase,
		baseURL:          parsedBase,
		basePrefix:       basePrefix,
		idleTimeout:      defaultStreamIdleTimeout,
		log:              log,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		DisableCompression:    true,
	}

	s.httpClient = &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxStreamRedirects {
				return apperror.NewBusinessError("Terlalu banyak pengalihan URL berkas digital")
			}
			if _, _, err := s.validateAndNormalizeURL(req.URL.String(), true); err != nil {
				return err
			}
			return nil
		},
	}

	return s
}

func parseBaseURL(raw string) (*url.URL, string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.ContainsAny(raw, "\\\x00\r\n") {
		return nil, "", ""
	}

	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, "", ""
	}

	scheme := strings.ToLower(u.Scheme)
	trimmedPath := strings.Trim(u.Path, "/")
	basePrefix := "/"
	if trimmedPath != "" {
		cleanBase := path.Clean("/" + trimmedPath)
		basePrefix = cleanBase + "/"
	}

	cleanURL := &url.URL{
		Scheme: scheme,
		Host:   u.Host,
		Path:   basePrefix,
	}
	return cleanURL, basePrefix, cleanURL.String()
}

func (s *service) GetMasterBerkas(ctx context.Context) ([]MasterBerkasDigital, error) {
	list, err := s.repo.FetchMasterBerkas(ctx)
	if err != nil {
		s.log.Error("gagal mengambil master berkas digital: %v", err)
		return nil, err
	}
	return list, nil
}

func (s *service) GetBerkasByNoRawat(ctx context.Context, noRawat string, kodeList []string) ([]BerkasDigitalPerawatan, error) {
	list, err := s.repo.FetchBerkasByNoRawat(ctx, noRawat, kodeList)
	if err != nil {
		s.log.Error("gagal mengambil berkas digital no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	return list, nil
}

func (s *service) GetBerkasStream(ctx context.Context, targetURL string) (*BerkasStream, error) {
	canonicalURL, filenameHint, err := s.validateAndNormalizeURL(targetURL, false)
	if err != nil {
		return nil, err
	}

	reqCtx, cancelReq := context.WithCancel(ctx)
	cleanup := true
	defer func() {
		if cleanup {
			cancelReq()
		}
	}()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, canonicalURL, nil)
	if err != nil {
		s.log.Error("gagal membuat request stream berkas %s: %v", canonicalURL, err)
		return nil, apperror.NewBusinessError("URL berkas digital tidak valid")
	}

	if rh, ok := ctx.Value(rangeContextKey{}).(rangeHeaders); ok && rh.rangeVal != "" {
		req.Header.Set("Range", rh.rangeVal)
		if rh.ifRangeVal != "" {
			req.Header.Set("If-Range", rh.ifRangeVal)
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Error("gagal fetch berkas dari storage %s: %v", canonicalURL, err)
		return nil, apperror.NewNotFoundError("Berkas digital tidak dapat diakses dari server penyimpanan")
	}

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		_ = resp.Body.Close()
		return &BerkasStream{
			Body:         http.NoBody,
			StatusCode:   http.StatusRequestedRangeNotSatisfiable,
			AcceptRanges: sanitizeHeaderValue(resp.Header.Get("Accept-Ranges")),
			ContentRange: sanitizeHeaderValue(resp.Header.Get("Content-Range")),
		}, nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		_ = resp.Body.Close()
		s.log.Warn("server berkas mengembalikan status %d untuk URL %s", resp.StatusCode, canonicalURL)
		return nil, apperror.NewNotFoundError("Berkas digital tidak ditemukan di server penyimpanan")
	}

	safeContentType, disposition, safeFilename := ResolveSafeContentDisposition(resp.Header.Get("Content-Type"), filenameHint)
	acceptRanges := sanitizeHeaderValue(resp.Header.Get("Accept-Ranges"))
	if acceptRanges == "" {
		acceptRanges = "bytes"
	}

	cleanup = false
	return &BerkasStream{
		Body:          newIdleTimeoutReadCloser(resp.Body, cancelReq, s.idleTimeout),
		StatusCode:    resp.StatusCode,
		ContentType:   safeContentType,
		ContentLength: sanitizeHeaderValue(resp.Header.Get("Content-Length")),
		Filename:      safeFilename,
		Disposition:   disposition,
		AcceptRanges:  acceptRanges,
		ContentRange:  sanitizeHeaderValue(resp.Header.Get("Content-Range")),
		ETag:          sanitizeHeaderValue(resp.Header.Get("ETag")),
		LastModified:  sanitizeHeaderValue(resp.Header.Get("Last-Modified")),
	}, nil
}

func (s *service) BuildFullURL(lokasiFile string) string {
	lokasiFile = strings.TrimSpace(lokasiFile)
	if lokasiFile == "" || s.baseURL == nil || strings.HasPrefix(lokasiFile, "//") {
		return ""
	}

	lower := strings.ToLower(lokasiFile)
	if strings.Contains(lokasiFile, "://") || strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:") {
		canonicalURL, _, err := s.validateAndNormalizeURL(lokasiFile, false)
		if err != nil {
			return ""
		}
		return canonicalURL
	}

	segs, ok := sanitizeRelativeSegments(lokasiFile)
	if !ok {
		return ""
	}

	escaped := make([]string, len(segs))
	for i, seg := range segs {
		escaped[i] = url.PathEscape(seg)
	}
	return s.urlBerkasDigital + strings.Join(escaped, "/")
}

func (s *service) validateAndNormalizeURL(rawTarget string, allowHTTPSUpgrade bool) (string, string, error) {
	rawTarget = strings.TrimSpace(rawTarget)
	if rawTarget == "" || s.baseURL == nil || strings.ContainsAny(rawTarget, "\\\x00\r\n") {
		return "", "", apperror.NewBusinessError("URL berkas digital tidak valid")
	}

	u, err := url.Parse(rawTarget)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", "", apperror.NewBusinessError("URL berkas digital tidak valid")
	}

	scheme := strings.ToLower(u.Scheme)
	baseScheme := strings.ToLower(s.baseURL.Scheme)
	if scheme != baseScheme {
		if !(allowHTTPSUpgrade && baseScheme == "http" && scheme == "https") {
			return "", "", apperror.NewBusinessError("URL berkas digital tidak valid")
		}
	}

	if !strings.EqualFold(u.Host, s.baseURL.Host) {
		return "", "", apperror.NewBusinessError("URL berkas digital tidak valid")
	}

	segs, ok := sanitizeRelativeSegments(u.EscapedPath())
	if !ok {
		return "", "", apperror.NewBusinessError("URL berkas digital tidak valid")
	}

	cleanUnescapedPath := "/" + strings.Join(segs, "/")
	if !strings.HasPrefix(cleanUnescapedPath, s.basePrefix) {
		return "", "", apperror.NewBusinessError("URL berkas digital tidak valid")
	}

	relAfterBase := strings.TrimPrefix(cleanUnescapedPath, s.basePrefix)
	if relAfterBase == "" {
		return "", "", apperror.NewBusinessError("URL berkas digital tidak valid")
	}

	escaped := make([]string, len(segs))
	for i, seg := range segs {
		escaped[i] = url.PathEscape(seg)
	}

	canonicalURL := scheme + "://" + s.baseURL.Host + "/" + strings.Join(escaped, "/")
	return canonicalURL, segs[len(segs)-1], nil
}

func sanitizeRelativeSegments(rawPath string) ([]string, bool) {
	if rawPath == "" || strings.ContainsAny(rawPath, "\\\x00\r\n?#") {
		return nil, false
	}

	parts := strings.Split(rawPath, "/")
	segs := make([]string, 0, len(parts))
	for _, seg := range parts {
		if seg == "" {
			continue
		}
		if seg == "." || seg == ".." {
			return nil, false
		}
		unesc, err := url.PathUnescape(seg)
		if err != nil || unesc == "" || unesc == "." || unesc == ".." {
			return nil, false
		}
		if strings.ContainsAny(unesc, "/\\\x00\r\n") {
			return nil, false
		}
		segs = append(segs, unesc)
	}

	if len(segs) == 0 {
		return nil, false
	}
	return segs, true
}

func ResolveSafeContentDisposition(rawContentType, filenameHint string) (string, string, string) {
	safeFilename := sanitizeFilename(filenameHint)

	mediaType := ""
	trimmedCT := strings.TrimSpace(rawContentType)
	if trimmedCT != "" {
		parsed, _, err := mime.ParseMediaType(trimmedCT)
		if err == nil {
			mediaType = strings.ToLower(strings.TrimSpace(parsed))
		} else {
			mediaType = strings.ToLower(strings.TrimSpace(strings.Split(trimmedCT, ";")[0]))
		}
	}

	if mediaType == "" || mediaType == "application/octet-stream" || mediaType == "binary/octet-stream" {
		switch strings.ToLower(path.Ext(safeFilename)) {
		case ".pdf":
			mediaType = "application/pdf"
		case ".jpg", ".jpeg":
			mediaType = "image/jpeg"
		case ".png":
			mediaType = "image/png"
		case ".webp":
			mediaType = "image/webp"
		case ".gif":
			mediaType = "image/gif"
		default:
			mediaType = "application/octet-stream"
		}
	}

	safeContentType := "application/octet-stream"
	disposition := "attachment"
	switch mediaType {
	case "application/pdf", "image/jpeg", "image/png", "image/webp", "image/gif":
		safeContentType = mediaType
		disposition = "inline"
	}

	if safeFilename == "" {
		if safeContentType == "application/pdf" {
			safeFilename = "dokumen.pdf"
		} else {
			safeFilename = "berkas-digital"
		}
	}

	return safeContentType, disposition, safeFilename
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if idx := strings.LastIndexAny(name, "/\\"); idx >= 0 {
		name = name[idx+1:]
	}

	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if r < 0x20 || r == 0x7f || r == '"' || r == '\'' || r == ';' || r == '/' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}

	cleaned := strings.TrimSpace(b.String())
	if cleaned == "." || cleaned == ".." {
		return ""
	}
	return cleaned
}

func sanitizeHeaderValue(val string) string {
	val = strings.TrimSpace(val)
	if strings.ContainsAny(val, "\r\n\x00") {
		return ""
	}
	return val
}

type idleTimeoutReadCloser struct {
	mu        sync.Mutex
	body      io.ReadCloser
	cancel    context.CancelFunc
	idleLimit time.Duration
	timer     *time.Timer
	closed    bool
	closeErr  error
}

func newIdleTimeoutReadCloser(body io.ReadCloser, cancel context.CancelFunc, idleLimit time.Duration) io.ReadCloser {
	if idleLimit <= 0 {
		idleLimit = defaultStreamIdleTimeout
	}
	return &idleTimeoutReadCloser{
		body:      body,
		cancel:    cancel,
		idleLimit: idleLimit,
		timer:     time.AfterFunc(idleLimit, cancel),
	}
}

func (r *idleTimeoutReadCloser) resetTimer() {
	r.mu.Lock()
	if !r.closed {
		r.timer.Reset(r.idleLimit)
	}
	r.mu.Unlock()
}

func (r *idleTimeoutReadCloser) Read(p []byte) (int, error) {
	r.resetTimer()
	n, err := r.body.Read(p)
	if n > 0 {
		r.resetTimer()
	}
	return n, err
}

func (r *idleTimeoutReadCloser) Close() error {
	r.mu.Lock()
	if r.closed {
		err := r.closeErr
		r.mu.Unlock()
		return err
	}
	r.closed = true
	r.timer.Stop()
	r.mu.Unlock()

	err := r.body.Close()
	r.cancel()

	r.mu.Lock()
	r.closeErr = err
	r.mu.Unlock()
	return err
}
