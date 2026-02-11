package simbaContext

import (
	"context"
	"log/slog"
	"time"
)

// ContextCopier builds a new context by copying selected values from a source context.
type ContextCopier struct {
	src     context.Context
	dst     context.Context
	timeout *time.Duration
}

// NewContextCopier creates a new context copier starting with a background context.
func NewContextCopier(src context.Context) *ContextCopier {
	return &ContextCopier{
		src: src,
		dst: context.Background(),
	}
}

// WithTimeout sets a timeout for the context.
func (c *ContextCopier) WithTimeout(timeout time.Duration) *ContextCopier {
	c.timeout = &timeout
	return c
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
// If a timeout was set, returns a context with deadline that will automatically cancel after the timeout.
func (c *ContextCopier) Build() context.Context {
	if c.timeout != nil {
		var cancel context.CancelFunc
		c.dst, cancel = context.WithTimeout(c.dst, *c.timeout)
		_ = cancel
	}
	return c.dst
}

// CopyDefault creates a new context by copying the values added by simba by default, such as trace ID and logger.
// If a timeout is provided, the returned context will automatically cancel after the specified duration.
func CopyDefault(src context.Context, timeout ...time.Duration) context.Context {
	ctx := context.Background()

	// Copy trace ID if present
	if traceID := GetTraceID(src); traceID != "" {
		ctx = WithTraceID(ctx, traceID)
	}

	// Copy logger if present
	if logger, ok := src.Value(LoggerKey).(*slog.Logger); ok && logger != nil {
		ctx = context.WithValue(ctx, LoggerKey, logger)
	}

	// If a timeout is provided, create a context with deadline
	if len(timeout) > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout[0])
		defer cancel()
	}

	return ctx
}
