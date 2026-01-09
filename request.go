package simba

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"reflect"
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

	// Handle setting defaults on request body fields
	errs := setDefaultsFromTags(req)
	if len(errs) > 0 {
		return simbaErrors.NewSimbaError(
			http.StatusInternalServerError,
			"invalid default value(s)",
			nil,
		).WithDetails(errs)
	}

	var validationTarget any = req
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		elem := v.Elem()
		if elem.Kind() == reflect.Ptr {
			validationTarget = elem.Interface()
		}
	}

	if validationErrors := ValidateStruct(validationTarget); len(validationErrors) > 0 {
		return simbaErrors.NewSimbaError(
			http.StatusBadRequest,
			"request validation failed",
			nil,
		).WithDetails(validationErrors)
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

// setDefaultsFromTags sets default values for all zero-valued fields in a struct.
func setDefaultsFromTags(model any) []ValidationError {
	var errs []ValidationError
	v := reflect.ValueOf(model)

	// Dereference all pointer levels until we reach the struct
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return errs
		}
		v = v.Elem()
	}

	// Ensure we have a struct
	if v.Kind() != reflect.Struct {
		return errs
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)
		structField := t.Field(i)
		if fieldValue.CanSet() && fieldValue.IsZero() {
			if err := setDefaultValue(fieldValue, structField); err != nil {
				errs = append(errs, *err)
			}
		}
	}
	return errs
}
