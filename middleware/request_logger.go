package middleware

import (
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/sillen102/simba/logging"
)

var (
	excludePaths  = map[string]struct{}{}
	pathLogLevels = map[string]slog.Level{
		"/health": slog.LevelDebug,
	}
)

// LogRequests logs the incoming requests
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip logging for excluded paths
		if _, ok := excludePaths[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		wrapped := wrapResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Get duration
		duration := roundDuration(time.Since(start))

		// Log request details after processing
		logLevel := slog.LevelInfo // Default log level
		if level, ok := pathLogLevels[r.URL.Path]; ok {
			logLevel = level
		}

		logging.From(r.Context()).
			Log(r.Context(), logLevel, "request processed",
				"remoteIp", r.RemoteAddr,
				"userAgent", r.UserAgent(),
				"method", r.Method,
				"path", r.URL.Path,
				"protocol", r.Proto,
				"host", r.Host,
				"referer", r.Referer(),
				"status", wrapped.Status(),
				"duration (ms)", duration,
			)
	})
}

// ExcludePaths adds paths to the exclusion list for request logging
func ExcludePaths(paths ...string) {
	for _, path := range paths {
		excludePaths[path] = struct{}{}
	}
}

// ExcludePath adds a single path to the exclusion list for request logging
func ExcludePath(path string) {
	excludePaths[path] = struct{}{}
}

// SetPathLogLevel sets the log level for a specific path other than the default info level
func SetPathLogLevel(path string, level slog.Level) {
	if level == slog.LevelInfo {
		delete(pathLogLevels, path)
		return
	}
	pathLogLevels[path] = level
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
