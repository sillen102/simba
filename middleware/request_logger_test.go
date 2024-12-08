package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sillen102/simba/middleware"
	"gotest.tools/v3/assert"
)

func TestLogRequests(t *testing.T) {
	t.Parallel()

	t.Run("logs request and response", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond) // Simulate some work
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"success"}`))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.LogRequests(handler).ServeHTTP(w, req)

		// Since we're using a custom logger, we can only verify the response was written
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"message":"success"}`, w.Body.String())
	})
}
