package logging

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type LogFormat string

// defaultLogger is a global logger that is used as a fallback if no logger is provided in the context or there is no context
var (
	defaultLoggerOnce sync.Once
	defaultLogger     zerolog.Logger
)

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

// Init sets the default logger
func Init(config LoggerConfig) {
	defaultLoggerOnce.Do(func() {
		defaultLogger = New(config)
	})
}

// GetDefault returns the default logger
func GetDefault() *zerolog.Logger {
	return &defaultLogger
}

// New returns a new logger
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
	zerolog.DefaultContextLogger = &logger
	return logger
}

// FromCtx returns a logger from the context
func FromCtx(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		return &defaultLogger
	}
	return logger
}
