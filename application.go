package simba

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/middleware"
)

// Application is the main application struct that holds the router and other application Settings
type Application[AuthModel any] struct {

	// Router is the main router for the application
	Router *Router

	// settings is the application settings
	settings *Settings

	// authFunc is the function used to authenticate and retrieve the authenticated model
	// from the Request
	authFunc AuthFunc[AuthModel]

	// logger is the logger used by the application and gets injected into the Request context
	// for each Request
	logger zerolog.Logger
}

// AuthFunc is a function type for authenticating and retrieving an authenticated model struct from a Request
type AuthFunc[AuthModel any] func(r *http.Request) (*AuthModel, error)

// Default returns a new [Application] application with default Settings
func Default() *Application[struct{}] {
	return DefaultWithAuth[struct{}](nil)
}

// New returns a new [Application] application
func New(settings ...Settings) *Application[struct{}] {
	return NewWithAuth[struct{}](nil, settings...)
}

// DefaultWithAuth returns a new [Application] application with default Settings and ability to have authenticated routes
// using the provided authFunc to authenticate and retrieve the user
func DefaultWithAuth[AuthModel any](authFunc AuthFunc[AuthModel]) *Application[AuthModel] {
	app := NewWithAuth(authFunc)
	app.Router.Extend(app.defaultMiddleware())
	app.addDefaultEndpoints()
	return app
}

// NewWithAuth returns a new [Application] application with ability to have authenticated routes
// using the provided [AuthFunc] to authenticate and retrieve the authenticated model
func NewWithAuth[User any](authFunc AuthFunc[User], provided ...Settings) *Application[User] {
	settings, err := loadConfig(provided...)
	if err != nil {
		panic(err)
	}

	logger := createLogger(settings)

	router := newRouter(logger, settings.Request)
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

// defaultMiddleware returns the middleware chain used in the default [Application] application
func (s *Application[AuthModel]) defaultMiddleware() alice.Chain {
	requestIdConfig := middleware.RequestIdConfig{
		AcceptFromHeader: s.settings.Request.RequestIdMode == middleware.AcceptFromHeader,
	}

	return alice.New(
		requestIdConfig.RequestID,
		middleware.LogRequests,
	)
}

// createLogger creates a new logger with the provided Settings
func createLogger(settings *Settings) zerolog.Logger {
	return logging.New(logging.Config{
		Format: settings.Logging.Format,
		Level:  settings.Logging.Level,
		Output: settings.Logging.Output,
	})
}
