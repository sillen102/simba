package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/sillen102/simba/simbaContext"
)

// WebSocketTraceID generates a fresh trace ID for each WebSocket callback invocation.
// This allows each message, connection event, and disconnect to be traced independently
// while maintaining the same connectionID throughout the connection lifecycle.
//
// Returns a WebSocketMiddleware function that can be used with simba.WithMiddleware().
func WebSocketTraceID() func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		// Generate new trace ID
		id, err := uuid.NewV7()
		if err != nil || id == uuid.Nil {
			id = uuid.New()
		}

		// Add trace ID to context
		return context.WithValue(ctx, simbaContext.TraceIDKey, id.String())
	}
}
