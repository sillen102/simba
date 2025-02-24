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
	Description string `json:"description" description:"Description of the user" example:"A test user"`
}

type respBody struct {
	ID          string `json:"id" description:"ID of the user" example:"1"`
	Name        string `json:"name" description:"Name of the user" example:"John Doe"`
	Age         int    `json:"age" description:"Age of the user" example:"30"`
	Description string `json:"description" description:"Description of the user" example:"A test user"`
}

// handler is a test handler for the POST /test/{id} route
// It returns a response with the request body and params
// It also sets a custom header and cookie
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

func TestOpenAPI(t *testing.T) {
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
	app.GenerateDocs()
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
}
