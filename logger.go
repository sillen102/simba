package simba

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	once sync.Once
)

func init() {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = TimeFormat
}

// New creates a completely new logger instance for each request
func NewLogger(cfg *LoggingConfig) zerolog.Logger {
	if cfg == nil {
		// Default config
		cfg = &LoggingConfig{
			Level:  zerolog.InfoLevel,
			Format: TextFormat,
			output: Stdout,
			Output: os.Stdout,
		}
	}

	if cfg.Output == nil {
		cfg.Output = getOutput(cfg.output)
	}

	var logger zerolog.Logger
	if cfg.Format == JsonFormat {
		logger = zerolog.New(cfg.Output).Level(cfg.Level).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(cfg.Output).Level(cfg.Level).Output(zerolog.ConsoleWriter{
			Out:          cfg.Output,
			TimeLocation: time.UTC,
			TimeFormat:   TimeFormat,
		}).With().Timestamp().Logger()
	}

	once.Do(func() {
		zerolog.DefaultContextLogger = &logger
		zerolog.SetGlobalLevel(cfg.Level)
	})

	return logger
}

// With returns a new context with the provided logger or with the default logger if no logger is provided
func (a *Application[AuthModel]) With(ctx context.Context) context.Context {
	return a.logger.WithContext(ctx)
}

// GetLogger returns the default Simba logger
func (a *Application[AuthModel]) GetLogger() *zerolog.Logger {
	return &a.logger
}

// LoggerFrom returns the logger from the context.
// Returns a new logger if no logger is found in the context.
func LoggerFrom(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)

	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		l := zerolog.New(os.Stdout).Level(zerolog.GlobalLevel()).Output(zerolog.ConsoleWriter{
			Out:          os.Stdout,
			TimeLocation: time.UTC,
			TimeFormat:   TimeFormat,
		}).With().Timestamp().Logger()
		logger = &l

		logger.Warn().Msg("no logger found in context, using default logger")
	}

	return logger
}
