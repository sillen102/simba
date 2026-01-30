package simba

import (
	"fmt"
	"net/http"

	"github.com/sillen102/simba/middleware"
	"github.com/sillen102/simba/settings"
)

// Application is the main application struct that holds the Mux and other application Settings
type Application struct {

	// ApplicationName is the name of the application
	ApplicationName string `yaml:"application-name" env:"APPLICATION_NAME" default:"Simba Application"`

	// ApplicationVersion is the version of the application
	ApplicationVersion string `yaml:"application-version" env:"APPLICATION_VERSION" default:"0.1.0"`

	// Server is the HTTP server for the application
	Server *http.Server

	// Router is the main Mux for the application
	Router *Router

	// Settings is the application Settings
	Settings *settings.Simba

	// telemetryProvider manages tracing and metrics via a pluggable interface
	telemetryProvider TelemetryProvider
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

	router := newRouter(cfg.Request, cfg.Docs)
	router.Use(func(next http.Handler) http.Handler {
		return injectRequestSettings(next, &cfg.Request)
	})

	// Support modular telemetry config if provided; fallback for legacy settings
	telemetryProvider := NoOpTelemetryProvider{}

	return &Application{
		Server:            &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), Handler: router},
		Router:            router,
		Settings:          cfg,
		telemetryProvider: telemetryProvider,
	}
}

// SetTelemetryProvider allows injection or replacement of the TelemetryProvider after application creation.
func (a *Application) SetTelemetryProvider(tp TelemetryProvider) {
	a.telemetryProvider = tp
}

// defaultMiddleware returns the middleware chain used in the default [Application] application
func (a *Application) defaultMiddleware() []func(http.Handler) http.Handler {
	middlewares := []func(http.Handler) http.Handler{
		a.telemetryProvider.TracingMiddleware(),
		a.telemetryProvider.MetricsMiddleware(),
		middleware.TraceID,
		middleware.Logger{Logger: a.Settings.Logger}.ContextLogger,
		middleware.PanicRecovery,
		middleware.LogRequests,
	}

	return middlewares
}
