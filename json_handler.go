package simba

import (
	"context"
	"net/http"

	"github.com/sillen102/simba/simbaContext"
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
//	func(ctx context.Context, req *simba.Request[RequestBody, Params]) (*simba.Response, error) {
//		// Access the Request body and params fields
//		req.Body.Test
//		req.Params.Name
//		req.Params.ID
//		req.Params.Page
//		req.Params.Size
//
//		// Return a response
//		return &simba.Response{
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
func JsonHandler[RequestBody any, Params any](h JsonHandlerFunc[RequestBody, Params]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// JsonHandlerFunc is a function type for handling routes with Request body and params
type JsonHandlerFunc[RequestBody any, Params any] func(ctx context.Context, req *Request[RequestBody, Params]) (*Response, error)

// ServeHTTP implements the http.Handler interface for JsonHandlerFunc
func (h JsonHandlerFunc[RequestBody, Params]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := handleRequest[RequestBody, Params](r)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	resp, err := h(ctx, req)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
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
//	func(ctx context.Context, req *simba.Request[RequestBody, Params], authModel *AuthModel) (*simba.Response, error) {
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
//		return &simba.Response{
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
func AuthJsonHandler[RequestBody any, Params any, AuthModel any](h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// AuthenticatedJsonHandlerFunc is a function type for handling authenticated routes with Request body and params
type AuthenticatedJsonHandlerFunc[RequestBody any, Params any, AuthModel any] func(ctx context.Context, req *Request[RequestBody, Params], authModel *AuthModel) (*Response, error)

// ServeHTTP implements the http.Handler interface for AuthenticatedJsonHandlerFunc
func (h AuthenticatedJsonHandlerFunc[RequestBody, Params, AuthModel]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authFunc := r.Context().Value(simbaContext.AuthFuncKey).(AuthFunc[AuthModel])
	if authFunc == nil {
		HandleError(w, r, NewHttpError(http.StatusUnauthorized, "auth function is not set", nil))
		return
	}

	authModel, err := authFunc(r)
	if err != nil {
		HandleError(w, r, NewHttpError(http.StatusUnauthorized, "failed to authenticate", err))
		return
	}

	req, err := handleRequest[RequestBody, Params](r)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	resp, err := h(ctx, req, authModel)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}

// handleRequest handles extracting body and params from the Request
func handleRequest[RequestBody any, Params any](r *http.Request) (*Request[RequestBody, Params], error) {
	var reqBody RequestBody
	err := handleJsonBody(r, &reqBody)
	if err != nil {
		return nil, err
	}

	params, err := parseAndValidateParams[Params](r)
	if err != nil {
		return nil, err
	}

	return &Request[RequestBody, Params]{
		Cookies: r.Cookies(),
		Body:    reqBody,
		Params:  params,
	}, nil
}
