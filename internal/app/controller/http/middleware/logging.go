package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	log "github.com/go-kit/log"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

type LoggingMiddleware struct {
	log.Logger
}

func (m *LoggingMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.Log(
					"err", err,
					"trace", debug.Stack(),
				)
			}
		}()
		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		m.Log(
			"status", wrapped.status,
			"method", r.Method,
			"path", r.URL.EscapedPath(),
			"duration", time.Since(start),
		)
	}
	return http.HandlerFunc(fn)
}

func NewLoggingMiddleware(l log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{l}
}