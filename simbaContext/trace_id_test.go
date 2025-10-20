package simbaContext_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestWithExistingOrNewTraceID(t *testing.T) {
	t.Parallel()

	t.Run("generates new trace ID if none exists", func(t *testing.T) {
		ctx := context.Background()
		ctxWithTraceID := simbaContext.WithExistingOrNewTraceID(ctx)

		traceID, ok := ctxWithTraceID.Value(simbaContext.TraceIDKey).(string)
		assert.Assert(t, ok)
		assert.Assert(t, traceID != "")

		// Check if the trace ID is a valid UUID
		_, err := uuid.Parse(traceID)
		assert.NoError(t, err)
	})

	t.Run("generates new trace ID if existing is empty", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), simbaContext.TraceIDKey, "")
		ctxWithTraceID := simbaContext.WithExistingOrNewTraceID(ctx)

		traceID, ok := ctxWithTraceID.Value(simbaContext.TraceIDKey).(string)
		assert.Assert(t, ok)
		assert.Assert(t, traceID != "")

		// Check if the trace ID is a valid UUID
		_, err := uuid.Parse(traceID)
		assert.NoError(t, err)
	})

	t.Run("reuses existing trace ID", func(t *testing.T) {
		existingTraceID := "existing-trace-id"
		ctx := context.WithValue(context.Background(), simbaContext.TraceIDKey, existingTraceID)
		ctxWithTraceID := simbaContext.WithExistingOrNewTraceID(ctx)

		traceID, ok := ctxWithTraceID.Value(simbaContext.TraceIDKey).(string)
		assert.Assert(t, ok)
		assert.Equal(t, existingTraceID, traceID)
	})
}

func TestWithTraceID(t *testing.T) {
	t.Parallel()

	t.Run("sets provided trace ID", func(t *testing.T) {
		providedTraceID := "provided-trace-id"
		ctx := context.Background()
		ctxWithTraceID := simbaContext.WithTraceID(ctx, providedTraceID)

		traceID, ok := ctxWithTraceID.Value(simbaContext.TraceIDKey).(string)
		assert.Assert(t, ok)
		assert.Equal(t, providedTraceID, traceID)
	})
}

func TestGetTraceID(t *testing.T) {
	t.Parallel()

	t.Run("returns empty string if no trace ID", func(t *testing.T) {
		ctx := context.Background()
		traceID := simbaContext.GetTraceID(ctx)
		assert.Equal(t, "", traceID)
	})

	t.Run("returns existing trace ID", func(t *testing.T) {
		existingTraceID := "existing-trace-id"
		ctx := context.WithValue(context.Background(), simbaContext.TraceIDKey, existingTraceID)
		traceID := simbaContext.GetTraceID(ctx)
		assert.Equal(t, existingTraceID, traceID)
	})
}
