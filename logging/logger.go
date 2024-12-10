package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
)

type loggerKey string

const (
	TimeFormat string    = "2006-01-02T15:04:05.000000"
	LoggerKey  loggerKey = "logger"
)

// Config holds the settings for the logger
type Config struct {

	// Level is the log Level for the logger that will be used
	Level slog.Level `default:"info"`

	// Format is the log Format for the logger that will be used
	Format LogFormat `default:"text"`

	// Output is the output writer that the logger will use
	Output io.Writer
}

type LogFormat string

const (
	JsonFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

func NewLogger(provided ...Config) *slog.Logger {
	var config Config
	if len(provided) > 0 {
		config = provided[0]
	} else {
		// Default logger settings
		config = Config{
			Level:  slog.LevelInfo,
			Format: TextFormat,
			Output: os.Stdout,
		}
	}

	opts := slog.HandlerOptions{
		Level: config.Level,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				attr.Value = slog.StringValue(attr.Value.Time().UTC().Format(TimeFormat))
			}
			return attr
		},
	}

	var handler slog.Handler
	switch config.Format {
	case TextFormat:
		handler = slog.NewTextHandler(config.Output, &opts)
	case JsonFormat:
		handler = slog.NewJSONHandler(config.Output, &opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// ParseLogLevel parses a string into a slog.Level
func ParseLogLevel(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New("invalid log level")
	}
}

// With returns a new context with the logger added to it.
func With(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// From returns the logger from the context.
// Returns a new logger if no logger is found in the context.
func From(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	} else {
		return slog.Default()
	}
}
