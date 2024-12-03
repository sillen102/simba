package logging

import (
	"context"
	"io"
	"time"

	"github.com/rs/zerolog"
)

type LogFormat string

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

	switch config.Format {
	case JsonFormat:
		logger = zerolog.New(config.Output).Level(config.Level).With().Timestamp().Logger()
	case TextFormat:
		logger = zerolog.New(config.Output).Level(config.Level).Output(zerolog.ConsoleWriter{
			Out:          config.Output,
			TimeLocation: time.UTC,
			TimeFormat:   TimeFormat,
		}).With().Timestamp().Logger()
	default:
		logger = zerolog.New(config.Output).With().Timestamp().Logger()
	}
	zerolog.DefaultContextLogger = &logger

	return &logger
}

// Get returns the global logger
func Get() *zerolog.Logger {
	return zerolog.Ctx(context.Background())
}

// FromCtx returns a logger from the context
func FromCtx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
