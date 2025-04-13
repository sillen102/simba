package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaTestAssert"
)

type testHandler struct {
	logs []string
}

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	var sb strings.Builder
	r.Attrs(func(a slog.Attr) bool {
		sb.WriteString(a.Key)
		sb.WriteString("=")
		sb.WriteString(a.Value.String())
		sb.WriteString(" ")
		return true
	})
	h.logs = append(h.logs, sb.String())
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func TestPanicRecovery(t *testing.T) {
	t.Parallel()

	t.Run("recovers from panic and logs stack trace (text format)", func(t *testing.T) {
		handler := &testHandler{}
		logger := slog.New(handler)
		ctx := context.WithValue(context.Background(), simbaContext.LoggerKey, logger)

		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.PanicRecovery(httpHandler).ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusInternalServerError, w.Code)
		simbaTestAssert.Equal(t, "Internal Server Error\n", w.Body.String())
		simbaTestAssert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

		simbaTestAssert.Assert(t, len(handler.logs) > 0, "Expected logs to be recorded")
		logMsg := handler.logs[0]
		simbaTestAssert.Assert(t, strings.Contains(logMsg, "error=test panic"), "Log should contain panic message")
		simbaTestAssert.Assert(t, strings.Contains(logMsg, "stacktrace="), "Log should contain stack trace")
	})

	t.Run("recovers from panic and logs stack trace (JSON format)", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, nil))
		ctx := context.WithValue(context.Background(), simbaContext.LoggerKey, logger)

		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		middleware.PanicRecovery(httpHandler).ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusInternalServerError, w.Code)

		// Parse the JSON log output
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		simbaTestAssert.NoError(t, err, "Should be valid JSON")

		// Verify the structure
		simbaTestAssert.Equal(t, "test panic", logEntry["error"])
		simbaTestAssert.Assert(t, logEntry["stacktrace"] != nil, "Should have stacktrace")
	})

	t.Run("does not interfere with normal requests", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.PanicRecovery(handler).ServeHTTP(w, req)

		simbaTestAssert.Equal(t, http.StatusOK, w.Code)
		simbaTestAssert.Equal(t, "success", w.Body.String())
	})
}
