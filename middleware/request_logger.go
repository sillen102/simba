package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sillen102/simba/logging"
)

// RequestLoggerConfig is a middleware that logs the incoming requests.
type RequestLoggerConfig struct {
	LogRequestBody bool
}

// LogRequests logs the incoming requests.
func (rl *RequestLoggerConfig) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		// Get logger from context
		logger := logging.FromCtx(r.Context())

		if rl.LogRequestBody {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				var bodyJson map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyJson); err == nil {
					bodyLogger := logger.With().
						Interface("requestBody", bodyJson).
						Logger()
					logger = &bodyLogger
				}
			}
		}

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
// written HTTP status code to be captured for logging.
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
