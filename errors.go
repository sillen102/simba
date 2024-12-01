package simba

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sillen102/simba/middleware"
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

// HasErrors checks if there are validation errors
func (e *HTTPError) HasValidationErrors() bool {
	return len(e.ValidationErrors) > 0
}

// NewApiError creates a new ApiError
func NewHttpError(httpStatusCode int, publicMessage string, err error, validationErrors ...ValidationError) *HTTPError {
	return &HTTPError{
		HttpStatusCode:   httpStatusCode,
		Message:          publicMessage,
		ValidationErrors: validationErrors,
		err:              err,
	}
}

// IsHTTPError checks if the error is an [HTTPError].
func IsHTTPError(err error) bool {
	var httpError *HTTPError
	ok := errors.As(err, &httpError)
	return ok
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
	// Path of the request
	Path string `json:"path"`
	// Method of the request
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
	if id := r.Context().Value(middleware.RequestIDKey); id != nil {
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

// ValidationError defines the interface for a validation error
// @Description Detailed information about a validation error
type ValidationError struct {
	// Field that failed validation or request if the whole request body is invalid
	Field string `json:"field"`
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
	return fmt.Sprintf("request validation failed: %d errors", len(ve))
}

// WriteJSONError writes a JSON error response to the response writer
func WriteJSONError(w http.ResponseWriter, errorResponse *ErrorResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorResponse.Status)
	return json.NewEncoder(w).Encode(errorResponse)
}

// HandleError is a helper function for handling errors in HTTP handlers
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
