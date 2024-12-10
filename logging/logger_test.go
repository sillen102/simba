package logging_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/sillen102/simba/logging"
)

func TestNewLogger_DefaultConfig(t *testing.T) {
	logger := logging.NewLogger()
	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
}

func TestNewLogger_CustomConfig(t *testing.T) {
	var buf bytes.Buffer
	config := logging.Config{
		Level:  slog.LevelDebug,
		Format: logging.JsonFormat,
		Output: &buf,
	}
	logger := logging.NewLogger(config)
	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}

	// Log a test message
	logger.Debug("Test message")

	// Check if the buffer contains the logged message
	if !bytes.Contains(buf.Bytes(), []byte("Test message")) {
		t.Fatal("Expected buffer to contain logged message, but it did not")
	}
}

func TestWithAndFrom(t *testing.T) {
	logger := logging.NewLogger()
	ctx := logging.With(context.Background(), logger)
	retrievedLogger := logging.From(ctx)
	if retrievedLogger != logger {
		t.Fatal("Expected to retrieve the same logger from context")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
	}

	for _, test := range tests {
		level, err := logging.ParseLogLevel(test.input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if level != test.expected {
			t.Fatalf("Expected %v, got %v", test.expected, level)
		}
	}
}

func TestParseLogFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected logging.LogFormat
	}{
		{"text", logging.TextFormat},
		{"json", logging.JsonFormat},
	}

	for _, test := range tests {
		format, err := logging.ParseLogFormat(test.input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if format != test.expected {
			t.Fatalf("Expected %v, got %v", test.expected, format)
		}
	}
}
