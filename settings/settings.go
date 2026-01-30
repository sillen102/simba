package settings

import (
	"log/slog"
	"os"

	configloader "github.com/sillen102/config-loader"
	"github.com/sillen102/simba/simbaModels"
)

// Simba is a struct that holds the application settings
type Simba struct {

	// Application settings
	Application `yaml:"application"`

	// Server settings
	Server `yaml:"server"`

	// Request settings
	Request `yaml:"request"`

	// Docs settings
	Docs `yaml:"docs"`

	// Telemetry settings
	Telemetry `yaml:"telemetry"`

	// Logger settings
	Logger *slog.Logger `yaml:"-" env:"-"`

	envGetter func(string) string
}

type Application struct {
	Name    string `yaml:"name" env:"APPLICATION_NAME" default:"Simba Application"`
	Version string `yaml:"version" env:"APPLICATION_VERSION" default:"0.1.0"`
}

// Server holds the Simba for the application server
type Server struct {

	// Host is the host the server will listen on
	Host string `yaml:"host" env:"SIMBA_SERVER_HOST" default:"0.0.0.0"`

	// Addr is the address the server will listen on
	Port int `yaml:"port" env:"SIMBA_SERVER_PORT" default:"9999"`
}

// Request holds the Simba for the Request processing
type Request struct {

	// AllowUnknownFields will set the behavior for unknown fields in the Request body,
	// resulting in a 400 Bad Request response if a field is present that cannot be
	// mapped to the model struct.
	AllowUnknownFields bool `yaml:"allow-unknown-fields" env:"SIMBA_REQUEST_ALLOW_UNKNOWN_FIELDS" default:"true"`

	// LogRequestBody will determine if the Request body will be logged
	// If set to "disabled", the Request body will not be logged, which is also the default
	LogRequestBody bool `yaml:"log-request-body" env:"SIMBA_REQUEST_LOG_REQUEST_BODY" default:"false"`

	// TraceIDMode determines how the Trace ID will be handled
	TraceIDMode simbaModels.TraceIDMode `yaml:"trace-id-mode" env:"SIMBA_TRACE_ID_MODE" default:"AcceptFromHeader"`
}

type Docs struct {

	// GenerateOpenAPIDocs will determine if the API documentation (YAML or JSON) will be generated
	GenerateOpenAPIDocs bool `yaml:"generate-docs" env:"SIMBA_DOCS_GENERATE" default:"true"`

	// MountDocsUIEndpoint will determine if the documentation UI will be mounted
	MountDocsUIEndpoint bool `yaml:"mount-docs-endpoint" env:"SIMBA_DOCS_MOUNT_DOCS_UI_ENDPOINT" default:"true"`

	// OpenAPIFilePath is the path to the OpenAPI YAML file
	OpenAPIFilePath string `yaml:"open-api-file-path" env:"SIMBA_DOCS_OPENAPI_FILE_PATH" default:"/openapi.json"`

	// DocsUIPath is the path to the API documentation
	DocsUIPath string `yaml:"docs-path" env:"SIMBA_DOCS_UI_PATH" default:"/docs"`

	// ServiceName is the name of the service
	ServiceName string
}

// Telemetry holds the settings for OpenTelemetry integration
type Telemetry struct {
	// Enabled determines if telemetry is enabled (opt-in, default: false)
	Enabled bool `yaml:"enabled" env:"SIMBA_TELEMETRY_ENABLED" default:"false"`

	// Tracing configuration
	Tracing TracingConfig `yaml:"tracing"`

	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics"`

	// ServiceName is the name of the service for telemetry (defaults to Application.Name)
	ServiceName string `yaml:"service-name" env:"SIMBA_TELEMETRY_SERVICE_NAME"`

	// ServiceVersion is the version of the service for telemetry (defaults to Application.Version)
	ServiceVersion string `yaml:"service-version" env:"SIMBA_TELEMETRY_SERVICE_VERSION"`

	// Environment is the deployment environment (development, staging, production, etc.)
	Environment string `yaml:"environment" env:"SIMBA_TELEMETRY_ENVIRONMENT" default:"development"`
}

// TracingConfig holds the configuration for distributed tracing
type TracingConfig struct {
	// Enabled determines if tracing is enabled (default: true when telemetry is enabled)
	Enabled bool `yaml:"enabled" env:"SIMBA_TELEMETRY_TRACING_ENABLED" default:"true"`

	// Exporter is the type of exporter to use (otlp, stdout)
	Exporter string `yaml:"exporter" env:"SIMBA_TELEMETRY_TRACING_EXPORTER" default:"otlp"`

	// Endpoint is the endpoint for the trace exporter
	Endpoint string `yaml:"endpoint" env:"SIMBA_TELEMETRY_TRACING_ENDPOINT" default:"localhost:4317"`

	// Insecure determines if the connection should be insecure (default: true for local development)
	Insecure bool `yaml:"insecure" env:"SIMBA_TELEMETRY_TRACING_INSECURE" default:"true"`

	// SamplingRate is the sampling rate for traces (0.0 to 1.0, default: 1.0 = 100%)
	SamplingRate float64 `yaml:"sampling-rate" env:"SIMBA_TELEMETRY_TRACING_SAMPLING_RATE" default:"1.0"`
}

// MetricsConfig holds the configuration for metrics collection
type MetricsConfig struct {
	// Enabled determines if metrics collection is enabled (default: true when telemetry is enabled)
	Enabled bool `yaml:"enabled" env:"SIMBA_TELEMETRY_METRICS_ENABLED" default:"true"`

	// Exporter is the type of exporter to use (otlp, stdout)
	Exporter string `yaml:"exporter" env:"SIMBA_TELEMETRY_METRICS_EXPORTER" default:"otlp"`

	// Endpoint is the endpoint for the metrics exporter
	Endpoint string `yaml:"endpoint" env:"SIMBA_TELEMETRY_METRICS_ENDPOINT" default:"localhost:4317"`

	// Insecure determines if the connection should be insecure (default: true for local development)
	Insecure bool `yaml:"insecure" env:"SIMBA_TELEMETRY_METRICS_INSECURE" default:"true"`

	// ExportInterval is the interval in seconds for exporting metrics (default: 60 seconds)
	ExportInterval int `yaml:"export-interval" env:"SIMBA_TELEMETRY_METRICS_EXPORT_INTERVAL" default:"60"`
}

// Option is a function that configures a Simba application settings struct.
type Option func(*Simba)

// WithApplicationName sets the application name
func WithApplicationName(name string) Option {
	return func(s *Simba) {
		s.Name = name
	}
}

// WithApplicationVersion sets the application version
func WithApplicationVersion(version string) Option {
	return func(s *Simba) {
		s.Version = version
	}
}

// WithServerPort sets the server port
func WithServerPort(port int) Option {
	return func(s *Simba) {
		s.Port = port
	}
}

// WithServerHost sets the server host
func WithServerHost(host string) Option {
	return func(s *Simba) {
		s.Host = host
	}
}

// WithAllowUnknownFields sets whether to allow unknown fields
func WithAllowUnknownFields(allow bool) Option {
	return func(s *Simba) {
		s.AllowUnknownFields = allow
	}
}

// WithLogRequestBody sets whether to log request bodies
func WithLogRequestBody(log bool) Option {
	return func(s *Simba) {
		s.LogRequestBody = log
	}
}

// WithTraceIDMode sets the trace ID mode
func WithTraceIDMode(mode simbaModels.TraceIDMode) Option {
	return func(s *Simba) {
		s.TraceIDMode = mode
	}
}

// WithGenerateOpenAPIDocs sets whether to generate OpenAPI docs
func WithGenerateOpenAPIDocs(generate bool) Option {
	return func(s *Simba) {
		s.GenerateOpenAPIDocs = generate
	}
}

// WithMountDocsUIEndpoint sets whether to mount the docs endpoint
func WithMountDocsUIEndpoint(mount bool) Option {
	return func(s *Simba) {
		s.MountDocsUIEndpoint = mount
	}
}

// WithOpenAPIFilePath sets the OpenAPI JSON file path
func WithOpenAPIFilePath(path string) Option {
	return func(s *Simba) {
		s.OpenAPIFilePath = path
	}
}

// WithDocsUIPath sets the docs UI path
func WithDocsUIPath(path string) Option {
	return func(s *Simba) {
		s.DocsUIPath = path
	}
}

// WithLogger sets the logger
func WithLogger(logger *slog.Logger) Option {
	return func(s *Simba) {
		if logger != nil {
			s.Logger = logger
		}
	}
}

// WithTelemetryEnabled sets whether telemetry is enabled
func WithTelemetryEnabled(enabled bool) Option {
	return func(s *Simba) {
		s.Enabled = enabled
	}
}

// WithTracingEndpoint sets the tracing endpoint
func WithTracingEndpoint(endpoint string) Option {
	return func(s *Simba) {
		s.Tracing.Endpoint = endpoint
	}
}

// WithTracingExporter sets the tracing exporter type
func WithTracingExporter(exporter string) Option {
	return func(s *Simba) {
		s.Tracing.Exporter = exporter
	}
}

// WithMetricsEnabled sets whether metrics collection is enabled
func WithMetricsEnabled(enabled bool) Option {
	return func(s *Simba) {
		s.Metrics.Enabled = enabled
	}
}

// WithMetricsExporter sets the metrics exporter type
func WithMetricsExporter(exporter string) Option {
	return func(s *Simba) {
		s.Metrics.Exporter = exporter
	}
}

// WithMetricsEndpoint sets the metrics endpoint
func WithMetricsEndpoint(endpoint string) Option {
	return func(s *Simba) {
		s.Metrics.Endpoint = endpoint
	}
}

// WithTelemetryEnvironment sets the telemetry environment
func WithTelemetryEnvironment(environment string) Option {
	return func(s *Simba) {
		s.Environment = environment
	}
}

// WithTelemetryServiceName sets the telemetry service name
func WithTelemetryServiceName(serviceName string) Option {
	return func(s *Simba) {
		s.Telemetry.ServiceName = serviceName
	}
}

// WithTelemetryServiceVersion sets the telemetry service version
func WithTelemetryServiceVersion(serviceVersion string) Option {
	return func(s *Simba) {
		s.ServiceVersion = serviceVersion
	}
}

// WithEnvGetter is a test-only option to mock environment variable retrieval
func WithEnvGetter(getter func(string) string) Option {
	return func(s *Simba) {
		s.envGetter = getter
	}
}

// Load loads the application settings
func Load(opts ...Option) (*Simba, error) {
	// Initialize settings with defaults
	settings := &Simba{
		envGetter: os.Getenv, // Set default environment getter
	}

	// Apply user options first
	for _, opt := range opts {
		opt(settings)
	}

	// Save logger reference before config loading potentially resets it
	savedLogger := settings.Logger

	// Load config from files and environment
	err := configloader.NewLoader(&configloader.LoaderOpts{
		EnvGetter: settings.envGetter,
	}).Load(settings)

	if err != nil {
		return nil, err
	}

	// Restore the logger if it was set via options
	if savedLogger != nil {
		settings.Logger = savedLogger
	}

	// Reapply options to override any config values
	for _, opt := range opts {
		opt(settings)
	}

	// Ensure we have a logger (only set default if no logger is configured)
	if settings.Logger == nil {
		settings.Logger = slog.Default()
	}

	// Set the service name for Docs
	settings.Docs.ServiceName = settings.Name

	return settings, nil
}

// LoadWithOptions loads settings using the options pattern
func LoadWithOptions(opts ...Option) (*Simba, error) {
	return Load(opts...)
}
