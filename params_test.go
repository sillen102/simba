package simba_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	t.Run("validation errors", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, TestAllParamTypes]) (*simba.Response, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test/"+uuid.New().String()+"/invalid/test", nil)
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.GET("/test/:uuid/:id/:slug", simba.HandlerFunc(handler))
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse simba.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
		assert.Equal(t, "invalid parameter value", errorResponse.Message)

		assert.Equal(t, 1, len(errorResponse.ValidationErrors))
		assert.Equal(t, "PathID", errorResponse.ValidationErrors[0].Parameter)
		assert.Equal(t, simba.ParameterTypePath, errorResponse.ValidationErrors[0].Type)
		assert.Equal(t, "invalid parameter value: invalid", errorResponse.ValidationErrors[0].Message)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, TestAllParamTypes]) (*simba.Response, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test/invalid-uuid/123/test", nil)
		req.Header.Set("X-String", "test")
		req.Header.Set("X-Int", "1")

		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.GET("/test/:uuid/:id/:slug", simba.HandlerFunc(handler))
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse simba.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
		assert.Equal(t, "invalid UUID parameter value", errorResponse.Message)

		assert.Equal(t, 1, len(errorResponse.ValidationErrors))
		assert.Equal(t, "PathUUID", errorResponse.ValidationErrors[0].Parameter)
		assert.Equal(t, simba.ParameterTypePath, errorResponse.ValidationErrors[0].Type)
		assert.Equal(t, "invalid UUID parameter value: invalid-uuid", errorResponse.ValidationErrors[0].Message)
	})

	t.Run("invalid type conversion", func(t *testing.T) {
		handler := func(ctx context.Context, req *simba.Request[simba.NoBody, TestAllParamTypes]) (*simba.Response, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test/"+uuid.New().String()+"/123/test?page=invalid", nil)
		req.Header.Set("X-String", "test")
		req.Header.Set("X-Int", "not-a-number")

		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.GET("/test/:uuid/:id/:slug", simba.HandlerFunc(handler))
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse simba.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
		assert.Equal(t, "invalid parameter value", errorResponse.Message)

		assert.Equal(t, 1, len(errorResponse.ValidationErrors))
		assert.Equal(t, "HeaderInt", errorResponse.ValidationErrors[0].Parameter)
		assert.Equal(t, simba.ParameterTypeHeader, errorResponse.ValidationErrors[0].Type)
		assert.Equal(t, "invalid parameter value: not-a-number", errorResponse.ValidationErrors[0].Message)
	})
}
