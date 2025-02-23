package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sillen102/simba"
)

type RequestBody struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

func handler(ctx context.Context, req *simba.Request[RequestBody, simba.NoParams]) (*simba.Response[ResponseBody], error) {

	// Access the request body fields
	// req.Body.Age
	// req.Body.Name

	// Access the request cookies
	// req.Cookies

	// Access the request headers
	// req.Headers

	return &simba.Response[ResponseBody]{
		Headers: map[string][]string{"My-Header": {"header-value"}},
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are %d years old", req.Body.Name, req.Body.Age),
		},
		Status: http.StatusOK, // Can be omitted, defaults to 200 if there's a body, 204 if there's no body
	}, nil
}

func noBodyHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{}, nil // Returns 204 since there is no body in the response
}

func main() {
	app := simba.Default()
	app.Router.POST("/users", simba.JsonHandler(handler))
	app.Router.GET("/no-body", simba.JsonHandler(noBodyHandler))
	app.Start()
}
