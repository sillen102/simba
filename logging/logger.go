package logging

import (
	"context"
	"io"
	"sync"
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

var (
	defaultLogger zerolog.Logger
	once          sync.Once
)

func init() {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = TimeFormat
}

func New(config LoggerConfig) zerolog.Logger {
	var logger zerolog.Logger

	if config.Format == JsonFormat {
		logger = zerolog.New(config.Output).Level(config.Level).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(config.Output).Level(config.Level).Output(zerolog.ConsoleWriter{
			Out:          config.Output,
			TimeLocation: time.UTC,
			TimeFormat:   TimeFormat,
		}).With().Timestamp().Logger()
	}
	once.Do(func() {
		defaultLogger = logger
		zerolog.DefaultContextLogger = &logger
	})
	return logger
}

func FromCtx(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		return &defaultLogger
	}
	return logger
}

// GetDefault returns the default logger
func GetDefault() *zerolog.Logger {
	return &defaultLogger
}
