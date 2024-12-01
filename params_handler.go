package simba

import (
	"context"
	"net/http"

	"github.com/uptrace/bunrouter"
)

// ParamsHandler handles a request with the request body and params.
//
// Example usage:
//
//	Define a handler function that returns a ParamsHandler:
//
//	func MyParamsHandler() handlers.ParamsHandler[RequestBody, Params] {
//		return handlers.ParamsHandler[RequestBody, Params](func(ctx context.Context, req *Request[RequestBody], params Params) (*Response, error) {
//			return &Response{
//				Body:   map[string]string{"message": "success"},
//				Status: http.StatusOK,
//			}, nil
//		})
//	}
//
// Register the handler:
//
//	router.GET("/test/:id", MyHandler())
type ParamsHandler[RequestBody any, Params any] func(ctx context.Context, req *Request[RequestBody], params Params) (*Response, error)

func (h ParamsHandler[RequestBody, Params]) ServeHTTP(w http.ResponseWriter, r bunrouter.Request) error {
	// Decode request body
	var reqBody RequestBody
	err := decodeBodyIfNeeded(r, &reqBody)
	if err != nil {
		return err
	}

	var params Params
	// Extract params from route

	ctx := r.Context()
	req := &Request[RequestBody]{
		Headers: r.Header,
		Cookies: r.Cookies(),
		Body:    reqBody,
	}

	resp, err := h(ctx, req, params)
	if err != nil {
		return err
	}

	writeResponse(w, r, resp, nil)
	return nil
}
