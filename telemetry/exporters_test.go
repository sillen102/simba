package telemetry

import (
	"context"
	"testing"

	"github.com/sillen102/simba/telemetry/config"
)

func TestNewTraceExporter_OTLP(t *testing.T) {
	cfg := &config.TracingConfig{
		Exporter: "otlp",
		Endpoint: "localhost:4317",
		Insecure: true,
	}

	exporter, err := newTraceExporter(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create OTLP trace exporter: %v", err)
	}
	defer exporter.Shutdown(context.Background())

	if exporter == nil {
		t.Error("Expected non-nil exporter")
	}
}

func TestNewTraceExporter_Stdout(t *testing.T) {
	cfg := &config.TracingConfig{
		Exporter: "stdout",
	}

	exporter, err := newTraceExporter(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create stdout trace exporter: %v", err)
	}
	defer exporter.Shutdown(context.Background())

	if exporter == nil {
		t.Error("Expected non-nil exporter")
	}
}

func TestNewTraceExporter_Invalid(t *testing.T) {
	cfg := &config.TracingConfig{
		Exporter: "invalid",
	}

	exporter, err := newTraceExporter(context.Background(), cfg)
	if err == nil {
		if exporter != nil {
			exporter.Shutdown(context.Background())
		}
		t.Fatal("Expected error for invalid exporter")
	}

	if exporter != nil {
		t.Error("Expected nil exporter for invalid type")
	}
}

func TestNewMetricExporter_OTLP(t *testing.T) {
	cfg := &config.MetricsConfig{
		Exporter: "otlp",
		Endpoint: "localhost:4317",
		Insecure: true,
	}

	exporter, err := newMetricExporter(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create OTLP metric exporter: %v", err)
	}
	defer exporter.Shutdown(context.Background())

	if exporter == nil {
		t.Error("Expected non-nil exporter")
	}
}

func TestNewMetricExporter_Stdout(t *testing.T) {
	cfg := &config.MetricsConfig{
		Exporter: "stdout",
	}

	exporter, err := newMetricExporter(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create stdout metric exporter: %v", err)
	}
	defer exporter.Shutdown(context.Background())

	if exporter == nil {
		t.Error("Expected non-nil exporter")
	}
}

func TestNewMetricExporter_Invalid(t *testing.T) {
	cfg := &config.MetricsConfig{
		Exporter: "invalid",
	}

	exporter, err := newMetricExporter(context.Background(), cfg)
	if err == nil {
		if exporter != nil {
			exporter.Shutdown(context.Background())
		}
		t.Fatal("Expected error for invalid exporter")
	}

	if exporter != nil {
		t.Error("Expected nil exporter for invalid type")
	}
}
