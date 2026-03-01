package simbaTest

import (
	"context"
	"errors"
	"net/http"

	"github.com/swaggest/openapi-go"

	"github.com/sillen102/simba/auth"
	"github.com/sillen102/simba/constants"
	"github.com/sillen102/simba/models"
	"github.com/sillen102/simba/simbaErrors"
)

// Receiver A dummy struct to test the OpenAPI generation with receiver functions.
type Receiver struct{}

// NoTagsHandler A dummy function to test the OpenAPI generation without any tags in the comment.
func (h *Receiver) NoTagsHandler(_ context.Context, req *models.Request[RequestBody, Params]) (*models.Response[ResponseBody], error) {
	return &models.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Trace-ID": []string{req.Params.TraceID}},
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
func (h *Receiver) TagsHandler(_ context.Context, req *models.Request[RequestBody, Params]) (*models.Response[ResponseBody], error) {
	return &models.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Trace-ID": []string{req.Params.TraceID}},
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
func (h *Receiver) DeprecatedHandler(ctx context.Context, req *models.Request[RequestBody, Params]) (*models.Response[ResponseBody], error) {
	return &models.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Trace-ID": []string{req.Params.TraceID}},
		Body: ResponseBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
	}, nil
}

// NoTagsHandler A dummy function to test the OpenAPI generation without any tags in the comment.
func NoTagsHandler(ctx context.Context, req *models.Request[RequestBody, Params]) (*models.Response[ResponseBody], error) {
	return &models.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Trace-ID": []string{req.Params.TraceID}},
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
func TagsHandler(_ context.Context, req *models.Request[RequestBody, Params]) (*models.Response[ResponseBody], error) {
	return &models.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Trace-ID": []string{req.Params.TraceID}},
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
func DeprecatedHandler(ctx context.Context, req *models.Request[RequestBody, Params]) (*models.Response[ResponseBody], error) {
	return &models.Response[ResponseBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Trace-ID": []string{req.Params.TraceID}},
		Body: ResponseBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
	}, nil
}

func BasicAuthFunc(_ context.Context, username, password string) (*User, error) {
	if username != "user" || password != "password" {
		return nil, simbaErrors.NewSimbaError(
			http.StatusUnauthorized,
			constants.UnauthorizedErrMsg,
			errors.New("invalid username or password"),
		)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var BasicAuthAuthenticationHandler = auth.BasicAuth[*User](
	BasicAuthFunc,
	auth.BasicAuthConfig{
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
func BasicAuthHandler(_ context.Context, req *models.Request[models.NoBody, models.NoParams], auth *User) (*models.Response[models.NoBody], error) {
	return &models.Response[models.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// ApiKeyAuthFunc A dummy function to test the OpenAPI generation with api key auth.
// @APIKeyAuth "User" "sessionid" "cookie" "Session cookie"
func ApiKeyAuthFunc(_ context.Context, apiKey string) (*User, error) {
	if apiKey != "valid-key" {
		return nil, simbaErrors.NewSimbaError(
			http.StatusUnauthorized,
			constants.UnauthorizedErrMsg,
			errors.New("invalid api key"),
		)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var ApiKeyAuthAuthenticationHandler = auth.APIKeyAuth[*User](
	ApiKeyAuthFunc,
	auth.APIKeyAuthConfig{
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
func ApiKeyAuthHandler(_ context.Context, req *models.Request[models.NoBody, models.NoParams], auth *User) (*models.Response[models.NoBody], error) {
	return &models.Response[models.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

// BearerAuthFunc A dummy function to test the OpenAPI generation with bearer token auth.
// @BearerAuth "admin" "jwt" "Bearer token"
func BearerAuthFunc(_ context.Context, token string) (*User, error) {
	if token != "token" {
		return nil, simbaErrors.NewSimbaError(
			http.StatusUnauthorized,
			constants.UnauthorizedErrMsg,
			errors.New("invalid token"),
		)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var BearerAuthAuthenticationHandler = auth.BearerAuth[*User](
	BearerAuthFunc,
	auth.BearerAuthConfig{
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
func BearerTokenAuthHandler(ctx context.Context, req *models.Request[models.NoBody, models.NoParams], auth *User) (*models.Response[models.NoBody], error) {
	return &models.Response[models.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

func SessionCookieAuthFunc(_ context.Context, sessionID string) (*User, error) {
	if sessionID != "valid-cookie" {
		return nil, simbaErrors.NewSimbaError(
			http.StatusUnauthorized,
			constants.UnauthorizedErrMsg,
			errors.New("invalid session cookie"),
		)
	}

	return &User{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var SessionCookieAuthAuthenticationHandler = auth.SessionCookieAuth[*User](
	SessionCookieAuthFunc,
	auth.SessionCookieAuthConfig[*User]{
		CookieName:  "session",
		Description: "Session cookie",
	},
)

// SessionCookieAuthHandler A dummy function to test the OpenAPI generation with session cookie auth.
// @ID sessionCookieAuthHandler
// @Summary session cookie handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func SessionCookieAuthHandler(_ context.Context, req *models.Request[models.NoBody, models.NoParams], auth *User) (*models.Response[models.NoBody], error) {
	return &models.Response[models.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}
