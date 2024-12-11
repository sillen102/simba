package simba

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func (a *Application[AuthModel]) Start() {
	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Run server in a goroutine
	go func() {
		a.logger.Info("server listening on " + a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("error starting server", "error", err)
			panic(err)
		}
	}()
	<-stop

	// Gracefully shutdown the server
	a.logger.Info("shutting down server...")
	if err := a.Stop(); err != nil {
		a.logger.Error("error shutting down server", "error", err)
		panic(err)
	}
}

func (a *Application[AuthModel]) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.Server.Shutdown(ctx)
}
