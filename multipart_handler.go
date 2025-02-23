package simba

import (
	"context"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"

	"github.com/sillen102/simba/simbaContext"
)

// MultipartHandler handles a MultipartRequest with params.
// // The MultipartRequest holds a MultipartReader and the parsed params.
// // The reason to provide the reader is to allow the logic for processing the parts to be handled by the handler function.
//
//	Example usage:
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
//		// Access the Multipart reader and params fields
//		req.Params.Name
//		req.Params.ID
//		req.Params.Page
//		req.Params.Size
//		req.Reader // Multipart reader
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
//	Mux.POST("/test/{id}", simba.MultipartHandler(handler))
func MultipartHandler[Params any, ResponseBody any](h MultipartHandlerFunc[Params, ResponseBody]) Handler {
	return h
}

// MultipartHandlerFunc is a function type for handling routes with Request body and params
type MultipartHandlerFunc[Params any, ResponseBody any] func(ctx context.Context, req *MultipartRequest[Params]) (*Response[ResponseBody], error)

// ServeHTTP implements the http.Handler interface for JsonHandlerFunc
func (h MultipartHandlerFunc[Params, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := handleMultipartRequest[Params](r)
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

func (h MultipartHandlerFunc[Params, ResponseBody]) getTypes() (reflect.Type, reflect.Type, reflect.Type) {
	var file multipart.File
	var p Params
	var resb ResponseBody

	bodyType := reflect.TypeOf(&file).Elem()
	paramsType := reflect.TypeOf(p)
	responseType := reflect.TypeOf(resb)

	slog.Debug("type information",
		"bodyType", bodyType,
		"paramsType", paramsType,
		"responseType", responseType,
	)

	return bodyType, paramsType, responseType
}

// AuthMultipartHandler handles a MultipartRequest with params and an authenticated model.
// The MultipartRequest holds a MultipartReader and the parsed params.
// The reason to provide the reader is to allow the logic for processing the parts to be handled by the handler function.
//
//	Example usage:
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
//	func(ctx context.Context, req *simba.MultipartRequest[Params], authModel *AuthModel) (*simba.Response[map[string]string], error) {
//		// Access the Multipart reader and params fields
//		req.Params.Name
//		req.Params.ID
//		req.Params.Page
//		req.Params.Size
//		req.Reader // Multipart reader
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
//	Mux.POST("/test/{id}", simba.AuthMultipartHandler(handler))
func AuthMultipartHandler[Params, AuthModel, ResponseBody any](h AuthenticatedMultipartHandlerFunc[Params, AuthModel, ResponseBody]) Handler {
	return h
}

// AuthenticatedMultipartHandlerFunc is a function type for handling a MultipartRequest with params and an authenticated model
type AuthenticatedMultipartHandlerFunc[Params, AuthModel, ResponseBody any] func(ctx context.Context, req *MultipartRequest[Params], authModel *AuthModel) (*Response[ResponseBody], error)

func (h AuthenticatedMultipartHandlerFunc[Params, AuthModel, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authFunc := r.Context().Value(simbaContext.AuthFuncKey).(AuthFunc[AuthModel])
	if authFunc == nil {
		WriteError(w, r, NewHttpError(http.StatusUnauthorized, "auth function is not set", nil))
		return
	}

	authModel, err := authFunc(r)
	if err != nil {
		WriteError(w, r, NewHttpError(http.StatusUnauthorized, "failed to authenticate", err))
		return
	}

	req, err := handleMultipartRequest[Params](r)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp, err := h(ctx, req, authModel)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	writeResponse(w, r, resp, nil)
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthModel, ResponseBody]) getTypes() (reflect.Type, reflect.Type, reflect.Type) {
	var file multipart.File
	var p Params
	var resb ResponseBody

	bodyType := reflect.TypeOf(&file).Elem()
	paramsType := reflect.TypeOf(p)
	responseType := reflect.TypeOf(resb)

	slog.Debug("type information",
		"bodyType", bodyType,
		"paramsType", paramsType,
		"responseType", responseType,
	)

	return bodyType, paramsType, responseType
}

// handleMultipartRequest handles extracting the [multipart.Reader] and params from the MultiPart Request
func handleMultipartRequest[Params any](r *http.Request) (*MultipartRequest[Params], error) {

	contentType := r.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "multipart/form-data") {
		return nil, NewHttpError(http.StatusBadRequest, "invalid content type", nil)
	}

	reqParams, err := parseAndValidateParams[Params](r)
	if err != nil {
		return nil, err
	}

	if _, params, err := mime.ParseMediaType(contentType); err != nil || params["boundary"] == "" {
		return nil, NewHttpError(http.StatusBadRequest, "invalid content type", err, ValidationError{
			Parameter: "Content-Type",
			Type:      ParameterTypeHeader,
			Message:   "multipart form-data requests must include a boundary parameter",
		})
	}

	multipartReader, err := r.MultipartReader()
	if err != nil {
		return nil, NewHttpError(http.StatusBadRequest, "invalid request body", err)
	}

	return &MultipartRequest[Params]{
		Cookies: r.Cookies(),
		Reader:  multipartReader,
		Params:  reqParams,
	}, nil
}
