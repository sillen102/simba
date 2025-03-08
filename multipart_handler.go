package simba

import (
	"context"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// MultipartHandlerFunc is a function type for handling routes with Request body and params
type MultipartHandlerFunc[Params any, ResponseBody any] func(ctx context.Context, req *simbaModels.MultipartRequest[Params]) (*simbaModels.Response[ResponseBody], error)

// AuthenticatedMultipartHandlerFunc is a function type for handling a MultipartRequest with params and an authenticated model
type AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody any] struct {
	handler     func(ctx context.Context, req *simbaModels.MultipartRequest[Params], authModel *AuthModel) (*simbaModels.Response[ResponseBody], error)
	authHandler AuthHandler[AuthParams, AuthModel]
}

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

// ServeHTTP implements the http.Handler interface for JsonHandlerFunc
func (h MultipartHandlerFunc[Params, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := handleMultipartRequest[Params](r)
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

func (h MultipartHandlerFunc[Params, ResponseBody]) getRequestBody() any {
	var file multipart.File
	return &file
}

func (h MultipartHandlerFunc[Params, ResponseBody]) getParams() any {
	var p Params
	return p
}

func (h MultipartHandlerFunc[Params, ResponseBody]) getResponseBody() any {
	var resb ResponseBody
	return resb
}

func (h MultipartHandlerFunc[Params, ResponseBody]) getAccepts() string {
	return mimetypes.MultipartForm
}

func (h MultipartHandlerFunc[Params, ResponseBody]) getProduces() string {
	return mimetypes.ApplicationJSON
}

func (h MultipartHandlerFunc[Params, ResponseBody]) getHandler() any {
	return h
}

func (h MultipartHandlerFunc[Params, ResponseBody]) getAuthModel() any {
	return nil
}

func (h MultipartHandlerFunc[Params, ResponseBody]) getAuthHandler() any {
	return nil
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
func AuthMultipartHandler[Params, AuthParams, AuthModel, ResponseBody any](
	handler func(ctx context.Context, req *simbaModels.MultipartRequest[Params], authModel *AuthModel) (*simbaModels.Response[ResponseBody], error),
	authHandler AuthHandler[AuthParams, AuthModel],
) Handler {
	return AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]{
		handler:     handler,
		authHandler: authHandler,
	}
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authModel, err := handleAuthRequest[AuthParams, AuthModel](h.authHandler, r)
	if err != nil {
		simbaErrors.WriteError(w, r, simbaErrors.NewHttpError(http.StatusUnauthorized, "failed to authenticate", err))
		return
	}

	req, err := handleMultipartRequest[Params](r)
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

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getRequestBody() any {
	var file multipart.File
	return &file
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getResponseBody() any {
	var resb ResponseBody
	return resb
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getParams() any {
	var p Params
	return p
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getAccepts() string {
	return mimetypes.MultipartForm
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getProduces() string {
	return mimetypes.ApplicationJSON
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getHandler() any {
	return h.handler
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getAuthModel() any {
	var am AuthModel
	return am
}

func (h AuthenticatedMultipartHandlerFunc[Params, AuthParams, AuthModel, ResponseBody]) getAuthHandler() any {
	return h.authHandler
}

// handleMultipartRequest handles extracting the [multipart.Reader] and params from the MultiPart Request
func handleMultipartRequest[Params any](r *http.Request) (*simbaModels.MultipartRequest[Params], error) {

	contentType := r.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "multipart/form-data") {
		return nil, simbaErrors.NewHttpError(http.StatusBadRequest, "invalid content type", nil)
	}

	reqParams, err := parseAndValidateParams[Params](r)
	if err != nil {
		return nil, err
	}

	if _, params, err := mime.ParseMediaType(contentType); err != nil || params["boundary"] == "" {
		return nil, simbaErrors.NewHttpError(http.StatusBadRequest, "invalid content type", err, simbaErrors.ValidationError{
			Parameter: "Content-Type",
			Type:      simbaErrors.ParameterTypeHeader,
			Message:   "multipart form-data requests must include a boundary parameter",
		})
	}

	multipartReader, err := r.MultipartReader()
	if err != nil {
		return nil, simbaErrors.NewHttpError(http.StatusBadRequest, "invalid request body", err)
	}

	return &simbaModels.MultipartRequest[Params]{
		Reader: multipartReader,
		Params: reqParams,
	}, nil
}
