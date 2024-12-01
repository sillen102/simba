package simba_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/test"
	"gotest.tools/v3/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful handler", func(t *testing.T) {

		handler := func(ctx context.Context, req *simba.Request[test.RequestBody, test.Params]) (*simba.Response, error) {
			assert.Equal(t, "John", req.Params.Name)
			assert.Equal(t, 1, req.Params.ID)
			assert.Equal(t, int64(1), req.Params.Page)
			assert.Equal(t, int64(10), req.Params.Size)

			assert.Equal(t, "test", req.Body.Test)

			return &simba.Response{
				Headers: map[string][]string{"My-Header": {"header-value"}},
				Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
				Body:    map[string]string{"message": "success"},
				Status:  http.StatusOK,
			}, nil
		}

		body := strings.NewReader(`{"test": "test"}`)
		req := httptest.NewRequest(http.MethodPost, "/test/1?page=1&size=10", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("name", "John")
		w := httptest.NewRecorder()

		router := simba.NewRouter()
		router.POST("/test/:id", simba.HandlerFunc(handler))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "{\"message\":\"success\"}\n", w.Body.String())

		assert.Equal(t, "header-value", w.Header().Get("My-Header"))

		cookie := w.Result().Cookies()[0].Value
		assert.Equal(t, "cookie-value", cookie)
	})
}
