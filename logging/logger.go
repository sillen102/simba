package logging

import (
	"context"
	"log/slog"

	"github.com/sillen102/simba/simbaContext"
)

// From returns the logger from the context.
// Returns a new logger if no logger is found in the context.
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(simbaContext.LoggerKey).(*slog.Logger); ok {
		return l
	} else {
		return slog.Default()
	}
}
