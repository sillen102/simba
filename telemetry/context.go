package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan creates a new span with the given name and options
// This is a convenience wrapper around the OTEL trace API
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("simba").Start(ctx, name, opts...)
}

// GetTracer returns the global tracer for Simba
func GetTracer(ctx context.Context) trace.Tracer {
	return otel.Tracer("simba")
}

// GetMeter returns the global meter for Simba
func GetMeter(ctx context.Context) metric.Meter {
	return otel.Meter("simba")
}
