package simba

import (
	"context"
	"net/http"
)

// HandlerFunc returns an [http.Handler] that can be used for non-authenticated routes
func HandlerFunc[RequestBody any, Params any](h Handler[RequestBody, Params]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// Handler handles a request with the request body and params.
//
//	Example usage:
//
// Define a request body struct:
//
//	type RequestBody struct {
//		Test string `json:"test" validate:"required"`
//	}
//
// Define a request params struct:
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
//		// Access the request body and params fields
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
//	router.POST("/test/:id", simba.HandlerFunc(handler))
type Handler[RequestBody any, Params any] func(ctx context.Context, req *Request[RequestBody, Params]) (*Response, error)

// ServeHTTP implements the http.Handler interface for Handler
func (h Handler[RequestBody, Params]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

// AuthenticatedHandlerFunc returns an [http.Handler] that can be used for authenticated routes
func AuthenticatedHandlerFunc[RequestBody any, Params any, AuthModel any](h AuthenticatedHandler[RequestBody, Params, AuthModel]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// AuthenticatedHandler handles a request with the request body and params.
//
//	Example usage:
//
// Define a request body struct:
//
//	type RequestBody struct {
//		Test string `json:"test" validate:"required"`
//	}
//
// Define a request params struct:
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
//	func(ctx context.Context, req *simba.Request[RequestBody, Params], user *AuthModel) (*simba.Response, error) {
//		// Access the request body and params fields
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
//	router.POST("/test/:id", simba.AuthenticatedHandlerFunc(handler))
type AuthenticatedHandler[RequestBody any, Params any, AuthModel any] func(ctx context.Context, req *Request[RequestBody, Params], user *AuthModel) (*Response, error)

// ServeHTTP implements the http.Handler interface for AuthenticatedHandler
func (h AuthenticatedHandler[RequestBody, Params, AuthModel]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authFunc := r.Context().Value(authFuncKey).(AuthFunc[AuthModel])
	if authFunc == nil {
		HandleError(w, r, NewHttpError(http.StatusUnauthorized, "auth function is not set", nil))
		return
	}

	user, err := authFunc(r)
	if err != nil {
		HandleError(w, r, NewHttpError(http.StatusUnauthorized, "failed to authenticate", err))
		return
	}

	req, err := handleRequest[RequestBody, Params](r)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	resp, err := h(ctx, req, user)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}

// handleRequest handles extracting body and params from the request
func handleRequest[RequestBody any, Params any](r *http.Request) (*Request[RequestBody, Params], error) {
	var reqBody RequestBody
	err := decodeBodyIfNeeded(r, &reqBody)
	if err != nil {
		return nil, err
	}

	if validationErrors := validateStruct(reqBody); len(validationErrors) > 0 {
		return nil, NewHttpError(http.StatusBadRequest, "invalid request body", nil, validationErrors...)
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
