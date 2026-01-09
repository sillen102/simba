package simba_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
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

		handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, TestParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
			// Assert that the header was set by the middleware in the handler
			assert.Equal(t, req.Params.CustomHeader, "middleware-applied")
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
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

func TestRouter_POST(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
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

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
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

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
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

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
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

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
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

func TestRouter_OPTIONS(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
			Body:   map[string]string{"message": "options handled"},
			Status: http.StatusOK,
		}, nil
	}

	router.OPTIONS("/test-options", simba.JsonHandler(handler))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test-options", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"message":"options handled"}`, strings.Trim(w.Body.String(), "\n"))
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-options", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestRouter_HEAD(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	router.HEAD("/test-head", simba.JsonHandler(handler))

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/test-head", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-head", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestRouter_Use(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	// Middleware to add a custom header
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Test-Middleware", "true")
			next.ServeHTTP(w, r)
		})
	}

	router.Use(middleware)

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, struct {
		TestMiddleware string `header:"X-Test-Middleware"`
	}]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
			Body: map[string]string{"middleware": req.Params.TestMiddleware},
		}, nil
	}

	router.GET("/test-middleware", simba.JsonHandler(handler))

	t.Run("middleware applied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-middleware", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"middleware":"true"}`, strings.Trim(w.Body.String(), "\n"))
	})
}

func TestRouter_Extend(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	// Middleware to add a custom header
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware-1", "true")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware-2", "true")
			next.ServeHTTP(w, r)
		})
	}

	router.Extend([]func(http.Handler) http.Handler{middleware1, middleware2})

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, struct {
		Middleware1 string `header:"X-Middleware-1"`
		Middleware2 string `header:"X-Middleware-2"`
	}]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
			Body: map[string]string{
				"middleware1": req.Params.Middleware1,
				"middleware2": req.Params.Middleware2,
			},
		}, nil
	}

	router.GET("/test-extend", simba.JsonHandler(handler))

	t.Run("extended middleware applied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-extend", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"middleware1":"true","middleware2":"true"}`, strings.Trim(w.Body.String(), "\n"))
	})
}

func TestRouter_WithMiddleware(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	type TestBody struct {
		Middleware1 string `json:"middleware1"`
		Middleware2 string `json:"middleware2"`
	}

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Test-Middleware", "one")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Another-Middleware", "two")
			next.ServeHTTP(w, r)
		})
	}

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, struct {
		TestMiddleware    string `header:"X-Test-Middleware"`
		AnotherMiddleware string `header:"X-Another-Middleware"`
	}]) (*simbaModels.Response[TestBody], error) {
		return &simbaModels.Response[TestBody]{
			Body: TestBody{
				Middleware1: req.Params.TestMiddleware,
				Middleware2: req.Params.AnotherMiddleware,
			},
		}, nil
	}

	router.WithMiddleware("GET", "/with-one", simba.JsonHandler(handler), middleware1)
	router.WithMiddleware("GET", "/with-both", simba.JsonHandler(handler), middleware1, middleware2)
	router.GET("/with-none", simba.JsonHandler(handler))

	t.Run("one middleware applied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/with-one", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var body TestBody
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.Nil(t, err)
		assert.Equal(t, "one", body.Middleware1)
		assert.Equal(t, "", body.Middleware2)
	})

	t.Run("both middleware applied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/with-both", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var body TestBody
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.Nil(t, err)
		assert.Equal(t, "one", body.Middleware1)
		assert.Equal(t, "two", body.Middleware2)
	})

	t.Run("no middleware applied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/with-none", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var body TestBody
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.Nil(t, err)
		assert.Equal(t, "", body.Middleware1)
		assert.Equal(t, "", body.Middleware2)
	})
}

func TestRouter_HandlerWithPointerRequestBody(t *testing.T) {
	t.Parallel()

	router := simba.Default().Router

	type PatchBody struct {
		Name   *string  `json:"name"`
		Active *bool    `json:"active" default:"true"`
		Count  *int     `json:"count" default:"10"`
		Rate   *float64 `json:"rate" default:"3.14"`
		Status *string  `json:"status" default:"pending"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[*PatchBody, simbaModels.NoParams]) (*simbaModels.Response[PatchBody], error) {
		// Return the body to verify defaults were applied
		return &simbaModels.Response[PatchBody]{
			Body: PatchBody{
				Name:   req.Body.Name,
				Active: req.Body.Active,
				Count:  req.Body.Count,
				Rate:   req.Body.Rate,
				Status: req.Body.Status,
			},
		}, nil
	}

	router.PATCH("/test-pointer-body", simba.JsonHandler(handler))

	t.Run("pointer request body with defaults applied", func(t *testing.T) {
		body := `{"name":"test"}`
		req := httptest.NewRequest(http.MethodPatch, "/test-pointer-body", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp PatchBody
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Nil(t, err)
		assert.NotNil(t, resp.Name)
		assert.Equal(t, "test", *resp.Name)
		assert.NotNil(t, resp.Active)
		assert.Equal(t, true, *resp.Active)
		assert.NotNil(t, resp.Count)
		assert.Equal(t, 10, *resp.Count)
		assert.NotNil(t, resp.Rate)
		assert.Equal(t, 3.14, *resp.Rate)
		assert.NotNil(t, resp.Status)
		assert.Equal(t, "pending", *resp.Status)
	})

	t.Run("pointer request body empty with all defaults", func(t *testing.T) {
		body := `{}`
		req := httptest.NewRequest(http.MethodPatch, "/test-pointer-body", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp PatchBody
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Nil(t, err)
		assert.Nil(t, resp.Name)
		assert.NotNil(t, resp.Active)
		assert.Equal(t, true, *resp.Active)
		assert.NotNil(t, resp.Count)
		assert.Equal(t, 10, *resp.Count)
		assert.NotNil(t, resp.Rate)
		assert.Equal(t, 3.14, *resp.Rate)
		assert.NotNil(t, resp.Status)
		assert.Equal(t, "pending", *resp.Status)
	})

	t.Run("pointer request body with partial overrides", func(t *testing.T) {
		body := `{"active":false,"count":20}`
		req := httptest.NewRequest(http.MethodPatch, "/test-pointer-body", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp PatchBody
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Nil(t, err)
		assert.Nil(t, resp.Name)
		assert.NotNil(t, resp.Active)
		assert.Equal(t, false, *resp.Active)
		assert.NotNil(t, resp.Count)
		assert.Equal(t, 20, *resp.Count)
		assert.NotNil(t, resp.Rate)
		assert.Equal(t, 3.14, *resp.Rate)
		assert.NotNil(t, resp.Status)
		assert.Equal(t, "pending", *resp.Status)
	})
}
