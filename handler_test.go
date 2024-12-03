package simba_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/test"
	"gotest.tools/v3/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	t.Run("body and params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params]) (*simba.Response, error) {
			assert.Equal(t, "John", req.Params.Name)
			assert.Equal(t, 1, req.Params.ID)
			assert.Equal(t, true, req.Params.Active)
			assert.Equal(t, int64(0), req.Params.Page)
			assert.Equal(t, int64(10), req.Params.Size)

			assert.Equal(t, "test", req.Body.Test)

			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusOK,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?page=0&size=10&active=true", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		router := simba.Default()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"message\":\"success\"}\n", w.Body.String())

		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no body", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, test.Params]) (*simba.Response, error) {
			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1&size=10&active=true", nil)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no content type set, expect 400", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params]) (*simba.Response, error) {
			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1&size=10&active=true", body)
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResponse simba.ErrorResponse
		// Read json into errorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)

		assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
		assert.Equal(t, "Bad Request", errorResponse.Error)
		assert.Equal(t, "/test/1", errorResponse.Path)
		assert.Equal(t, http.MethodPost, errorResponse.Method)
		assert.Equal(t, "invalid content type", errorResponse.Message)
	})

	t.Run("no params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, simba.NoParams]) (*simba.Response, error) {
			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1&size=10&active=true", body) // Params should be ignored
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("wrong method, expect 405", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, simba.NoParams]) (*simba.Response, error) {
			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodGet, "/test/1?page=1&size=10&active=true", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("uuid params", func(t *testing.T) {
		type UuidParams struct {
			ID       uuid.UUID `path:"id"`
			HeaderID uuid.UUID `header:"Header-ID"`
			QueryID  uuid.UUID `query:"queryId"`
		}
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, UuidParams]) (*simba.Response, error) {
			assert.Equal(t, "123e4567-e89b-12d3-a456-426655440000", req.Params.ID.String())
			assert.Equal(t, "248ccd0e-4bdf-4c41-a125-92ef3a416251", req.Params.HeaderID.String())
			assert.Equal(t, "ccf586b9-6fc9-4c1b-a3a6-8b89ac25ab84", req.Params.QueryID.String())
			return &simba.Response{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/123e4567-e89b-12d3-a456-426655440000?queryId=ccf586b9-6fc9-4c1b-a3a6-8b89ac25ab84", body)
		req.Header.Set("Header-ID", "248ccd0e-4bdf-4c41-a125-92ef3a416251")
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("float params", func(t *testing.T) {
		type FloatParams struct {
			Page float64 `query:"page"`
			Size float64 `query:"size"`
		}
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, FloatParams]) (*simba.Response, error) {
			assert.Equal(t, 1.1, req.Params.Page)
			assert.Equal(t, 2.2, req.Params.Size)
			return &simba.Response{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1.1&size=2.2&active=true", body)
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("default values on params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, test.Params]) (*simba.Response, error) {
			assert.Equal(t, int64(0), req.Params.Page)  // default value
			assert.Equal(t, int64(10), req.Params.Size) // default value
			assert.Equal(t, 10.0, req.Params.Score)
			return &simba.Response{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?active=true", body)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestAuthenticatedHandler(t *testing.T) {
	t.Parallel()

	authFunc := func(r *http.Request) (*test.User, error) {
		return &test.User{
			ID:   1,
			Name: "John Doe",
			Role: "admin",
		}, nil
	}

	errorAuthFunc := func(r *http.Request) (*test.User, error) {
		return nil, errors.New("user not found")
	}

	t.Run("authenticated handler", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params], user *test.User) (*simba.Response, error) {
			assert.Equal(t, 1, user.ID)
			assert.Equal(t, "John Doe", user.Name)
			assert.Equal(t, "admin", user.Role)

			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1&size=10&active=true", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		router := simba.DefaultWithAuth[test.User](authFunc)
		router.POST("/test/:id", simba.AuthenticatedHandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("auth func error", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params], user *test.User) (*simba.Response, error) {
			return &simba.Response{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1", body)
		w := httptest.NewRecorder()

		router := simba.NewRouterWithAuth[test.User](errorAuthFunc)
		router.POST("/test/:id", simba.AuthenticatedHandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResponse simba.HTTPError
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, "unauthorized", errorResponse.Message)
	})

	t.Run("router with auth with unauthenticated handler", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params]) (*simba.Response, error) {
			assert.Equal(t, "John", req.Params.Name)
			assert.Equal(t, 1, req.Params.ID)
			assert.Equal(t, true, req.Params.Active)
			assert.Equal(t, int64(0), req.Params.Page)
			assert.Equal(t, int64(10), req.Params.Size)

			assert.Equal(t, "test", req.Body.Test)

			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusOK,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?page=0&size=10&active=true", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		router := simba.NewRouterWithAuth[test.User](authFunc)
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"message\":\"success\"}\n", w.Body.String())

		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})
}
