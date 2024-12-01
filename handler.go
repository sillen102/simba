package simba

import (
	"context"
	"net/http"
)

func HandlerFunc[RequestBody any, Params any](h SimpleHandler[RequestBody, Params]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// SimpleHandler handles a request with the request body and params.
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
//		Page   int64  `query:"page" validate:"required"`
//		Size   int64  `query:"size" validate:"required"`
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
type SimpleHandler[RequestBody any, Params any] func(ctx context.Context, req *Request[RequestBody, Params]) (*Response, error)

func (h SimpleHandler[RequestBody, Params]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody RequestBody
	err := decodeBodyIfNeeded(r, &reqBody)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	params, err := parseAndValidateParams[Params](r)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	req := &Request[RequestBody, Params]{
		Cookies: r.Cookies(),
		Body:    reqBody,
		Params:  *params,
	}

	resp, err := h(ctx, req)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}
