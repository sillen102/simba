package simba_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/enums"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/test"
	"gotest.tools/v3/assert"
)

func TestJsonHandler(t *testing.T) {
	t.Parallel()

	t.Run("body and params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params]) (*simba.Response[map[string]string], error) {
			assert.Equal(t, "John", req.Params.Name)
			assert.Equal(t, 1, req.Params.ID)
			assert.Equal(t, true, req.Params.Active)
			assert.Equal(t, 0, req.Params.Page)
			assert.Equal(t, int64(10), req.Params.Size)

			assert.Equal(t, "test", req.Body.Test)

			return &simba.Response[map[string]string]{
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

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.Config{Logger: logger})
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"message\":\"success\"}\n", w.Body.String())
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))
		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no body", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, test.Params]) (*simba.Response[simba.NoBody], error) {
			return &simba.Response[simba.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1&size=10&active=true", nil)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.Config{Logger: logger})
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
			return &simba.Response[simba.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test?page=1&size=10&active=true", body) // Params should be ignored
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.Config{Logger: logger})
		app.Router.POST("/test", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("default values on params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, test.Params]) (*simba.Response[simba.NoBody], error) {
			assert.Equal(t, 1, req.Params.Page)         // default value
			assert.Equal(t, int64(10), req.Params.Size) // default value
			assert.Equal(t, 10.0, req.Params.Score)
			return &simba.Response[simba.NoBody]{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?active=true", body)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.Config{Logger: logger})
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("override default values with query params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, test.Params]) (*simba.Response[simba.NoBody], error) {
			assert.Equal(t, 5, req.Params.Page)         // overridden value
			assert.Equal(t, int64(20), req.Params.Size) // overridden value
			assert.Equal(t, 15.5, req.Params.Score)     // overridden value
			return &simba.Response[simba.NoBody]{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?active=true&page=5&size=20&score=15.5", body)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.Config{Logger: logger})
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestHandlerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		path           string
		contentType    string
		body           string
		expectedStatus int
		expectedError  string
		expectedMsg    string
	}{
		{
			name:           "missing content type",
			method:         http.MethodPost,
			path:           "/test",
			contentType:    "",
			body:           `{"test": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Bad Request",
			expectedMsg:    "invalid content type",
		},
		{
			name:           "invalid json body",
			method:         http.MethodPost,
			path:           "/test",
			contentType:    "application/json",
			body:           `{"test": invalid}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Unprocessable Entity",
			expectedMsg:    "invalid request body",
		},
		{
			name:           "missing required field",
			method:         http.MethodPost,
			path:           "/test",
			contentType:    "application/json",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Bad Request",
			expectedMsg:    "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := func(ctx context.Context, req *simba.Request[test.RequestBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
				return &simba.Response[simba.NoBody]{Status: http.StatusOK}, nil
			}

			body := strings.NewReader(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			app := simba.New()
			app.Router.POST("/test", simba.JsonHandler(handler))
			app.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var errorResponse simba.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			assert.NilError(t, err)

			assert.Equal(t, tt.expectedStatus, errorResponse.Status)
			assert.Equal(t, tt.expectedError, errorResponse.Error)
			assert.Equal(t, tt.path, errorResponse.Path)
			assert.Equal(t, tt.method, errorResponse.Method)
			if tt.expectedMsg != "" {
				assert.Equal(t, tt.expectedMsg, errorResponse.Message)
			}
		})
	}
}

func TestAuthenticatedJsonHandler(t *testing.T) {
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
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params], user *test.User) (*simba.Response[simba.NoBody], error) {
			assert.Equal(t, 1, user.ID)
			assert.Equal(t, "John Doe", user.Name)
			assert.Equal(t, "admin", user.Role)

			return &simba.Response[simba.NoBody]{
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

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.Config{Logger: logger})
		app.Router.POST("/test/{id}", simba.AuthJsonHandler(handler, authFunc))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("auth func error", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params], user *test.User) (*simba.Response[simba.NoBody], error) {
			return &simba.Response[simba.NoBody]{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1", body)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.Config{Logger: logger})
		app.Router.POST("/test/{id}", simba.AuthJsonHandler(handler, errorAuthFunc))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResponse simba.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, "unauthorized", errorResponse.Message)
	})
}

type TestRequestBody struct {
	Test string `json:"test"`
}

func TestReadJson_DisallowUnknownFields(t *testing.T) {
	t.Parallel()

	handler := func(ctx context.Context, req *simba.Request[TestRequestBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
		return &simba.Response[simba.NoBody]{Status: http.StatusOK}, nil
	}

	body := strings.NewReader(`{"test": "value", "unknown": "field"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	logBuffer := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
	app := simba.New(settings.Config{
		Logger: logger,
		Request: settings.Request{
			AllowUnknownFields: enums.Disallow,
		},
	})
	app.Router.POST("/test", simba.JsonHandler(handler))
	app.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var errorResponse simba.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResponse)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errorResponse.Status)
	assert.Equal(t, "invalid request body", errorResponse.Message)
	assert.Equal(t, 1, len(errorResponse.ValidationErrors))

	validationError := errorResponse.ValidationErrors[0]
	assert.Equal(t, "body", validationError.Parameter)
	assert.Equal(t, simba.ParameterTypeBody, validationError.Type)
	assert.Equal(t, "json: unknown field \"unknown\"", validationError.Message)
}
