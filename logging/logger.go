package logging

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	// RequestIDKey is the key used to store the request ID in the context
	LoggerKey = "logger"
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
	slog.SetDefault(slog.New(getHandler(cfg)))
}

// Get returns the default logger.
func Get() *slog.Logger {
	return slog.Default()
}

// FromCtx returns a logger from the context.
func FromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// getHandler returns a slog.Handler based on the configuration.
func getHandler(cfg LoggerConfig) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: getLevel(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(a.Key, a.Value.Time().UTC())
			}
			return a
		},
	}

	if cfg.Format == "json" {
		return slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.NewTextHandler(os.Stdout, opts)
}

// getLevel returns the slog.Level based on the configuration.
func getLevel() slog.Level {
	switch cfg.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
