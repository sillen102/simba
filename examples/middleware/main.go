package main

import (
	"context"
	"net/http"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/logging"
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
	app := simba.Default()
	app.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware", "123")
			next.ServeHTTP(w, r)
		})
	})
	app.Router.POST("/users", simba.HandlerFunc(handler))
	logging.GetDefault().Info().Msg("Listening on http://localhost:9999")
	http.ListenAndServe(":9999", app)
}
