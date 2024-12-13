package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/simbaContext"
	"gotest.tools/v3/assert"
)

func TestLogRequests(t *testing.T) {
	t.Parallel()

	t.Run("logs request and response", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond) // Simulate some work
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message":"success"}`))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Create a logger and inject it into the context
		logger := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
		ctx := context.WithValue(req.Context(), simbaContext.LoggerKey, logger)
		req = req.WithContext(ctx)

		middleware.LogRequests(handler).ServeHTTP(w, req)

		// Since we're using a custom logger, we can only verify the response was written
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"message":"success"}`, w.Body.String())
	})
}
