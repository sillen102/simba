package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/middleware"
	"gotest.tools/v3/assert"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	t.Run("adds request id when none exists", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		config := &middleware.RequestIdConfig{
			AcceptFromHeader: false,
		}
		config.RequestID(handler).ServeHTTP(w, req)

		requestID := w.Header().Get(middleware.RequestIDHeader)
		assert.Assert(t, len(requestID) > 0, "expected request ID in response header")
	})

	t.Run("preserves existing request id", func(t *testing.T) {
		existingID := "test-request-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middleware.RequestIDHeader, existingID)
		w := httptest.NewRecorder()

		config := &middleware.RequestIdConfig{
			AcceptFromHeader: true,
		}
		config.RequestID(handler).ServeHTTP(w, req)

		requestID := w.Header().Get(middleware.RequestIDHeader)
		assert.Equal(t, existingID, requestID, "expected original request ID in response header")
	})

	t.Run("generates new id even when one exists", func(t *testing.T) {
		existingID := "test-request-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middleware.RequestIDHeader, existingID)
		w := httptest.NewRecorder()

		config := &middleware.RequestIdConfig{
			AcceptFromHeader: false,
		}
		config.RequestID(handler).ServeHTTP(w, req)

		requestID := w.Header().Get(middleware.RequestIDHeader)
		assert.Assert(t, requestID != existingID, "expected new request ID in response header")
		assert.Assert(t, len(requestID) > 0, "expected non-empty request ID")
	})

	t.Run("request id is available in context", func(t *testing.T) {
		var contextRequestID string
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextRequestID = r.Context().Value(middleware.RequestIDKey).(string)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		config := &middleware.RequestIdConfig{
			AcceptFromHeader: false,
		}
		config.RequestID(handler).ServeHTTP(w, req)

		responseRequestID := w.Header().Get(middleware.RequestIDHeader)
		assert.Equal(t, responseRequestID, contextRequestID, "request ID in context should match response header")
	})
}
