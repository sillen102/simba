package simba

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type contextKey string

const (
	routerOptionsKey contextKey = "routerOptionsKey"
)

// injectConfiguration injects the RouterOptions into the request context
func injectConfiguration(next http.Handler, options *RouterOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), routerOptionsKey, options)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getConfigurationFromContext retrieves RouterOptions from the given context.
// Returns the options stored in the context or zero value for RouterOptions if not found in the context.
func getConfigurationFromContext(ctx context.Context) *RouterOptions {
	options, ok := ctx.Value(routerOptionsKey).(*RouterOptions)
	if !ok {
		// Return a default or zero value, or handle the absence of RouterOptions appropriately
		return &RouterOptions{}
	}
	return options
}

// decodeBodyIfNeeded decodes the request body if it is not of NoBody type
func decodeBodyIfNeeded[RequestBody any](r *http.Request, req *RequestBody) error {
	if _, isNoBody := any(*req).(NoBody); isNoBody {
		return nil
	}

	if r.Header.Get("Content-Type") != "application/json" {
		return NewHttpError(http.StatusBadRequest, "invalid content type", nil)
	}

	return readJson(r.Body, getConfigurationFromContext(r.Context()), req)
}

// readJson reads the JSON body and unmarshalls it into the model
func readJson(body io.ReadCloser, options *RouterOptions, model any) error {
	decoder := json.NewDecoder(body)
	if options.RequestDisallowUnknownFields != nil && *options.RequestDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(&model)
	if err != nil {
		return NewHttpError(http.StatusBadRequest, "invalid request body", err)
	}
	return nil
}
