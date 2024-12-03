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
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000000"
	switch config.Format {
	case JsonFormat:
		logger = zerolog.New(config.Output).Level(config.Level).With().Timestamp().Logger()
	case TextFormat:
		logger = zerolog.New(config.Output).Level(config.Level).Output(zerolog.ConsoleWriter{
			Out: config.Output,
		}).With().Timestamp().Logger()
	default:
		logger = zerolog.New(config.Output).With().Timestamp().Logger()
	}
	return &logger
}

// FromCtx returns a logger from the context
func FromCtx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
