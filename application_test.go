package simba_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestDefaultApplication(t *testing.T) {
	t.Parallel()

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
	}

	app := simba.Default()
	app.Router.GET("/test", simba.JsonHandler(handler))

	t.Run("creates default application", func(t *testing.T) {
		assert.Assert(t, app != nil)
		assert.Assert(t, app.Server != nil)
		assert.Assert(t, app.Router != nil)
		assert.Assert(t, app.Settings != nil)
	})

	t.Run("adds default endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		app.Router.Mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"status\":\"ok\"}", w.Body.String())
	})

	t.Run("applies default middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		app.Router.Mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Check if the middleware is applied by verifying the presence of a request ID header
		assert.Assert(t, w.Header().Get(simbaContext.TraceIDHeader) != "")
	})
}

func TestNewApplication(t *testing.T) {
	t.Parallel()

	t.Run("creates new application with provided Simba", func(t *testing.T) {
		opts := []settings.Option{
			settings.WithServerHost("localhost"),
			settings.WithServerPort(8080),
		}
		app := simba.New(opts...)

		assert.Assert(t, app != nil)
		assert.Assert(t, app.Server != nil)
		assert.Assert(t, app.Router != nil)
		assert.Assert(t, app.Settings != nil)
		assert.Equal(t, "localhost:8080", app.Server.Addr)
	})
}

func TestApplicationRegisterShutdownHook(t *testing.T) {
	t.Parallel()

	t.Run("runs registered hooks during stop", func(t *testing.T) {
		app := simba.New()

		var order []string
		app.RegisterShutdownHook(func(ctx context.Context) error {
			order = append(order, "first")
			return nil
		})
		app.RegisterShutdownHook(func(ctx context.Context) error {
			order = append(order, "second")
			return nil
		})

		err := app.Stop()
		assert.Nil(t, err)
		assert.Equal(t, 2, len(order))
		assert.Equal(t, "first", order[0])
		assert.Equal(t, "second", order[1])
	})

	t.Run("returns hook errors", func(t *testing.T) {
		app := simba.New()
		expectedErr := errors.New("shutdown failed")

		app.RegisterShutdownHook(func(ctx context.Context) error {
			return expectedErr
		})

		err := app.Stop()
		assert.Assert(t, err != nil)
		assert.Assert(t, errors.Is(err, expectedErr))
	})
}
