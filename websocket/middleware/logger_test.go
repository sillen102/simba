package middleware_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaTest/assert"
	"github.com/sillen102/simba/websocket/middleware"
)

func TestWebSocketLogger(t *testing.T) {
	t.Parallel()

	t.Run("injects logger into context", func(t *testing.T) {
		mw := middleware.Logger()
		ctx := context.Background()

		// Apply middleware
		newCtx := mw(ctx)

		// Check that logger was added
		logger := newCtx.Value(simbaContext.LoggerKey)
		assert.NotNil(t, logger)

		// Verify it's a *slog.Logger
		_, ok := logger.(*slog.Logger)
		assert.True(t, ok, "logger should be *slog.Logger")
	})

	t.Run("logger includes connectionID when present", func(t *testing.T) {
		// Set up logger to capture output
		var buf bytes.Buffer
		testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		slog.SetDefault(testLogger)

		mw := middleware.Logger()
		connID := "conn-12345"
		ctx := context.WithValue(context.Background(), simbaContext.ConnectionIDKey, connID)

		// Apply middleware
		newCtx := mw(ctx)

		// Get logger and log a message
		logger := newCtx.Value(simbaContext.LoggerKey).(*slog.Logger)
		logger.Info("test message")

		// Check output includes connectionID
		output := buf.String()
		assert.True(t, strings.Contains(output, "connectionId="+connID), "log should include connectionId")
		assert.True(t, strings.Contains(output, "test message"), "log should include message")
	})

	t.Run("logger includes traceID when present", func(t *testing.T) {
		// Set up logger to capture output
		var buf bytes.Buffer
		testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		slog.SetDefault(testLogger)

		mw := middleware.Logger()
		traceID := "trace-67890"
		ctx := context.WithValue(context.Background(), simbaContext.TraceIDKey, traceID)

		// Apply middleware
		newCtx := mw(ctx)

		// Get logger and log a message
		logger := newCtx.Value(simbaContext.LoggerKey).(*slog.Logger)
		logger.Info("test message")

		// Check output includes traceID
		output := buf.String()
		assert.True(t, strings.Contains(output, "traceId="+traceID), "log should include traceId")
		assert.True(t, strings.Contains(output, "test message"), "log should include message")
	})

	t.Run("logger includes both connectionID and traceID", func(t *testing.T) {
		// Set up logger to capture output
		var buf bytes.Buffer
		testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		slog.SetDefault(testLogger)

		mw := middleware.Logger()
		connID := "conn-12345"
		traceID := "trace-67890"
		ctx := context.WithValue(context.Background(), simbaContext.ConnectionIDKey, connID)
		ctx = context.WithValue(ctx, simbaContext.TraceIDKey, traceID)

		// Apply middleware
		newCtx := mw(ctx)

		// Get logger and log a message
		logger := newCtx.Value(simbaContext.LoggerKey).(*slog.Logger)
		logger.Info("test message")

		// Check output includes both IDs
		output := buf.String()
		assert.True(t, strings.Contains(output, "connectionId="+connID), "log should include connectionId")
		assert.True(t, strings.Contains(output, "traceId="+traceID), "log should include traceId")
		assert.True(t, strings.Contains(output, "test message"), "log should include message")
	})

	t.Run("works without connectionID or traceID", func(t *testing.T) {
		mw := middleware.Logger()
		ctx := context.Background()

		// Apply middleware
		newCtx := mw(ctx)

		// Should still add a logger
		logger := newCtx.Value(simbaContext.LoggerKey)
		assert.NotNil(t, logger)

		// Should be able to use it
		loggerTyped := logger.(*slog.Logger)
		loggerTyped.Info("test message") // Should not panic
	})

	t.Run("handles empty string connectionID", func(t *testing.T) {
		// Set up logger to capture output
		var buf bytes.Buffer
		testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		slog.SetDefault(testLogger)

		mw := middleware.Logger()
		ctx := context.WithValue(context.Background(), simbaContext.ConnectionIDKey, "")

		// Apply middleware
		newCtx := mw(ctx)

		// Get logger and log a message
		logger := newCtx.Value(simbaContext.LoggerKey).(*slog.Logger)
		logger.Info("test message")

		// Empty connectionID should not be added to logs
		output := buf.String()
		assert.False(t, strings.Contains(output, "connectionId="), "empty connectionId should not be added")
	})

	t.Run("handles empty string traceID", func(t *testing.T) {
		// Set up logger to capture output
		var buf bytes.Buffer
		testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		slog.SetDefault(testLogger)

		mw := middleware.Logger()
		ctx := context.WithValue(context.Background(), simbaContext.TraceIDKey, "")

		// Apply middleware
		newCtx := mw(ctx)

		// Get logger and log a message
		logger := newCtx.Value(simbaContext.LoggerKey).(*slog.Logger)
		logger.Info("test message")

		// Empty traceID should not be added to logs
		output := buf.String()
		assert.False(t, strings.Contains(output, "traceId="), "empty traceId should not be added")
	})

	t.Run("handles non-string connectionID gracefully", func(t *testing.T) {
		mw := middleware.Logger()
		ctx := context.WithValue(context.Background(), simbaContext.ConnectionIDKey, 12345)

		// Apply middleware - should not panic
		newCtx := mw(ctx)

		// Should still add a logger
		logger := newCtx.Value(simbaContext.LoggerKey)
		assert.NotNil(t, logger)
	})

	t.Run("handles non-string traceID gracefully", func(t *testing.T) {
		mw := middleware.Logger()
		ctx := context.WithValue(context.Background(), simbaContext.TraceIDKey, 67890)

		// Apply middleware - should not panic
		newCtx := mw(ctx)

		// Should still add a logger
		logger := newCtx.Value(simbaContext.LoggerKey)
		assert.NotNil(t, logger)
	})

	t.Run("preserves other context values", func(t *testing.T) {
		mw := middleware.Logger()

		// Create context with other values
		type testKey string
		ctx := context.WithValue(context.Background(), testKey("key"), "value")

		// Apply middleware
		newCtx := mw(ctx)

		// Check that other values are preserved
		value := newCtx.Value(testKey("key"))
		assert.Equal(t, "value", value)
	})

	t.Run("works in middleware chain", func(t *testing.T) {
		// Set up logger to capture output
		var buf bytes.Buffer
		testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		slog.SetDefault(testLogger)

		// Simulate middleware chain: TraceID -> Logger
		traceIDMw := middleware.TraceID()
		loggerMw := middleware.Logger()

		// Start with context containing connectionID
		connID := "conn-xyz"
		ctx := context.WithValue(context.Background(), simbaContext.ConnectionIDKey, connID)

		// Apply middleware chain
		ctx = traceIDMw(ctx)
		ctx = loggerMw(ctx)

		// Get logger and log
		logger := ctx.Value(simbaContext.LoggerKey).(*slog.Logger)
		logger.Info("middleware chain test")

		// Should have both connectionID and traceID
		output := buf.String()
		assert.True(t, strings.Contains(output, "connectionId="+connID), "should have connectionId")
		assert.True(t, strings.Contains(output, "traceId="), "should have traceId")
		assert.True(t, strings.Contains(output, "middleware chain test"), "should have message")
	})
}
