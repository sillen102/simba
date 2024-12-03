package logging

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type LogFormat string

// defaultLogger is a global logger that is used only if no logger is provided in the context
var defaultLogger zerolog.Logger

func init() {
	defaultLogger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{
		Out:          os.Stderr,
		TimeLocation: time.UTC,
		TimeFormat:   TimeFormat,
	}).With().Timestamp().Logger()
}

const (
	JsonFormat LogFormat = "json"
	TextFormat LogFormat = "text"
	TimeFormat string    = "2006-01-02T15:04:05.000000"
)

type LoggerConfig struct {
	Format LogFormat
	Level  zerolog.Level
	Output io.Writer
}

// New returns a new logger
func New(config LoggerConfig) *zerolog.Logger {
	var logger zerolog.Logger
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = TimeFormat
	zerolog.SetGlobalLevel(config.Level)

	if config.Format == JsonFormat {
		logger = zerolog.New(config.Output).Level(config.Level).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(config.Output).Level(config.Level).Output(zerolog.ConsoleWriter{
			Out:          config.Output,
			TimeLocation: time.UTC,
			TimeFormat:   TimeFormat,
		}).With().Timestamp().Logger()
	}
	zerolog.DefaultContextLogger = &logger
	defaultLogger = logger
	return &logger
}

// Get returns the global logger
func Get() *zerolog.Logger {
	return zerolog.Ctx(context.Background())
}

// FromCtx returns a logger from the context
func FromCtx(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		// Return the default logger instead of a no-op logger
		return &defaultLogger
	}
	return logger
}
