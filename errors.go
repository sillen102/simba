package simba

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sillen102/simba/logging"
)

type HTTPError struct {
	HttpStatusCode   int
	Message          string
	err              error
	ValidationErrors ValidationErrors
}

// Error implements the error interface and returns the full error details
func (e *HTTPError) Error() string {
	// If there's an underlying error, just return the message
	// as the underlying error will be accessible via Unwrap()
	return e.Message
}

// WrapErrorHTTP wraps an error with an HTTP status code
func WrapErrorHTTP(httpStatusCode int, err error, message string) *HTTPError {
	return &HTTPError{
		HttpStatusCode: httpStatusCode,
		Message:        message,
		err:            err,
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
		Message:          publicMessage,
		ValidationErrors: validationErrors,
		err:              err,
	}
}

// ErrorResponse defines the structure of an error message
// @Description Represents the structure of an error message returned by the API
type ErrorResponse struct {
	// Timestamp of the error
	Timestamp time.Time `json:"timestamp"`
	// HTTP status code
	Status int `json:"status"`
	// HTTP error type
	Error string `json:"error"`
	// Path of the Request
	Path string `json:"path"`
	// Method of the Request
	Method string `json:"method"`
	// Request ID
	RequestID string `json:"requestId,omitempty"`
	// Error message
	Message string `json:"message,omitempty"`
	// Validation errors
	ValidationErrors []ValidationError `json:"validationErrors,omitempty"`
} // @Name ErrorResponse

// NewErrorResponse creates a new ErrorResponse instance with the given status and message
func NewErrorResponse(r *http.Request, status int, message string, validationErrors ...ValidationError) *ErrorResponse {
	// Safely get RequestID from context
	var requestID string
	if id := r.Context().Value(RequestIDKey); id != nil {
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
	ParameterTypePath   ParameterType = "path"
	ParameterTypeQuery  ParameterType = "query"
	ParameterTypeBody   ParameterType = "body"
)

func (p ParameterType) String() string {
	return string(p)
}

// ValidationError defines the interface for a validation error
// @Description Detailed information about a validation error
type ValidationError struct {
	// Parameter that failed validation
	Parameter string `json:"parameter"`
	// Type indicates where the parameter was located (header, path, query, body)
	Type ParameterType `json:"type"`
	// Error message describing the validation error
	Message string `json:"message"`
} // @Name ValidationError

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	return fmt.Sprintf("Request validation failed: %d errors", len(ve))
}

// HandleError is a helper function for handling errors in HTTP handlers
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	logger := logging.From(r.Context())

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

	errorMessage := httpErr.Message
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
		// Set error message to "forbidden" for the response
		// to hide details of the error to a potential attacker
		// and reduce the amount of information that can be
		// leaked. The error is logged above.
		errorMessage = "forbidden"

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
		handleUnexpectedError(w)
		return
	}
}

// writeJSONError writes a JSON error response to the response writer
func writeJSONError(w http.ResponseWriter, errorResponse *ErrorResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorResponse.Status)
	return json.NewEncoder(w).Encode(errorResponse)
}

// handleUnexpectedError is a helper function for handling unexpected errors
func handleUnexpectedError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
}
