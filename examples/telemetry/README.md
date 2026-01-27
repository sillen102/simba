# Simba Telemetry Example

This example demonstrates how to use Simba's comprehensive OpenTelemetry integration with a full observability stack. It showcases automatic HTTP tracing and metrics collection, as well as custom instrumentation patterns.

## Overview

This example provides:

- **Automatic HTTP Instrumentation**: Traces and metrics for all HTTP requests via middleware
- **Custom Spans**: Adding detailed tracing to business logic operations
- **Custom Metrics**: Recording application-specific metrics
- **Error Tracking**: Proper error handling and span status management
- **Full Observability Stack**: Complete setup with Jaeger, Prometheus, and Grafana

## What's Being Demonstrated

### Automatic Features (Zero Code Required)
- HTTP request tracing with W3C trace context propagation
- Request duration histograms (for percentile calculations)
- Request count by endpoint and status code
- Response size tracking
- Trace ID injection into logs and response headers

### Custom Instrumentation Examples
- Creating custom spans for database operations
- Adding attributes to spans (user IDs, operation types, etc.)
- Nested span creation (parent-child relationships)
- Error recording and span status management
- Creating custom metrics (counters, histograms)

## Architecture

```
┌─────────────────┐
│  Simba App      │
│  (Port 9999)    │
└────────┬────────┘
         │ OTLP (gRPC)
         ▼
┌─────────────────┐
│ OTEL Collector  │
│  (Port 4317)    │
└────┬────────┬───┘
     │        │
     │        └──────────────┐
     ▼                       ▼
┌─────────────┐      ┌──────────────┐
│   Jaeger    │      │  Prometheus  │
│ (Port 16686)│      │  (Port 9090) │
└─────────────┘      └──────┬───────┘
                            │
                            ▼
                     ┌──────────────┐
                     │   Grafana    │
                     │  (Port 3000) │
                     └──────────────┘
```

## Prerequisites

- Docker
- Docker Compose
- curl (for testing)

## Quick Start

1. **Start the entire stack**:
   ```bash
   cd examples/telemetry
   docker-compose up --build
   ```

   This will start:
   - Simba application
   - OpenTelemetry Collector
   - Jaeger (tracing)
   - Prometheus (metrics)
   - Grafana (visualization)

2. **Wait for services to be ready** (approximately 30 seconds)

3. **Access the UIs**:
   - Simba App: http://localhost:9999
   - Jaeger UI: http://localhost:16686
   - Grafana: http://localhost:3000 (login: admin/admin)
   - Prometheus: http://localhost:9090

## Testing the Application

### Health Check (Built-in)
```bash
curl http://localhost:9999/health
```

Expected response:
```json
{"status":"ok"}
```

This endpoint is provided automatically by `simba.Default()` and demonstrates basic automatic tracing.

### Create a User (Custom Spans Demo)
```bash
curl -X POST http://localhost:9999/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'
```

Expected response:
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2025-11-30T..."
}
```

This endpoint demonstrates:
- Input validation span
- Database operation span with attributes
- Email sending span
- Nested span relationships

### Get a User (Error Handling Demo)
```bash
# Get existing user
curl http://localhost:9999/users/1

# Get non-existent user (404 error)
curl http://localhost:9999/users/999
```

This endpoint demonstrates:
- Database query span
- External API call span with random failures (20% chance)
- Error status tracking in spans
- Graceful degradation when external APIs fail

### Trigger Custom Metrics
```bash
curl http://localhost:9999/metrics-demo
```

This endpoint demonstrates:
- Custom counter increment
- Custom histogram recording

## Observing Telemetry Data

### In Jaeger (Distributed Tracing)

1. Open http://localhost:16686
2. Select service: **simba-telemetry-demo**
3. Click "Find Traces"
4. Click on any trace to see:
   - HTTP request span (automatic)
   - Custom spans for database operations
   - Custom spans for email sending
   - Custom spans for external API calls
   - Span attributes (user IDs, operation types, etc.)
   - Error states (for failed external API calls)

**What to look for**:
- Trace for `POST /users` showing nested spans: validation → database → email
- Trace for `GET /users/:id` showing database query + external API call
- Failed external API calls marked with error status

### In Grafana (Metrics Visualization)

1. Open http://localhost:3000
2. Login with **admin/admin**
3. Navigate to Dashboards → Simba → **Simba HTTP Metrics**

The dashboard shows:
- **HTTP Request Rate**: Requests per second by endpoint
- **HTTP Request Duration**: p50, p95, p99 latencies
- **HTTP Status Codes**: Distribution of 2xx, 4xx, 5xx responses
- **Response Size**: p95 response sizes by endpoint
- **Request Rate by Endpoint**: Stacked bar chart
- **Error Rate**: 4xx and 5xx error rates over time

### In Prometheus (Raw Metrics)

1. Open http://localhost:9090
2. Try these queries:

**Request rate**:
```promql
rate(simba_http_server_request_count_total[1m])
```

**95th percentile latency**:
```promql
histogram_quantile(0.95, rate(simba_http_server_request_duration_milliseconds_bucket[5m]))
```

**Custom counter**:
```promql
simba_custom_demo_counter_total
```

**Custom histogram**:
```promql
histogram_quantile(0.95, rate(simba_custom_demo_histogram_bucket[5m]))
```

## Configuration Options

### Environment Variables

The application can be configured via environment variables:

```bash
# Enable/disable telemetry
SIMBA_TELEMETRY_ENABLED=true

# Tracing configuration
SIMBA_TELEMETRY_TRACING_ENABLED=true
SIMBA_TELEMETRY_TRACING_ENDPOINT=otel-collector:4317
SIMBA_TELEMETRY_TRACING_EXPORTER=otlp  # or 'stdout'
SIMBA_TELEMETRY_TRACING_INSECURE=true
SIMBA_TELEMETRY_TRACING_SAMPLING_RATE=1.0  # 100% sampling

# Metrics configuration
SIMBA_TELEMETRY_METRICS_ENABLED=true
SIMBA_TELEMETRY_METRICS_ENDPOINT=otel-collector:4317
SIMBA_TELEMETRY_METRICS_EXPORTER=otlp  # or 'stdout'
SIMBA_TELEMETRY_METRICS_INSECURE=true
SIMBA_TELEMETRY_METRICS_EXPORT_INTERVAL=60  # seconds

# Service identification
SIMBA_TELEMETRY_SERVICE_NAME=my-service
SIMBA_TELEMETRY_SERVICE_VERSION=1.0.0
SIMBA_TELEMETRY_ENVIRONMENT=production
```

### YAML Configuration

See `config.yaml` for an example YAML configuration file.

### Programmatic Configuration

```go
app := simba.Default(
    settings.WithApplicationName("my-service"),
    settings.WithApplicationVersion("1.0.0"),
    settings.WithTelemetryEnabled(true),
    settings.WithTracingEndpoint("localhost:4317"),
    settings.WithMetricsEndpoint("localhost:4317"),
    settings.WithTelemetryEnvironment("production"),
)
```

## Code Examples

### Creating Custom Spans

```go
import "github.com/sillen102/simba/telemetry"
import "go.opentelemetry.io/otel/attribute"

func myHandler(ctx context.Context, req *Request) (*Response, error) {
    // Create a custom span
    ctx, span := telemetry.StartSpan(ctx, "my.operation")
    defer span.End()
    
    // Add attributes
    span.SetAttributes(
        attribute.String("user.id", "123"),
        attribute.String("operation.type", "query"),
    )
    
    // Do work...
    
    return &Response{}, nil
}
```

### Error Handling in Spans

```go
import "go.opentelemetry.io/otel/codes"

ctx, span := telemetry.StartSpan(ctx, "risky.operation")
defer span.End()

result, err := riskyOperation()
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, "Operation failed")
    return nil, err
}

span.SetStatus(codes.Ok, "Success")
```

### Creating Custom Metrics

```go
import "go.opentelemetry.io/otel/metric"

// In main() or setup function
meter := telemetry.GetMeter(context.Background())

counter, err := meter.Int64Counter(
    "my.custom.counter",
    metric.WithDescription("Description of the counter"),
    metric.WithUnit("1"),
)

// In handler
counter.Add(ctx, 1, 
    metric.WithAttributes(
        attribute.String("status", "success"),
    ),
)
```

## Troubleshooting

### Services Not Starting

Check service logs:
```bash
docker-compose logs app
docker-compose logs otel-collector
```

### No Traces in Jaeger

1. Verify OTEL Collector is receiving data:
   ```bash
   docker-compose logs otel-collector | grep -i trace
   ```

2. Check telemetry is enabled in the app:
   ```bash
   docker-compose logs app | grep -i telemetry
   ```

### No Metrics in Prometheus

1. Verify OTEL Collector Prometheus exporter:
   ```bash
   curl http://localhost:8889/metrics
   ```

2. Check Prometheus targets:
   - Open http://localhost:9090/targets
   - Ensure `otel-collector` target is UP

### Dashboard Not Loading in Grafana

1. Check datasource configuration:
   - Go to Configuration → Data Sources
   - Verify Prometheus and Jaeger are configured

2. Manually import dashboard:
   - Go to Dashboards → Import
   - Upload `grafana/dashboards/simba-http-metrics.json`

## Cleanup

Stop and remove all containers and volumes:

```bash
docker-compose down -v
```

This will remove all collected metrics and traces.

## Next Steps

- Explore the `main.go` code to see how custom spans and metrics are created
- Modify the endpoints to add your own instrumentation
- Experiment with different sampling rates
- Try the stdout exporter for local development
- Connect your own application to the observability stack

## Learn More

- [Simba Documentation](https://github.com/sillen102/simba)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
