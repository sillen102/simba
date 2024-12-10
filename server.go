package simba

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sillen102/simba/logging"
)

func (a *Application[AuthModel]) Start(ctx context.Context) {
	logger := logging.From(ctx)

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Run server in a goroutine
	go func() {
		logger.Info("server listening on " + a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error starting server", "error", err)
			panic(err)
		}
	}()
	<-stop

	// Gracefully shutdown the server
	logger.Info("shutting down server...")
	if err := a.Stop(); err != nil {
		logger.Error("error shutting down server", "error", err)
		panic(err)
	}
}

func (a *Application[AuthModel]) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.Server.Shutdown(ctx)
}
