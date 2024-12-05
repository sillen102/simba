package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/logging"
)

type RequestBody struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

func handler(ctx context.Context, req *simba.Request[RequestBody, simba.NoParams]) (*simba.Response, error) {

	// Access the request body fields
	// req.Body.Age
	// req.Body.Name

	// Access the request cookies
	// req.Cookies

	// Access the request headers
	// req.Headers

	return &simba.Response{
		Headers: map[string][]string{"My-Header": {"header-value"}},
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are %d years old", req.Body.Name, req.Body.Age),
		},
		Status: http.StatusOK, // Can be omitted, defaults to 200 if there's a body, 204 if there's no body
	}, nil
}

func noBodyHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
	return &simba.Response{}, nil // Returns 204 since there is no body in the response
}

func main() {
	app := simba.Default()
	app.Router.POST("/users", simba.HandlerFunc(handler))
	app.Router.GET("/no-body", simba.HandlerFunc(noBodyHandler))
	logging.GetDefault().Info().Msg("Listening on http://localhost:9999")
	http.ListenAndServe(":9999", app)
}
