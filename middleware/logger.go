package middleware

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func getLogger(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)

	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		l := zerolog.New(os.Stdout).Level(zerolog.GlobalLevel()).Output(zerolog.ConsoleWriter{
			Out:          os.Stdout,
			TimeLocation: time.UTC,
			TimeFormat:   "2006-01-02T15:04:05.000000",
		}).With().Timestamp().Logger()
		logger = &l

		logger.Warn().Msg("no logger found in context, using default logger")
	}

	return logger
}
