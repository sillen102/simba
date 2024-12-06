package simba

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/sillen102/simba/logging"
)

type requestContextKey string
type authContextKey string

const (
	requestSettingsKey requestContextKey = "requestSettings"
	authFuncKey        authContextKey    = "authFunc"
)

// injectLogger injects the logger into the Request context
func injectLogger(next http.Handler, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logging.WithLogger(r.Context(), logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// autoCloseRequestBody automatically closes the Request body
func autoCloseRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logging.FromCtx(r.Context()).Error().Err(err).Msg("error closing Request body")
			}
		}(r.Body)
		next.ServeHTTP(w, r)
	})
}

// injectRequestSettings injects the application Settings into the Request context
func injectRequestSettings(next http.Handler, settings RequestSettings) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), requestSettingsKey, settings)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// injectAuthFunc injects the AuthFunc into the Request context
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

// getConfigurationFromContext retrieves RequestSettings from the given context.
// Returns the request settings stored in the context or zero value for RequestSettings if not found in the context.
func getConfigurationFromContext(ctx context.Context) *RequestSettings {
	requestSettings, ok := ctx.Value(requestSettingsKey).(*RequestSettings)
	if !ok {
		// Return a default or zero value, or handle the absence of RequestSettings appropriately
		return &RequestSettings{}
	}
	return requestSettings
}

// handleJsonBody decodes the request body if it is not of NoBody type and unmarshalls it into the model
// If the content type is not "application/json", returns an error
// If the request body is of NoBody type, returns nil
// If there are validation errors for the request body, returns an error
func handleJsonBody[RequestBody any](r *http.Request, req *RequestBody) error {
	if _, isNoBody := any(*req).(NoBody); isNoBody {
		return nil
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return NewHttpError(http.StatusBadRequest, "invalid content type", nil)
	}

	requestSettings := getConfigurationFromContext(r.Context())
	if requestSettings.LogRequestBody != zerolog.Disabled {
		logging.FromCtx(r.Context()).
			WithLevel(requestSettings.LogRequestBody).
			Interface("body", r.Body).
			Msg("request body")
	}

	err = readJson(r.Body, requestSettings, req)
	if err != nil {
		return err
	}

	if validationErrors := validateStruct(req); len(validationErrors) > 0 {
		return NewHttpError(http.StatusBadRequest, "invalid Request body", nil, validationErrors...)
	}

	return nil
}

// readJson reads the JSON body and unmarshalls it into the model
func readJson(body io.ReadCloser, requestSettings *RequestSettings, model any) error {
	decoder := json.NewDecoder(body)
	if requestSettings.UnknownFields == Disallow {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(&model)
	if err != nil {
		return NewHttpError(http.StatusBadRequest, "invalid Request body", err)
	}
	return nil
}
