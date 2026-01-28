package middleware_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestWebSocketTraceID(t *testing.T) {
	t.Parallel()

	t.Run("generates trace ID and adds to context", func(t *testing.T) {
		t.Parallel()

		mw := middleware.WebSocketTraceID()
		ctx := context.Background()

		// Apply middleware
		newCtx := mw(ctx)

		// Check that trace ID was added
		traceID := newCtx.Value(simbaContext.TraceIDKey)
		assert.NotNil(t, traceID)

		// Verify it's a valid UUID string
		traceIDStr, ok := traceID.(string)
		assert.True(t, ok, "traceID should be a string")
		assert.True(t, traceIDStr != "", "traceID should not be empty")

		// Verify it's a valid UUID
		_, err := uuid.Parse(traceIDStr)
		assert.NoError(t, err, "traceID should be a valid UUID")
	})

	t.Run("generates different trace IDs on each invocation", func(t *testing.T) {
		t.Parallel()

		mw := middleware.WebSocketTraceID()
		ctx := context.Background()

		// Apply middleware multiple times
		ctx1 := mw(ctx)
		ctx2 := mw(ctx)
		ctx3 := mw(ctx)

		// Get trace IDs
		traceID1 := ctx1.Value(simbaContext.TraceIDKey).(string)
		traceID2 := ctx2.Value(simbaContext.TraceIDKey).(string)
		traceID3 := ctx3.Value(simbaContext.TraceIDKey).(string)

		// Verify they're all different
		assert.True(t, traceID1 != traceID2, "trace IDs should be different")
		assert.True(t, traceID2 != traceID3, "trace IDs should be different")
		assert.True(t, traceID1 != traceID3, "trace IDs should be different")
	})

	t.Run("overwrites existing trace ID", func(t *testing.T) {
		t.Parallel()

		mw := middleware.WebSocketTraceID()

		// Create context with existing trace ID
		oldTraceID := "old-trace-id"
		ctx := context.WithValue(context.Background(), simbaContext.TraceIDKey, oldTraceID)

		// Apply middleware
		newCtx := mw(ctx)

		// Check that trace ID was replaced
		newTraceID := newCtx.Value(simbaContext.TraceIDKey).(string)
		assert.True(t, newTraceID != oldTraceID, "trace ID should be replaced")

		// Verify new one is a valid UUID
		_, err := uuid.Parse(newTraceID)
		assert.NoError(t, err, "new traceID should be a valid UUID")
	})

	t.Run("preserves other context values", func(t *testing.T) {
		t.Parallel()

		mw := middleware.WebSocketTraceID()

		// Create context with other values
		type testKey string
		ctx := context.WithValue(context.Background(), testKey("key"), "value")

		// Apply middleware
		newCtx := mw(ctx)

		// Check that other values are preserved
		value := newCtx.Value(testKey("key"))
		assert.Equal(t, "value", value)
	})

	t.Run("preserves connectionID from context", func(t *testing.T) {
		t.Parallel()

		mw := middleware.WebSocketTraceID()

		// Create context with connectionID
		connID := "test-connection-id"
		ctx := context.WithValue(context.Background(), simbaContext.ConnectionIDKey, connID)

		// Apply middleware
		newCtx := mw(ctx)

		// Check that connectionID is preserved
		preservedConnID, ok := newCtx.Value(simbaContext.ConnectionIDKey).(string)
		assert.True(t, ok)
		assert.Equal(t, connID, preservedConnID)

		// Check that traceID was added
		traceID := newCtx.Value(simbaContext.TraceIDKey)
		assert.NotNil(t, traceID)
	})

	t.Run("can be called multiple times for same connection", func(t *testing.T) {
		t.Parallel()

		mw := middleware.WebSocketTraceID()

		// Simulate connection context
		connID := "connection-123"
		baseCtx := context.WithValue(context.Background(), simbaContext.ConnectionIDKey, connID)

		// Simulate multiple message callbacks
		var traceIDs []string
		for i := 0; i < 5; i++ {
			ctx := mw(baseCtx)
			traceID := ctx.Value(simbaContext.TraceIDKey).(string)
			traceIDs = append(traceIDs, traceID)
		}

		// Verify all trace IDs are unique
		seen := make(map[string]bool)
		for _, id := range traceIDs {
			assert.False(t, seen[id], "each trace ID should be unique")
			seen[id] = true
		}

		// Verify all generated valid UUIDs
		for _, id := range traceIDs {
			_, err := uuid.Parse(id)
			assert.NoError(t, err, "each traceID should be a valid UUID")
		}
	})
}
