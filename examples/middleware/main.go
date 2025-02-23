package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
)

type ResponseBody struct {
	Message string `json:"message"`
}

type Params struct {
	MiddlewareHeader string `header:"X-Middleware"`
}

func handler(ctx context.Context, req *simba.Request[simba.NoBody, Params]) (*simba.Response[ResponseBody], error) {
	return &simba.Response[ResponseBody]{
		Body: ResponseBody{
			Message: "Hello " + req.Params.MiddlewareHeader,
		},
	}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	app := simba.Default(settings.Config{Logger: logger})
	app.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware", "123")
			next.ServeHTTP(w, r)
		})
	})
	app.Router.POST("/users", simba.JsonHandler(handler))
	app.Start()
}
