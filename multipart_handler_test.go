package simba_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestMultipartHandler(t *testing.T) {
	t.Parallel()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("lastName", "Connor")
	_ = writer.WriteField("alive", "true")
	_ = writer.Close()

	id := uuid.NewString()

	t.Run("multipart file and params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaTest.Params]) (*simbaModels.Response[map[string]string], error) {
			assert.Equal(t, id, req.Params.ID.String())
			assert.Equal(t, true, req.Params.Active)
			assert.Equal(t, 0, req.Params.Page)
			assert.Equal(t, int64(10), req.Params.Size)

			return &simbaModels.Response[map[string]string]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusAccepted,
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/multipart-test/%s?page=0&size=10&active=true", id), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/multipart-test/{id}", simba.MultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"message\":\"success\"}\n", w.Body.String())
		assert.Equal(t, "header-value", w.Header().Get("My-Header"))
		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})

	t.Run("no params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
			return &simbaModels.Response[map[string]string]{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusAccepted,
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/multipart-test", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/multipart-test", simba.MultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("default values on params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaTest.Params]) (*simbaModels.Response[simbaModels.NoBody], error) {
			assert.Equal(t, 1, req.Params.Page)         // default value
			assert.Equal(t, int64(10), req.Params.Size) // default value
			assert.Equal(t, 10.0, req.Params.Score)
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/multipart-test/%s?active=true", id), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/multipart-test/{id}", simba.MultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("override default values with query params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaTest.Params]) (*simbaModels.Response[simbaModels.NoBody], error) {
			assert.Equal(t, 5, req.Params.Page)         // overridden value
			assert.Equal(t, int64(20), req.Params.Size) // overridden value
			assert.Equal(t, 15.5, req.Params.Score)     // overridden value
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/multipart-test/%s?active=true&page=5&size=20&score=15.5", id), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/multipart-test/{id}", simba.MultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestMultipartHandlerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		path           string
		contentType    string
		body           func() *bytes.Buffer
		expectedStatus int
		expectedError  string
		expectedMsg    string
	}{
		{
			name:        "missing content type",
			method:      http.MethodPost,
			path:        "/test",
			contentType: "",
			body: func() *bytes.Buffer {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				_ = writer.Close()
				return body
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Bad Request",
			expectedMsg:    "invalid content type",
		},
		{
			name:        "missing boundary",
			method:      http.MethodPost,
			path:        "/test",
			contentType: "multipart/form-data",
			body: func() *bytes.Buffer {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				_ = writer.Close()
				return body
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Bad Request",
			expectedMsg:    "invalid content type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
				return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
			}

			req := httptest.NewRequest(tt.method, tt.path, tt.body())
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			app := simba.New()
			app.Router.POST("/test", simba.MultipartHandler(handler))
			app.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var errorResponse simbaErrors.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			assert.NoError(t, err)

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

func TestAuthenticatedMultipartHandler(t *testing.T) {
	t.Parallel()

	authFunc := func(ctx context.Context, token string) (*simbaTest.User, error) {
		assert.Equal(t, "token", token)
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

	customAuthFunc := func(ctx context.Context, token string) (*simbaTest.User, error) {
		return nil, simbaErrors.NewSimbaError(http.StatusForbidden, "forbidden", nil)
	}

	customAuthHandler := simba.BearerAuthType[simbaTest.User]{
		Handler: customAuthFunc,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("lastName", "Connor")
	_ = writer.WriteField("alive", "true")
	_ = writer.Close()

	id := uuid.NewString()

	t.Run("authenticated multipart handler", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaTest.Params], authModel *simbaTest.User) (*simbaModels.Response[simbaModels.NoBody], error) {
			assert.Equal(t, 1, authModel.ID)
			assert.Equal(t, "John Doe", authModel.Name)
			assert.Equal(t, "admin", authModel.Role)
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/test/%s?page=1&size=10&active=true", id), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer token")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.AuthMultipartHandler(handler, authHandler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("authenticated multipart handler with error", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaTest.Params], user *simbaTest.User) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/test/1", body)
		req.Header.Set("Authorization", "Bearer token")
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.AuthMultipartHandler(handler, errorAuthHandler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResponse simbaErrors.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errorResponse.Message)
	})

	t.Run("authenticated multipart handler with custom error", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.MultipartRequest[simbaTest.Params], user *simbaTest.User) (*simbaModels.Response[simbaModels.NoBody], error) {
			return &simbaModels.Response[simbaModels.NoBody]{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/test/1", body)
		req.Header.Set("Authorization", "Bearer token")
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		app := simba.New(settings.WithLogger(logger))
		app.Router.POST("/test/{id}", simba.AuthMultipartHandler(handler, customAuthHandler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var errorResponse simbaErrors.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "forbidden", errorResponse.Message)
	})
}
