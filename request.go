package simba

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/sillen102/simba/logging"
)

type optionsContextKey string
type authContextKey string

const (
	routerOptionsKey optionsContextKey = "routerOptions"
	authFuncKey      authContextKey    = "authFunc"
)

// injectLogger injects the logger into the request context
func injectLogger(next http.Handler, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// autoCloseRequestBody automatically closes the request body
func autoCloseRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logging.FromCtx(r.Context()).Error().Err(err).Msg("error closing request body")
			}
		}(r.Body)
		next.ServeHTTP(w, r)
	})
}

// injectOptions injects the RouterOptions into the request context
func injectOptions(next http.Handler, options Options) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), routerOptionsKey, options)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// injectAuthFunc injects the AuthFunc into the request context
func injectAuthFunc[User any](next http.Handler, authFunc AuthFunc[User]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authFunc == nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), authFuncKey, authFunc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getConfigurationFromContext retrieves RouterOptions from the given context.
// Returns the options stored in the context or zero value for RouterOptions if not found in the context.
func getConfigurationFromContext(ctx context.Context) *Options {
	options, ok := ctx.Value(routerOptionsKey).(*Options)
	if !ok {
		// Return a default or zero value, or handle the absence of RouterOptions appropriately
		return &Options{}
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
func readJson(body io.ReadCloser, options *Options, model any) error {
	decoder := json.NewDecoder(body)
	if options.RequestDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(&model)
	if err != nil {
		return NewHttpError(http.StatusBadRequest, "invalid request body", err)
	}
	return nil
}
