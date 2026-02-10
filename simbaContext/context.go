package simbaContext

import (
	"context"
	"log/slog"
)

// ContextCopier builds a new context by copying selected values from a source context.
type ContextCopier struct {
	src context.Context
	dst context.Context
}

// NewContextCopier creates a new context copier starting with a background context.
func NewContextCopier(src context.Context) *ContextCopier {
	return &ContextCopier{
		src: src,
		dst: context.Background(),
	}
}

// WithTraceID copies the trace ID from the source context.
func (c *ContextCopier) WithTraceID() *ContextCopier {
	if traceID := GetTraceID(c.src); traceID != "" {
		c.dst = WithTraceID(c.dst, traceID)
	}
	return c
}

// WithLogger copies the logger from the source context.
func (c *ContextCopier) WithLogger() *ContextCopier {
	if logger, ok := c.src.Value(LoggerKey).(*slog.Logger); ok && logger != nil {
		c.dst = context.WithValue(c.dst, LoggerKey, logger)
	}
	return c
}

// WithValue copies a specific value from the source context if it exists.
func (c *ContextCopier) WithValue(key any) *ContextCopier {
	if value := c.src.Value(key); value != nil {
		c.dst = context.WithValue(c.dst, key, value)
	}
	return c
}

// Build finalizes the context copying and returns the new context.
func (c *ContextCopier) Build() context.Context {
	return c.dst
}

// CopyDefault creates a new context by copying the values added by simba by default, such as trace ID and logger.
func CopyDefault(src context.Context) context.Context {
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
