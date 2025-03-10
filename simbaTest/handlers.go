package simbaTest

import (
	"context"
	"net/http"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
	"github.com/swaggest/openapi-go"
)

// Receiver A dummy struct to test the OpenAPI generation with receiver functions.
type Receiver struct{}

// NoTagsHandler A dummy function to test the OpenAPI generation without any tags in the comment.
func (h *Receiver) NoTagsHandler(ctx context.Context, req *simbaModels.Request[RequestBody, Params]) (*simbaModels.Response[ResponseBody], error) {
	return &simbaModels.Response[ResponseBody]{
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
func (h *Receiver) TagsHandler(ctx context.Context, req *simbaModels.Request[RequestBody, Params]) (*simbaModels.Response[ResponseBody], error) {
	return &simbaModels.Response[ResponseBody]{
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

// DeprecatedHandler A dummy function to test the OpenAPI generation with deprecated tag.
// @Deprecated
func (h *Receiver) DeprecatedHandler(ctx context.Context, req *simbaModels.Request[RequestBody, Params]) (*simbaModels.Response[ResponseBody], error) {
	return &simbaModels.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Request-ID": []string{req.Params.RequestID}},
		Body: ResponseBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
	}, nil
}

// NoTagsHandler A dummy function to test the OpenAPI generation without any tags in the comment.
func NoTagsHandler(ctx context.Context, req *simbaModels.Request[RequestBody, Params]) (*simbaModels.Response[ResponseBody], error) {
	return &simbaModels.Response[ResponseBody]{
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

// TagsHandler A dummy function to test the OpenAPI generation with tags in the comment.
// @ID testHandler
// @Tag Test
// @Tag User
// @Summary test handler
// @Description this is a multiline
// @StatusCode 201
//
// description for the handler
// @Error 409 Resource already exists
func TagsHandler(ctx context.Context, req *simbaModels.Request[RequestBody, Params]) (*simbaModels.Response[ResponseBody], error) {
	return &simbaModels.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Request-ID": []string{req.Params.RequestID}},
		Body: ResponseBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
	}, nil
}

// DeprecatedHandler A dummy function to test the OpenAPI generation with deprecated tag.
// @Deprecated
func DeprecatedHandler(ctx context.Context, req *simbaModels.Request[RequestBody, Params]) (*simbaModels.Response[ResponseBody], error) {
	return &simbaModels.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Request-ID": []string{req.Params.RequestID}},
		Body: ResponseBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
	}, nil
}

func BasicAuthFunc(ctx context.Context, username, password string) (*User, error) {
	if !(username == "user" && password == "password") {
		return nil, simbaErrors.NewHttpError(http.StatusUnauthorized, "invalid username or password", nil)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var BasicAuthAuthenticationHandler = simba.BasicAuth[User](
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
func BasicAuthHandler(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams], auth *User) (*simbaModels.Response[simbaModels.NoBody], error) {
	return &simbaModels.Response[simbaModels.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// ApiKeyAuthFunc A dummy function to test the OpenAPI generation with api key auth.
// @APIKeyAuth "User" "sessionid" "cookie" "Session cookie"
func ApiKeyAuthFunc(ctx context.Context, apiKey string) (*User, error) {
	if apiKey != "valid-key" {
		return nil, simbaErrors.NewHttpError(http.StatusUnauthorized, "invalid api key", nil)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var ApiKeyAuthAuthenticationHandler = simba.APIKeyAuth[User](
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
func ApiKeyAuthHandler(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams], auth *User) (*simbaModels.Response[simbaModels.NoBody], error) {
	return &simbaModels.Response[simbaModels.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// BearerAuthFunc A dummy function to test the OpenAPI generation with bearer token auth.
// @BearerAuth "admin" "jwt" "Bearer token"
func BearerAuthFunc(ctx context.Context, token string) (*User, error) {
	if token != "token" {
		return nil, simbaErrors.NewHttpError(http.StatusUnauthorized, "invalid token", nil)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var BearerAuthAuthenticationHandler = simba.BearerAuth[User](
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
func BearerTokenAuthHandler(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams], auth *User) (*simbaModels.Response[simbaModels.NoBody], error) {
	return &simbaModels.Response[simbaModels.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}
