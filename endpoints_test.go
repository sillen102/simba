package simba_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"gotest.tools/v3/assert"
)

func TestAddDefaultEndpoints(t *testing.T) {
	t.Parallel()

	t.Run("health check endpoint", func(t *testing.T) {
		app := simba.Default()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		app.Router.Mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"status\":\"ok\"}", w.Body.String())
	})
}
