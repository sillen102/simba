package simbaErrors_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaTestAssert"
)

func TestHandleError(t *testing.T) {
	t.Parallel()

	t.Run("log wrapped error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		ctx := context.WithValue(req.Context(), simbaContext.LoggerKey, logger)
		req = req.WithContext(ctx)

		simbaErrors.WriteError(w, req, simbaErrors.WrapError(
			http.StatusInternalServerError,
			fmt.Errorf("outermost error: %w", fmt.Errorf("wrapping error: %w", errors.New("original error"))),
			"Internal server error"))

		simbaTestAssert.Equal(t, http.StatusInternalServerError, w.Code)
		simbaTestAssert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResponse simbaErrors.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		simbaTestAssert.NoError(t, err)
		simbaTestAssert.Equal(t, http.StatusInternalServerError, errorResponse.Status)
		simbaTestAssert.Equal(t, "Internal server error", errorResponse.Message)

		expectedLog := "wrapping error: original error"
		simbaTestAssert.Assert(t, strings.Contains(logBuffer.String(), expectedLog))
	})

	t.Run("unauthorized does not show wrapped error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		logBuffer := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(logBuffer, &slog.HandlerOptions{}))
		ctx := context.WithValue(req.Context(), simbaContext.LoggerKey, logger)
		req = req.WithContext(ctx)

		simbaErrors.WriteError(w, req, simbaErrors.WrapError(http.StatusUnauthorized, errors.New("wrapped error"), "Internal server error"))

		simbaTestAssert.Equal(t, http.StatusUnauthorized, w.Code)
		simbaTestAssert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResponse simbaErrors.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		simbaTestAssert.NoError(t, err)
		simbaTestAssert.Equal(t, http.StatusUnauthorized, errorResponse.Status)
		simbaTestAssert.Equal(t, "unauthorized", errorResponse.Message) // hide details of the error

		expectedLog := "wrapped error"
		simbaTestAssert.Assert(t, strings.Contains(logBuffer.String(), expectedLog))
	})
}
