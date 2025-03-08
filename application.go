package simba

import (
	"fmt"
	"net/http"

	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaOpenapi"
)

// Application is the main application struct that holds the Mux and other application Settings
type Application struct {

	// Server is the HTTP server for the application
	Server *http.Server

	// Router is the main Mux for the application
	Router *Router

	// Settings is the application Settings
	Settings *settings.Simba
}

// Default returns a new [Application] application with default Simba
func Default(opts ...settings.Option) *Application {
	app := New(opts...)
	app.Router.Extend(app.defaultMiddleware())
	app.addDefaultEndpoints()
	return app
}

// New returns a new [Application] application
func New(opts ...settings.Option) *Application {
	cfg, err := settings.LoadWithOptions(opts...)
	if err != nil {
		panic(err)
	}

	router := newRouter(cfg.Request, cfg.Docs, simbaOpenapi.NewOpenAPIGenerator())
	router.Use(func(next http.Handler) http.Handler {
		return injectRequestSettings(next, &cfg.Request)
	})

	return &Application{
		Server:   &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), Handler: router},
		Router:   router,
		Settings: cfg,
	}
}

// defaultMiddleware returns the middleware chain used in the default [Application] application
func (a *Application) defaultMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.Logger{Logger: a.Settings.Logger}.ContextLogger,
		middleware.PanicRecovery,
		middleware.LogRequests,
	}
}
