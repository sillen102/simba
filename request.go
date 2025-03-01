package simba

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/sillen102/simba/enums"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
)

// TODO: Request process testing
// 	1. Request timeouts
//  2. Large request body
//  3. Malformed JSON
//  4. Header validation edge cases

// closeRequestBody automatically closes the Request body after processing
func closeRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logging.From(r.Context()).Error("error closing Request body", "error", err)
			}
		}(r.Body)
		next.ServeHTTP(w, r)
	})
}

// injectRequestSettings injects the application Config into the Request context
func injectRequestSettings(next http.Handler, settings *settings.Request) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), simbaContext.RequestSettingsKey, settings)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getConfigurationFromContext retrieves Request from the given context.
// Returns the request Config stored in the context or zero value for Request if not found in the context.
func getConfigurationFromContext(ctx context.Context) *settings.Request {
	requestSettings, ok := ctx.Value(simbaContext.RequestSettingsKey).(*settings.Request)
	if !ok {
		// Return a default or zero value, or handle the absence of Request appropriately
		return &settings.Request{}
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
	if requestSettings.LogRequestBody != enums.Disabled {
		logging.From(r.Context()).Info("request body", "body", r.Body)
	}

	err = readJson(r.Body, requestSettings, req)
	if err != nil {
		return err
	}

	if validationErrors := ValidateStruct(req, ParameterTypeBody); len(validationErrors) > 0 {
		return NewHttpError(http.StatusBadRequest, "invalid request body", nil, validationErrors...)
	}

	return nil
}

// readJson reads the JSON body and unmarshalls it into the model
func readJson(body io.ReadCloser, requestSettings *settings.Request, model any) error {
	decoder := json.NewDecoder(body)
	if requestSettings.AllowUnknownFields == enums.Disallow {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(&model)
	if err != nil {
		return NewHttpError(
			http.StatusUnprocessableEntity,
			"invalid request body",
			err,
			ValidationError{
				Parameter: "body",
				Type:      ParameterTypeBody,
				Message:   err.Error(),
			},
		)
	}
	return nil
}
