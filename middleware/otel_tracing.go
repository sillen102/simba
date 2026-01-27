package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/telemetry"
)

// OtelTracing returns middleware that adds OpenTelemetry tracing to HTTP requests.
// It returns a pass-through handler if telemetry is disabled for zero overhead.
func OtelTracing(provider *telemetry.Provider, telemetrySettings *settings.Telemetry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// If telemetry is disabled or tracing is disabled, return pass-through
		if provider == nil || !telemetrySettings.Enabled || !telemetrySettings.Tracing.Enabled {
			return next
		}

		// Use otelhttp for automatic instrumentation with W3C trace context propagation
		return otelhttp.NewHandler(next, "simba.http.server",
			otelhttp.WithTracerProvider(provider.TracerProvider()),
		)
	}
}
