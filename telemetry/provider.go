package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/sillen102/simba/settings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Provider manages OpenTelemetry tracer and meter providers
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	settings       *settings.Telemetry
}

// NewProvider creates and initializes a new telemetry provider
func NewProvider(ctx context.Context, appSettings *settings.Simba) (*Provider, error) {
	telemetrySettings := &appSettings.Telemetry

	// Use application name as service name if not explicitly set
	serviceName := telemetrySettings.ServiceName
	if serviceName == "" {
		serviceName = appSettings.Application.Name
	}

	// Use application version as service version if not explicitly set
	serviceVersion := telemetrySettings.ServiceVersion
	if serviceVersion == "" {
		serviceVersion = appSettings.Application.Version
	}

	// Create resource
	res, err := newResource(serviceName, serviceVersion, telemetrySettings.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := &Provider{
		settings: telemetrySettings,
	}

	// Initialize tracer provider if tracing is enabled
	if telemetrySettings.Tracing.Enabled {
		traceExporter, err := newTraceExporter(ctx, &telemetrySettings.Tracing)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}

		// Create tracer provider with sampling
		samplerOption := sdktrace.WithSampler(
			sdktrace.TraceIDRatioBased(telemetrySettings.Tracing.SamplingRate),
		)

		provider.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			samplerOption,
		)

		// Set global tracer provider
		otel.SetTracerProvider(provider.tracerProvider)
	}

	// Initialize meter provider if metrics are enabled
	if telemetrySettings.Metrics.Enabled {
		metricExporter, err := newMetricExporter(ctx, &telemetrySettings.Metrics)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}

		// Create meter provider with periodic reader
		reader := sdkmetric.NewPeriodicReader(
			metricExporter,
			sdkmetric.WithInterval(time.Duration(telemetrySettings.Metrics.ExportInterval)*time.Second),
		)

		provider.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(reader),
		)

		// Set global meter provider
		otel.SetMeterProvider(provider.meterProvider)
	}

	return provider, nil
}

// Shutdown gracefully shuts down the telemetry provider
func (p *Provider) Shutdown(ctx context.Context) error {
	var err error

	if p.tracerProvider != nil {
		if shutdownErr := p.tracerProvider.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("failed to shutdown tracer provider: %w", shutdownErr)
		}
	}

	if p.meterProvider != nil {
		if shutdownErr := p.meterProvider.Shutdown(ctx); shutdownErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to shutdown meter provider: %w", err, shutdownErr)
			} else {
				err = fmt.Errorf("failed to shutdown meter provider: %w", shutdownErr)
			}
		}
	}

	return err
}

// TracerProvider returns the underlying tracer provider
func (p *Provider) TracerProvider() trace.TracerProvider {
	if p.tracerProvider == nil {
		return otel.GetTracerProvider()
	}
	return p.tracerProvider
}

// MeterProvider returns the underlying meter provider
func (p *Provider) MeterProvider() metric.MeterProvider {
	if p.meterProvider == nil {
		return otel.GetMeterProvider()
	}
	return p.meterProvider
}

// Tracer returns a tracer with the given name
func (p *Provider) Tracer(name string) trace.Tracer {
	if p.tracerProvider == nil {
		return otel.Tracer(name)
	}
	return p.tracerProvider.Tracer(name)
}

// Meter returns a meter with the given name
func (p *Provider) Meter(name string) metric.Meter {
	if p.meterProvider == nil {
		return otel.Meter(name)
	}
	return p.meterProvider.Meter(name)
}
