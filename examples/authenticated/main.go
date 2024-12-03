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
	return &simba.Response{
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are an %s", user.Name, user.Role),
		},
		Status: http.StatusOK, // We can omit this and it will default to 200 OK if the body is not nil and there is no error
	}, nil
}

func main() {
	// the router will use the authFunc to authenticate and retrieve the user
	// for each request that uses the AuthenticatedHandlerFunc and pass it to the handler
	router := simba.DefaultWithAuth[User](authFunc)
	router.GET("/user", simba.AuthenticatedHandlerFunc(authenticatedHandler))
	logging.Get().Info().Msg("Listening on http://localhost:9999")
	http.ListenAndServe(":9999", router)
}
