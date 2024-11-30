package handlers

import (
	"context"
	"net/http"

	"github.com/sillen102/simba"
)

// ParamsHandler handles a request with the request body and params.
//
// Example usage:
//
//	Define a handler function that returns a ParamsHandler:
//
//	func MyHandler() handlers.ParamsHandler[RequestBody, Params] {
//		return func(ctx context.Context, req *simba.Request[RequestBody], params Params) (*simba.Response, error) {
//
//			return &simba.Response{
//				Body:   map[string]string{"message": "success"},
//				Status: http.StatusOK,
//			}, nil
//		}
//	}
//
//	Register the handler: router.GET("/test/{id}", simba.Handle[RequestBody, Params](MyHandler()))
type ParamsHandler[RequestBody any, Params any] func(ctx context.Context, req *simba.Request[RequestBody], params Params) (*simba.Response, error)

func (h ParamsHandler[RequestBody, Params]) Handle(w http.ResponseWriter, r *http.Request) {
	// Decode request body and params
	var reqBody RequestBody
	var params Params
	// Perform body and params decoding

	// Create request context
	ctx := r.Context()
	req := &simba.Request[RequestBody]{
		Body: reqBody,
		// other request details
	}

	// Call the handler
	resp, err := h(ctx, req, params)

	// Write response
	writeResponse(w, r, resp, err)
}
