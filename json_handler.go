package simba

import (
	"context"
	"net/http"

	"github.com/sillen102/simba/auth"
	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/models"
	"github.com/sillen102/simba/simbaErrors"
)

// JsonHandlerFunc is a function type for handling routes with Request body and params
type JsonHandlerFunc[RequestBody, Params, ResponseBody any] func(ctx context.Context, req *models.Request[RequestBody, Params]) (*models.Response[ResponseBody], error)

// AuthenticatedJsonHandlerFunc is a function type for handling authenticated routes with Request body and params
type AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody any] struct {
	handler     func(ctx context.Context, req *models.Request[RequestBody, Params], authModel AuthModel) (*models.Response[ResponseBody], error)
	authHandler auth.Handler[AuthModel]
}

// JsonHandler handles a Request with the Request body and params.
//
//	Example usage:
//
// Define a Request body struct:
//
//	type RequestBody struct {
//		Test string `json:"test" validate:"required"`
//	}
//
// Define a Request params struct:
//
//	type Params struct {
//		Name   string `header:"name" validate:"required"`
//		ID     int    `path:"id" validate:"required"`
//		Active bool   `query:"active" validate:"required"`
//		Page   int64  `query:"page" validate:"min=0"`
//		Size   int64  `query:"size" validate:"min=0"`
//	}
//
// Define a handler function:
//
//	func(ctx context.Context, req *simba.Request[RequestBody, Params]) (*simba.Response[map[string]string], error) {
//		// Access the Request body and params fields
//		req.Body.Test
//		req.Params.Name
//		req.Params.ID
//		req.Params.Page
//		req.Params.Size
//
//		// Return a response
//		return &simba.Response[map[string]string]{
//			Headers: map[string][]string{"My-Header": {"header-value"}},
//			Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
//			Body:    map[string]string{"message": "success"},
//			Status:  http.StatusOK,
//		}, nil
//	}
//
// Register the handler:
//
//	Mux.POST("/test/{id}", simba.JsonHandler(handler))
func JsonHandler[RequestBody, Params, ResponseBody any](h JsonHandlerFunc[RequestBody, Params, ResponseBody]) Handler {
	return h
}

// ServeHTTP implements the http.Handler interface for JsonHandlerFunc
func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := handleJsonRequest[RequestBody, Params](r)
	if err != nil {
		simbaErrors.WriteError(w, r, err)
		return
	}

	resp, err := h(ctx, req)
	if err != nil {
		simbaErrors.WriteError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetRequestBody() any {
	var rb RequestBody
	return rb
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetResponseBody() any {
	var resb ResponseBody
	return resb
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetParams() any {
	var p Params
	return p
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetAccepts() string {
	return mimetypes.ApplicationJSON
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetProduces() string {
	return mimetypes.ApplicationJSON
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetHandler() any {
	return h
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetAuthModel() any {
	return nil
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) GetAuthHandler() any {
	return nil
}

// AuthJsonHandler handles a Request with the Request body and params.
//
//	Example usage:
//
// Define a Request body struct:
//
//	type RequestBody struct {
//		Test string `json:"test" validate:"required"`
//	}
//
// Define a Request params struct:
//
//	type Params struct {
//		Name   string `header:"name" validate:"required"`
//		ID     int    `path:"id" validate:"required"`
//		Active bool   `query:"active" validate:"required"`
//		Page   int64  `query:"page" validate:"min=0"`
//		Size   int64  `query:"size" validate:"min=0"`
//	}
//
// Define a user struct:
//
//	type AuthModel struct {
//		ID   int
//		Name string
//		Role string
//	}
//
// Define a handler function:
//
//	func(ctx context.Context, req *simba.Request[RequestBody, Params], authModel AuthModel) (*simba.Response[map[string]string], error) {
//		// Access the Request body and params fields
//		req.Body.Test
//		req.Params.Name
//		req.Params.ID
//		req.Params.Page
//		req.Params.Size
//
//		// Access the user fields
//		user.ID
//		user.Name
//		user.Role
//
//		// Return a response
//		return &simba.Response[map[string]string]{
//			Headers: map[string][]string{"My-Header": {"header-value"}},
//			Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
//			Body:    map[string]string{"message": "success"},
//			Status:  http.StatusOK,
//		}, nil
//	}
//
// Register the handler:
//
//	Mux.POST("/test/{id}", simba.AuthJsonHandler(handler))
func AuthJsonHandler[RequestBody, Params, AuthModel, ResponseBody any](
	handler func(ctx context.Context, req *models.Request[RequestBody, Params], authModel AuthModel) (*models.Response[ResponseBody], error),
	authHandler auth.Handler[AuthModel],
) Handler {
	return AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]{
		handler:     handler,
		authHandler: authHandler,
	}
}

// ServeHTTP implements the http.Handler interface for AuthenticatedJsonHandlerFunc
func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authModel, err := auth.HandleAuthRequest[AuthModel](h.authHandler, r)
	if err != nil {
		statusCode := http.StatusUnauthorized // Default status code for unauthorized access
		if statusCoder, ok := err.(simbaErrors.StatusCodeProvider); ok {
			statusCode = statusCoder.StatusCode()
		}

		errorMessage := "unauthorized" // Default error message for unauthorized access
		if msgProvider, ok := err.(simbaErrors.PublicMessageProvider); ok {
			errorMessage = msgProvider.PublicMessage()
		}

		simbaErrors.WriteError(w, r, simbaErrors.NewSimbaError(statusCode, errorMessage, err))
		return
	}

	req, err := handleJsonRequest[RequestBody, Params](r)
	if err != nil {
		simbaErrors.WriteError(w, r, err)
		return
	}

	resp, err := h.handler(ctx, req, authModel)
	if err != nil {
		simbaErrors.WriteError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetRequestBody() any {
	var rb RequestBody
	return rb
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetParams() any {
	var p Params
	return p
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetResponseBody() any {
	var resb ResponseBody
	return resb
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetAccepts() string {
	return mimetypes.ApplicationJSON
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetProduces() string {
	return mimetypes.ApplicationJSON
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetHandler() any {
	return h.handler
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetAuthModel() any {
	var am AuthModel
	return am
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) GetAuthHandler() any {
	return h.authHandler
}

// handleJsonRequest handles extracting body and params from the Request
func handleJsonRequest[RequestBody any, Params any](r *http.Request) (*models.Request[RequestBody, Params], error) {
	params, err := ParseAndValidateParams[Params](r)
	if err != nil {
		return nil, err
	}

	var reqBody RequestBody
	err = handleJsonBody(r, &reqBody)
	if err != nil {
		return nil, err
	}

	return &models.Request[RequestBody, Params]{
		Body:   reqBody,
		Params: params,
	}, nil
}
