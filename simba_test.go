package simba_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/logging"
	"gotest.tools/v3/assert"
)

func TestSettingOptions(t *testing.T) {
	t.Parallel()

	t.Run("default options", func(t *testing.T) {
		router := simba.New()
		assert.Equal(t, router.GetOptions().RequestDisallowUnknownFields, true)
	})

	t.Run("set disallow unknown fields", func(t *testing.T) {
		options := simba.Options{
			RequestDisallowUnknownFields: false,
		}
		router := simba.New(options)

		assert.Equal(t, router.GetOptions().RequestDisallowUnknownFields, options.RequestDisallowUnknownFields)
	})

	t.Run("set request id accept header", func(t *testing.T) {
		options := simba.Options{
			RequestIdAcceptHeader: true,
		}
		router := simba.New(options)
		assert.Equal(t, router.GetOptions().RequestIdAcceptHeader, options.RequestIdAcceptHeader)
	})

	t.Run("set log request body", func(t *testing.T) {
		options := simba.Options{
			LogRequestBody: true,
		}
		router := simba.New(options)
		assert.Equal(t, router.GetOptions().LogRequestBody, options.LogRequestBody)
	})

	t.Run("set log level", func(t *testing.T) {
		options := simba.Options{
			LogLevel: zerolog.DebugLevel,
		}
		router := simba.New(options)
		assert.Equal(t, router.GetOptions().LogLevel, options.LogLevel)
	})

	t.Run("set log format", func(t *testing.T) {
		options := simba.Options{
			LogFormat: logging.TextFormat,
		}
		router := simba.New(options)
		assert.Equal(t, router.GetOptions().LogFormat, options.LogFormat)
	})

	t.Run("set log output", func(t *testing.T) {
		options := simba.Options{
			LogOutput: os.Stderr,
		}
		router := simba.New(options)
		assert.Equal(t, router.GetOptions().LogOutput, options.LogOutput)
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
		app.Use(middleware)
		app.GET("/test", simba.HandlerFunc(handler))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		app.GetRouter().ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
