package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/swaggest/openapi-go"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/auth"
	"github.com/sillen102/simba/models"
	"github.com/sillen102/simba/simbaErrors"
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
func authFunc(ctx context.Context, apiKey string) (*User, error) {
	if apiKey != "valid-key" {
		return nil, simbaErrors.NewSimbaError(http.StatusUnauthorized, "invalid api key", nil)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
		Role: "admin",
	}, nil
}

var authHandler = auth.APIKeyAuth[*User](
	authFunc,
	auth.APIKeyAuthConfig{
		Name:        "admin",
		FieldName:   "sessionid",
		In:          openapi.InHeader,
		Description: "admin access only",
	},
)

// @ID authenticatedHandler
// @Summary authenticated handler
// @Description this is a handler that requires authentication
func authenticatedHandler(
	ctx context.Context,
	req *models.Request[models.NoBody, struct {
		UserID int `path:"userId"`
	}],
	user *User,
) (*models.Response[ResponseBody], error) {

	// Access the request cookies
	// req.Cookies

	// Access the request headers
	// req.Headers

	return &models.Response[ResponseBody]{
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are an %s", user.Name, user.Role),
		},
	}, nil
}

func main() {
	// the app will use the authFunc to authenticate and retrieve the user
	// for each request that uses the AuthJsonHandler and pass it to the handler
	app := simba.Default()
	app.Router.GET("/users/{userId}", simba.AuthJsonHandler(authenticatedHandler, authHandler))
	app.Start()
}
