package simba_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"github.com/stretchr/testify/require"
	"github.com/swaggest/openapi-go"
)

type params struct {
	ID        string `path:"id" description:"ID of the user" example:"1"`
	Gender    string `query:"gender" description:"Gender of the user" example:"male"`
	RequestID string `header:"X-Request-ID" description:"Request ID" example:"1234"`
}

type reqBody struct {
	Name        string `json:"name" description:"Name of the user" example:"John Doe"`
	Age         int    `json:"age" description:"Age of the user" example:"30"`
	Description string `json:"description" description:"description of the user" example:"A test user"`
}

type respBody struct {
	ID          string `json:"id" description:"ID of the user" example:"1"`
	Name        string `json:"name" description:"Name of the user" example:"John Doe"`
	Age         int    `json:"age" description:"Age of the user" example:"30"`
	Description string `json:"description" description:"description of the user" example:"A test user"`
}

type basicAuthParams struct {
	Username string `header:"Authorization" description:"Basic auth username"`
}

type apiKeyAuthParams struct {
	APIKey string `header:"Authorization" description:"API key"`
}

type bearerTokenAuthParams struct {
	Token string `header:"Authorization" description:"Bearer token"`
}

type user struct {
	ID   int
	Name string
}

// @ID testHandler
// @Deprecated
// @Tag Test
// @Tag User
// @Summary test handler
// @Description this is a multiline
//
// description for the handler
// @Error 409 Resource already exists
func handler(ctx context.Context, req *simba.Request[reqBody, params]) (*simba.Response[respBody], error) {
	return &simba.Response[respBody]{
		Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
		Headers: http.Header{"X-Request-ID": []string{req.Params.RequestID}},
		Body: respBody{
			ID:          req.Params.ID,
			Name:        req.Body.Name,
			Age:         req.Body.Age,
			Description: req.Body.Description,
		},
		Status: http.StatusCreated,
	}, nil
}

func TestOpenAPIDocsGen(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(&reqBody{
		Name:        "John Doe",
		Age:         30,
		Description: "A test user",
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/test/1?gender=male", io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "1234")
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: "Bearer token"})
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test/{id}", simba.JsonHandler(handler))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	yamlContent := getW.Body.String()
	require.Contains(t, yamlContent, "/test/{id}")
	require.Contains(t, yamlContent, "X-Request-ID")
	require.Contains(t, yamlContent, "description: ID of the user")
	require.Contains(t, yamlContent, "description: Name of the user")
	require.Contains(t, yamlContent, "description: Age of the user")
	require.Contains(t, yamlContent, "description: Gender of the user")
	require.Contains(t, yamlContent, "description: description of the user")
	require.Contains(t, yamlContent, "description: Request body contains invalid data")
	require.Contains(t, yamlContent, "description: Request body could not be processed")
	require.Contains(t, yamlContent, "description: Unexpected error")
	require.Contains(t, yamlContent, "description: Resource already exists")
	require.Contains(t, yamlContent, `
      description: |-
        this is a multiline

        description for the handler`,
	)
	require.Contains(t, yamlContent, "operationId: testHandler")
	require.Contains(t, yamlContent, "summary: test handler")
	require.Contains(t, yamlContent, "deprecated: true")
	require.Contains(t, yamlContent, "\"201\":")
	require.Contains(t, yamlContent, "tags:", "- Test", "- User")
	require.Contains(t, yamlContent, "- User")
	require.Contains(t, yamlContent, "- Test")
}

func basicAuthFunc(ctx context.Context, req *simba.AuthRequest[basicAuthParams]) (*user, error) {
	return &user{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var basicAuthAuthenticationHandler = simba.BasicAuth[basicAuthParams, user](
	basicAuthFunc,
	simba.BasicAuthConfig{
		Name:        "admin",
		Description: "admin access only",
	})

// @ID basicAuthHandler
// @Summary basic auth handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func basicAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *user) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

func TestOpenAPIDocsGenBasicAuthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(basicAuthHandler, basicAuthAuthenticationHandler))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	yamlContent := getW.Body.String()
	require.Contains(t, yamlContent, "/test")
	require.Contains(t, yamlContent, `
      description: |
        this is a multiline

        description for the handler`,
	)
	require.Contains(t, yamlContent, `
  securitySchemes:
    admin:
      description: admin access only
      scheme: basic
      type: http`,
	)
	require.Contains(t, yamlContent, `
      security:
      - admin: []`,
	)
	require.Contains(t, yamlContent, "operationId: basicAuthHandler")
	require.Contains(t, yamlContent, "summary: basic auth handler")
	require.NotContains(t, yamlContent, "deprecated: true")
	require.Contains(t, yamlContent, "\"204\":")
}

func TestMultipleAuthHandlers(t *testing.T) {
	t.Parallel()

	app := simba.Default()

	req1 := httptest.NewRequest(http.MethodPost, "/test1", nil)
	w1 := httptest.NewRecorder()
	app.Router.POST("/test1", simba.AuthJsonHandler(basicAuthHandler, basicAuthAuthenticationHandler))
	app.Router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/test2", nil)
	w2 := httptest.NewRecorder()
	app.Router.POST("/test2", simba.AuthJsonHandler(basicAuthHandler, basicAuthAuthenticationHandler))
	app.Router.ServeHTTP(w2, req2)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	yamlContent := getW.Body.String()
	require.Contains(t, yamlContent, "/test1")
	require.Contains(t, yamlContent, "/test2")
	require.Contains(t, yamlContent, `
  securitySchemes:
    admin:
      description: admin access only
      scheme: basic
      type: http`,
	)
	require.Contains(t, yamlContent, `
      security:
      - admin: []`,
	)
	require.Contains(t, yamlContent, "operationId: basicAuthHandler")
}

func apiKeyAuthFunc(ctx context.Context, req *simba.AuthRequest[apiKeyAuthParams]) (*user, error) {
	return &user{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var apiKeyAuthAuthenticationHandler = simba.APIKeyAuth[apiKeyAuthParams, user](
	apiKeyAuthFunc,
	simba.APIKeyAuthConfig{
		Name:        "User",
		FieldName:   "sessionid",
		In:          openapi.InCookie,
		Description: "Session cookie",
	})

// @ID apiKeyAuthHandler
// @Summary api key handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func apiKeyAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *user) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

func TestOpenAPIDocsGenAPIKeyAuthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "APIKey token")
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(apiKeyAuthHandler, apiKeyAuthAuthenticationHandler))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	yamlContent := getW.Body.String()
	require.Contains(t, yamlContent, "/test")
	require.Contains(t, yamlContent, `
      description: |
        this is a multiline

        description for the handler`,
	)
	require.Contains(t, yamlContent, `
  securitySchemes:
    User:
      description: Session cookie
      in: cookie
      name: sessionid
      type: apiKey`,
	)
	require.Contains(t, yamlContent, `
      security:
      - User: []`,
	)
	require.Contains(t, yamlContent, "operationId: apiKeyAuthHandler")
	require.Contains(t, yamlContent, "summary: api key handler")
}

func bearerAuthFunc(ctx context.Context, req *simba.AuthRequest[bearerTokenAuthParams]) (*user, error) {
	return &user{
		ID:   1,
		Name: "John Doe",
	}, nil
}

var bearerAuthAuthenticationHandler = simba.BearerAuth[bearerTokenAuthParams, user](
	bearerAuthFunc,
	simba.BearerAuthConfig{
		Name:        "admin",
		Format:      "jwt",
		Description: "Bearer token",
	})

// @ID  bearerTokenAuthHandler
// @Summary  bearer token handler
// @Description this is a multiline
//
// description for the handler
//
// @Error  409 	Resource already exists
func bearerTokenAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *user) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

func TestOpenAPIDocsGenBearerTokenAuthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(bearerTokenAuthHandler, bearerAuthAuthenticationHandler))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	yamlContent := getW.Body.String()
	require.Contains(t, yamlContent, "/test")
	require.Contains(t, yamlContent, `
      description: |
        this is a multiline

        description for the handler`,
	)
	require.Contains(t, yamlContent, `
  securitySchemes:
    admin:
      bearerFormat: jwt
      description: Bearer token
      scheme: bearer
      type: http`,
	)
	require.Contains(t, yamlContent, `
      security:
      - admin: []`,
	)
	require.Contains(t, yamlContent, "operationId: bearerTokenAuthHandler")
	require.Contains(t, yamlContent, "summary: bearer token handler")
}

// A dummy function to test the OpenAPI generation without any tags.
func noTagsHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

func TestOpenAPIGenNoTags(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.JsonHandler(noTagsHandler))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	yamlContent := getW.Body.String()
	require.Contains(t, yamlContent, "/test")
	require.Contains(t, yamlContent, "description: A dummy function to test the OpenAPI generation without any tags.")
	require.Contains(t, yamlContent, "operationId: no-tags-handler")
	require.Contains(t, yamlContent, "summary: No tags handler")
	require.Contains(t, yamlContent, "tags:")
	require.Contains(t, yamlContent, "- SimbaTest")
}
