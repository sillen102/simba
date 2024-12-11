package simba_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba"
	"gotest.tools/v3/assert"
)

// TODO: Add more tests
// 	1. Route conflicts
// 	2. Wildcard routes
//  3. Middleware chain ordering
//  4. Route parameter validation
//  5. OPTIONS requests handling
//  6. HEAD requests handling

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
		app.Router.GET("/test", simba.JsonHandler(handler))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestRouter_Handle(t *testing.T) {
	t.Parallel()

	t.Run("Handle registers a handler and serves requests", func(t *testing.T) {
		router := simba.Default().Router

		// Define a simple handler
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("handled"))
		}

		// Register the handler
		router.Handle("GET /test", handler)

		// Create a request to the registered path
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Serve the request
		router.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "handled", w.Body.String())
	})

	t.Run("Handle applies middleware to the handler", func(t *testing.T) {
		router := simba.Default().Router

		// Define a middleware that sets a header
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Custom-Header", "middleware-applied")
				next.ServeHTTP(w, r)
			})
		}

		// Use the middleware
		router.Use(middleware)

		// Define a simple handler
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		// Register the handler
		router.Handle("/test-middleware", handler)

		// Create a request to the registered path
		req := httptest.NewRequest(http.MethodGet, "/test-middleware", nil)
		w := httptest.NewRecorder()

		// Serve the request
		router.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "middleware-applied", w.Header().Get("X-Custom-Header"))
	})
}

func TestRouter_POST(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
		return &simba.Response{
			Body:   map[string]string{"message": "post handled"},
			Status: http.StatusCreated,
		}, nil
	}

	router.POST("/test-post", simba.JsonHandler(handler))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test-post", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, `{"message":"post handled"}`, strings.Trim(w.Body.String(), "\n"))
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-post", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestRouter_GET(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
		return &simba.Response{
			Body: map[string]string{"message": "get handled"},
		}, nil
	}

	router.GET("/test-get", simba.JsonHandler(handler))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-get", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"message":"get handled"}`, strings.Trim(w.Body.String(), "\n"))
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test-get", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestRouter_PUT(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
		return &simba.Response{
			Body:   map[string]string{"message": "put handled"},
			Status: http.StatusAccepted,
		}, nil
	}

	router.PUT("/test-put", simba.JsonHandler(handler))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/test-put", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, `{"message":"put handled"}`, strings.Trim(w.Body.String(), "\n"))
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-put", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestRouter_DELETE(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
		return &simba.Response{}, nil
	}

	router.DELETE("/test-delete", simba.JsonHandler(handler))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/test-delete", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-delete", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestRouter_PATCH(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
		return &simba.Response{
			Body:   map[string]string{"message": "patch handled"},
			Status: http.StatusAccepted,
		}, nil
	}

	router.PATCH("/test-patch", simba.JsonHandler(handler))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/test-patch", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, `{"message":"patch handled"}`, strings.Trim(w.Body.String(), "\n"))
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-patch", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
