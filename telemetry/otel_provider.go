package telemetry

import (
	"context"
	"net/http"

	"github.com/sillen102/simba"
	config "github.com/sillen102/simba/telemetry/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"time"
)

// OtelTelemetryProvider implements simba.TelemetryProvider using OpenTelemetry SDK
// (wraps a full OTel Provider instance for tracing/metrics).
type OtelTelemetryProvider struct {
	provider        *Provider
	telemetryConfig *config.TelemetryConfig
}

func NewOtelTelemetryProvider(ctx context.Context, cfg *config.TelemetryConfig) (simba.TelemetryProvider, error) {
	if cfg == nil || !cfg.Enabled {
		return simba.NoOpTelemetryProvider{}, nil
	}
	prov, err := NewProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &OtelTelemetryProvider{provider: prov, telemetryConfig: cfg}, nil
}

// TracingMiddleware injects OTel tracing handler
func (o *OtelTelemetryProvider) TracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if o.provider == nil || !o.telemetryConfig.Enabled || !o.telemetryConfig.Tracing.Enabled {
			return next
		}
		return otelhttp.NewHandler(next, "simba.http.server",
			otelhttp.WithTracerProvider(o.provider.TracerProvider()),
		)
	}
}

// MetricsMiddleware injects OTel metrics handler
func (o *OtelTelemetryProvider) MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if o.provider == nil || !o.telemetryConfig.Enabled || !o.telemetryConfig.Metrics.Enabled {
			return next
		}
		meter := o.provider.Meter("simba")
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
			wrappedWriter := &metricsResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(wrappedWriter, r)
			duration := float64(time.Since(start).Milliseconds())
			attrs := []attribute.KeyValue{
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.Int("http.status_code", wrappedWriter.statusCode),
			}
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

// Shutdown delegates to the underlying Otel Provider shutdown
func (o *OtelTelemetryProvider) Shutdown(ctx context.Context) error {
	if o.provider != nil {
		return o.provider.Shutdown(ctx)
	}
	return nil
}

// Provider exposes the underlying OTel Provider (for custom metrics/tracing).
func (o *OtelTelemetryProvider) Provider() *Provider {
	return o.provider
}
