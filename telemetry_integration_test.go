package simba

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaModels"

	"go.opentelemetry.io/otel/trace"
)

func TestTelemetry_Integration_Disabled(t *testing.T) {
	app := New(
		settings.WithTelemetryEnabled(false),
	)

	// Add a simple handler
	app.Router.GET("/test", JsonHandler(func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
			Status: http.StatusOK,
			Body:   map[string]string{"status": "ok"},
		}, nil
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestTelemetry_Integration_EnabledWithTracing(t *testing.T) {
	app := Default(
		settings.WithTelemetryEnabled(true),
		settings.WithTracingExporter("stdout"),
		settings.WithMetricsEnabled(false),
	)
	defer func() {
		if app.telemetryProvider != nil {
			app.telemetryProvider.Shutdown(context.Background())
		}
	}()

	spanFound := false

	// Add a handler that checks for span context
	app.Router.GET("/test", JsonHandler(func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		// Check if span context exists
		spanCtx := trace.SpanContextFromContext(ctx)
		if spanCtx.IsValid() {
			spanFound = true
		}

		return &simbaModels.Response[map[string]string]{
			Status: http.StatusOK,
			Body:   map[string]string{"status": "ok"},
		}, nil
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !spanFound {
		t.Error("Expected span to be present in handler context")
	}
}

func TestTelemetry_Integration_TraceIDFromOTEL(t *testing.T) {
	app := Default(
		settings.WithTelemetryEnabled(true),
		settings.WithTracingExporter("stdout"),
		settings.WithMetricsEnabled(false),
	)
	defer func() {
		if app.telemetryProvider != nil {
			app.telemetryProvider.Shutdown(context.Background())
		}
	}()

	var capturedTraceID string

	// Add a handler that captures the trace ID
	app.Router.GET("/test", JsonHandler(func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		// Get trace ID from context (should be OTEL trace ID)
		capturedTraceID = simbaContext.GetTraceID(ctx)

		return &simbaModels.Response[map[string]string]{
			Status: http.StatusOK,
			Body:   map[string]string{"status": "ok"},
		}, nil
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify trace ID was set
	if capturedTraceID == "" {
		t.Error("Expected trace ID to be set")
	}

	// Verify X-Trace-Id header was set
	responseTraceID := w.Header().Get("X-Trace-Id")
	if responseTraceID == "" {
		t.Error("Expected X-Trace-Id header to be set in response")
	}

	// Verify they match
	if capturedTraceID != responseTraceID {
		t.Errorf("Expected context trace ID %q to match response header %q", capturedTraceID, responseTraceID)
	}

	// OTEL trace IDs are 32 characters (hex representation of 16 bytes)
	if len(capturedTraceID) != 32 {
		t.Errorf("Expected OTEL trace ID length 32, got %d (%s)", len(capturedTraceID), capturedTraceID)
	}
}

func TestTelemetry_Integration_BothTracingAndMetrics(t *testing.T) {
	app := Default(
		settings.WithTelemetryEnabled(true),
		settings.WithTracingExporter("stdout"),
		settings.WithMetricsEnabled(true),
	)
	defer func() {
		if app.telemetryProvider != nil {
			app.telemetryProvider.Shutdown(context.Background())
		}
	}()

	// Add a handler
	app.Router.GET("/test", JsonHandler(func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		return &simbaModels.Response[map[string]string]{
			Status: http.StatusOK,
			Body:   map[string]string{"result": "success"},
		}, nil
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestTelemetry_Integration_FallbackToUUIDWhenDisabled(t *testing.T) {
	app := Default(
		settings.WithTelemetryEnabled(false),
	)

	var capturedTraceID string

	// Add a handler that captures the trace ID
	app.Router.GET("/test", JsonHandler(func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[map[string]string], error) {
		capturedTraceID = simbaContext.GetTraceID(ctx)

		return &simbaModels.Response[map[string]string]{
			Status: http.StatusOK,
			Body:   map[string]string{"status": "ok"},
		}, nil
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify trace ID was set (should be UUID v7, which is 36 characters)
	if capturedTraceID == "" {
		t.Error("Expected trace ID to be set")
	}

	// UUID format is 36 characters (with dashes)
	if len(capturedTraceID) != 36 {
		t.Errorf("Expected UUID trace ID length 36, got %d (%s)", len(capturedTraceID), capturedTraceID)
	}
}
