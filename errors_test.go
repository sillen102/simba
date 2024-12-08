package simba_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba"
	"gotest.tools/v3/assert"
)

func TestHandleError(t *testing.T) {
	t.Parallel()

	t.Run("log wrapped error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := simba.NewLogger(&simba.LoggingConfig{
			Output: logBuffer,
			Format: simba.TextFormat,
		})
		ctx := logger.WithContext(req.Context())
		req = req.WithContext(ctx)

		simba.HandleError(w, req, simba.WrapErrorHTTP(http.StatusInternalServerError, errors.New("wrapped error"), "Internal server error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResponse simba.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Status)
		assert.Equal(t, "Internal server error", errorResponse.Message)

		expectedLog := "wrapped error"
		assert.Assert(t, strings.Contains(logBuffer.String(), expectedLog))
	})

	t.Run("unauthorized does not show wrapped error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := simba.NewLogger(&simba.LoggingConfig{
			Output: logBuffer,
			Format: simba.TextFormat,
		})
		ctx := logger.WithContext(req.Context())
		req = req.WithContext(ctx)

		simba.HandleError(w, req, simba.WrapErrorHTTP(http.StatusUnauthorized, errors.New("wrapped error"), "Internal server error"))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResponse simba.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, http.StatusUnauthorized, errorResponse.Status)
		assert.Equal(t, "unauthorized", errorResponse.Message) // hide details of the error

		expectedLog := "wrapped error"
		assert.Assert(t, strings.Contains(logBuffer.String(), expectedLog))
	})

	t.Run("forbidden does not show wrapped error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := simba.NewLogger(&simba.LoggingConfig{
			Output: logBuffer,
			Format: simba.TextFormat,
		})
		ctx := logger.WithContext(req.Context())
		req = req.WithContext(ctx)

		simba.HandleError(w, req, simba.WrapErrorHTTP(http.StatusForbidden, errors.New("wrapped error"), "Internal server error"))

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResponse simba.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NilError(t, err)
		assert.Equal(t, http.StatusForbidden, errorResponse.Status)
		assert.Equal(t, "forbidden", errorResponse.Message) // hide details of the error

		expectedLog := "wrapped error"
		assert.Assert(t, strings.Contains(logBuffer.String(), expectedLog))
	})
}
