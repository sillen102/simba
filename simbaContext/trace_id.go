package simbaContext

import (
	"context"

	"github.com/google/uuid"
)

// WithExistingOrNewTraceID returns a context with a trace ID. If the context already has a trace ID, it is reused.
func WithExistingOrNewTraceID(ctx context.Context) context.Context {
	var traceID string
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if !ok || traceID == "" {
		id, err := uuid.NewV7()
		if err != nil || id == uuid.Nil {
			traceID = uuid.NewString()
		} else {
			traceID = id.String()
		}
	}
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithTraceID returns a context with the provided trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID retrieves the trace ID from the context. If no trace ID is present, it returns an empty string.
func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if !ok {
		return ""
	}
	return traceID
}
