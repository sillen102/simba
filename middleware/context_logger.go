package middleware

import (
	"log/slog"
	"net/http"

	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/simbaContext"
)

type CtxLogger struct {
	Logger *slog.Logger
}

func (c CtxLogger) ContextLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(logging.With(r.Context(), c.Logger.With(
			"method", r.Method,
			"path", r.URL.Path,
			"requestId", r.Context().Value(simbaContext.RequestIDKey),
		))))
	})
}
