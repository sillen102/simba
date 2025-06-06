package simba_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest"
	"github.com/sillen102/simba/simbaTest/assert"
)

type TestAllParamTypes struct {
	// Header parameters
	HeaderString string    `header:"X-String" validate:"required"`
	HeaderInt    int       `header:"X-Int" validate:"required"`
	HeaderBool   bool      `header:"X-Bool"`
	HeaderUUID   uuid.UUID `header:"X-UUID"`

	// path parameters
	PathID   int       `path:"id" validate:"required"`
	PathSlug string    `path:"slug"`
	PathUUID uuid.UUID `path:"uuid"`

	// Query parameters
	QueryPage    int       `query:"page" default:"1"`
	QuerySize    int       `query:"size" default:"10"`
	QueryFilter  string    `query:"filter"`
	QueryEnabled bool      `query:"enabled" default:"true"`
	QueryDate    time.Time `query:"date"`
	QuerySlice1  []string  `query:"slice1"`
	QuerySlice2  []string  `query:"slice2"`
}

func TestParamParsing(t *testing.T) {
	t.Parallel()

	t.Run("all parameter types", func(t *testing.T) {
		testUUID := uuid.New()
		testDate := time.Now().UTC().Truncate(time.Second)

		handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, TestAllParamTypes]) (*simbaModels.Response[simbaModels.NoBody], error) {
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

			// Verify query slice
			assert.Equal(t, []string{"one", "two"}, req.Params.QuerySlice1)
			assert.Equal(t, []string{"three", "four"}, req.Params.QuerySlice2)

			return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
		}

		// Create request with all parameter types
		path := fmt.Sprintf("/test/%s/%d/%s?page=%d&size=%d&filter=%s&enabled=%t&date=%s&slice1=one&slice1=two&slice2=three,four",
			testUUID.String(),
			123,
			"test-slug",
			2,
			20,
			"active",
			true,
			testDate.Format(time.RFC3339),
		)
		req := httptest.NewRequest(http.MethodGet, path, nil)

		// Set headers
		req.Header.Set("X-String", "test-string")
		req.Header.Set("X-Int", "42")
		req.Header.Set("X-Bool", "true")
		req.Header.Set("X-UUID", testUUID.String())

		w := httptest.NewRecorder()

		app := simbaTest.New()
		app.Router.GET("/test/{uuid}/{id}/{slug}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("default values", func(t *testing.T) {
		handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, TestAllParamTypes]) (*simbaModels.Response[simbaModels.NoBody], error) {
			assert.Equal(t, 1, req.Params.QueryPage)
			assert.Equal(t, 10, req.Params.QuerySize)
			assert.Equal(t, true, req.Params.QueryEnabled)
			return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test/"+uuid.New().String()+"/123/test", nil)
		req.Header.Set("X-String", "test")
		req.Header.Set("X-Int", "1")

		w := httptest.NewRecorder()

		app := simbaTest.New()
		app.Router.GET("/test/{uuid}/{id}/{slug}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("embedded struct parameters", func(t *testing.T) {
		type EmbeddedParams struct {
			QueryParam   string `query:"queryParam" validate:"required"`
			DefaultParam string `query:"defaultParam" default:"default-value"`
		}

		handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, struct {
			EmbeddedParams
		}]) (*simbaModels.Response[simbaModels.NoBody], error) {
			assert.Equal(t, "test-param", req.Params.QueryParam)
			assert.Equal(t, "default-value", req.Params.DefaultParam)
			return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test/"+uuid.New().String()+"?queryParam=test-param", nil)

		w := httptest.NewRecorder()

		app := simbaTest.New()
		app.Router.GET("/test/{uuid}", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

type ValidationTestParams struct {
	Email     string `query:"email" validate:"omitempty,email"`
	Password  string `query:"password" validate:"omitempty,min=8,max=32"`
	Code      string `query:"code" validate:"omitempty,len=6"`
	Pin       string `query:"pin" validate:"omitempty,numeric"`
	Username  string `query:"username" validate:"omitempty,alphanum"`
	Page      int    `query:"page" default:"1"`
	Size      int    `query:"size" default:"10"`
	SortOrder string `query:"sort" default:"asc"`
}

func TestValidationRules(t *testing.T) {
	t.Parallel()

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, ValidationTestParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	tests := []struct {
		name            string
		query           string
		expectedError   string
		expectedDetails any
		parameter       string
	}{
		{
			name:            "invalid email",
			query:           "?email=notanemail",
			expectedError:   "'notanemail' is not a valid email address",
			expectedDetails: []any{"'notanemail' is not a valid email address"},
			parameter:       "email",
		},
		{
			name:            "password too short",
			query:           "?password=short",
			expectedError:   "password must be at least 8 characters long",
			expectedDetails: []any{"password must be at least 8 characters long"},
			parameter:       "password",
		},
		{
			name:            "password too long",
			query:           "?password=thispasswordiswaytoolongandshouldfail",
			expectedError:   "password must not exceed 32 characters",
			expectedDetails: []any{"password must not exceed 32 characters"},
			parameter:       "password",
		},
		{
			name:            "invalid code length",
			query:           "?code=12345",
			expectedError:   "code must be exactly 6 characters long",
			expectedDetails: []any{"code must be exactly 6 characters long"},
			parameter:       "code",
		},
		{
			name:            "non-numeric pin",
			query:           "?pin=abc123",
			expectedError:   "'abc123' must be a valid number",
			expectedDetails: []any{"'abc123' must be a valid number"},
			parameter:       "pin",
		},
		{
			name:            "non-alphanumeric username",
			query:           "?username=user@name",
			expectedError:   "'user@name' must contain only letters and numbers",
			expectedDetails: []any{"'user@name' must contain only letters and numbers"},
			parameter:       "username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			w := httptest.NewRecorder()

			app := simba.New()
			app.Router.GET("/test", simba.JsonHandler(handler))
			app.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errorResponse simbaErrors.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
			assert.Equal(t, "request validation failed, 1 validation error", errorResponse.Message)
			assert.Equal(t, tt.expectedDetails, errorResponse.Details)
		})
	}
}

func TestDefaultValues(t *testing.T) {
	t.Parallel()

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, ValidationTestParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		assert.Equal(t, 1, req.Params.Page)
		assert.Equal(t, 10, req.Params.Size)
		assert.Equal(t, "asc", req.Params.SortOrder)
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusNoContent}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	app := simba.New()
	app.Router.GET("/test", simba.JsonHandler(handler))
	app.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestUUIDParameters(t *testing.T) {
	t.Parallel()

	type UUIDParams struct {
		ID       uuid.UUID `path:"id"`
		HeaderID uuid.UUID `header:"Header-ID"`
		QueryID  uuid.UUID `query:"queryId"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, UUIDParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
	}

	tests := []struct {
		name      string
		path      string
		parameter string
		headerID  string
		wantMsg   string
	}{
		{
			name:      "invalid uuid in path",
			path:      "/test/invalid-uuid",
			parameter: "id",
			wantMsg:   "invalid UUID parameter value: invalid-uuid",
		},
		{
			name:      "invalid uuid in header",
			path:      "/test/123e4567-e89b-12d3-a456-426655440000",
			parameter: "Header-ID",
			headerID:  "invalid-uuid",
			wantMsg:   "invalid UUID parameter value: invalid-uuid",
		},
		{
			name:      "invalid uuid in query",
			path:      "/test/123e4567-e89b-12d3-a456-426655440000?queryId=invalid-uuid",
			parameter: "queryId",
			headerID:  "248ccd0e-4bdf-4c41-a125-92ef3a416251",
			wantMsg:   "invalid UUID parameter value: invalid-uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.headerID != "" {
				req.Header.Set("Header-ID", tt.headerID)
			}
			w := httptest.NewRecorder()

			app := simba.New()
			app.Router.GET("/test/{id}", simba.JsonHandler(handler))
			app.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errorResponse simbaErrors.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
			assert.Equal(t, "Bad Request", errorResponse.Error)
			expectedPath := tt.path
			if idx := strings.Index(expectedPath, "?"); idx != -1 {
				expectedPath = expectedPath[:idx]
			}
			assert.Equal(t, expectedPath, errorResponse.Path)
			assert.Equal(t, http.MethodGet, errorResponse.Method)
			assert.Equal(t, "request validation failed, 1 validation error", errorResponse.Message)
		})
	}
}

func TestFloatParameters(t *testing.T) {
	t.Parallel()

	type FloatParams struct {
		ID   int     `path:"id"`
		Page float64 `query:"page"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, FloatParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test/1?page=invalid", nil)
	w := httptest.NewRecorder()

	app := simba.New()
	app.Router.GET("/test/{id}", simba.JsonHandler(handler))
	app.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResponse simbaErrors.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResponse)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
	assert.Equal(t, "Bad Request", errorResponse.Error)
	assert.Equal(t, "/test/1", errorResponse.Path)
	assert.Equal(t, http.MethodGet, errorResponse.Method)
	assert.Equal(t, "request validation failed, 1 validation error", errorResponse.Message)
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
		Cookie  string    `cookie:"Cookie"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, Params]) (*simbaModels.Response[simbaModels.NoBody], error) {
		t.Error("handler should not be called")
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	tests := []struct {
		name         string
		path         string
		paramName    string
		errorMessage string
	}{
		{
			name:         "invalid page parameter",
			path:         "/test/1?active=true&page=invalid",
			paramName:    "page",
			errorMessage: "invalid int parameter value: invalid",
		},
		{
			name:         "invalid size parameter",
			path:         "/test/1?active=true&size=invalid",
			paramName:    "size",
			errorMessage: "invalid int parameter value: invalid",
		},
		{
			name:         "invalid score parameter",
			path:         "/test/1?active=true&score=invalid",
			paramName:    "score",
			errorMessage: "invalid float parameter value: invalid",
		},
		{
			name:         "invalid active parameter",
			path:         "/test/1?active=notbool",
			paramName:    "active",
			errorMessage: "invalid bool parameter value: notbool",
		},
		{
			name:         "invalid id parameter",
			path:         "/test/notint?active=true",
			paramName:    "id",
			errorMessage: "invalid int parameter value: notint",
		},
	}

	app := simba.New()
	app.Router.GET("/test/{id}", simba.JsonHandler(handler))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var errorResponse simbaErrors.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
			assert.Equal(t, "Bad Request", errorResponse.Error)
			expectedPath := tt.path
			if idx := strings.Index(expectedPath, "?"); idx != -1 {
				expectedPath = expectedPath[:idx]
			}
			assert.Equal(t, expectedPath, errorResponse.Path)
			assert.Equal(t, http.MethodGet, errorResponse.Method)
			assert.Equal(t, "request validation failed, 1 validation error", errorResponse.Message)
		})
	}
}

func TestTimeParameters(t *testing.T) {
	t.Parallel()

	type TimeParams struct {
		CustomTime  time.Time `query:"customTime" format:"2006-01-02"`
		DefaultTime time.Time `query:"defaultTime"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, TimeParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		expectedDefaultTime, _ := time.Parse(time.RFC3339, "2023-10-15T14:00:00Z")
		expectedCustomTime, _ := time.Parse("2006-01-02", "2023-10-15")

		assert.Equal(t, expectedDefaultTime, req.Params.DefaultTime)
		assert.Equal(t, expectedCustomTime, req.Params.CustomTime)
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test?defaultTime=2023-10-15T14:00:00Z&customTime=2023-10-15", nil)
	w := httptest.NewRecorder()

	app := simba.New()
	app.Router.GET("/test", simba.JsonHandler(handler))
	app.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// CustomID is a type that implements TextMarshaler and TextUnmarshaler
type CustomID string

// MarshalText implements the TextMarshaler interface
func (c CustomID) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("custom-%s", string(c))), nil
}

// UnmarshalText implements the TextUnmarshaler interface
func (c *CustomID) UnmarshalText(text []byte) error {
	if !bytes.HasPrefix(text, []byte("custom-")) {
		return fmt.Errorf("invalid CustomID format: %s", text)
	}
	*c = CustomID(bytes.TrimPrefix(text, []byte("custom-")))
	return nil
}

func TestTextUnmarshalerParameters(t *testing.T) {
	t.Parallel()

	type TextUnmarshalerParams struct {
		ID CustomID `query:"id"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, TextUnmarshalerParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		assert.Equal(t, CustomID("123"), req.Params.ID)
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
	}

	// Test valid format
	t.Run("valid custom format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test?id=custom-123", nil)
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.GET("/test", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test invalid format
	t.Run("invalid custom format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test?id=invalid-123", nil)
		w := httptest.NewRecorder()

		app := simba.New()
		app.Router.GET("/test", simba.JsonHandler(handler))
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse simbaErrors.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
		assert.Equal(t, "request validation failed, 1 validation error", errorResponse.Message)
		assert.Equal(t, []any{"invalid value invalid-123 for id"}, errorResponse.Details.([]any))
	})
}

func TestCookieParams(t *testing.T) {
	t.Parallel()

	type cookieParams struct {
		Toke string `cookie:"token" validate:"required"`
	}

	validToken := "test-token"

	handler := func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, cookieParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{Status: http.StatusOK}, nil
	}

	app := simba.New()
	app.Router.GET("/test", simba.JsonHandler(handler))

	testCases := []struct {
		name           string
		cookie         *http.Cookie
		expectedStatus int
	}{
		{
			name:           "valid cookie",
			cookie:         &http.Cookie{Name: "token", Value: validToken},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing cookie",
			cookie:         nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty cookie",
			cookie:         &http.Cookie{Name: "token", Value: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not present cookie",
			cookie:         &http.Cookie{Name: "not-present", Value: "test"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
