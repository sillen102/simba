package simba

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

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

// injectRequestSettings injects the application Simba into the Request context
func injectRequestSettings(next http.Handler, requestSettings *settings.Request) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), simbaContext.RequestSettingsKey, requestSettings)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getConfigurationFromContext retrieves Request from the given context.
// Returns the request Simba stored in the context or zero value for Request if not found in the context.
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
	if _, isNoBody := any(*req).(simbaModels.NoBody); isNoBody {
		return nil
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return simbaErrors.ErrInvalidContentType.
			WithDetails("expected application/json, got: " + contentType)
	}

	requestSettings := getConfigurationFromContext(r.Context())
	if requestSettings.LogRequestBody {
		logging.From(r.Context()).Info("request body", "body", r.Body)
	}

	err = readJson(r.Body, requestSettings, req)
	if err != nil {
		return err
	}

	if validationErrors := ValidateStruct(req); len(validationErrors) > 0 {
		return simbaErrors.ErrInvalidRequest.WithDetails(validationErrors)
	}

	return nil
}

// readJson reads the JSON body and unmarshalls it into the model
func readJson(body io.ReadCloser, requestSettings *settings.Request, model any) error {
	decoder := json.NewDecoder(body)
	if !requestSettings.AllowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(&model)
	if err != nil {

		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			return simbaErrors.NewSimbaError(
				http.StatusUnprocessableEntity,
				"invalid request body",
				unmarshalTypeError,
			).WithDetails("invalid type for field: " + unmarshalTypeError.Field + ", expected " + unmarshalTypeError.Type.String())
		}

		var jsonSyntaxError *json.SyntaxError
		if errors.As(err, &jsonSyntaxError) {
			return simbaErrors.NewSimbaError(
				http.StatusUnprocessableEntity,
				"invalid request body",
				jsonSyntaxError,
			).WithDetails("invalid syntax at offset: " + strconv.Itoa(int(jsonSyntaxError.Offset)))
		}

		var invalidUnmarshalError *time.ParseError
		if errors.As(err, &invalidUnmarshalError) {
			return simbaErrors.NewSimbaError(
				http.StatusUnprocessableEntity,
				"invalid request body",
				invalidUnmarshalError,
			).WithDetails("invalid time format: " + invalidUnmarshalError.Value)
		}

		// Default case for JSON decoding errors
		return simbaErrors.NewSimbaError(
			http.StatusUnprocessableEntity,
			"invalid request body",
			err,
		).WithDetails("error decoding JSON")
	}
	return nil
}
