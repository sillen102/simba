package simbaErrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/simbaContext"
)

type HTTPError struct {
	HttpStatusCode   int
	PublicMessage    string
	ValidationErrors ValidationErrors
	err              error
}

// Error implements the error interface and returns the full error details
func (e *HTTPError) Error() string {
	if e.err != nil {
		if e.PublicMessage != "" {
			return e.PublicMessage + ": " + e.err.Error()
		} else {
			return e.err.Error()
		}
	}
	return e.PublicMessage
}

// WrapError wraps an error with an HTTP status code
func WrapError(httpStatusCode int, err error, publicMessage string, validationErrors ...ValidationError) *HTTPError {
	return &HTTPError{
		HttpStatusCode:   httpStatusCode,
		PublicMessage:    publicMessage,
		ValidationErrors: validationErrors,
		err:              err,
	}
}

// Unwrap returns the underlying error
func (e *HTTPError) Unwrap() error {
	return e.err
}

// HasValidationErrors checks if there are validation errors
func (e *HTTPError) HasValidationErrors() bool {
	return len(e.ValidationErrors) > 0
}

// NewHttpError creates a new ApiError
func NewHttpError(httpStatusCode int, publicMessage string, err error, validationErrors ...ValidationError) *HTTPError {
	return &HTTPError{
		HttpStatusCode:   httpStatusCode,
		PublicMessage:    publicMessage,
		ValidationErrors: validationErrors,
		err:              err,
	}
}

// ErrorResponse defines the structure of an error message
type ErrorResponse struct {
	// Timestamp of the error
	Timestamp time.Time `json:"timestamp" example:"2021-01-01T12:00:00Z"`
	// HTTP status code
	Status int `json:"status" example:"400"`
	// HTTP error type
	Error string `json:"error" example:"Bad Request"`
	// Path of the Request
	Path string `json:"path" example:"/api/v1/users"`
	// Method of the Request
	Method string `json:"method" example:"GET"`
	// Request ID
	RequestID string `json:"requestId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000" required:"false"`
	// Error message
	Message string `json:"message,omitempty" example:"Validation failed"`
	// Validation errors
	ValidationErrors []ValidationError `json:"validationErrors,omitempty" required:"false"`
}

// NewErrorResponse creates a new ErrorResponse instance with the given status and message
func NewErrorResponse(r *http.Request, status int, message string, validationErrors ...ValidationError) *ErrorResponse {
	// Safely get RequestID from context
	var requestID string
	if id := r.Context().Value(simbaContext.RequestIDKey); id != nil {
		if strID, ok := id.(string); ok {
			requestID = strID
		}
	}

	return &ErrorResponse{
		Timestamp:        time.Now().UTC(),
		Status:           status,
		Error:            http.StatusText(status),
		Path:             r.URL.Path,
		Method:           r.Method,
		RequestID:        requestID,
		Message:          message,
		ValidationErrors: validationErrors,
	}
}

type ParameterType string

const (
	ParameterTypeHeader ParameterType = "header"
	ParameterTypeCookie ParameterType = "cookie"
	ParameterTypePath   ParameterType = "path"
	ParameterTypeQuery  ParameterType = "query"
	ParameterTypeBody   ParameterType = "body"
)

func (p ParameterType) String() string {
	return string(p)
}

// ValidationError defines the interface for a validation error
type ValidationError struct {
	// Parameter or field that failed validation
	Parameter string `json:"parameter" example:"name"`
	// Type indicates where the parameter was located (header, path, query, body)
	Type ParameterType `json:"type" example:"query"`
	// Error message describing the validation error
	Message string `json:"message" example:"name is required"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	return fmt.Sprintf("Request validation failed: %d errors", len(ve))
}

// WriteError is a helper function for handling errors in HTTP handlers
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	logger := logging.From(r.Context())
	if logger == nil {
		logger = slog.Default()
	}

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		// Log unexpected errors as they are always serious
		logger.Error("unexpected error encountered", "error", err)
		err = writeJSONError(w, NewErrorResponse(r, http.StatusInternalServerError, "Internal server error"))
		if err != nil {
			logger.Error("failed to write error response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	errorMessage := httpErr.PublicMessage
	if errorMessage == "" {
		errorMessage = "an error occurred"
	}

	switch httpErr.HttpStatusCode {
	case http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusConflict,
		http.StatusUnprocessableEntity:
		// Log debug for 400, 404, 405, 409 and 422 errors as they are not serious
		// and, are returned before reaching the handler and can usually be fixed
		// by the user.
		logger.Debug(errorMessage, "error", httpErr.Unwrap())

	case http.StatusUnauthorized:
		// Log warnings for 401 errors
		logger.Warn(errorMessage, "error", httpErr.Unwrap())
		// Set error message to "unauthorized" for the response
		// to hide details of the error to a potential attacker
		// and reduce the amount of information that can be
		// leaked. The error is logged above.
		errorMessage = "unauthorized"

	case http.StatusForbidden:
		// Log warnings for 403 errors
		logger.Warn(errorMessage, "error", httpErr.Unwrap())

	default:
		// Log errors for other HTTP status codes
		logger.Error(errorMessage, "error", httpErr.Unwrap())
	}

	var errorResponse *ErrorResponse
	if httpErr.HasValidationErrors() {
		errorResponse = NewErrorResponse(r, httpErr.HttpStatusCode, errorMessage, httpErr.ValidationErrors...)
	} else {
		errorResponse = NewErrorResponse(r, httpErr.HttpStatusCode, errorMessage)
	}

	err = writeJSONError(w, errorResponse)
	if err != nil {
		logger.Error("failed to write error response", "error", err)
		HandleUnexpectedError(w)
		return
	}
}

// HandleUnexpectedError is a helper function for handling unexpected errors
func HandleUnexpectedError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
}

// writeJSONError writes a JSON error response to the response writer
func writeJSONError(w http.ResponseWriter, errorResponse *ErrorResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorResponse.Status)
	return json.NewEncoder(w).Encode(errorResponse)
}
