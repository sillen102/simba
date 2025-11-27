package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/telemetry"

	"go.opentelemetry.io/otel/trace"
)

func TestOtelTracing_Disabled(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: false,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := OtelTracing(nil, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestOtelTracing_TracingDisabled(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: true,
		Tracing: settings.TracingConfig{
			Enabled: false,
		},
	}

	appSettings := &settings.Simba{
		Application: settings.Application{
			Name:    "test-app",
			Version: "1.0.0",
		},
		Telemetry: *telemetrySettings,
	}

	provider, err := telemetry.NewProvider(context.Background(), appSettings)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := OtelTracing(provider, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestOtelTracing_Enabled(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: true,
		Tracing: settings.TracingConfig{
			Enabled:  true,
			Exporter: "stdout",
		},
	}

	appSettings := &settings.Simba{
		Application: settings.Application{
			Name:    "test-app",
			Version: "1.0.0",
		},
		Telemetry: *telemetrySettings,
	}

	provider, err := telemetry.NewProvider(context.Background(), appSettings)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	spanCreated := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if a span was created in the context
		spanCtx := trace.SpanContextFromContext(r.Context())
		if spanCtx.IsValid() {
			spanCreated = true
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := OtelTracing(provider, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !spanCreated {
		t.Error("Expected span to be created when tracing is enabled")
	}
}

func TestOtelTracing_NilProvider(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: true,
		Tracing: settings.TracingConfig{
			Enabled: true,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// With nil provider, should return pass-through
	middleware := OtelTracing(nil, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
