package simba

import (
	"context"
	"io"
	"net/http"

	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// RawBodyHandlerFunc is a function type for handling routes with Request body and params
type RawBodyHandlerFunc[Params, ResponseBody any] func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, Params]) (*simbaModels.Response[ResponseBody], error)

// AuthenticatedRawBodyHandlerFunc is a function type for handling authenticated routes with Request body and params
type AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody any] struct {
	handler     func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, Params], authModel AuthModel) (*simbaModels.Response[ResponseBody], error)
	authHandler AuthHandler[AuthModel]
}

// RawBodyHandler handles a Request with the Request body and params.
//
// Register the handler:
//
//	Mux.POST("/test/{id}", simba.RawBodyHandler(handler))
func RawBodyHandler[Params, ResponseBody any](h RawBodyHandlerFunc[Params, ResponseBody]) Handler {
	return h
}

// ServeHTTP implements the http.Handler interface for RawBodyHandlerFunc
func (h RawBodyHandlerFunc[Params, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := handleRawRequest[Params](r)
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

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetRequestBody() any {
	var rb []byte
	return rb
}

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetResponseBody() any {
	var resb ResponseBody
	return resb
}

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetParams() any {
	var p Params
	return p
}

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetAccepts() string {
	return mimetypes.ApplicationJSON
}

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetProduces() string {
	return mimetypes.ApplicationJSON
}

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetHandler() any {
	return h
}

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetAuthModel() any {
	return nil
}

func (h RawBodyHandlerFunc[Params, ResponseBody]) GetAuthHandler() any {
	return nil
}

// AuthRawBodyHandler handles a Request with the Request body and params.
//
// Register the handler:
//
//	Mux.POST("/test/{id}", simba.AuthRawBodyHandler(handler))
func AuthRawBodyHandler[Params, AuthModel, ResponseBody any](
	handler func(ctx context.Context, req *simbaModels.Request[io.ReadCloser, Params], authModel AuthModel) (*simbaModels.Response[ResponseBody], error),
	authHandler AuthHandler[AuthModel],
) Handler {
	return AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]{
		handler:     handler,
		authHandler: authHandler,
	}
}

// ServeHTTP implements the http.Handler interface for AuthenticatedRawBodyHandlerFunc
func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authModel, err := HandleAuthRequest[AuthModel](h.authHandler, r)
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

	req, err := handleRawRequest[Params](r)
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

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetRequestBody() any {
	var rb []byte
	return rb
}

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetParams() any {
	var p Params
	return p
}

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetResponseBody() any {
	var resb ResponseBody
	return resb
}

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetAccepts() string {
	return mimetypes.ApplicationJSON
}

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetProduces() string {
	return mimetypes.ApplicationJSON
}

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetHandler() any {
	return h.handler
}

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetAuthModel() any {
	var am AuthModel
	return am
}

func (h AuthenticatedRawBodyHandlerFunc[Params, AuthModel, ResponseBody]) GetAuthHandler() any {
	return h.authHandler
}

// handleRequest handles extracting body and params from the Request
func handleRawRequest[Params any](r *http.Request) (*simbaModels.Request[io.ReadCloser, Params], error) {
	params, err := ParseAndValidateParams[Params](r)
	if err != nil {
		return nil, err
	}

	return &simbaModels.Request[io.ReadCloser, Params]{
		Body:   r.Body,
		Params: params,
	}, nil
}
