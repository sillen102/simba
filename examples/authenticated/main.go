package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sillen102/simba"
)

type ResponseBody struct {
	Message string `json:"message" example:"Hello John Doe, you are an admin"`
}

// User represents a user
// in this example we use a simple struct to represent a user
type User struct {
	ID   int
	Name string
	Role string
}

// authFunc is a function that authenticates and returns a user
// in this example we just return a hard-coded user
//
// @BasicAuth "admin" "admin access only"
func authFunc(r *http.Request) (*User, error) {
	return &User{
		ID:   1,
		Name: "John Doe",
		Role: "admin",
	}, nil
}

// @ID authenticatedHandler
// @Summary authenticated handler
// @Description this is a handler that requires authentication
func authenticatedHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], user *User) (*simba.Response[ResponseBody], error) {

	// Access the request cookies
	// req.Cookies

	// Access the request headers
	// req.Headers

	return &simba.Response[ResponseBody]{
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are an %s", user.Name, user.Role),
		},
	}, nil
}

func main() {
	// the app will use the authFunc to authenticate and retrieve the user
	// for each request that uses the AuthJsonHandler and pass it to the handler
	app := simba.Default()
	app.Router.GET("/user", simba.AuthJsonHandler(authenticatedHandler, authFunc))
	app.Start()
}
