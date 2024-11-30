package handlers

import (
	"context"
	"net/http"

	"github.com/sillen102/simba"
)

// BaseHandler handles a request with only the request body.
//
// Example usage:
//
//	Define a handler function that returns a BaseHandler:
//
//	func MyHandler() handlers.BaseHandler[RequestBody] {
//		return func(ctx context.Context, req *simba.Request[RequestBody]) (*simba.Response, error) {
//
//			return &simba.Response{
//				Body:   map[string]string{"message": "success"},
//				Status: http.StatusOK,
//			}, nil
//		}
//	}
//
//	Register the handler:
//
//	router.POST("/test", simba.Handle[RequestBody, simba.NoParams](MyHandler()))
type BaseHandler[RequestBody any] func(ctx context.Context, req *simba.Request[RequestBody]) (*simba.Response, error)

func (h BaseHandler[RequestBody]) Handle(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var reqBody RequestBody
	// Perform body decoding

	// Create request context
	ctx := r.Context()
	req := &simba.Request[RequestBody]{
		Body: reqBody,
		// other request details
	}

	// Call the handler
	resp, err := h(ctx, req)

	// Write response
	writeResponse(w, r, resp, err)
}
