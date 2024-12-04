package simba

import (
	"io"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/middleware"
)

type Application[AuthModel any] struct {
	router     *httprouter.Router
	options    Options
	middleware alice.Chain
	authFunc   AuthFunc[AuthModel]
	logger     zerolog.Logger
}

type Options struct {
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
func New(opts ...Options) *Application[struct{}] {
	return NewWithAuth[struct{}](nil, opts...)
}

// DefaultWithAuth returns a new [Application] application with default settings and ability to have authenticated routes
// using the provided authFunc to authenticate and retrieve the user
func DefaultWithAuth[AuthModel any](authFunc AuthFunc[AuthModel]) *Application[AuthModel] {
	router := NewWithAuth[AuthModel](authFunc)
	router.middleware = router.middleware.Extend(defaultMiddleware(router.options))
	router.addDefaultEndpoints()
	return router
}

// NewWithAuth returns a new [Application] application with ability to have authenticated routes
// using the provided authFunc to authenticate and retrieve the authenticated model
func NewWithAuth[User any](authFunc AuthFunc[User], opts ...Options) *Application[User] {
	options := defaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	logger := logging.New(logging.Config{
		Format: options.LogFormat,
		Level:  options.LogLevel,
		Output: options.LogOutput,
	})

	return &Application[User]{
		router:  httprouter.New(),
		options: options,
		middleware: alice.New().
			Append(func(handler http.Handler) http.Handler {
				return injectLogger(handler, logger)
			}).
			Append(autoCloseRequestBody).
			Append(func(next http.Handler) http.Handler {
				return injectAuthFunc(next, authFunc)
			}).
			Append(func(next http.Handler) http.Handler {
				return injectOptions(next, options)
			}),
		authFunc: authFunc,
		logger:   logger,
	}
}

// GetOptions returns the options for the application
func (s *Application[AuthModel]) GetOptions() Options {
	return s.options
}

// GetRouter returns the underlying [httprouter.Router]
func (s *Application[AuthModel]) GetRouter() *httprouter.Router {
	return s.router
}

// Use registers a middleware handler
func (s *Application[AuthModel]) Use(middleware func(next http.Handler) http.Handler) {
	s.middleware = s.middleware.Append(middleware)
}

// POST registers a handler for POST requests to the given pattern
func (s *Application[AuthModel]) POST(path string, handler http.Handler) {
	s.router.Handler(http.MethodPost, path, s.middleware.Then(handler))
}

// GET registers a handler for GET requests to the given pattern
func (s *Application[AuthModel]) GET(path string, handler http.Handler) {
	s.router.Handler(http.MethodGet, path, s.middleware.Then(handler))
}

// PUT registers a handler for PUT requests to the given pattern
func (s *Application[AuthModel]) PUT(path string, handler http.Handler) {
	s.router.Handler(http.MethodPut, path, s.middleware.Then(handler))
}

// DELETE registers a handler for DELETE requests to the given pattern
func (s *Application[AuthModel]) DELETE(path string, handler http.Handler) {
	s.router.Handler(http.MethodDelete, path, s.middleware.Then(handler))
}

// PATCH registers a handler for PATCH requests to the given pattern
func (s *Application[AuthModel]) PATCH(path string, handler http.Handler) {
	s.router.Handler(http.MethodPatch, path, s.middleware.Then(handler))
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (s *Application[AuthModel]) OPTIONS(path string, handler http.Handler) {
	s.router.Handler(http.MethodOptions, path, s.middleware.Then(handler))
}

// HEAD registers a handler for HEAD requests to the given pattern
func (s *Application[AuthModel]) HEAD(path string, handler http.Handler) {
	s.router.Handler(http.MethodHead, path, s.middleware.Then(handler))
}

// CONNECT registers a handler for CONNECT requests to the given pattern
func (s *Application[AuthModel]) CONNECT(path string, handler http.Handler) {
	s.router.Handler(http.MethodConnect, path, s.middleware.Then(handler))
}

// TRACE registers a handler for TRACE requests to the given pattern
func (s *Application[AuthModel]) TRACE(path string, handler http.Handler) {
	s.router.Handler(http.MethodTrace, path, s.middleware.Then(handler))
}

// defaultOptions creates a [Options] with default values
func defaultOptions() Options {
	return Options{
		RequestDisallowUnknownFields: true,
		RequestIdAcceptHeader:        false,
		LogRequestBody:               false,
		LogLevel:                     zerolog.InfoLevel,
		LogFormat:                    logging.TextFormat,
		LogOutput:                    os.Stdout,
	}
}

// defaultMiddleware returns the middleware chain used in the default [Application] application
func defaultMiddleware(opts Options) alice.Chain {
	requestIdConfig := middleware.RequestIdConfig{
		AcceptFromHeader: opts.RequestIdAcceptHeader,
	}

	requestLoggerConfig := middleware.RequestLoggerConfig{
		LogRequestBody: opts.LogRequestBody,
	}

	return alice.New(
		middleware.PanicRecover,
		requestIdConfig.RequestID,
		requestLoggerConfig.LogRequests,
	)
}
