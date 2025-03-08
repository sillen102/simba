package simba

import (
	"context"
	"net/http"

	"github.com/sillen102/simba/simbaOpenapi/openapiModels"
	oapi "github.com/swaggest/openapi-go"
)

type AuthHandler[AuthParams, AuthModel any] interface {
	GetType() openapiModels.AuthType
	GetName() string
	GetFieldName() string
	GetFormat() string
	GetDescription() string
	GetIn() oapi.In
	GetHandler() AuthHandlerFunc[AuthParams, AuthModel]
}

type BasicAuthConfig struct {
	Name        string
	Description string
}

// BasicAuth creates a basic auth handler with configuration
func BasicAuth[AuthParams, AuthModel any](
	handler AuthHandlerFunc[AuthParams, AuthModel],
	config BasicAuthConfig,
) AuthHandler[AuthParams, AuthModel] {
	return BasicAuthType[AuthParams, AuthModel]{
		Name:        config.Name,
		Description: config.Description,
		Handler:     handler,
	}
}

type BasicAuthType[AuthParams, AuthModel any] struct {
	Name        string
	Description string
	Handler     AuthHandlerFunc[AuthParams, AuthModel]
}

func (t BasicAuthType[AuthParams, AuthModel]) GetType() openapiModels.AuthType {
	return openapiModels.AuthTypeBasic
}

func (t BasicAuthType[AuthParams, AuthModel]) GetName() string {
	return t.Name
}

func (t BasicAuthType[AuthParams, AuthModel]) GetFieldName() string {
	return "Authorization"
}

func (t BasicAuthType[AuthParams, AuthModel]) GetFormat() string {
	return "Basic"
}

func (t BasicAuthType[AuthParams, AuthModel]) GetDescription() string {
	return t.Description
}

func (t BasicAuthType[AuthParams, AuthModel]) GetIn() oapi.In {
	return oapi.InHeader
}

func (t BasicAuthType[AuthParams, AuthModel]) GetHandler() AuthHandlerFunc[AuthParams, AuthModel] {
	return t.Handler
}

type APIKeyAuthConfig struct {
	Name        string
	FieldName   string
	In          oapi.In
	Description string
}

// APIKeyAuth creates an API key auth handler with configuration
func APIKeyAuth[AuthParams, AuthModel any](
	handler AuthHandlerFunc[AuthParams, AuthModel],
	config APIKeyAuthConfig,
) AuthHandler[AuthParams, AuthModel] {
	return APIKeyAuthType[AuthParams, AuthModel]{
		Name:        config.Name,
		FieldName:   config.FieldName,
		In:          config.In,
		Description: config.Description,
		Handler:     handler,
	}
}

type APIKeyAuthType[AuthParams, AuthModel any] struct {
	Name        string
	FieldName   string
	In          oapi.In
	Description string
	Handler     AuthHandlerFunc[AuthParams, AuthModel]
}

func (t APIKeyAuthType[AuthParams, AuthModel]) GetType() openapiModels.AuthType {
	return openapiModels.AuthTypeAPIKey
}

func (t APIKeyAuthType[AuthParams, AuthModel]) GetName() string {
	return t.Name
}

func (t APIKeyAuthType[AuthParams, AuthModel]) GetFieldName() string {
	return t.FieldName
}

func (t APIKeyAuthType[AuthParams, AuthModel]) GetFormat() string {
	return ""
}

func (t APIKeyAuthType[AuthParams, AuthModel]) GetDescription() string {
	return t.Description
}

func (t APIKeyAuthType[AuthParams, AuthModel]) GetIn() oapi.In {
	return t.In
}

func (t APIKeyAuthType[AuthParams, AuthModel]) GetHandler() AuthHandlerFunc[AuthParams, AuthModel] {
	return t.Handler
}

type BearerAuthConfig struct {
	Name        string
	Format      string
	Description string
}

// BearerAuth creates a bearer auth handler with configuration
func BearerAuth[AuthParams, AuthModel any](
	handler AuthHandlerFunc[AuthParams, AuthModel],
	config BearerAuthConfig,
) AuthHandler[AuthParams, AuthModel] {
	return BearerAuthType[AuthParams, AuthModel]{
		Name:        config.Name,
		Format:      config.Format,
		Description: config.Description,
		Handler:     handler,
	}
}

type BearerAuthType[AuthParams, AuthModel any] struct {
	Name        string
	Format      string
	Description string
	Handler     AuthHandlerFunc[AuthParams, AuthModel]
}

func (t BearerAuthType[AuthParams, AuthModel]) GetType() openapiModels.AuthType {
	return openapiModels.AuthTypeBearer
}

func (t BearerAuthType[AuthParams, AuthModel]) GetName() string {
	return t.Name
}

func (t BearerAuthType[AuthParams, AuthModel]) GetFieldName() string {
	return "Authorization"
}

func (t BearerAuthType[AuthParams, AuthModel]) GetFormat() string {
	return t.Format
}

func (t BearerAuthType[AuthParams, AuthModel]) GetDescription() string {
	return t.Description
}

func (t BearerAuthType[AuthParams, AuthModel]) GetIn() oapi.In {
	return oapi.InHeader
}

func (t BearerAuthType[AuthParams, AuthModel]) GetHandler() AuthHandlerFunc[AuthParams, AuthModel] {
	return t.Handler
}

// AuthRequest is a request object that contains the parameters needed to authenticate a request.
// The parameters are taken from the headers, cookies, or query parameters of the request and should be
// specified using tags in the struct definition.
type AuthRequest[AuthParams any] struct {
	Params AuthParams
}

// AuthHandlerFunc is a function that handles authentication for a route.
type AuthHandlerFunc[AuthParams, AuthModel any] func(ctx context.Context, req *AuthRequest[AuthParams]) (*AuthModel, error)

// handleAuthRequest is a helper function that handles parses the parameters and calls the authentication
// function with the parsed parameters.
func handleAuthRequest[AuthParams, AuthModel any](
	authHandler AuthHandler[AuthParams, AuthModel],
	r *http.Request,
) (*AuthModel, error) {
	params, err := parseAndValidateParams[AuthParams](r)
	if err != nil {
		return nil, err
	}

	return authHandler.GetHandler()(r.Context(), &AuthRequest[AuthParams]{
		Params: params,
	})
}
