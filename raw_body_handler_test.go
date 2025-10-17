package simba_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestRawBodyHandler(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()

	t.Run("body and params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, simbaTest.Params]) (*simbaModels.Response[map[string]string], error) {
			bodyBytes, _ := io.ReadAll(req.Body)
			assert.Equal(t, "John", string(bodyBytes))
			assert.Equal(t, id, req.Params.ID.String())
			assert.Equal(t, true, req.Params.Active)
			assert.Equal(t, 0, req.Params.Page)
			assert.Equal(t, int64(10), req.Params.Size)
			return &simbaModels.Response[map[string]string]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusOK,
			}, nil
		}

		body := strings.NewReader("John")
		req := httptest.NewRequest(http.MethodPost, "/test/"+id+"?page=0&size=10&active=true", body)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.POST("/test/{id}", simba.RawBodyHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "{\"message\":\"success\"}\n", w.Body.String())
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))
		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no body", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, simbaTest.Params]) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/test/"+id+"?page=1&size=10&active=true", nil)
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.POST("/test/{id}", simba.RawBodyHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))
		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader("John")
		req := httptest.NewRequest(http.MethodPost, "/test", body)
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.POST("/test", simba.RawBodyHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))
		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})
}

func TestAuthenticatedRawBodyHandler(t *testing.T) {
	t.Parallel()

	authFunc := func(ctx context.Context, token string) (*simbaTest.User, error) {
		assert.Equal(t, "token", token)
		return &simbaTest.User{
			ID:   1,
			Name: "John Doe",
			Role: "admin",
		}, nil
	}
	authHandler := simba.BearerAuthType[*simbaTest.User]{Handler: authFunc}

	id := uuid.NewString()

	t.Run("authenticated handler", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, simbaTest.Params], user *simbaTest.User) (*simbaModels.Response[simbaModels.NoBody], error) {
			assert.Equal(t, 1, user.ID)
			assert.Equal(t, "John Doe", user.Name)
			assert.Equal(t, "admin", user.Role)
			bodyBytes, _ := io.ReadAll(req.Body)
			assert.Equal(t, "John", string(bodyBytes))
			return &simbaModels.Response[simbaModels.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader("John")
		req := httptest.NewRequest(http.MethodPost, "/test/"+id+"?page=1&size=10&active=true", body)
		req.Header.Set("Authorization", "Bearer token")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.POST("/test/{id}", simba.AuthRawBodyHandler(handler, authHandler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))
		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("auth func error", func(t *testing.T) {
		errorAuthFunc := func(ctx context.Context, token string) (*simbaTest.User, error) {
			return nil, simbaErrors.NewSimbaError(http.StatusUnauthorized, "unauthorized", nil)
		}
		errorAuthHandler := simba.BearerAuthType[*simbaTest.User]{Handler: errorAuthFunc}

		handler := func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, simbaTest.Params], user *simbaTest.User) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		body := strings.NewReader("John")
		req := httptest.NewRequest(http.MethodPost, "/test/"+id, body)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.POST("/test/{id}", simba.AuthRawBodyHandler(handler, errorAuthHandler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var errorResponse simbaErrors.ErrorResponse
		_ = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.Equal(t, "unauthorized", errorResponse.Message)
	})
}
