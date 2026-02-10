package simbaContext_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/sillen102/simba/simbaContext"
)

func TestCopyContextValues(t *testing.T) {
	t.Run("copies trace ID and logger", func(t *testing.T) {
		// Create source context with trace ID and logger
		logger := slog.Default()
		src := simbaContext.WithTraceID(context.Background(), "test-trace-123")
		src = context.WithValue(src, simbaContext.LoggerKey, logger)

		// Copy the context
		copied := simbaContext.CopyContextValues(src)

		// Verify trace ID was copied
		if got := simbaContext.GetTraceID(copied); got != "test-trace-123" {
			t.Errorf("GetTraceID() = %q, want %q", got, "test-trace-123")
		}

		// Verify logger was copied
		if got, ok := copied.Value(simbaContext.LoggerKey).(*slog.Logger); !ok || got != logger {
			t.Errorf("logger was not copied correctly")
		}
	})

	t.Run("returns background context when source has no values", func(t *testing.T) {
		src := context.Background()
		copied := simbaContext.CopyContextValues(src)

		// Verify trace ID is empty
		if got := simbaContext.GetTraceID(copied); got != "" {
			t.Errorf("GetTraceID() = %q, want empty string", got)
		}
	})
}
