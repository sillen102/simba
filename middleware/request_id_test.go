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

	t.Run("generates new request ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(simbaContext.RequestIDKey).(string)
			assert.Assert(t, requestID != "")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequestID(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Assert(t, w.Header().Get(simbaContext.RequestIDHeader) != "")

		// Check if the request ID is a valid UUID
		_, err := uuid.Parse(w.Header().Get(simbaContext.RequestIDHeader))
		assert.NoError(t, err)
	})

	t.Run("accepts request ID from header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(simbaContext.RequestIDKey).(string)
			assert.Equal(t, "test-request-id", requestID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(simbaContext.RequestIDHeader, "test-request-id")
		req = req.WithContext(context.WithValue(req.Context(), simbaContext.RequestSettingsKey, &settings.Request{
			RequestIdMode: simbaModels.AcceptFromHeader,
		}))
		w := httptest.NewRecorder()

		middleware.RequestID(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test-request-id", w.Header().Get(simbaContext.RequestIDHeader))
	})
}
