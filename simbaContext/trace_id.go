package simbaContext

import (
	"context"
	"log/slog"

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

// WithTraceIDLogger returns a context with a logger that includes the trace ID.
func WithTraceIDLogger(ctx context.Context) context.Context {
	logger := slog.Default().With(
		"traceId", GetTraceID(ctx),
	)
	return context.WithValue(ctx, LoggerKey, logger)
}

// WithTraceIDAndLogger returns a context with a trace ID (existing or new) and a logger that includes the trace ID.
func WithTraceIDAndLogger(ctx context.Context) context.Context {
	ctx = WithExistingOrNewTraceID(ctx)
	return WithTraceIDLogger(ctx)
}
