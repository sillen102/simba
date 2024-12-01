package simba_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba"
	"gotest.tools/v3/assert"
)

func TestBaseHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("successful basic handler", func(t *testing.T) {

		handler := simba.BaseHandler[RequestBody](func(ctx context.Context, req *simba.Request[RequestBody]) (*simba.Response, error) {
			assert.Equal(t, "test", req.Body.Test)

			return &simba.Response{
				Body:   map[string]string{"message": "success"},
				Status: http.StatusOK,
			}, nil
		})

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test", body)
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test", handler)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
