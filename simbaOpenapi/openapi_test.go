package simbaOpenapi_test

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
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIDocsGen(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(&simbaTest.RequestBody{
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

	app := simbaTest.Default()
	app.Router.POST("/test/{id}", simba.JsonHandler(simbaTest.TagsHandler))

	app.RunTest(func() {
		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)

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
	})
}

func TestValidateRequiredField(t *testing.T) {
	t.Parallel()

	type reqBody struct {
		Name string `json:"name" validate:"required"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[reqBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	app := simbaTest.Default()
	app.Router.POST("/test", simba.JsonHandler(handler))

	body, err := json.Marshal(&reqBody{Name: "John Doe"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/test", io.NopCloser(bytes.NewReader(body)))

	w := httptest.NewRecorder()

	app.RunTest(func() {
		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)

		require.Contains(t, yamlContent, `
      required:
      - name`)
	})
}

func TestValidateMinField(t *testing.T) {
	t.Parallel()

	type reqBody struct {
		Size int `json:"size" validate:"min=5"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[reqBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	app := simbaTest.Default()
	app.Router.POST("/test", simba.JsonHandler(handler))

	body, err := json.Marshal(&reqBody{Size: 5})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/test", io.NopCloser(bytes.NewReader(body)))

	w := httptest.NewRecorder()

	app.RunTest(func() {
		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)
		fmt.Println(yamlContent)

		require.Contains(t, yamlContent, `
    SimbaOpenapiTestReqBody:
      properties:
        size:
          minimum: 5
          type: integer
      type: object`)
	})
}

func TestValidateMaxField(t *testing.T) {
	t.Parallel()

	type reqBody struct {
		Size int `json:"size" validate:"max=5"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[reqBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	app := simbaTest.Default()
	app.Router.POST("/test", simba.JsonHandler(handler))

	app.RunTest(func() {
		body, err := json.Marshal(&reqBody{Size: 5})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/test", io.NopCloser(bytes.NewReader(body)))
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)

		require.Contains(t, yamlContent, `
    SimbaOpenapiTestReqBody:
      properties:
        size:
          maximum: 5
          type: integer
      type: object`)
	})
}

func TestOpenAPIDocsGenBasicAuthHandler(t *testing.T) {
	t.Parallel()

	app := simbaTest.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(simbaTest.BasicAuthHandler, simbaTest.BasicAuthAuthenticationHandler))

	app.RunTest(func() {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)
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
	})
}

func TestMultipleAuthHandlers(t *testing.T) {
	t.Parallel()

	app := simbaTest.Default()
	app.Router.POST("/test1", simba.AuthJsonHandler(simbaTest.BasicAuthHandler, simbaTest.BasicAuthAuthenticationHandler))
	app.Router.POST("/test2", simba.AuthJsonHandler(simbaTest.BasicAuthHandler, simbaTest.BasicAuthAuthenticationHandler))

	app.RunTest(func() {
		w1 := httptest.NewRecorder()
		w2 := httptest.NewRecorder()

		req1 := httptest.NewRequest(http.MethodPost, "/test1", nil)
		req2 := httptest.NewRequest(http.MethodPost, "/test2", nil)

		app.Router.ServeHTTP(w1, req1)
		app.Router.ServeHTTP(w2, req2)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)
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
	})
}

func TestOpenAPIDocsGenAPIKeyAuthHandler(t *testing.T) {
	t.Parallel()

	app := simbaTest.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(simbaTest.ApiKeyAuthHandler, simbaTest.ApiKeyAuthAuthenticationHandler))

	app.RunTest(func() {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("Authorization", "APIKey token")
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)
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
	})
}

func TestOpenAPIDocsGenBearerTokenAuthHandler(t *testing.T) {
	t.Parallel()

	app := simbaTest.Default()
	app.Router.POST("/test", simba.AuthJsonHandler(simbaTest.BearerTokenAuthHandler, simbaTest.BearerAuthAuthenticationHandler))

	app.RunTest(func() {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)
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
	})
}

func TestOpenAPIGenNoTags(t *testing.T) {
	t.Parallel()

	app := simbaTest.Default()
	app.Router.POST("/test", simba.JsonHandler(simbaTest.NoTagsHandler))

	app.RunTest(func() {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)
		require.Contains(t, yamlContent, "/test")
		require.Contains(t, yamlContent, "description: A dummy function to test the OpenAPI generation without any tags")
		require.Contains(t, yamlContent, "operationId: no-tags-handler")
		require.Contains(t, yamlContent, "summary: No tags handler")
		require.Contains(t, yamlContent, "tags:")
		require.Contains(t, yamlContent, "- SimbaTest")
		require.Contains(t, yamlContent, "\"202\":")
	})
}

func TestOpenAPIGenNoTagsReceiverFuncHandler(t *testing.T) {
	t.Parallel()

	receiver := simbaTest.Receiver{}
	app := simbaTest.Default()
	app.Router.POST("/test", simba.JsonHandler(receiver.NoTagsHandler))

	app.RunTest(func() {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		yamlContent := fetchOpenAPIDocumentation(t, app.Application)
		require.Contains(t, yamlContent, "/test")
		require.Contains(t, yamlContent, "description: A dummy function to test the OpenAPI generation without any tags")
		require.Contains(t, yamlContent, "operationId: no-tags-handler")
		require.Contains(t, yamlContent, "summary: No tags handler")
		require.Contains(t, yamlContent, "tags:")
		require.Contains(t, yamlContent, "- SimbaTest")
		require.Contains(t, yamlContent, "\"202\":")
	})
}

func fetchOpenAPIDocumentation(t *testing.T, app *simba.Application) string {
	t.Helper()

	getReq := httptest.NewRequest(http.MethodGet, "/openapi.yml", nil)
	getReq.Header.Set("Accept", "application/yaml")
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)
	require.Equal(t, "application/yaml", getW.Header().Get("Content-Type"))

	return getW.Body.String()
}
