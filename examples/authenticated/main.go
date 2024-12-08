package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sillen102/simba"
)

type ResponseBody struct {
	Message string `json:"message"`
}

// User is a struct that represents a user
// in this example we use a simple struct to represent a user
type User struct {
	ID   int
	Name string
	Role string
}

// authFunc is a function that authenticates and returns a user
// in this example we just return a hard-coded user
func authFunc(r *http.Request) (*User, error) {
	return &User{
		ID:   1,
		Name: "John Doe",
		Role: "admin",
	}, nil
}

func authenticatedHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], user *User) (*simba.Response, error) {

	// Access the request cookies
	// req.Cookies

	// Access the request headers
	// req.Headers

	return &simba.Response{
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are an %s", user.Name, user.Role),
		},
	}, nil
}

func main() {
	// the app will use the authFunc to authenticate and retrieve the user
	// for each request that uses the AuthHandlerFunc and pass it to the handler
	app := simba.DefaultAuthWith(authFunc)
	app.Router.GET("/user", simba.AuthHandlerFunc(authenticatedHandler))

	app.GetLogger().Info().Msg("Listening on http://localhost:9999")
	http.ListenAndServe(":9999", app)
}
