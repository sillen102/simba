package logging

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	defaultLogger zerolog.Logger
	once          sync.Once
)

// Init initializes the logger for the application with the provided config
func Init(config Config) {
	if config.Output == nil {
		config.Output = getOutput(config.output)
	}

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
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().UTC()
		}
		zerolog.TimeFieldFormat = TimeFormat
		defaultLogger = logger
		zerolog.DefaultContextLogger = &defaultLogger
	})
}

// With returns a new context with the provided logger or with the default logger if no logger is provided
func With(ctx context.Context, logger ...zerolog.Logger) context.Context {
	if len(logger) == 0 {
		return defaultLogger.WithContext(ctx)
	}
	return logger[0].WithContext(ctx)
}

// Get returns the logger from the context
// Returns the default logger if no logger is found in the context or no context is provided
func Get(ctx ...context.Context) *zerolog.Logger {
	if len(ctx) == 0 {
		return &defaultLogger
	}

	logger := zerolog.Ctx(ctx[0])
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		return &defaultLogger
	}

	return logger
}
