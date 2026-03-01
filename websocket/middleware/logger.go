package middleware

import (
	"context"
	"log/slog"

	"github.com/sillen102/simba/simbaContext"
)

// Logger injects a logger with connectionID and traceID into the context.
// This ensures all log statements within WebSocket callbacks automatically include
// both IDs for proper correlation.
//
// Returns a WebSocketMiddleware function that can be used with websocket.WithMiddleware().
func Logger() func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		// Get connectionID and traceID from context
		connID := ctx.Value(simbaContext.ConnectionIDKey)
		traceID := ctx.Value(simbaContext.TraceIDKey)

		// Create logger with both IDs
		logger := slog.Default()

		if connIDStr, ok := connID.(string); ok && connIDStr != "" {
			logger = logger.With("connectionId", connIDStr)
		}

		if traceIDStr, ok := traceID.(string); ok && traceIDStr != "" {
			logger = logger.With("traceId", traceIDStr)
		}

		// Add logger to context
		return context.WithValue(ctx, simbaContext.LoggerKey, logger)
	}
}
