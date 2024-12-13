package simba_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
	"gotest.tools/v3/assert"
)

func TestDefaultApplication(t *testing.T) {
	t.Parallel()

	app := simba.Default()
	app.Router.Handle("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
		assert.Assert(t, w.Header().Get(simbaContext.RequestIDHeader) != "")
	})
}

func TestNewApplication(t *testing.T) {
	t.Parallel()

	t.Run("creates new application with provided Config", func(t *testing.T) {
		cfg := settings.Config{
			Server: settings.Server{
				Host: "localhost",
				Port: 8080,
			},
		}
		app := simba.New(cfg)

		assert.Assert(t, app != nil)
		assert.Assert(t, app.Server != nil)
		assert.Assert(t, app.Router != nil)
		assert.Assert(t, app.Settings != nil)
		assert.Equal(t, "localhost:8080", app.Server.Addr)
	})
}
