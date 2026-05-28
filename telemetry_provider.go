package simba

import (
	"context"
	"net/http"
)

// TelemetryProvider defines the pluggable interface for telemetry (tracing/metrics) in Simba.
// It is intentionally provider-agnostic: concrete implementations (e.g., OpenTelemetry) live in a subpackage or plugin.
type TelemetryProvider interface {
	// TracingMiddleware returns an http.Handler middleware for request tracing.
	TracingMiddleware() func(http.Handler) http.Handler

	// MetricsMiddleware returns an http.Handler middleware for request metrics.
	MetricsMiddleware() func(http.Handler) http.Handler

	// Shutdown cleanly shuts down the telemetry provider (noop for NoOpTelemetryProvider)
	Shutdown(ctx context.Context) error
}

// NoOpTelemetryProvider implements the TelemetryProvider interface with no-op handlers,
// for use when no telemetry is desired.
type NoOpTelemetryProvider struct{}

func (NoOpTelemetryProvider) TracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}

func (NoOpTelemetryProvider) MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}

func (NoOpTelemetryProvider) Shutdown(ctx context.Context) error {
	return nil
}
