package simba

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaOpenapi/openapiModels"
	oapi "github.com/swaggest/openapi-go"
)

type AuthHandler[AuthModel any] interface {
	GetType() openapiModels.AuthType
	GetName() string
	GetFieldName() string
	GetFormat() string
	GetDescription() string
	GetIn() oapi.In
	GetHandler() AuthHandlerFunc[AuthModel]
}

type BasicAuthConfig struct {
	Name        string
	Description string
}

// BasicAuthHandlerFunc is a function that handles basic auth. This is the function that should be implemented by the user.
// It should return the user model if the username and password are valid, otherwise it should return an error.
type BasicAuthHandlerFunc[AuthModel any] func(ctx context.Context, username, password string) (AuthModel, error)

// BasicAuth creates a basic auth handler with configuration
func BasicAuth[AuthModel any](
	handler BasicAuthHandlerFunc[AuthModel],
	config BasicAuthConfig,
) AuthHandler[AuthModel] {
	return BasicAuthType[AuthModel]{
		Name:        config.Name,
		Description: config.Description,
		Handler:     handler,
	}
}

type BasicAuthType[AuthModel any] struct {
	Name        string
	Description string
	Handler     BasicAuthHandlerFunc[AuthModel]
}

func (t BasicAuthType[AuthModel]) GetType() openapiModels.AuthType {
	return openapiModels.AuthTypeBasic
}

func (t BasicAuthType[AuthModel]) GetName() string {
	return t.Name
}

func (t BasicAuthType[AuthModel]) GetFieldName() string {
	return AuthHeader
}

func (t BasicAuthType[AuthModel]) GetFormat() string {
	return "Basic"
}

func (t BasicAuthType[AuthModel]) GetDescription() string {
	return t.Description
}

func (t BasicAuthType[AuthModel]) GetIn() oapi.In {
	return oapi.InHeader
}

func (t BasicAuthType[AuthModel]) GetHandler() AuthHandlerFunc[AuthModel] {
	return func(r *http.Request) (AuthModel, error) {
		var zero AuthModel

		authHeader := r.Header.Get(AuthHeader)
		if authHeader == "" {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("missing Authorization header"),
			)
		}

		if !strings.HasPrefix(authHeader, BasicPrefix) {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("invalid Authorization header format, expected Basic authentication"),
			)
		}

		encoded := authHeader[len(BasicPrefix):]

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("invalid base64 in Authorization header"),
			)
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("invalid Basic auth format, expected 'username:password'"),
			)
		}

		username := credentials[0]
		password := credentials[1]

		return t.Handler(r.Context(), username, password)
	}
}

type APIKeyAuthConfig struct {
	Name        string
	FieldName   string
	In          oapi.In
	Description string
}

// APIKeyAuthHandlerFunc is a function that handles API key authentication. This is the function that should be implemented by the user.
// It should return the user model if the API key is valid, otherwise it should return an error.
type APIKeyAuthHandlerFunc[AuthModel any] func(ctx context.Context, apiKey string) (AuthModel, error)

// APIKeyAuth creates an API key auth handler with configuration
func APIKeyAuth[AuthModel any](
	handler APIKeyAuthHandlerFunc[AuthModel],
	config APIKeyAuthConfig,
) AuthHandler[AuthModel] {
	return APIKeyAuthType[AuthModel]{
		Name:        config.Name,
		FieldName:   config.FieldName,
		In:          config.In,
		Description: config.Description,
		Handler:     handler,
	}
}

type APIKeyAuthType[AuthModel any] struct {
	Name        string
	FieldName   string
	In          oapi.In
	Description string
	Handler     APIKeyAuthHandlerFunc[AuthModel]
}

func (t APIKeyAuthType[AuthModel]) GetType() openapiModels.AuthType {
	return openapiModels.AuthTypeAPIKey
}

func (t APIKeyAuthType[AuthModel]) GetName() string {
	return t.Name
}

func (t APIKeyAuthType[AuthModel]) GetFieldName() string {
	return t.FieldName
}

func (t APIKeyAuthType[AuthModel]) GetFormat() string {
	return ""
}

func (t APIKeyAuthType[AuthModel]) GetDescription() string {
	return t.Description
}

func (t APIKeyAuthType[AuthModel]) GetIn() oapi.In {
	return t.In
}

func (t APIKeyAuthType[AuthModel]) GetHandler() AuthHandlerFunc[AuthModel] {
	return func(r *http.Request) (AuthModel, error) {
		var zero AuthModel

		apiKey := r.Header.Get(t.FieldName)
		if apiKey == "" {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("missing API key"),
			)
		}

		return t.Handler(r.Context(), apiKey)
	}
}

type BearerAuthConfig struct {
	Name        string
	Format      string
	Description string
}

// BearerAuthHandlerFunc is a function that handles bearer token authentication.
// This is the function that should be implemented by the user. It should return the user model
// if the token is valid, otherwise it should return an error.
type BearerAuthHandlerFunc[AuthModel any] func(ctx context.Context, token string) (AuthModel, error)

// BearerAuth creates a bearer auth handler with configuration
func BearerAuth[AuthModel any](
	handler BearerAuthHandlerFunc[AuthModel],
	config BearerAuthConfig,
) AuthHandler[AuthModel] {
	return BearerAuthType[AuthModel]{
		Name:        config.Name,
		Format:      config.Format,
		Description: config.Description,
		Handler:     handler,
	}
}

type BearerAuthType[AuthModel any] struct {
	Name        string
	Format      string
	Description string
	Handler     BearerAuthHandlerFunc[AuthModel]
}

func (t BearerAuthType[AuthModel]) GetType() openapiModels.AuthType {
	return openapiModels.AuthTypeBearer
}

func (t BearerAuthType[AuthModel]) GetName() string {
	return t.Name
}

func (t BearerAuthType[AuthModel]) GetFieldName() string {
	return AuthHeader
}

func (t BearerAuthType[AuthModel]) GetFormat() string {
	return t.Format
}

func (t BearerAuthType[AuthModel]) GetDescription() string {
	return t.Description
}

func (t BearerAuthType[AuthModel]) GetIn() oapi.In {
	return oapi.InHeader
}

func (t BearerAuthType[AuthModel]) GetHandler() AuthHandlerFunc[AuthModel] {
	return func(r *http.Request) (AuthModel, error) {
		var zero AuthModel

		authHeader := r.Header.Get(AuthHeader)
		if authHeader == "" {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("missing Authorization header"),
			)
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("invalid Authorization header format, expected Bearer authentication"),
			)
		}

		token := authHeader[len(BearerPrefix):]
		if token == "" {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("missing token"),
			)
		}

		return t.Handler(r.Context(), token)
	}
}

type SessionCookieAuthConfig[AuthModel any] struct {
	CookieName  string
	Description string
}

// SessionCookieAuthHandlerFunc is a function that handles session cookie authentication.
// This is the function that should be implemented by the user.
// It should return the user model if the API key is valid, otherwise it should return an error.
type SessionCookieAuthHandlerFunc[AuthModel any] func(ctx context.Context, cookie string) (AuthModel, error)

// SessionCookieAuth creates a session cookie auth handler with configuration
func SessionCookieAuth[AuthModel any](
	handler SessionCookieAuthHandlerFunc[AuthModel],
	config SessionCookieAuthConfig[AuthModel],
) AuthHandler[AuthModel] {
	return SessionCookieAuthType[AuthModel]{
		CookieName:  config.CookieName,
		Description: config.Description,
		Handler:     handler,
	}
}

type SessionCookieAuthType[AuthModel any] struct {
	CookieName  string
	Description string
	Handler     SessionCookieAuthHandlerFunc[AuthModel]
}

func (t SessionCookieAuthType[AuthModel]) GetType() openapiModels.AuthType {
	return openapiModels.AuthTypeSessionCookie
}

func (t SessionCookieAuthType[AuthModel]) GetName() string {
	return "SessionCookie"
}

func (t SessionCookieAuthType[AuthModel]) GetFieldName() string {
	return t.CookieName
}

func (t SessionCookieAuthType[AuthModel]) GetFormat() string {
	return ""
}

func (t SessionCookieAuthType[AuthModel]) GetDescription() string {
	return t.Description
}

func (t SessionCookieAuthType[AuthModel]) GetIn() oapi.In {
	return oapi.InCookie
}

func (t SessionCookieAuthType[AuthModel]) GetHandler() AuthHandlerFunc[AuthModel] {
	return func(r *http.Request) (AuthModel, error) {
		var zero AuthModel

		cookie, err := r.Cookie(t.CookieName)
		if err != nil || cookie == nil {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("missing session cookie"),
			)
		}
		if cookie.Value == "" {
			return zero, simbaErrors.NewSimbaError(
				http.StatusUnauthorized,
				UnauthorizedErrMsg,
				errors.New("empty session cookie"),
			)
		}
		token := cookie.Value

		return t.Handler(r.Context(), token)
	}
}

// AuthHandlerFunc is a function that handles authentication for a route.
type AuthHandlerFunc[AuthModel any] func(r *http.Request) (AuthModel, error)

// handleAuthRequest is a helper function that handles parses the parameters and calls the authentication
// function with the parsed parameters.
func handleAuthRequest[AuthModel any](
	authHandler AuthHandler[AuthModel],
	r *http.Request,
) (AuthModel, error) {
	return authHandler.GetHandler()(r)
}
