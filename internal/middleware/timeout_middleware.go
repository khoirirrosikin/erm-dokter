package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"erm-dokter/internal/pkg/response"
	"erm-dokter/internal/shared/apperror"
)

type timeoutResponseWriter struct {
	mu          sync.Mutex
	header      http.Header
	buf         bytes.Buffer
	statusCode  int
	wroteHeader bool
	timedOut    bool
}

func newTimeoutResponseWriter(initialHeader http.Header) *timeoutResponseWriter {
	var cloned http.Header
	if initialHeader != nil {
		cloned = initialHeader.Clone()
	} else {
		cloned = make(http.Header)
	}
	return &timeoutResponseWriter{
		header:     cloned,
		statusCode: http.StatusOK,
	}
}

func (tw *timeoutResponseWriter) Header() http.Header {
	return tw.header
}

func (tw *timeoutResponseWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.statusCode = code
	tw.wroteHeader = true
}

func (tw *timeoutResponseWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.statusCode = http.StatusOK
		tw.wroteHeader = true
	}
	return tw.buf.Write(b)
}

func (tw *timeoutResponseWriter) markTimedOut() {
	tw.mu.Lock()
	tw.timedOut = true
	tw.mu.Unlock()
}

func (tw *timeoutResponseWriter) flushTo(w http.ResponseWriter) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	dst := w.Header()
	for k := range dst {
		delete(dst, k)
	}
	for k, vals := range tw.header {
		copied := make([]string, len(vals))
		copy(copied, vals)
		dst[k] = copied
	}

	if tw.wroteHeader {
		w.WriteHeader(tw.statusCode)
	}
	if tw.buf.Len() > 0 {
		_, _ = w.Write(tw.buf.Bytes())
	}
}

type recoveredPanic struct {
	val   any
	stack []byte
}

func handleRecoveredPanic(w http.ResponseWriter, p recoveredPanic) {
	if p.val == http.ErrAbortHandler {
		panic(http.ErrAbortHandler)
	}
	apperror.HandleError(w, fmt.Errorf("panic recovered: %v\n%s", p.val, p.stack))
}

func TimeoutMiddleware(timeout time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			tw := newTimeoutResponseWriter(w.Header())
			done := make(chan struct{})
			panicCh := make(chan recoveredPanic, 1)

			go func() {
				defer func() {
					if rec := recover(); rec != nil {
						panicCh <- recoveredPanic{
							val:   rec,
							stack: debug.Stack(),
						}
					}
				}()
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case p := <-panicCh:
				handleRecoveredPanic(w, p)
			case <-done:
				tw.flushTo(w)
			case <-ctx.Done():
				select {
				case p := <-panicCh:
					handleRecoveredPanic(w, p)
				case <-done:
					tw.flushTo(w)
				default:
					tw.markTimedOut()
					if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						response.Error(w, http.StatusGatewayTimeout, "Request timeout", nil)
					}
				}
			}
		}
	}
}
