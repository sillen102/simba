package simba_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/test"
	"gotest.tools/v3/assert"
)

func TestMultipartHandler(t *testing.T) {
	t.Parallel()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("lastName", "Connor")
	writer.WriteField("alive", "true")
	writer.Close()

	t.Run("multipart file and params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.MultipartRequest[test.Params]) (*simba.Response, error) {
			assert.Equal(t, "John", req.Params.Name)
			assert.Equal(t, 1, req.Params.ID)
			assert.Equal(t, true, req.Params.Active)
			assert.Equal(t, 0, req.Params.Page)
			assert.Equal(t, int64(10), req.Params.Size)

			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusAccepted,
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/multipart-test/1?page=0&size=10&active=true", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		app := simba.New(settings.Settings{
			Logging: logging.Config{
				Output: logBuffer,
			},
		})
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
		handler := func(ctx context.Context, req *simba.MultipartRequest[simba.NoParams]) (*simba.Response, error) {
			return &simba.Response{
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
		app := simba.New(settings.Settings{
			Logging: logging.Config{
				Output: logBuffer,
			},
		})
		app.Router.POST("/multipart-test", simba.MultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("default values on params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.MultipartRequest[test.Params]) (*simba.Response, error) {
			assert.Equal(t, 1, req.Params.Page)         // default value
			assert.Equal(t, int64(10), req.Params.Size) // default value
			assert.Equal(t, 10.0, req.Params.Score)
			return &simba.Response{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/multipart-test/1?active=true", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		app := simba.New(settings.Settings{
			Logging: logging.Config{
				Output: logBuffer,
			},
		})
		app.Router.POST("/multipart-test/{id}", simba.MultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("override default values with query params", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.MultipartRequest[test.Params]) (*simba.Response, error) {
			assert.Equal(t, 5, req.Params.Page)         // overridden value
			assert.Equal(t, int64(20), req.Params.Size) // overridden value
			assert.Equal(t, 15.5, req.Params.Score)     // overridden value
			return &simba.Response{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/multipart-test/1?active=true&page=5&size=20&score=15.5", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		app := simba.New(settings.Settings{
			Logging: logging.Config{
				Output: logBuffer,
			},
		})
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
			path:        "/test/1",
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
			path:        "/test/1",
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

			handler := func(ctx context.Context, req *simba.MultipartRequest[simba.NoParams]) (*simba.Response, error) {
				return &simba.Response{Status: http.StatusOK}, nil
			}

			req := httptest.NewRequest(tt.method, tt.path, tt.body())
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			logBuffer := &bytes.Buffer{}
			app := simba.New(settings.Settings{
				Logging: logging.Config{
					Output: logBuffer,
				},
			})
			app.Router.POST("/test/{id}", simba.MultipartHandler(handler))
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

func TestAuthenticatedMultipartHandler(t *testing.T) {
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

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("lastName", "Connor")
	writer.WriteField("alive", "true")
	writer.Close()

	t.Run("authenticated multipart handler", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.MultipartRequest[test.Params], authModel *test.User) (*simba.Response, error) {
			assert.Equal(t, 1, authModel.ID)
			assert.Equal(t, "John Doe", authModel.Name)
			assert.Equal(t, "admin", authModel.Role)
			return &simba.Response{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1&size=10&active=true", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		app := simba.NewAuthWith(authFunc, settings.Settings{
			Logging: logging.Config{
				Output: logBuffer,
			},
		})
		app.Router.POST("/test/{id}", simba.AuthMultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("authenticated multipart handler with error", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.MultipartRequest[test.Params], user *test.User) (*simba.Response, error) {
			return &simba.Response{}, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/test/1", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		app := simba.NewAuthWith(errorAuthFunc, settings.Settings{
			Logging: logging.Config{
				Output: logBuffer,
			},
		})
		app.Router.POST("/test/{id}", simba.AuthMultipartHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResponse simba.HTTPError
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, "unauthorized", errorResponse.Message)
	})
}
