package simba_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sillen102/simba"
	"gotest.tools/v3/assert"
)

type TestAllParamTypes struct {
	// Header parameters
	HeaderString string    `header:"X-String" validate:"required"`
	HeaderInt    int       `header:"X-Int" validate:"required"`
	HeaderBool   bool      `header:"X-Bool"`
	HeaderUUID   uuid.UUID `header:"X-UUID"`

	// Path parameters
	PathID   int       `path:"id" validate:"required"`
	PathSlug string    `path:"slug"`
	PathUUID uuid.UUID `path:"uuid"`

	// Query parameters
	QueryPage    int       `query:"page" default:"1"`
	QuerySize    int       `query:"size" default:"10"`
	QueryFilter  string    `query:"filter"`
	QueryEnabled bool      `query:"enabled" default:"true"`
	QueryDate    time.Time `query:"date"`
}

func TestParamParsing(t *testing.T) {
	t.Parallel()

	t.Run("all parameter types", func(t *testing.T) {
		testUUID := uuid.New()
		testDate := time.Now().UTC().Truncate(time.Second)

		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, TestAllParamTypes]) (*simba.Response, error) {
			// Verify header parameters
			assert.Equal(t, "test-string", req.Params.HeaderString)
			assert.Equal(t, 42, req.Params.HeaderInt)
			assert.Equal(t, true, req.Params.HeaderBool)
			assert.Equal(t, testUUID, req.Params.HeaderUUID)

			// Verify path parameters
			assert.Equal(t, 123, req.Params.PathID)
			assert.Equal(t, "test-slug", req.Params.PathSlug)
			assert.Equal(t, testUUID, req.Params.PathUUID)

			// Verify query parameters
			assert.Equal(t, 2, req.Params.QueryPage)
			assert.Equal(t, 20, req.Params.QuerySize)
			assert.Equal(t, "active", req.Params.QueryFilter)
			assert.Equal(t, true, req.Params.QueryEnabled)
			assert.Equal(t, testDate, req.Params.QueryDate)

			return &simba.Response{Status: http.StatusOK}, nil
		}

		// Create request with all parameter types
		path := "/test/" + testUUID.String() + "/123/test-slug"
		req := httptest.NewRequest(http.MethodGet, path+"?page=2&size=20&filter=active&date="+testDate.Format(time.RFC3339), nil)

		// Set headers
		req.Header.Set("X-String", "test-string")
		req.Header.Set("X-Int", "42")
		req.Header.Set("X-Bool", "true")
		req.Header.Set("X-UUID", testUUID.String())

		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.GET("/test/:uuid/:id/:slug", simba.HandlerFunc(handler))
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("default values", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, TestAllParamTypes]) (*simba.Response, error) {
			assert.Equal(t, 1, req.Params.QueryPage)
			assert.Equal(t, 10, req.Params.QuerySize)
			assert.Equal(t, true, req.Params.QueryEnabled)
			return &simba.Response{Status: http.StatusOK}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test/"+uuid.New().String()+"/123/test", nil)
		req.Header.Set("X-String", "test")
		req.Header.Set("X-Int", "1")

		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.GET("/test/:uuid/:id/:slug", simba.HandlerFunc(handler))
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestUUIDParameters(t *testing.T) {
	t.Parallel()

	type UUIDParams struct {
		ID       uuid.UUID `path:"id"`
		HeaderID uuid.UUID `header:"Header-ID"`
		QueryID  uuid.UUID `query:"queryId"`
	}

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, UUIDParams]) (*simba.Response, error) {
		return &simba.Response{Status: http.StatusOK}, nil
	}

	tests := []struct {
		name     string
		path     string
		headerID string
		wantMsg  string
	}{
		{
			name:    "invalid uuid in path",
			path:    "/test/invalid-uuid",
			wantMsg: "invalid UUID parameter value",
		},
		{
			name:     "invalid uuid in header",
			path:     "/test/123e4567-e89b-12d3-a456-426655440000",
			headerID: "invalid-uuid",
			wantMsg:  "invalid UUID parameter value",
		},
		{
			name:     "invalid uuid in query",
			path:     "/test/123e4567-e89b-12d3-a456-426655440000?queryId=invalid-uuid",
			headerID: "248ccd0e-4bdf-4c41-a125-92ef3a416251",
			wantMsg:  "invalid UUID parameter value",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.headerID != "" {
				req.Header.Set("Header-ID", tt.headerID)
			}
			w := httptest.NewRecorder()

			app := simba.New()
			app.Router.GET("/test/:id", simba.HandlerFunc(handler))
			app.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var errorResponse simba.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			assert.NilError(t, err)

			assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
			assert.Equal(t, "Bad Request", errorResponse.Error)
			expectedPath := tt.path
			if idx := strings.Index(expectedPath, "?"); idx != -1 {
				expectedPath = expectedPath[:idx]
			}
			assert.Equal(t, expectedPath, errorResponse.Path)
			assert.Equal(t, http.MethodGet, errorResponse.Method)
			assert.Equal(t, tt.wantMsg, errorResponse.Message)
		})
	}
}

func TestFloatParameters(t *testing.T) {
	t.Parallel()

	type FloatParams struct {
		Page float64 `query:"page"`
	}

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, FloatParams]) (*simba.Response, error) {
		return &simba.Response{Status: http.StatusOK}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test/1?page=invalid", nil)
	w := httptest.NewRecorder()

	app := simba.New()
	app.Router.GET("/test/:id", simba.HandlerFunc(handler))
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResponse simba.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResponse)
	assert.NilError(t, err)

	assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
	assert.Equal(t, "Bad Request", errorResponse.Error)
	assert.Equal(t, "/test/1", errorResponse.Path)
	assert.Equal(t, http.MethodGet, errorResponse.Method)
	assert.Equal(t, "invalid parameter value", errorResponse.Message)
}

func TestInvalidParameterTypes(t *testing.T) {
	t.Parallel()

	type Params struct {
		Page    int       `query:"page"`
		Size    int       `query:"size"`
		Score   float64   `query:"score"`
		Active  bool      `query:"active"`
		ID      int       `path:"id"`
		Header  string    `header:"Header"`
		Header2 uuid.UUID `header:"Header2"`
	}

	handler := func(ctx context.Context, req *simba.Request[simba.NoBody, Params]) (*simba.Response, error) {
		t.Error("handler should not be called")
		return &simba.Response{}, nil
	}

	tests := []struct {
		name string
		path string
	}{
		{
			name: "invalid page parameter",
			path: "/test/1?active=true&page=invalid",
		},
		{
			name: "invalid size parameter",
			path: "/test/1?active=true&size=invalid",
		},
		{
			name: "invalid score parameter",
			path: "/test/1?active=true&score=invalid",
		},
		{
			name: "invalid active parameter",
			path: "/test/1?active=notbool",
		},
		{
			name: "invalid id parameter",
			path: "/test/notint?active=true",
		},
	}

	app := simba.New()
	app.Router.GET("/test/:id", simba.HandlerFunc(handler))

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var errorResponse simba.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			assert.NilError(t, err)

			assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
			assert.Equal(t, "Bad Request", errorResponse.Error)
			expectedPath := tt.path
			if idx := strings.Index(expectedPath, "?"); idx != -1 {
				expectedPath = expectedPath[:idx]
			}
			assert.Equal(t, expectedPath, errorResponse.Path)
			assert.Equal(t, http.MethodGet, errorResponse.Method)
			assert.Equal(t, "invalid parameter value", errorResponse.Message)
		})
	}
}
