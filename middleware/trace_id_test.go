package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	t.Run("generates new trace ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Context().Value(simbaContext.TraceIDKey).(string)
			assert.Assert(t, traceID != "")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.TraceID(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Assert(t, w.Header().Get(simbaContext.TraceIDHeader) != "")

		// Check if the request ID is a valid UUID
		_, err := uuid.Parse(w.Header().Get(simbaContext.TraceIDHeader))
		assert.NoError(t, err)
	})

	t.Run("accepts trace ID from header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Context().Value(simbaContext.TraceIDKey).(string)
			assert.Equal(t, "test-trace-id", traceID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(simbaContext.TraceIDHeader, "test-trace-id")
		req = req.WithContext(context.WithValue(req.Context(), simbaContext.RequestSettingsKey, &settings.Request{
			TraceIDMode: simbaModels.AcceptFromHeader,
		}))
		w := httptest.NewRecorder()

		middleware.TraceID(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test-trace-id", w.Header().Get(simbaContext.TraceIDHeader))
	})
}
