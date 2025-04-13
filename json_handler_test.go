package simba_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest"
	"github.com/sillen102/simba/simbaTestAssert"
)

func TestJsonHandler(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()

	t.Run("body and params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaTest.RequestBody, simbaTest.Params]) (*simbaModels.Response[map[string]string], error) {
			simbaTestAssert.Equal(t, "John", req.Body.Name)
			simbaTestAssert.Equal(t, id, req.Params.ID.String())
			simbaTestAssert.Equal(t, true, req.Params.Active)
			simbaTestAssert.Equal(t, 0, req.Params.Page)
			simbaTestAssert.Equal(t, int64(10), req.Params.Size)

			return &simbaModels.Response[map[string]string]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusOK,
			}, nil
		}

		body := strings.NewReader(`{"name": "John"}`)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/test/%s?page=0&size=10&active=true", id), body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusOK, w.Code)
		simbaTestAssert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		simbaTestAssert.Equal(t, "{\"message\":\"success\"}\n", w.Body.String())
		simbaTestAssert.Equal(t, "header-value", w.Header().Get("My-Header"))
		cookie := w.Result().Cookies()[0].Value
		simbaTestAssert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no body", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaTest.Params]) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/test/%s?page=1&size=10&active=true", id), nil)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusNoContent, w.Code)
		simbaTestAssert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		simbaTestAssert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaTest.RequestBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader(`{"name": "John"}`)
		req := httptest.NewRequest(http.MethodPost, "/test?page=1&size=10&active=true", body) // Params should be ignored
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusNoContent, w.Code)
		simbaTestAssert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		simbaTestAssert.Equal(t, "cookie-value", cookie)
	})

	t.Run("default values on params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaTest.Params]) (*simbaModels.Response[simbaModels.NoBody], error) {
			simbaTestAssert.Equal(t, 1, req.Params.Page)         // default value
			simbaTestAssert.Equal(t, int64(10), req.Params.Size) // default value
			simbaTestAssert.Equal(t, 10.0, req.Params.Score)
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		body := strings.NewReader(`{"name": "John"}`)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/test/%s?active=true", id), body)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("override default values with query params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaTest.Params]) (*simbaModels.Response[simbaModels.NoBody], error) {
			simbaTestAssert.Equal(t, 5, req.Params.Page)         // overridden value
			simbaTestAssert.Equal(t, int64(20), req.Params.Size) // overridden value
			simbaTestAssert.Equal(t, 15.5, req.Params.Score)     // overridden value
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		body := strings.NewReader(`{"name": "John"}`)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/test/%s?active=true&page=5&size=20&score=15.5", id), body)
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusNoContent, w.Code)
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

			handler := func(ctx context.Context, req *simbaModels.Request[simbaTest.RequestBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
				return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
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

			simbaTestAssert.Equal(t, tt.expectedStatus, w.Code)
			simbaTestAssert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var errorResponse simbaErrors.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			simbaTestAssert.NoError(t, err)

			simbaTestAssert.Equal(t, tt.expectedStatus, errorResponse.Status)
			simbaTestAssert.Equal(t, tt.expectedError, errorResponse.Error)
			simbaTestAssert.Equal(t, tt.path, errorResponse.Path)
			simbaTestAssert.Equal(t, tt.method, errorResponse.Method)
			if tt.expectedMsg != "" {
				simbaTestAssert.Equal(t, tt.expectedMsg, errorResponse.Message)
			}
		})
	}
}

func TestAuthenticatedJsonHandler(t *testing.T) {
	t.Parallel()

	authFunc := func(ctx context.Context, token string) (*simbaTest.User, error) {
		simbaTestAssert.Equal(t, "token", token)
		return &simbaTest.User{
			ID:   1,
			Name: "John Doe",
			Role: "admin",
		}, nil
	}

	authHandler := simba.BearerAuthType[simbaTest.User]{
		Handler: authFunc,
	}

	errorAuthFunc := func(ctx context.Context, token string) (*simbaTest.User, error) {
		return nil, errors.New("user not found")
	}

	errorAuthHandler := simba.BearerAuthType[simbaTest.User]{
		Handler: errorAuthFunc,
	}

	id := uuid.NewString()

	t.Run("authenticated handler", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaTest.RequestBody, simbaTest.Params], user *simbaTest.User) (*simbaModels.Response[simbaModels.NoBody], error) {
			simbaTestAssert.Equal(t, 1, user.ID)
			simbaTestAssert.Equal(t, "John Doe", user.Name)
			simbaTestAssert.Equal(t, "admin", user.Role)

			return &simbaModels.Response[simbaModels.NoBody]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Status:  http.StatusNoContent,
			}, nil
		}

		body := strings.NewReader(`{"name": "John"}`)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/test/%s?page=1&size=10&active=true", id), body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer token")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.AuthJsonHandler(handler, authHandler))
		app.Router.ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusNoContent, w.Code)
		simbaTestAssert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		simbaTestAssert.Equal(t, "cookie-value", cookie)
	})

	t.Run("auth func error", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaTest.RequestBody, simbaTest.Params], user *simbaTest.User) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1", body)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.AuthJsonHandler(handler, errorAuthHandler))
		app.Router.ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResponse simbaErrors.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		simbaTestAssert.NoError(t, err)
		simbaTestAssert.Equal(t, "unauthorized", errorResponse.Message)
	})
}

type TestRequestBody struct {
	Test string `json:"test"`
}

func TestReadJson_DisallowUnknownFields(t *testing.T) {
	t.Parallel()

	handler := func(ctx context.Context, req *simbaModels.Request[TestRequestBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
	}

	body := strings.NewReader(`{"test": "value", "unknown": "field"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	logBuffer := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
	app := simba.New(settings.WithLogger(logger), settings.WithAllowUnknownFields(false))
	app.Router.POST("/test", simba.JsonHandler(handler))
	app.Router.ServeHTTP(w, req)

	simbaTestAssert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var errorResponse simbaErrors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResponse)
	simbaTestAssert.NoError(t, err)
	simbaTestAssert.Equal(t, http.StatusUnprocessableEntity, errorResponse.Status)
	simbaTestAssert.Equal(t, "invalid request body", errorResponse.Message)
	simbaTestAssert.Equal(t, 1, len(errorResponse.ValidationErrors))

	validationError := errorResponse.ValidationErrors[0]
	simbaTestAssert.Equal(t, "body", validationError.Parameter)
	simbaTestAssert.Equal(t, simbaErrors.ParameterTypeBody, validationError.Type)
	simbaTestAssert.Equal(t, "json: unknown field \"unknown\"", validationError.Message)
}
