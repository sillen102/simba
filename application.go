package simba

import (
	"io"
	"net/http"
	"os"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/middleware"
)

// Application is the main application struct that holds the router and other application settings
type Application[AuthModel any] struct {
	// Router is the main router for the application
	Router *Router
	// settings is the application settings
	settings Settings
	// authFunc is the function used to authenticate and retrieve the authenticated model
	// from the request
	authFunc AuthFunc[AuthModel]
	// logger is the logger used by the application and gets injected into the request context
	// for each request
	logger zerolog.Logger
}

// Settings is a struct that holds the application settings
type Settings struct {
	// RequestDisallowUnknownFields will disallow unknown fields in the request body,
	// resulting in a 400 Bad Request response if a field is present that cannot be
	// mapped to the model struct.
	RequestDisallowUnknownFields bool

	// RequestIdHeader will determine if the request ID should be read from the
	// request header. If not set, the request ID will be generated.
	RequestIdAcceptHeader bool

	// LogRequestBody will determine if the request body will be logged
	LogRequestBody bool

	// LogLevel is the log level for the logger that will be used
	LogLevel zerolog.Level

	// LogFormat is the log format for the logger that will be used
	LogFormat logging.LogFormat

	// LogOutput is the output for the logger that will be used
	// If not set, the output will be [os.Stdout]
	LogOutput io.Writer
}

// AuthFunc is a function type for authenticating and retrieving an authenticated model struct from a request
type AuthFunc[AuthModel any] func(r *http.Request) (*AuthModel, error)

// Default returns a new [Application] application with default settings
func Default() *Application[struct{}] {
	return DefaultWithAuth[struct{}](nil)
}

// New returns a new [Application] application
func New(settings ...Settings) *Application[struct{}] {
	return NewWithAuth[struct{}](nil, settings...)
}

// DefaultWithAuth returns a new [Application] application with default settings and ability to have authenticated routes
// using the provided authFunc to authenticate and retrieve the user
func DefaultWithAuth[AuthModel any](authFunc AuthFunc[AuthModel]) *Application[AuthModel] {
	app := NewWithAuth(authFunc)
	app.Router.Extend(app.defaultMiddleware())
	app.addDefaultEndpoints()
	return app
}

// NewWithAuth returns a new [Application] application with ability to have authenticated routes
// using the provided [AuthFunc] to authenticate and retrieve the authenticated model
func NewWithAuth[User any](authFunc AuthFunc[User], st ...Settings) *Application[User] {
	settings := defaultSettings()
	if len(st) > 0 {
		settings = st[0]
	}

	logger := createLogger(settings)

	router := newRouter(logger, settings)
	router.Use(func(next http.Handler) http.Handler {
		return injectAuthFunc(next, authFunc)
	})

	return &Application[User]{
		Router:   router,
		settings: settings,
		authFunc: authFunc,
		logger:   logger,
	}
}

// ServeHTTP implements the [http.Handler] interface for the Simba [Application]
func (s *Application[AuthModel]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Router.ServeHTTP(w, req)
}

// defaultSettings creates a [Settings] with default values
func defaultSettings() Settings {
	return Settings{
		RequestDisallowUnknownFields: true,
		RequestIdAcceptHeader:        false,
		LogRequestBody:               false,
		LogLevel:                     zerolog.InfoLevel,
		LogFormat:                    logging.TextFormat,
		LogOutput:                    os.Stdout,
	}
}

// defaultMiddleware returns the middleware chain used in the default [Application] application
func (s *Application[AuthModel]) defaultMiddleware() alice.Chain {
	requestIdConfig := middleware.RequestIdConfig{
		AcceptFromHeader: s.settings.RequestIdAcceptHeader,
	}

	requestLoggerConfig := middleware.RequestLoggerConfig{
		LogRequestBody: s.settings.LogRequestBody,
	}

	return alice.New(
		middleware.PanicRecover,
		requestIdConfig.RequestID,
		requestLoggerConfig.LogRequests,
	)
}

// createLogger creates a new logger with the provided settings
func createLogger(settings Settings) zerolog.Logger {
	return logging.New(logging.Config{
		Format: settings.LogFormat,
		Level:  settings.LogLevel,
		Output: settings.LogOutput,
	})
}
