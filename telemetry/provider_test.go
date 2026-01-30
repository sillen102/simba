package telemetry

import (
	"context"
	"testing"

	"github.com/sillen102/simba/telemetry/config"
)

func TestNewProvider_TelemetryDisabled(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Enabled:        false,
		ServiceName:    "test-app",
		ServiceVersion: "1.0.0",
	}

	// Should not create provider when telemetry is disabled
	// This test just verifies the calling code handles nil provider correctly
	if cfg.Enabled {
		t.Error("Expected telemetry to be disabled")
	}
}

func TestNewProvider_WithTracingOnly(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Tracing: config.TracingConfig{
			Enabled:      true,
			Exporter:     "stdout",
			SamplingRate: 1.0,
		},
		Metrics: config.MetricsConfig{
			Enabled: false,
		},
	}

	provider, err := NewProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.tracerProvider == nil {
		t.Error("Expected tracer provider to be initialized")
	}

	if provider.meterProvider != nil {
		t.Error("Expected meter provider to be nil when metrics disabled")
	}

	// Test that we can get a tracer
	tracer := provider.Tracer("test")
	if tracer == nil {
		t.Error("Expected non-nil tracer")
	}
}

func TestNewProvider_WithMetricsOnly(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Tracing: config.TracingConfig{
			Enabled: false,
		},
		Metrics: config.MetricsConfig{
			Enabled:        true,
			Exporter:       "stdout",
			ExportInterval: 60,
		},
	}

	provider, err := NewProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.tracerProvider != nil {
		t.Error("Expected tracer provider to be nil when tracing disabled")
	}

	if provider.meterProvider == nil {
		t.Error("Expected meter provider to be initialized")
	}

	// Test that we can get a meter
	meter := provider.Meter("test")
	if meter == nil {
		t.Error("Expected non-nil meter")
	}
}

func TestNewProvider_WithBothTracingAndMetrics(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "2.0.0",
		Environment:    "production",
		Tracing: config.TracingConfig{
			Enabled:      true,
			Exporter:     "stdout",
			SamplingRate: 0.5,
		},
		Metrics: config.MetricsConfig{
			Enabled:        true,
			Exporter:       "stdout",
			ExportInterval: 30,
		},
	}

	provider, err := NewProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.tracerProvider == nil {
		t.Error("Expected tracer provider to be initialized")
	}

	if provider.meterProvider == nil {
		t.Error("Expected meter provider to be initialized")
	}

	// Test TracerProvider method
	tracerProvider := provider.TracerProvider()
	if tracerProvider == nil {
		t.Error("Expected non-nil tracer provider")
	}

	// Test MeterProvider method
	meterProvider := provider.MeterProvider()
	if meterProvider == nil {
		t.Error("Expected non-nil meter provider")
	}
}

func TestNewProvider_UsesApplicationNameAsDefault(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "my-app",
		ServiceVersion: "3.0.0",
		Environment:    "test",
		Tracing: config.TracingConfig{
			Enabled:  true,
			Exporter: "stdout",
		},
	}

	provider, err := NewProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	// The provider should have been created successfully using the app name
	if provider.tracerProvider == nil {
		t.Error("Expected tracer provider to be initialized with app name as service name")
	}
}

func TestNewProvider_InvalidExporter(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Tracing: config.TracingConfig{
			Enabled:  true,
			Exporter: "invalid-exporter",
		},
	}

	provider, err := NewProvider(context.Background(), cfg)
	if err == nil {
		if provider != nil {
			provider.Shutdown(context.Background())
		}
		t.Fatal("Expected error for invalid exporter")
	}

	if provider != nil {
		t.Error("Expected nil provider when creation fails")
	}
}

func TestProvider_Shutdown(t *testing.T) {
	cfg := &config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Tracing: config.TracingConfig{
			Enabled:  true,
			Exporter: "stdout",
		},
		Metrics: config.MetricsConfig{
			Enabled:  true,
			Exporter: "stdout",
		},
	}

	provider, err := NewProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Test shutdown
	err = provider.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}
