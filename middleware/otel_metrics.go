package middleware

import (
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/telemetry"
)

// OtelMetrics returns middleware that collects HTTP metrics using OpenTelemetry.
// It returns a pass-through handler if telemetry is disabled for zero overhead.
func OtelMetrics(provider *telemetry.Provider, telemetrySettings *settings.Telemetry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// If telemetry is disabled or metrics are disabled, return pass-through
		if provider == nil || !telemetrySettings.Enabled || !telemetrySettings.Metrics.Enabled {
			return next
		}

		// Get meter for creating instruments
		meter := provider.Meter("simba")

		// Create metric instruments
		requestDuration, _ := meter.Float64Histogram(
			"http.server.request.duration",
			metric.WithDescription("Duration of HTTP requests in milliseconds"),
			metric.WithUnit("ms"),
		)

		requestCount, _ := meter.Int64Counter(
			"http.server.request.count",
			metric.WithDescription("Total number of HTTP requests"),
		)

		responseSize, _ := meter.Int64Histogram(
			"http.server.response.size",
			metric.WithDescription("Size of HTTP response in bytes"),
			metric.WithUnit("By"),
		)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status and size
			wrappedWriter := &metricsResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(wrappedWriter, r)

			// Calculate duration in milliseconds
			duration := float64(time.Since(start).Milliseconds())

			// Common attributes for all metrics
			attrs := []attribute.KeyValue{
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.Int("http.status_code", wrappedWriter.statusCode),
			}

			// Record metrics
			requestDuration.Record(r.Context(), duration, metric.WithAttributes(attrs...))
			requestCount.Add(r.Context(), 1, metric.WithAttributes(attrs...))

			if wrappedWriter.bytesWritten > 0 {
				responseSize.Record(r.Context(), wrappedWriter.bytesWritten, metric.WithAttributes(attrs...))
			}
		})
	}
}

// metricsResponseWriter wraps http.ResponseWriter to capture status code and bytes written
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *metricsResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Ensure metricsResponseWriter implements http.Flusher if the underlying ResponseWriter does
func (w *metricsResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// String returns the status code as a string
func (w *metricsResponseWriter) Status() string {
	return strconv.Itoa(w.statusCode)
}
