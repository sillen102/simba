package middleware

import (
	"net/http"
	"time"
)

// LogRequests logs the incoming requests
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		// Get logger from context
		logger := getLogger(r.Context())

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request details after processing
		duration := time.Since(start).Round(time.Microsecond)
		logger.Info().
			Str("remoteIp", r.RemoteAddr).
			Str("userAgent", r.UserAgent()).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.Status()).
			Dur("duration (ms)", duration).
			Msg("request processed")
	})
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging
type responseWriter struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}
