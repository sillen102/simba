package simba_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"gotest.tools/v3/assert"
)

func TestEndpoints(t *testing.T) {
	t.Parallel()

	t.Run("test health endpoint", func(t *testing.T) {
		router := simba.Default().Router

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"status\":\"ok\"}", w.Body.String())
	})

	t.Run("use middleware", func(t *testing.T) {
		// Define a middleware that sets a header
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-Custom-Header", "middleware-applied")
				next.ServeHTTP(w, r)
			})
		}

		type TestParams struct {
			CustomHeader string `header:"X-Custom-Header"`
		}

		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, TestParams]) (*simba.Response, error) {
			// Assert that the header was set by the middleware in the handler
			assert.Equal(t, req.Params.CustomHeader, "middleware-applied")
			return &simba.Response{}, nil
		}

		app := simba.New()
		app.Router.Use(middleware)
		app.Router.GET("/test", simba.HandlerFunc(handler))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
