package simba

import (
	"context"
	"net/http"

	"github.com/uptrace/bunrouter"
)

// BaseHandler handles a request with only the request body.
//
// Example usage:
//
//	Define a handler function that returns a BaseHandler:
//
//	func MyParamsHandler() handlers.BaseHandler[RequestBody] {
//		return handlers.BaseHandler[RequestBody](func(ctx context.Context, req *Request[RequestBody]) (*Response, error) {
//			return &Response{
//				Body:   map[string]string{"message": "success"},
//				Status: http.StatusOK,
//			}, nil
//		})
//	}
//
// Register the handler:
//
//	router.POST("/test", MyHandler())
type BaseHandler[RequestBody any] func(ctx context.Context, req *Request[RequestBody]) (*Response, error)

func (h BaseHandler[RequestBody]) ServeHTTP(w http.ResponseWriter, r bunrouter.Request) error {
	// Decode request body
	var reqBody RequestBody
	err := decodeBodyIfNeeded(r, &reqBody)
	if err != nil {
		return err
	}

	// Create request context
	ctx := r.Context()
	req := &Request[RequestBody]{
		Headers: r.Header,
		Cookies: r.Cookies(),
		Body:    reqBody,
	}

	// Call the handler
	resp, err := h(ctx, req)
	if err != nil {
		return err
	}

	// Write response
	writeResponse(w, r, resp, nil)
	return nil
}
