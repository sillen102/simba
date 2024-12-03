package logging

import (
	"context"
	"os"

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
}

// New returns a new logger
func New(config LoggerConfig) *zerolog.Logger {
	var logger zerolog.Logger
	switch config.Format {
	case JsonFormat:
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	default:
		logger = zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	}
	return &logger
}

// FromCtx returns a logger from the context
func FromCtx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
