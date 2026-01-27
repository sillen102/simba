package simba

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func (a *Application) Start() {
	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	log := a.Settings.Logger

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a goroutine to watch for the stop signal and cancel the context
	go func() {
		<-stop
		cancel()
	}()

	// Generate OpenAPI documentation in a goroutine
	go func() {
		log.Debug("generating OpenAPI documentation...")
		err := a.Router.GenerateOpenAPIDocumentation(ctx, a.Settings.Name, a.Settings.Version)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Debug("OpenAPI documentation generation cancelled")
			} else {
				log.Error("error generating OpenAPI documentation", "error", err)
			}
			return
		}
		log.Debug("OpenAPI documentation generated")
	}()

	// Run server in a goroutine
	go func() {
		log.Info("server listening on " + a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("error starting server", "error", err)
			panic(err)
		}
	}()

	// Wait for context cancellation (triggered by the stop signal)
	<-ctx.Done()

	// Gracefully shutdown the server
	log.Info("shutting down server...")
	if err := a.Stop(); err != nil {
		log.Error("error shutting down server", "error", err)
		panic(err)
	}
}

func (a *Application) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown telemetry provider first to ensure all spans and metrics are exported
	if a.telemetryProvider != nil {
		if err := a.telemetryProvider.Shutdown(ctx); err != nil {
			a.Settings.Logger.Error("Failed to shutdown telemetry provider", "error", err)
		} else {
			a.Settings.Logger.Debug("Telemetry provider shutdown successfully")
		}
	}

	// Then shutdown the HTTP server
	return a.Server.Shutdown(ctx)
}
