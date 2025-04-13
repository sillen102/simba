package simbaOpenapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaOpenapi"
	"github.com/sillen102/simba/simbaOpenapi/openapiModels"
	"github.com/sillen102/simba/simbaTest"
	"github.com/sillen102/simba/simbaTest/assert"
	"github.com/swaggest/openapi-go/openapi31"
)

type openAPIDoc struct {
	Info       openapi31.Info       `json:"info"`
	Paths      openapi31.Paths      `json:"paths"`
	Components openapi31.Components `json:"components"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func TestTitle(t *testing.T) {
	t.Parallel()

	generator := simbaOpenapi.NewOpenAPIGenerator()
	routeInfo := []openapiModels.RouteInfo{
		{
			Method:   http.MethodPost,
			Path:     "/test/{id}",
			Accepts:  mimetypes.ApplicationJSON,
			Produces: mimetypes.ApplicationJSON,
			Handler:  simbaTest.NoTagsHandler,
			ReqBody:  simbaTest.RequestBody{},
			RespBody: simbaTest.ResponseBody{},
			Params:   simbaTest.Params{},
		},
	}

	schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", routeInfo)
	assert.NoError(t, err)
	doc := unmarshalJSON(t, schema)

	assert.Equal(t, "Test", doc.Info.Title)
}

func TestVersion(t *testing.T) {
	t.Parallel()

	generator := simbaOpenapi.NewOpenAPIGenerator()
	routeInfo := []openapiModels.RouteInfo{
		{
			Method:   http.MethodPost,
			Path:     "/test/{id}",
			Accepts:  mimetypes.ApplicationJSON,
			Produces: mimetypes.ApplicationJSON,
			Handler:  simbaTest.NoTagsHandler,
			ReqBody:  simbaTest.RequestBody{},
			RespBody: simbaTest.ResponseBody{},
			Params:   simbaTest.Params{},
		},
	}

	schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", routeInfo)
	assert.NoError(t, err)
	doc := unmarshalJSON(t, schema)

	assert.Equal(t, "1.0.0", doc.Info.Version)
}

func TestDescription(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()
	receiver := simbaTest.Receiver{}

	testCases := []struct {
		name      string
		routeInfo []openapiModels.RouteInfo
		expected  string
	}{
		{
			name: "handler with tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "this is a multiline\n\ndescription for the handler",
		},
		{
			name: "handler with no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "A dummy function to test the OpenAPI generation without any tags in the comment.",
		},
		{
			name: "handler with receiver and tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "this is a multiline\n\ndescription for the handler",
		},
		{
			name: "handler with receiver and no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "A dummy function to test the OpenAPI generation without any tags in the comment.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			assert.Equal(t, tc.expected, *doc.Paths.MapOfPathItemValues[path].Post.Description)
		})
	}
}

func TestResponseCode(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()
	receiver := simbaTest.Receiver{}

	testCases := []struct {
		name      string
		routeInfo []openapiModels.RouteInfo
		expected  int
	}{
		{
			name: "handler with tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: http.StatusCreated,
		},
		{
			name: "handler with no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: http.StatusCreated,
		},
		{
			name: "handler with receiver and tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: http.StatusCreated,
		},
		{
			name: "handler with receiver and no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			assert.NotNil(
				t,
				*doc.Paths.MapOfPathItemValues[path].Post.Responses.MapOfResponseOrReferenceValues[strconv.Itoa(tc.expected)].Response,
				fmt.Sprintf("response code %d not found", tc.expected),
			)
		})
	}
}

func TestTags(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()
	receiver := simbaTest.Receiver{}

	testCases := []struct {
		name          string
		routeInfo     []openapiModels.RouteInfo
		expected      []string
		expectedError error
	}{
		{
			name: "handler with tag tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected:      []string{"Test", "User"},
			expectedError: nil,
		},
		{
			name: "handler with no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: []string{"SimbaTest"},
		},
		{
			name: "handler with receiver and tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: []string{"Test", "User"},
		},
		{
			name: "handler with receiver and no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: []string{"SimbaTest"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			assert.ContainsInAnyOrder(t, tc.expected, doc.Paths.MapOfPathItemValues[path].Post.Tags)
		})
	}
}

func TestOperationID(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()
	receiver := simbaTest.Receiver{}

	testCases := []struct {
		name      string
		routeInfo []openapiModels.RouteInfo
		expected  string
	}{
		{
			name: "handler with tag tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "testHandler",
		},
		{
			name: "handler with no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "no-tags-handler",
		},
		{
			name: "handler with receiver and tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "testHandler",
		},
		{
			name: "handler with receiver and no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "no-tags-handler",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			assert.Equal(t, tc.expected, *doc.Paths.MapOfPathItemValues[path].Post.ID)
		})
	}
}

func TestSummary(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()
	receiver := simbaTest.Receiver{}

	testCases := []struct {
		name      string
		routeInfo []openapiModels.RouteInfo
		expected  string
	}{
		{
			name: "handler with tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "test handler",
		},
		{
			name: "handler with no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "No tags handler",
		},
		{
			name: "handler with receiver and tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "test handler",
		},
		{
			name: "handler with receiver and no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: "No tags handler",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			assert.Equal(t, tc.expected, *doc.Paths.MapOfPathItemValues[path].Post.Summary)
		})
	}
}

func TestCustomError(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()
	receiver := simbaTest.Receiver{}

	testCases := []struct {
		name      string
		routeInfo []openapiModels.RouteInfo
		expected  int
	}{
		{
			name: "handler with tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: http.StatusConflict,
		},
		{
			name: "handler with no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: 0,
		},
		{
			name: "handler with receiver and tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: http.StatusConflict,
		},
		{
			name: "handler with receiver and no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			if tc.expected != 0 {
				assert.NotNil(
					t,
					*doc.Paths.MapOfPathItemValues[path].Post.Responses.MapOfResponseOrReferenceValues[strconv.Itoa(tc.expected)].Response,
					fmt.Sprintf("response code %d not found", tc.expected),
				)
			}
		})
	}
}

func TestDeprecated(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()
	receiver := simbaTest.Receiver{}

	testCases := []struct {
		name      string
		routeInfo []openapiModels.RouteInfo
		expected  bool
	}{
		{
			name: "handler without deprecated tag",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: false,
		},
		{
			name: "handler with deprecated tag",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.DeprecatedHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: true,
		},
		{
			name: "handler with no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  simbaTest.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: false,
		},
		{
			name: "handler with receiver, tags but no deprecated tag",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.TagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: false,
		},
		{
			name: "handler with receiver, tags and deprecated tag",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.DeprecatedHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: true,
		},
		{
			name: "handler with receiver and no tags",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:   http.MethodPost,
					Path:     path,
					Accepts:  mimetypes.ApplicationJSON,
					Produces: mimetypes.ApplicationJSON,
					Handler:  receiver.NoTagsHandler,
					ReqBody:  simbaTest.RequestBody{},
					RespBody: simbaTest.ResponseBody{},
					Params:   simbaTest.Params{},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			assert.Equal(t, tc.expected, *doc.Paths.MapOfPathItemValues[path].Post.Deprecated)
		})
	}
}

func TestValidateRequiredField(t *testing.T) {
	t.Parallel()

	type reqBody struct {
		Name string `json:"name" validate:"required"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[reqBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	generator := simbaOpenapi.NewOpenAPIGenerator()
	routeInfo := []openapiModels.RouteInfo{
		{
			Method:   http.MethodPost,
			Path:     "/test",
			Accepts:  mimetypes.ApplicationJSON,
			Produces: mimetypes.ApplicationJSON,
			Handler:  handler,
			ReqBody:  reqBody{},
			RespBody: simbaModels.NoParams{},
			Params:   simbaModels.NoBody{},
		},
	}

	schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", routeInfo)
	assert.NoError(t, err)
	doc := unmarshalJSON(t, schema)

	assert.Contains(t, []string{"name"}, doc.Components.Schemas["SimbaOpenapiTestReqBody"]["required"])
}

func TestValidateMinField(t *testing.T) {
	t.Parallel()

	type reqBody struct {
		Size   int      `json:"size" validate:"min=5"`
		Length string   `json:"length" validate:"min=5"`
		Items  []string `json:"items" validate:"min=5"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[reqBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	generator := simbaOpenapi.NewOpenAPIGenerator()
	routeInfo := []openapiModels.RouteInfo{
		{
			Method:   http.MethodPost,
			Path:     "/test",
			Accepts:  mimetypes.ApplicationJSON,
			Produces: mimetypes.ApplicationJSON,
			Handler:  handler,
			ReqBody:  reqBody{},
			RespBody: simbaModels.NoParams{},
			Params:   simbaModels.NoBody{},
		},
	}

	schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", routeInfo)
	assert.NoError(t, err)
	doc := unmarshalJSON(t, schema)

	assert.Equal(
		t,
		5.0,
		doc.Components.Schemas["SimbaOpenapiTestReqBody"]["properties"].(map[string]interface{})["size"].(map[string]interface{})["minimum"],
	)
	assert.Equal(
		t,
		5.0,
		doc.Components.Schemas["SimbaOpenapiTestReqBody"]["properties"].(map[string]interface{})["length"].(map[string]interface{})["minLength"],
	)
	assert.Equal(
		t,
		5.0,
		doc.Components.Schemas["SimbaOpenapiTestReqBody"]["properties"].(map[string]interface{})["items"].(map[string]interface{})["minItems"],
	)

	valid := reqBody{Size: 5, Length: "12345", Items: []string{"1", "2", "3", "4", "5"}}
	err = validate.Struct(valid)
	assert.NoError(t, err)

	invalid := reqBody{Size: 4, Length: "1234", Items: []string{"1", "2", "3", "4"}}
	err = validate.Struct(invalid)
	assert.Error(t, err)
}

func TestValidateMaxField(t *testing.T) {
	t.Parallel()

	type reqBody struct {
		Size   int      `json:"size" validate:"max=5"`
		Length string   `json:"length" validate:"max=5"`
		Items  []string `json:"items" validate:"max=5"`
	}

	handler := func(ctx context.Context, req *simbaModels.Request[reqBody, simbaModels.NoParams]) (*simbaModels.Response[simbaModels.NoBody], error) {
		return &simbaModels.Response[simbaModels.NoBody]{}, nil
	}

	generator := simbaOpenapi.NewOpenAPIGenerator()
	routeInfo := []openapiModels.RouteInfo{
		{
			Method:   http.MethodPost,
			Path:     "/test",
			Accepts:  mimetypes.ApplicationJSON,
			Produces: mimetypes.ApplicationJSON,
			Handler:  handler,
			ReqBody:  reqBody{},
			RespBody: simbaModels.NoParams{},
			Params:   simbaModels.NoBody{},
		},
	}

	schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", routeInfo)
	assert.NoError(t, err)
	doc := unmarshalJSON(t, schema)

	assert.Equal(
		t,
		5.0,
		doc.Components.Schemas["SimbaOpenapiTestReqBody"]["properties"].(map[string]interface{})["size"].(map[string]interface{})["maximum"],
	)
	assert.Equal(
		t,
		5.0,
		doc.Components.Schemas["SimbaOpenapiTestReqBody"]["properties"].(map[string]interface{})["length"].(map[string]interface{})["maxLength"],
	)
	assert.Equal(
		t,
		5.0,
		doc.Components.Schemas["SimbaOpenapiTestReqBody"]["properties"].(map[string]interface{})["items"].(map[string]interface{})["maxItems"],
	)

	valid := reqBody{Size: 5, Length: "12345", Items: []string{"1", "2", "3", "4", "5"}}
	err = validate.Struct(valid)
	assert.NoError(t, err)

	invalid := reqBody{Size: 6, Length: "123456", Items: []string{"1", "2", "3", "4", "5", "6"}}
	err = validate.Struct(invalid)
	assert.Error(t, err)
}

func TestAuthHandler(t *testing.T) {
	t.Parallel()

	path := "/test/{id}"
	generator := simbaOpenapi.NewOpenAPIGenerator()

	testCases := []struct {
		name                string
		schemeName          string
		routeInfo           []openapiModels.RouteInfo
		expectedDescription string
	}{
		{
			name:       "basic auth handler",
			schemeName: "admin",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:      http.MethodPost,
					Path:        path,
					Accepts:     mimetypes.ApplicationJSON,
					Produces:    mimetypes.ApplicationJSON,
					Handler:     simbaTest.BasicAuthHandler,
					ReqBody:     simbaTest.RequestBody{},
					RespBody:    simbaTest.ResponseBody{},
					Params:      simbaTest.Params{},
					AuthHandler: simbaTest.BasicAuthAuthenticationHandler,
					AuthModel:   simbaTest.User{},
				},
			},
			expectedDescription: "admin access only",
		},
		{
			name:       "api key auth handler",
			schemeName: "User",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:      http.MethodPost,
					Path:        path,
					Accepts:     mimetypes.ApplicationJSON,
					Produces:    mimetypes.ApplicationJSON,
					Handler:     simbaTest.ApiKeyAuthHandler,
					ReqBody:     simbaTest.RequestBody{},
					RespBody:    simbaTest.ResponseBody{},
					Params:      simbaTest.Params{},
					AuthHandler: simbaTest.ApiKeyAuthAuthenticationHandler,
					AuthModel:   simbaTest.User{},
				},
			},
			expectedDescription: "Session cookie",
		},
		{
			name:       "bearer token auth handler",
			schemeName: "admin",
			routeInfo: []openapiModels.RouteInfo{
				{
					Method:      http.MethodPost,
					Path:        path,
					Accepts:     mimetypes.ApplicationJSON,
					Produces:    mimetypes.ApplicationJSON,
					Handler:     simbaTest.BearerTokenAuthHandler,
					ReqBody:     simbaTest.RequestBody{},
					RespBody:    simbaTest.ResponseBody{},
					Params:      simbaTest.Params{},
					AuthHandler: simbaTest.BearerAuthAuthenticationHandler,
					AuthModel:   simbaTest.User{},
				},
			},
			expectedDescription: "Bearer token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := generator.GenerateDocumentation(context.Background(), "Test", "1.0.0", tc.routeInfo)
			assert.NoError(t, err)
			doc := unmarshalJSON(t, schema)

			securityScheme := doc.Components.SecuritySchemes[tc.schemeName].SecurityScheme

			switch {
			case securityScheme.HTTP != nil:
				assert.Equal(t, "basic", securityScheme.HTTP.Scheme)
			case securityScheme.APIKey != nil:
				assert.Equal(t, "sessionid", securityScheme.APIKey.Name)
				assert.Equal(t, openapi31.SecuritySchemeAPIKeyInCookie, securityScheme.APIKey.In)
			case securityScheme.HTTPBearer != nil:
				assert.Equal(t, "bearer", securityScheme.HTTPBearer.Scheme)
				assert.Equal(t, "jwt", *securityScheme.HTTPBearer.BearerFormat)
			}

			assert.Equal(t, tc.expectedDescription, *securityScheme.Description)
		})
	}
}

func unmarshalJSON(t *testing.T, schema []byte) openAPIDoc {
	t.Helper()

	var jsonDoc openAPIDoc
	err := json.Unmarshal(schema, &jsonDoc)
	assert.NoError(t, err)

	return jsonDoc
}
