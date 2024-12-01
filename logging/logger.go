package logging

import (
	"context"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
)

// LoggerConfig is the configuration for the logger
type LoggerConfig struct {
	Format string `env:"LOGGING_FORMAT" env-default:"text"`
	Level  string `env:"LOGGING_LEVEL" env-default:"debug"`
}

var cfg LoggerConfig

func init() {
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		log.Fatalf("failed to parse logging level: %v", err)
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999999"
	zerolog.DefaultContextLogger = getLogger(cfg)
}

// Get returns the default logger
func Get() *zerolog.Logger {
	return zerolog.DefaultContextLogger
}

// WithLogger returns a context with the given logger
func WithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

// FromCtx returns a logger from the context
func FromCtx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// getLogger returns a [zerolog.Logger] based on the configuration
func getLogger(cfg LoggerConfig) *zerolog.Logger {
	var logger zerolog.Logger
	switch cfg.Format {
	case "json":
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	default:
		logger = zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	}
	return &logger
}
