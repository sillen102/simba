package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sillen102/simba/settings"
)

// newTraceExporter creates a trace exporter based on configuration
func newTraceExporter(ctx context.Context, cfg *settings.TracingConfig) (sdktrace.SpanExporter, error) {
	switch cfg.Exporter {
	case "otlp":
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
		}

		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
		}

		return otlptracegrpc.New(ctx, opts...)

	case "stdout":
		return stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)

	default:
		return nil, fmt.Errorf("unsupported trace exporter: %s", cfg.Exporter)
	}
}

// newMetricExporter creates a metric exporter based on configuration
func newMetricExporter(ctx context.Context, cfg *settings.MetricsConfig) (sdkmetric.Exporter, error) {
	switch cfg.Exporter {
	case "otlp":
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		}

		if cfg.Insecure {
			opts = append(opts, otlpmetricgrpc.WithDialOption(
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			))
		}

		return otlpmetricgrpc.New(ctx, opts...)

	case "stdout":
		return stdoutmetric.New()

	default:
		return nil, fmt.Errorf("unsupported metric exporter: %s", cfg.Exporter)
	}
}
