package middleware

import (
	"math"
	"net/http"
	"time"

	"github.com/sillen102/simba/logging"
)

// LogRequests logs the incoming requests
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Get duration
		duration := roundDuration(time.Since(start))

		// Log request details after processing
		logging.From(r.Context()).Info("request processed",
			"remoteIp", r.RemoteAddr,
			"userAgent", r.UserAgent(),
			"status", wrapped.Status(),
			"duration (ms)", duration,
		)
	})
}

// roundDuration returns the duration in milliseconds rounded to 3 decimal places
func roundDuration(d time.Duration) float64 {
	duration := d.Seconds() * 1000
	return math.Round(duration*1000) / 1000
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
