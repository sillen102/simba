package simba

import (
	"context"
	"log/slog"
	"net/http"
	"reflect"
)

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

// JsonHandlerFunc is a function type for handling routes with Request body and params
type JsonHandlerFunc[RequestBody, Params, ResponseBody any] func(ctx context.Context, req *Request[RequestBody, Params]) (*Response[ResponseBody], error)

// ServeHTTP implements the http.Handler interface for JsonHandlerFunc
func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := handleRequest[RequestBody, Params](r)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp, err := h(ctx, req)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}

func (h JsonHandlerFunc[RequestBody, Params, ResponseBody]) getTypes() (reflect.Type, reflect.Type, reflect.Type) {
	var rb RequestBody
	var p Params
	var resb ResponseBody

	bodyType := reflect.TypeOf(rb)
	paramsType := reflect.TypeOf(p)
	responseType := reflect.TypeOf(resb)

	slog.Debug("type information",
		"bodyType", bodyType,
		"paramsType", paramsType,
		"responseType", responseType,
	)

	return bodyType, paramsType, responseType
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
//	func(ctx context.Context, req *simba.Request[RequestBody, Params], authModel *AuthModel) (*simba.Response[map[string]string], error) {
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
	handler func(ctx context.Context, req *Request[RequestBody, Params], authModel *AuthModel) (*Response[ResponseBody], error),
	authFunc AuthFunc[AuthModel],
) Handler {
	return AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]{
		handler:  handler,
		authFunc: authFunc,
	}
}

// AuthenticatedJsonHandlerFunc is a function type for handling authenticated routes with Request body and params
type AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody any] struct {
	handler  func(ctx context.Context, req *Request[RequestBody, Params], authModel *AuthModel) (*Response[ResponseBody], error)
	authFunc AuthFunc[AuthModel]
}

// ServeHTTP implements the http.Handler interface for AuthenticatedJsonHandlerFunc
func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authModel, err := h.authFunc(r)
	if err != nil {
		WriteError(w, r, NewHttpError(http.StatusUnauthorized, "failed to authenticate", err))
		return
	}

	req, err := handleRequest[RequestBody, Params](r)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp, err := h.handler(ctx, req, authModel)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}

func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel, ResponseBody]) getTypes() (reflect.Type, reflect.Type, reflect.Type) {
	var rb RequestBody
	var p Params
	var resb ResponseBody

	bodyType := reflect.TypeOf(rb)
	paramsType := reflect.TypeOf(p)
	responseType := reflect.TypeOf(resb)

	slog.Debug("type information",
		"bodyType", bodyType,
		"paramsType", paramsType,
		"responseType", responseType,
	)

	return bodyType, paramsType, responseType
}

// handleRequest handles extracting body and params from the Request
func handleRequest[RequestBody any, Params any](r *http.Request) (*Request[RequestBody, Params], error) {
	params, err := parseAndValidateParams[Params](r)
	if err != nil {
		return nil, err
	}

	var reqBody RequestBody
	err = handleJsonBody(r, &reqBody)
	if err != nil {
		return nil, err
	}

	return &Request[RequestBody, Params]{
		Cookies: r.Cookies(),
		Body:    reqBody,
		Params:  params,
	}, nil
}
