package simba

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func (a *Application[AuthModel]) Start(ctx context.Context) {
	logger := LoggerFrom(ctx)

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Run server in a goroutine
	go func() {
		logger.Info().Msg("server listening on " + a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("error starting server")
			panic(err)
		}
	}()
	<-stop

	// Gracefully shutdown the server
	logger.Info().Msg("shutting down server...")
	if err := a.Stop(); err != nil {
		logger.Error().Err(err).Msg("error shutting down server")
		panic(err)
	}
}

func (a *Application[AuthModel]) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.Server.Shutdown(ctx)
}
