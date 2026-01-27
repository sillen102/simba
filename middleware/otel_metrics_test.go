package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/telemetry"
)

func TestOtelMetrics_Disabled(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: false,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := OtelMetrics(nil, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestOtelMetrics_MetricsDisabled(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: true,
		Metrics: settings.MetricsConfig{
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
		w.Write([]byte("test response"))
	})

	middleware := OtelMetrics(provider, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestOtelMetrics_Enabled(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: true,
		Metrics: settings.MetricsConfig{
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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := OtelMetrics(provider, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response was written
	if w.Body.Len() == 0 {
		t.Error("Expected response body to be written")
	}
}

func TestOtelMetrics_NilProvider(t *testing.T) {
	telemetrySettings := &settings.Telemetry{
		Enabled: true,
		Metrics: settings.MetricsConfig{
			Enabled: true,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	// With nil provider, should return pass-through
	middleware := OtelMetrics(nil, telemetrySettings)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestMetricsResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	mw := &metricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	mw.WriteHeader(http.StatusNotFound)

	if mw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", mw.statusCode)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected underlying writer status 404, got %d", w.Code)
	}
}

func TestMetricsResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	mw := &metricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	data := []byte("test data")
	n, err := mw.Write(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	if mw.bytesWritten != int64(len(data)) {
		t.Errorf("Expected bytesWritten %d, got %d", len(data), mw.bytesWritten)
	}

	if w.Body.String() != string(data) {
		t.Errorf("Expected body %q, got %q", data, w.Body.String())
	}
}

func TestMetricsResponseWriter_Status(t *testing.T) {
	mw := &metricsResponseWriter{
		statusCode: http.StatusAccepted,
	}

	status := mw.Status()
	if status != "202" {
		t.Errorf("Expected status %q, got %q", "202", status)
	}
}
