package main

import (
	"context"
	"net/http"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/settings"
)

type ResponseBody struct {
	Message string `json:"message"`
}

type Params struct {
	MiddlewareHeader string `header:"X-Middleware"`
}

func handler(ctx context.Context, req *simba.Request[simba.NoBody, Params]) (*simba.Response, error) {
	return &simba.Response{
		Body: ResponseBody{
			Message: "Hello " + req.Params.MiddlewareHeader,
		},
	}, nil
}

func main() {
	app := simba.Default(settings.Settings{
		Logging: logging.Config{
			Format: logging.JsonFormat,
		},
	})
	app.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware", "123")
			next.ServeHTTP(w, r)
		})
	})
	app.Router.POST("/users", simba.JsonHandler(handler))
	app.Start()
}
