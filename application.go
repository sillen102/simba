package simba

import (
	"fmt"
	"net/http"

	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/settings"
)

// Application is the main application struct that holds the Mux and other application Settings
type Application[AuthModel any] struct {

	// Server is the HTTP server for the application
	Server *http.Server

	// Router is the main Mux for the application
	Router *Router

	// Settings is the application Settings
	Settings *settings.Config

	// AuthFunc is the function used to authenticate and retrieve the authenticated model
	// from the Request
	AuthFunc AuthFunc[AuthModel]
}

// AuthFunc is a function type for authenticating and retrieving an authenticated model struct from a Request
type AuthFunc[AuthModel any] func(r *http.Request) (*AuthModel, error)

// Default returns a new [Application] application with default Config
func Default(settings ...settings.Config) *Application[struct{}] {
	return DefaultAuthWith[struct{}](nil, settings...)
}

// New returns a new [Application] application
func New(settings ...settings.Config) *Application[struct{}] {
	return NewAuthWith[struct{}](nil, settings...)
}

// DefaultAuthWith returns a new [Application] application with default Config and ability to have authenticated routes
// using the provided AuthFunc to authenticate and retrieve the user
func DefaultAuthWith[AuthModel any](authFunc AuthFunc[AuthModel], settings ...settings.Config) *Application[AuthModel] {
	app := NewAuthWith(authFunc, settings...)
	app.Router.Extend(app.defaultMiddleware())
	app.addDefaultEndpoints()
	return app
}

// NewAuthWith returns a new [Application] application with ability to have authenticated routes
// using the provided [AuthFunc] to authenticate and retrieve the authenticated model
func NewAuthWith[AuthModel any](authFunc AuthFunc[AuthModel], provided ...settings.Config) *Application[AuthModel] {
	cfg, err := settings.Load(provided...)
	if err != nil {
		panic(err)
	}

	router := newRouter(cfg.Request)
	router.Use(func(next http.Handler) http.Handler {
		return injectAuthFunc(next, authFunc)
	})
	router.Use(func(next http.Handler) http.Handler {
		return injectRequestSettings(next, &cfg.Request)
	})

	return &Application[AuthModel]{
		Server:   &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), Handler: router},
		Router:   router,
		Settings: cfg,
		AuthFunc: authFunc,
	}
}

// defaultMiddleware returns the middleware chain used in the default [Application] application
func (a *Application[AuthModel]) defaultMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.Logger{Logger: a.Settings.Logger}.ContextLogger,
		middleware.PanicRecovery,
		middleware.LogRequests,
	}
}
