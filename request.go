package simba

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/google/uuid"
	"github.com/sillen102/simba/enums"
	"github.com/sillen102/simba/logging"
)

// TODO: Request process testing
// 	1. Request timeouts
//  2. Large request body
//  3. Malformed JSON
//  4. Header validation edge cases

type requestContextKey string
type authContextKey string
type requestIdContextKey string

const (
	RequestIDKey    requestIdContextKey = "requestId"
	RequestIDHeader string              = "X-Request-Id"
)

const (
	requestSettingsKey requestContextKey = "requestSettings"
	authFuncKey        authContextKey    = "authFunc"
)

// injectRequestID injects the Request ID into the [Request] context
func injectRequestID(next http.Handler, acceptFromHeader bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestID string
		if acceptFromHeader {
			requestID = r.Header.Get(RequestIDHeader)
		}
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// injectLogger injects the logger into the Request context
func injectLogger(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(logging.With(r.Context(),
			logger.With(
				"method", r.Method,
				"path", r.URL.Path,
				"requestId", r.Context().Value(RequestIDKey),
			),
		)))
	})
}

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
	if requestSettings.LogRequestBody != enums.Disabled {
		logging.From(r.Context()).Info("request body", "body", r.Body)
	}

	err = readJson(r.Body, requestSettings, req)
	if err != nil {
		return err
	}

	if validationErrors := validateStruct(req, ParameterTypeBody); len(validationErrors) > 0 {
		return NewHttpError(http.StatusBadRequest, "invalid request body", nil, validationErrors...)
	}

	return nil
}

// readJson reads the JSON body and unmarshalls it into the model
func readJson(body io.ReadCloser, requestSettings *RequestSettings, model any) error {
	decoder := json.NewDecoder(body)
	if requestSettings.AllowUnknownFields == enums.Disallow {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(&model)
	if err != nil {
		return NewHttpError(http.StatusBadRequest, "invalid request body", err)
	}
	return nil
}
