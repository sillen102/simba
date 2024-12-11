package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/sillen102/simba/simbaContext"
)

const (
	TimeFormat string = "2006-01-02T15:04:05.000000"
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

// NewLogger returns a new [slog.Logger] with the provided settings.
func NewLogger(config ...Config) *slog.Logger {
	return newLogger(config...)
}

// With returns a new context with the logger added to it.
func With(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, simbaContext.LoggerKey, logger)
}

// From returns the logger from the context.
// Returns a new logger if no logger is found in the context.
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(simbaContext.LoggerKey).(*slog.Logger); ok {
		return l
	} else {
		return slog.Default()
	}
}

func newLogger(provided ...Config) *slog.Logger {
	var config Config
	if len(provided) > 0 {
		config = provided[0]
	}

	if config.Level == 0 {
		config.Level = slog.LevelInfo
	}
	if config.Format == "" {
		config.Format = TextFormat
	}
	if config.Output == nil {
		config.Output = os.Stdout
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

	return slog.New(handler)
}

// MustParseLogLevel parses a string into a slog.Level and panics if it fails
func MustParseLogLevel(levelStr string) slog.Level {
	level, err := ParseLogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	return level
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

// MustParseLogFormat parses a string into a LogFormat and panics if it fails
func MustParseLogFormat(formatStr string) LogFormat {
	format, err := ParseLogFormat(formatStr)
	if err != nil {
		panic(err)
	}
	return format
}

// ParseLogFormat parses a string into a LogFormat
func ParseLogFormat(formatStr string) (LogFormat, error) {
	switch formatStr {
	case "text":
		return TextFormat, nil
	case "json":
		return JsonFormat, nil
	default:
		return TextFormat, errors.New("invalid log format")
	}
}
