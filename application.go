package simba

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/sillen102/simba/middleware"
)

// Application is the main application struct that holds the router and other application Settings
type Application[AuthModel any] struct {

	// Server is the HTTP server for the application
	Server *http.Server

	// Router is the main router for the application
	Router *Router

	// settings is the application settings
	settings *Settings

	// authFunc is the function used to authenticate and retrieve the authenticated model
	// from the Request
	authFunc AuthFunc[AuthModel]

	logger zerolog.Logger
}

// AuthFunc is a function type for authenticating and retrieving an authenticated model struct from a Request
type AuthFunc[AuthModel any] func(r *http.Request) (*AuthModel, error)

// Default returns a new [Application] application with default Settings
func Default(settings ...Settings) *Application[struct{}] {
	return DefaultAuthWith[struct{}](nil, settings...)
}

// New returns a new [Application] application
func New(settings ...Settings) *Application[struct{}] {
	return NewAuthWith[struct{}](nil, settings...)
}

// DefaultAuthWith returns a new [Application] application with default Settings and ability to have authenticated routes
// using the provided authFunc to authenticate and retrieve the user
func DefaultAuthWith[AuthModel any](authFunc AuthFunc[AuthModel], settings ...Settings) *Application[AuthModel] {
	app := NewAuthWith(authFunc, settings...)
	app.Router.Extend(app.defaultMiddleware())
	app.addDefaultEndpoints()
	return app
}

// NewAuthWith returns a new [Application] application with ability to have authenticated routes
// using the provided [AuthFunc] to authenticate and retrieve the authenticated model
func NewAuthWith[User any](authFunc AuthFunc[User], provided ...Settings) *Application[User] {
	settings, err := loadConfig(provided...)
	if err != nil {
		panic(err)
	}

	logger := NewLogger(&settings.Logging)

	router := newRouter(settings.Request, logger)
	router.Use(func(next http.Handler) http.Handler {
		return injectAuthFunc(next, authFunc)
	})

	return &Application[User]{
		Server:   &http.Server{Addr: fmt.Sprintf("%s:%s", settings.Server.Host, settings.Server.Port), Handler: router},
		Router:   router,
		settings: settings,
		authFunc: authFunc,
		logger:   logger,
	}
}

// ServeHTTP implements the [http.Handler] interface for the Simba [Application]
func (a *Application[AuthModel]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.Router.ServeHTTP(w, req)
}

// defaultMiddleware returns the middleware chain used in the default [Application] application
func (a *Application[AuthModel]) defaultMiddleware() alice.Chain {
	requestIdConfig := middleware.RequestIdConfig{
		AcceptFromHeader: a.settings.Request.RequestIdMode == middleware.AcceptFromHeader,
	}

	return alice.New(
		requestIdConfig.RequestID,
		middleware.LogRequests,
	)
}
