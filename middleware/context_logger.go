package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/sillen102/simba/simbaContext"
)

type Logger struct {
	Logger *slog.Logger
}

func (c Logger) ContextLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := context.WithValue(r.Context(), simbaContext.LoggerKey, c.Logger.With(
			"method", r.Method,
			"path", r.URL.Path,
			"traceId", r.Context().Value(simbaContext.TraceIDKey),
		))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
