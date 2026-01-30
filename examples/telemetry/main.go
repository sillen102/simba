package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/telemetry/config"

	// NOTE: Telemetry usage is now handled via the OtelTelemetryProvider explicitly constructed and injected below.
	// Imports for OTel interfaces only if needed for demonstration or metric/span creation in handlers.
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	telemetryPkg "github.com/sillen102/simba/telemetry" // used for provider setup in main
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Request/Response models
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	ExtraData any       `json:"extra_data,omitempty"`
}

type UserIDParams struct {
	ID string `path:"id" validate:"required"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// In-memory user storage for demo
var users = make(map[int]UserResponse)
var userIDCounter = 0

// Custom metrics and telemetry objects (initialized in main)
var (
	meter           metric.Meter
	tracer          trace.Tracer
	customCounter   metric.Int64Counter
	customHistogram metric.Float64Histogram
)

// createUserHandler demonstrates custom spans with attributes
func createUserHandler(ctx context.Context, req *simbaModels.Request[CreateUserRequest, simbaModels.NoParams]) (*simbaModels.Response[UserResponse], error) {
	// Create a custom span for validation
	ctx, validateSpan := tracer.Start(ctx, "validate.user.input")
	validateSpan.SetAttributes(
		attribute.String("user.email", req.Body.Email),
		attribute.String("user.name", req.Body.Name),
	)
	time.Sleep(10 * time.Millisecond) // Simulate validation work
	validateSpan.End()

	// Create a custom span for database operation
	if err := simulateDBOperation(ctx, "insert_user"); err != nil {
		return nil, err
	}

	// Create user
	userIDCounter++
	user := UserResponse{
		ID:        userIDCounter,
		Name:      req.Body.Name,
		Email:     req.Body.Email,
		CreatedAt: time.Now(),
	}
	users[user.ID] = user

	// Simulate sending welcome email
	ctx, emailSpan := tracer.Start(ctx, "send.welcome.email")
	emailSpan.SetAttributes(
		attribute.String("email.to", user.Email),
		attribute.String("email.type", "welcome"),
	)
	time.Sleep(20 * time.Millisecond) // Simulate email sending
	emailSpan.End()

	return &simbaModels.Response[UserResponse]{
		Body:   user,
		Status: http.StatusCreated,
	}, nil
}

// getUserHandler demonstrates nested spans and error handling
func getUserHandler(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, UserIDParams]) (*simbaModels.Response[UserResponse], error) {
	// Parse user ID
	userID, err := strconv.Atoi(req.Params.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Fetch from "database" with custom span
	ctx, dbSpan := tracer.Start(ctx, "database.get.user",
		trace.WithAttributes(attribute.Int("user.id", userID)),
	)
	time.Sleep(15 * time.Millisecond) // Simulate DB query
	user, exists := users[userID]
	dbSpan.End()

	if !exists {
		return nil, simbaErrors.NewSimbaError(http.StatusNotFound, fmt.Sprintf("user %d not found", userID), nil)
	}

	// Fetch additional data from "external API"
	extraData, err := simulateExternalAPICall(ctx, userID)
	if err != nil {
		// Note: We continue even if external API fails (graceful degradation)
		// The error will be tracked in the span
		user.ExtraData = map[string]string{"error": "external API unavailable"}
	} else {
		user.ExtraData = extraData
	}

	return &simbaModels.Response[UserResponse]{
		Body: user,
	}, nil
}

// metricsDemoHandler demonstrates creating custom metrics
func metricsDemoHandler(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[MessageResponse], error) {
	// Record custom metrics
	if err := recordCustomMetrics(ctx); err != nil {
		return nil, fmt.Errorf("failed to record metrics: %w", err)
	}

	return &simbaModels.Response[MessageResponse]{
		Body: MessageResponse{
			Message: "Custom metrics recorded successfully",
		},
	}, nil
}

// simulateDBOperation simulates a database operation with tracing
func simulateDBOperation(ctx context.Context, operation string) error {
	ctx, span := tracer.Start(ctx, "database."+operation)
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", "users"),
	)

	// Simulate DB latency
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)

	return nil
}

// simulateExternalAPICall simulates calling an external API with error handling
func simulateExternalAPICall(ctx context.Context, userID int) (map[string]any, error) {
	ctx, span := tracer.Start(ctx, "external.api.get_user_preferences")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", "https://api.example.com/preferences"),
		attribute.Int("user.id", userID),
	)

	// Simulate API latency
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// Randomly simulate API failures (20% chance)
	if rand.Float32() < 0.2 {
		err := fmt.Errorf("external API timeout")
		span.RecordError(err)
		span.SetStatus(codes.Error, "API call failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "Success")
	return map[string]any{
		"theme":       "dark",
		"language":    "en",
		"preferences": "loaded",
	}, nil
}

// recordCustomMetrics demonstrates recording custom metrics
func recordCustomMetrics(ctx context.Context) error {
	// Increment counter
	customCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("operation", "demo"),
		),
	)

	// Record histogram value (simulating some measurement)
	value := 100 + rand.Float64()*900 // Random value between 100-1000
	customHistogram.Record(ctx, value,
		metric.WithAttributes(
			attribute.String("measurement.type", "demo"),
		),
	)

	return nil
}

func main() {
	ctx := context.Background()

	// Setup the OpenTelemetry configuration
	tcfg := &config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "simba-telemetry-demo",
		ServiceVersion: "1.0.0",
		Environment:    "demo",
		Tracing: config.TracingConfig{
			Enabled:  true,
			Exporter: "otlp",
			Endpoint: "otel-collector:4317",
			Insecure: true,
		},
		Metrics: config.MetricsConfig{
			Enabled:        true,
			Exporter:       "otlp",
			Endpoint:       "otel-collector:4317",
			Insecure:       true,
			ExportInterval: 30,
		},
	}

	// Build the Simba application
	// No longer use settings.WithTelemetry, wiring is now explicit with provider injection
	app := simba.Default()

	// Explicitly construct and inject the OtelTelemetryProvider
	prov, err := telemetryPkg.NewOtelTelemetryProvider(ctx, tcfg)
	if err != nil {
		panic("Failed to create OtelTelemetryProvider: " + err.Error())
	}
	app.SetTelemetryProvider(prov)

	// Type assert so we can access implementation details to create meters/tracers as needed in the demo
	otelProv := prov.(*telemetryPkg.OtelTelemetryProvider)
	meter := otelProv.Provider().Meter("simba.demo")
	tracer = otelProv.Provider().Tracer("simba.demo")

	// Initialize custom metrics
	customCounter, err = meter.Int64Counter(
		"custom.demo.counter",
		// metric.WithDescription("A custom counter for demonstration"),
		// metric.WithUnit("1"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create counter: %v", err))
	}

	customHistogram, err = meter.Float64Histogram(
		"custom.demo.histogram",
		// metric.WithDescription("A custom histogram for demonstration"),
		// metric.WithUnit("ms"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create histogram: %v", err))
	}

	// Register routes
	app.Router.POST("/users", simba.JsonHandler(createUserHandler))
	app.Router.GET("/users/{id}", simba.JsonHandler(getUserHandler))
	app.Router.GET("/metrics-demo", simba.JsonHandler(metricsDemoHandler))

	app.Start()
}
