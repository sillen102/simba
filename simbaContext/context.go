package simbaContext

import (
	"context"
	"log/slog"
)

// CopyContextValues creates a new background context with all values from the source context.
// This is useful for spawning goroutines that should outlive the parent context's cancellation
// but need to retain values like trace IDs.
func CopyContextValues(src context.Context) context.Context {
	ctx := context.Background()

	// Copy trace ID if present
	if traceID := GetTraceID(src); traceID != "" {
		ctx = WithTraceID(ctx, traceID)
	}

	// Copy logger if present
	if logger, ok := src.Value(LoggerKey).(*slog.Logger); ok && logger != nil {
		ctx = context.WithValue(ctx, LoggerKey, logger)
	}

	return ctx
}
