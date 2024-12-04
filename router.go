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

// Router holds the router and options
type Router[AuthModel any] struct {
	router     *httprouter.Router
	options    Options
	middleware alice.Chain
	authFunc   AuthFunc[AuthModel]
	logger     zerolog.Logger
}

// Options are the settings for the router
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

// AuthFunc is a function type for authenticating and retrieving a user from a request
type AuthFunc[User any] func(r *http.Request) (*User, error)

// Default returns a new Router with default settings
func Default() *Router[struct{}] {
	return DefaultWithAuth[struct{}](nil)
}

// DefaultWithAuth returns a new Router with default settings and ability to have authenticated routes using the provided authFunc to
// authenticate and retrieve the user
func DefaultWithAuth[User any](authFunc AuthFunc[User]) *Router[User] {
	router := NewRouterWithAuth[User](authFunc)
	router.middleware = router.middleware.Extend(defaultMiddleware(router.options))
	return router
}

// NewRouter returns a new Router
func NewRouter(opts ...Options) *Router[struct{}] {
	return NewRouterWithAuth[struct{}](nil, opts...)
}

// NewRouterWithAuth returns a new Router with ability to have authenticated routes using the provided authFunc to
// authenticate and retrieve the user
func NewRouterWithAuth[User any](authFunc AuthFunc[User], opts ...Options) *Router[User] {
	options := defaultRouterOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	logger := logging.New(logging.LoggerConfig{
		Format: options.LogFormat,
		Level:  options.LogLevel,
		Output: options.LogOutput,
	})

	return &Router[User]{
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

// GetOptions returns the options for the router
func (s *Router[AuthModel]) GetOptions() Options {
	return s.options
}

// ServeHTTP implements the http.Handler interface
func (s *Router[AuthModel]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Use registers a middleware handler
func (s *Router[AuthModel]) Use(middleware func(next http.Handler) http.Handler) {
	s.middleware = s.middleware.Append(middleware)
}

// POST registers a handler for POST requests to the given pattern
func (s *Router[AuthModel]) POST(path string, handler http.Handler) {
	s.router.Handler(http.MethodPost, path, s.middleware.Then(handler))
}

// GET registers a handler for GET requests to the given pattern
func (s *Router[AuthModel]) GET(path string, handler http.Handler) {
	s.router.Handler(http.MethodGet, path, s.middleware.Then(handler))
}

// PUT registers a handler for PUT requests to the given pattern
func (s *Router[AuthModel]) PUT(path string, handler http.Handler) {
	s.router.Handler(http.MethodPut, path, s.middleware.Then(handler))
}

// DELETE registers a handler for DELETE requests to the given pattern
func (s *Router[AuthModel]) DELETE(path string, handler http.Handler) {
	s.router.Handler(http.MethodDelete, path, s.middleware.Then(handler))
}

// PATCH registers a handler for PATCH requests to the given pattern
func (s *Router[AuthModel]) PATCH(path string, handler http.Handler) {
	s.router.Handler(http.MethodPatch, path, s.middleware.Then(handler))
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (s *Router[AuthModel]) OPTIONS(path string, handler http.Handler) {
	s.router.Handler(http.MethodOptions, path, s.middleware.Then(handler))
}

// HEAD registers a handler for HEAD requests to the given pattern
func (s *Router[AuthModel]) HEAD(path string, handler http.Handler) {
	s.router.Handler(http.MethodHead, path, s.middleware.Then(handler))
}

// CONNECT registers a handler for CONNECT requests to the given pattern
func (s *Router[AuthModel]) CONNECT(path string, handler http.Handler) {
	s.router.Handler(http.MethodConnect, path, s.middleware.Then(handler))
}

// TRACE registers a handler for TRACE requests to the given pattern
func (s *Router[AuthModel]) TRACE(path string, handler http.Handler) {
	s.router.Handler(http.MethodTrace, path, s.middleware.Then(handler))
}

// defaultRouterOptions creates a RouterOptions with default values
func defaultRouterOptions() Options {
	return Options{
		RequestDisallowUnknownFields: true,
		RequestIdAcceptHeader:        false,
		LogRequestBody:               false,
		LogLevel:                     zerolog.InfoLevel,
		LogFormat:                    logging.TextFormat,
		LogOutput:                    os.Stdout,
	}
}

// defaultMiddleware returns the middleware chain used in the default router
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
