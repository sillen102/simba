package simba_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"github.com/stretchr/testify/require"
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

// @ID testHandler
// @Deprecated
// @Summary test handler
// @Description this is a multiline
//
// description for the handler
//
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

	fmt.Println(getW.Body.String())

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
      description: |
        this is a multiline

        description for the handler`,
	)
	require.Contains(t, yamlContent, "operationId: testHandler")
	require.Contains(t, yamlContent, "summary: test handler")
	require.Contains(t, yamlContent, "deprecated: true")
}

type basicAuthModel struct {
	Username string `header:"Authorization" description:"Basic auth username"`
}

type apiKeyAuthModel struct {
	APIKey string `header:"Authorization" description:"API key"`
}

type bearerTokenAuthModel struct {
	Token string `header:"Authorization" description:"Bearer token"`
}

// @BasicAuth "admin" "admin access only"
func basicAuthFunc(r *http.Request) (*basicAuthModel, error) {
	return &basicAuthModel{
		Username: "admin",
	}, nil
}

// @ID basicAuthHandler
// @Summary basic auth handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func basicAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *basicAuthModel) (*simba.Response[simba.NoBody], error) {
	return &simba.Response[simba.NoBody]{
		Status: http.StatusAccepted,
	}, nil
}

func TestOpenAPIDocsGenBasicAuthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(basicAuthHandler, basicAuthFunc))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	fmt.Println(getW.Body.String())

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
}

func TestMultipleAuthHandlers(t *testing.T) {
	t.Parallel()

	app := simba.Default()

	req1 := httptest.NewRequest(http.MethodPost, "/test1", nil)
	w1 := httptest.NewRecorder()
	app.Router.POST("/test1", simba.AuthJsonHandler(basicAuthHandler, basicAuthFunc))
	app.Router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/test2", nil)
	w2 := httptest.NewRecorder()
	app.Router.POST("/test2", simba.AuthJsonHandler(basicAuthHandler, basicAuthFunc))
	app.Router.ServeHTTP(w2, req2)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	fmt.Println(getW.Body.String())

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

// @APIKeyAuth "User" "sessionid" "cookie" "Session cookie"
func apiKeyAuthFunc(r *http.Request) (*apiKeyAuthModel, error) {
	return &apiKeyAuthModel{
		APIKey: "token",
	}, nil
}

// @ID apiKeyAuthHandler
// @Summary api key handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func apiKeyAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *apiKeyAuthModel) (*simba.Response[simba.NoBody], error) {
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
	app.Router.POST("/test", simba.AuthJsonHandler(apiKeyAuthHandler, apiKeyAuthFunc))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	fmt.Println(getW.Body.String())

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

// @BearerAuth "admin" "jwt" "Bearer token"
func bearerAuthFunc(r *http.Request) (*bearerTokenAuthModel, error) {
	return &bearerTokenAuthModel{
		Token: "token",
	}, nil
}

// @ID bearerTokenAuthHandler
// @Summary bearer token handler
// @Description this is a multiline
//
// description for the handler
//
// @Error 409 Resource already exists
func bearerTokenAuthHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], auth *bearerTokenAuthModel) (*simba.Response[simba.NoBody], error) {
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
	app.Router.POST("/test", simba.AuthJsonHandler(bearerTokenAuthHandler, bearerAuthFunc))
	app.Router.ServeHTTP(w, req)

	// Fetch OpenAPI documentation
	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	fmt.Println(getW.Body.String())

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
