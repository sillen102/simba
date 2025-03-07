package settings

import (
	"log/slog"

	"github.com/sillen102/simba/config"
	"github.com/sillen102/simba/enums"
)

// Simba is a struct that holds the application settings
type Simba struct {

	// Server settings
	Server

	// Request settings
	Request

	// Docs settings
	Docs

	// Logger settings
	Logger *slog.Logger `yaml:"-" env:"-"`
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
	RequestIdMode enums.RequestIdMode `yaml:"request-id-mode" env:"SIMBA_REQUEST_ID_MODE" default:"AcceptFromHeader"`
}

type Docs struct {

	// GenerateOpenAPIDocs will determine if the API documentation (YAML or JSON) will be generated
	GenerateOpenAPIDocs bool `yaml:"generate-docs" env:"SIMBA_DOCS_GENERATE" default:"true"`

	// MountDocsEndpoint will determine if the documentation UI will be mounted
	MountDocsEndpoint bool `yaml:"mount-docs-endpoint" env:"SIMBA_DOCS_MOUNT_ENDPOINT" default:"true"`

	// OpenAPIFileType is the type of the OpenAPI file (YAML or JSON)
	OpenAPIFileType string `yaml:"open-api-file-type" env:"SIMBA_DOCS_OPENAPI_FILE_TYPE" default:"application/yaml"`

	// OpenAPIFilePath is the path to the OpenAPI YAML file
	OpenAPIFilePath string `yaml:"open-api-file-path" env:"SIMBA_DOCS_OPENAPI_MOUNT_PATH" default:"/openapi.yml"`

	// DocsPath is the path to the API documentation
	DocsPath string `yaml:"docs-path" env:"SIMBA_DOCS_MOUNT_PATH" default:"/docs"`

	// ServiceName is the name of the service
	ServiceName string `yaml:"service-name" env:"SIMBA_DOCS_SERVICE_NAME" default:"Simba Application"`
}

// Option is a function that configures a Simba application settings struct.
type Option func(*Simba)

// WithServerPort sets the server port
func WithServerPort(port int) Option {
	return func(s *Simba) {
		s.Server.Port = port
	}
}

// WithServerHost sets the server host
func WithServerHost(host string) Option {
	return func(s *Simba) {
		s.Server.Host = host
	}
}

// WithAllowUnknownFields sets whether to allow unknown fields
func WithAllowUnknownFields(allow bool) Option {
	return func(s *Simba) {
		s.Request.AllowUnknownFields = allow
	}
}

// WithLogRequestBody sets whether to log request bodies
func WithLogRequestBody(log bool) Option {
	return func(s *Simba) {
		s.Request.LogRequestBody = log
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

// Load loads the application settings
func Load(opts ...Option) (*Simba, error) {
	var settings Simba

	// Load defaults from config files/env vars
	if err := config.NewLoader(nil).Load(&settings); err != nil {
		return nil, err
	}

	// Apply options
	for _, opt := range opts {
		opt(&settings)
	}

	// Set default logger if none provided
	if settings.Logger == nil {
		settings.Logger = slog.Default()
	}

	return &settings, nil
}

// LoadWithOptions loads settings using the options pattern
func LoadWithOptions(opts ...Option) (*Simba, error) {
	return Load(opts...)
}
