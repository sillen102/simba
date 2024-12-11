package simba

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/justinas/alice"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/settings"
)

// Application is the main application struct that holds the Mux and other application Settings
type Application[AuthModel any] struct {

	// Server is the HTTP server for the application
	Server *http.Server

	// Router is the main Mux for the application
	Router *Router

	// settings is the application settings
	settings *settings.Settings

	// authFunc is the function used to authenticate and retrieve the authenticated model
	// from the Request
	authFunc AuthFunc[AuthModel]

	// logger is the logger used by the application
	logger *slog.Logger
}

// AuthFunc is a function type for authenticating and retrieving an authenticated model struct from a Request
type AuthFunc[AuthModel any] func(r *http.Request) (*AuthModel, error)

// Default returns a new [Application] application with default Settings
func Default(settings ...settings.Settings) *Application[struct{}] {
	return DefaultAuthWith[struct{}](nil, settings...)
}

// New returns a new [Application] application
func New(settings ...settings.Settings) *Application[struct{}] {
	return NewAuthWith[struct{}](nil, settings...)
}

// DefaultAuthWith returns a new [Application] application with default Settings and ability to have authenticated routes
// using the provided authFunc to authenticate and retrieve the user
func DefaultAuthWith[AuthModel any](authFunc AuthFunc[AuthModel], settings ...settings.Settings) *Application[AuthModel] {
	app := NewAuthWith(authFunc, settings...)
	app.Router.Extend(app.defaultMiddleware())
	app.addDefaultEndpoints()
	return app
}

// NewAuthWith returns a new [Application] application with ability to have authenticated routes
// using the provided [AuthFunc] to authenticate and retrieve the authenticated model
func NewAuthWith[User any](authFunc AuthFunc[User], provided ...settings.Settings) *Application[User] {
	cfg, err := settings.Load(provided...)
	if err != nil {
		panic(err)
	}

	logger := logging.NewLogger(cfg.Logging)
	router := newRouter(cfg.Request)
	router.Use(func(next http.Handler) http.Handler {
		return injectAuthFunc(next, authFunc)
	})

	return &Application[User]{
		Server:   &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), Handler: router},
		Router:   router,
		settings: cfg,
		authFunc: authFunc,
		logger:   logger,
	}
}

// defaultMiddleware returns the middleware chain used in the default [Application] application
func (a *Application[AuthModel]) defaultMiddleware() alice.Chain {
	return alice.New(
		middleware.RequestID,
		middleware.Logger{Logger: a.logger}.ContextLogger,
		middleware.PanicRecovery,
		middleware.LogRequests,
	)
}
