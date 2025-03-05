package test

import (
	"context"
	"net/http"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaAuth"
	"github.com/swaggest/openapi-go"
)

// Receiver A dummy struct to test the OpenAPI generation with receiver functions.
type Receiver struct{}

// NoTagsHandler A dummy function to test the OpenAPI generation without any tags in the comment.
func (h *Receiver) NoTagsHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// TagsHandler A dummy function to test the OpenAPI generation with tags in the comment with receiver functions.
// @ID testHandler
// @Deprecated
// @Tag Test
// @Tag User
// @Summary test handler
// @Description this is a multiline
//
// description for the handler
// @Error 409 Resource already exists
func (h *Receiver) TagsHandler(ctx context.Context, req *simba.Request[RequestBody, Params]) (*simba.Response[ResponseBody], error) {
	return &simba.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Request-ID": []string{req.Params.RequestID}},
		Body: ResponseBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
		Status: http.StatusCreated,
	}, nil
}

// NoTagsHandler A dummy function to test the OpenAPI generation without any tags in the comment.
func NoTagsHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// TagsHandler A dummy function to test the OpenAPI generation with tags in the comment.
// @ID testHandler
// @Deprecated
// @Tag Test
// @Tag User
// @Summary test handler
// @Description this is a multiline
//
// description for the handler
// @Error 409 Resource already exists
func TagsHandler(ctx context.Context, req *simba.Request[RequestBody, Params]) (*simba.Response[ResponseBody], error) {
	return &simba.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Request-ID": []string{req.Params.RequestID}},
		Body: ResponseBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
		Status: http.StatusCreated,
	}, nil
}

// BasicAuthFunc A dummy function to test the OpenAPI generation with basic auth.
// @BasicAuth "admin" "admin access only"
func BasicAuthFunc(ctx context.Context, req *simba.AuthRequest[BasicAuthParams]) (*User, error) {
	username, password, ok := simbaAuth.BasicAuthDecode(req.Params.Credentials)
	if !ok || username != "user" || password != "password" {
		return nil, simba.NewHttpError(http.StatusUnauthorized, "invalid credentials", nil)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var BasicAuthAuthenticationHandler = simba.BasicAuth[BasicAuthParams, User](
	BasicAuthFunc,
	simba.BasicAuthConfig{
		Name:        "admin",
		Description: "admin access only",
	})

// BasicAuthHandler A dummy function to test the OpenAPI generation with basic auth.
// @ID basicAuthHandler
// @Summary basic auth handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func BasicAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *User) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// ApiKeyAuthFunc A dummy function to test the OpenAPI generation with api key auth.
// @APIKeyAuth "User" "sessionid" "cookie" "Session cookie"
func ApiKeyAuthFunc(ctx context.Context, req *simba.AuthRequest[ApiKeyParams]) (*User, error) {
	if req.Params.APIKey != "valid-key" {
		return nil, simba.NewHttpError(http.StatusUnauthorized, "invalid api key", nil)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var ApiKeyAuthAuthenticationHandler = simba.APIKeyAuth[ApiKeyParams, User](
	ApiKeyAuthFunc,
	simba.APIKeyAuthConfig{
		Name:        "User",
		FieldName:   "sessionid",
		In:          openapi.InCookie,
		Description: "Session cookie",
	})

// ApiKeyAuthHandler A dummy function to test the OpenAPI generation with api key auth.
// @ID apiKeyAuthHandler
// @Summary api key handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func ApiKeyAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *User) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// BearerAuthFunc A dummy function to test the OpenAPI generation with bearer token auth.
// @BearerAuth "admin" "jwt" "Bearer token"
func BearerAuthFunc(ctx context.Context, req *simba.AuthRequest[BearerTokenParams]) (*User, error) {
	if req.Params.Token != "Bearer token" {
		return nil, simba.NewHttpError(http.StatusUnauthorized, "invalid token", nil)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var BearerAuthAuthenticationHandler = simba.BearerAuth[BearerTokenParams, User](
	BearerAuthFunc,
	simba.BearerAuthConfig{
		Name:        "admin",
		Format:      "jwt",
		Description: "Bearer token",
	})

// BearerTokenAuthHandler A dummy function to test the OpenAPI generation with bearer token auth.
// @ID  bearerTokenAuthHandler
// @Summary  bearer token handler
// @Description this is a multiline
//
// description for the handler
//
// @Error  409 	Resource already exists
func BearerTokenAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *User) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}
