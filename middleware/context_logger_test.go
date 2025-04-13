package middleware_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaTestAssert"
)

func TestContextLogger(t *testing.T) {
	t.Parallel()

	t.Run("adds logger to context", func(t *testing.T) {
		var buf bytes.Buffer
		log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
		contextLogger := middleware.Logger{Logger: log}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxLogger := logging.From(r.Context())
			simbaTestAssert.Assert(t, ctxLogger != nil)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		contextLogger.ContextLogger(handler).ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logs request details", func(t *testing.T) {
		var buf bytes.Buffer
		log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
		contextLogger := middleware.Logger{Logger: log}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logging.From(r.Context()).Info("test log")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.WithValue(req.Context(), simbaContext.RequestIDKey, "12345"))
		w := httptest.NewRecorder()

		contextLogger.ContextLogger(handler).ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusOK, w.Code)
		logOutput := buf.String()
		simbaTestAssert.Assert(t, bytes.Contains([]byte(logOutput), []byte(`"test log"`)))
	})
}
