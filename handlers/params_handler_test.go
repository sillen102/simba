package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/handlers"
	"github.com/sillen102/simba/internal/test/assert"
)

func TestParamsHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("successful params handler", func(t *testing.T) {

		handler := handlers.ParamsHandler[RequestBody, Params](func(ctx context.Context, req *simba.Request[RequestBody], params Params) (*simba.Response, error) {
			assert.Equal(t, "test", req.Body.Test)
			assert.Equal(t, int64(1), params.Page)
			assert.Equal(t, int64(10), params.Size)
			assert.Equal(t, "world", params.Name)

			return &simba.Response{
				Body:   map[string]string{"message": "success"},
				Status: http.StatusOK,
			}, nil
		})

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodGet, "/test/hello?name=world", body)
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.GET("/test/:id", handler)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
