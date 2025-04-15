package settings

import (
	"log/slog"
	"os"

	"github.com/sillen102/config-loader"
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

	// RequestIdMode determines how the Request ID will be handled
	RequestIdMode simbaModels.RequestIdMode `yaml:"request-id-mode" env:"SIMBA_REQUEST_ID_MODE" default:"AcceptFromHeader"`
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

// WithRequestIdMode sets the request ID mode
func WithRequestIdMode(mode simbaModels.RequestIdMode) Option {
	return func(s *Simba) {
		s.RequestIdMode = mode
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

	// Set the service name
	settings.ServiceName = settings.Name

	return settings, nil
}

// LoadWithOptions loads settings using the options pattern
func LoadWithOptions(opts ...Option) (*Simba, error) {
	return Load(opts...)
}
