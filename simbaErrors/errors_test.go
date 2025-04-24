package simbaErrors_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaTest/assert"
)

func TestNewSimbaError(t *testing.T) {
	t.Parallel()

	err := simbaErrors.NewSimbaError(http.StatusBadRequest, "test error", errors.New("internal error"))
	assert.Equal(t, http.StatusBadRequest, err.StatusCode())
	assert.Equal(t, "test error", err.PublicMessage())
	assert.Equal(t, "internal error", err.Error())
}

func TestSimbaErrorWithDetails(t *testing.T) {
	t.Parallel()

	details := map[string]string{"field": "value"}
	err := simbaErrors.NewSimbaError(http.StatusBadRequest, "test error", nil).WithDetails(details)
	if detailsMap, ok := err.Details().(map[string]string); ok {
		assert.Equal(t, details, detailsMap)
	} else {
		t.Errorf("expected details to be of type map[string]string")
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	err := simbaErrors.NewSimbaError(http.StatusBadRequest, "test error", errors.New("internal error"))
	simbaErrors.WriteError(w, req, err)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var errorResponse simbaErrors.ErrorResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&errorResponse))
	assert.Equal(t, http.StatusBadRequest, errorResponse.Status)
	assert.Equal(t, "Bad Request", errorResponse.Error)
	assert.Equal(t, "/test", errorResponse.Path)
	assert.Equal(t, http.MethodGet, errorResponse.Method)
	assert.Equal(t, "test error", errorResponse.Message)
}

func TestHandleUnexpectedError(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	simbaErrors.HandleUnexpectedError(w)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	assert.Equal(t, http.StatusBadRequest, simbaErrors.ErrInvalidContentType.StatusCode())
	assert.Equal(t, "invalid content type", simbaErrors.ErrInvalidContentType.PublicMessage())

	assert.Equal(t, http.StatusBadRequest, simbaErrors.ErrInvalidRequestBody.StatusCode())
	assert.Equal(t, "invalid request body", simbaErrors.ErrInvalidRequestBody.PublicMessage())

	assert.Equal(t, http.StatusUnauthorized, simbaErrors.ErrUnauthorized.StatusCode())
	assert.Equal(t, "unauthorized", simbaErrors.ErrUnauthorized.PublicMessage())

	assert.Equal(t, http.StatusInternalServerError, simbaErrors.ErrUnexpected.StatusCode())
	assert.Equal(t, "unexpected error", simbaErrors.ErrUnexpected.PublicMessage())
}
