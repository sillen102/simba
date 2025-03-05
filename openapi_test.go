package simba_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/test"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIDocsGen(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(&test.RequestBody{
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
	app.Router.POST("/test/{id}", simba.JsonHandler(test.TagsHandler))
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

func TestOpenAPIDocsGenBasicAuthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(test.BasicAuthHandler, test.BasicAuthFunc))
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
	require.Contains(t, yamlContent, "\"202\":")
}

func TestMultipleAuthHandlers(t *testing.T) {
	t.Parallel()

	app := simba.Default()

	req1 := httptest.NewRequest(http.MethodPost, "/test1", nil)
	w1 := httptest.NewRecorder()
	app.Router.POST("/test1", simba.AuthJsonHandler(test.BasicAuthHandler, test.BasicAuthFunc))
	app.Router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/test2", nil)
	w2 := httptest.NewRecorder()
	app.Router.POST("/test2", simba.AuthJsonHandler(test.BasicAuthHandler, test.BasicAuthFunc))
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

func TestOpenAPIDocsGenAPIKeyAuthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "APIKey token")
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(test.ApiKeyAuthHandler, test.ApiKeyAuthFunc))
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

func TestOpenAPIDocsGenBearerTokenAuthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(test.BearerTokenAuthHandler, test.BearerAuthFunc))
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
	require.Contains(t, yamlContent, "\"202\":")
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

func TestOpenAPIGenNoTags(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.JsonHandler(test.NoTagsHandler))
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
	require.Contains(t, yamlContent, "description: A dummy function to test the OpenAPI generation without any tags")
	require.Contains(t, yamlContent, "operationId: no-tags-handler")
	require.Contains(t, yamlContent, "summary: No tags handler")
	require.Contains(t, yamlContent, "tags:")
	require.Contains(t, yamlContent, "- Test")
	require.Contains(t, yamlContent, "\"202\":")
}

func TestOpenAPIGenNoTagsReceiverFuncHandler(t *testing.T) {
	t.Parallel()

	receiver := test.Receiver{}

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	app := simba.Default()
	app.Router.POST("/test", simba.JsonHandler(receiver.NoTagsHandler))
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
	require.Contains(t, yamlContent, "description: A dummy function to test the OpenAPI generation without any tags")
	require.Contains(t, yamlContent, "operationId: no-tags-handler")
	require.Contains(t, yamlContent, "summary: No tags handler")
	require.Contains(t, yamlContent, "tags:")
	require.Contains(t, yamlContent, "- Test")
	require.Contains(t, yamlContent, "\"202\":")
}
