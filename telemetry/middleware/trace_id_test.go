package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/simbaContext"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceIDFromOTel(t *testing.T) {
	t.Parallel()

	t.Run("injects simba trace id from valid span context", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := simbaContext.GetTraceID(r.Context()); got != "00112233445566778899aabbccddeeff" {
				t.Fatalf("trace id = %q, want %q", got, "00112233445566778899aabbccddeeff")
			}
			w.WriteHeader(http.StatusOK)
		})

		spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    trace.TraceID{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			SpanID:     trace.SpanID{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17},
			TraceFlags: trace.FlagsSampled,
			Remote:     true,
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(trace.ContextWithSpanContext(req.Context(), spanCtx))
		w := httptest.NewRecorder()

		TraceIDFromOTel(handler).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("does nothing when span context is invalid", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := simbaContext.GetTraceID(r.Context()); got != "" {
				t.Fatalf("trace id = %q, want empty", got)
			}
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		TraceIDFromOTel(handler).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("preserves existing simba trace id when span context is invalid", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := simbaContext.GetTraceID(r.Context()); got != "existing-trace-id" {
				t.Fatalf("trace id = %q, want %q", got, "existing-trace-id")
			}
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(simbaContext.WithTraceID(req.Context(), "existing-trace-id"))
		w := httptest.NewRecorder()

		TraceIDFromOTel(handler).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
		}
	})
}
