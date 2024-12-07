package logging

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	once          sync.Once
	config        *Config
	defaultLogger zerolog.Logger
)

func Init(cfg *Config) {
	config = cfg
	cfg.Output = getOutput(cfg.output)
	defaultLogger = New(cfg)
}

// New creates a completely new logger instance for each request
func New(cfg *Config) zerolog.Logger {
	if cfg == nil {
		// Default config
		cfg = &Config{
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
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().UTC()
		}
		zerolog.TimeFieldFormat = TimeFormat
		zerolog.DefaultContextLogger = &logger
		zerolog.SetGlobalLevel(cfg.Level)
	})

	return logger
}

// CtxWith returns a new context with the provided logger or with the default logger if no logger is provided
func CtxWith(ctx context.Context, logger zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

// Get returns the logger from the context
// Returns a new logger if no logger is found in the context or no context is provided
func Get(ctx ...context.Context) *zerolog.Logger {
	var logger *zerolog.Logger
	if len(ctx) == 0 {
		logger = &defaultLogger
	} else {
		logger = zerolog.Ctx(ctx[0])
	}

	// If logger is nil create a new one
	if logger == nil {
		l := New(config)
		logger = &l
	}

	return logger
}
