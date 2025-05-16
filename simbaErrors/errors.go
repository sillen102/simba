package simbaErrors

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/simbaContext"
)

type StatusCodeProvider interface {
	StatusCode() int
}

type ErrorCodeProvider interface {
	ErrorCode() string
}

type PublicMessageProvider interface {
	PublicMessage() string
}

type DetailProvider interface {
	Details() any
}

type SimbaError struct {
	statusCode    int
	publicMessage string
	err           error
	details       any
}

func NewSimbaError(statusCode int, publicMessage string, err error) *SimbaError {
	return &SimbaError{
		statusCode:    statusCode,
		publicMessage: publicMessage,
		err:           err,
	}
}

func (e *SimbaError) WithDetails(details any) *SimbaError {
	e.details = details
	return e
}

func (e *SimbaError) Unwrap() error {
	return e.err
}

func (e *SimbaError) Error() string {
	if e.err == nil {
		return e.publicMessage
	}
	return e.err.Error()
}

func (e *SimbaError) StatusCode() int {
	return e.statusCode
}

func (e *SimbaError) PublicMessage() string {
	return e.publicMessage
}

func (e *SimbaError) Details() any {
	return e.details
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
	// Error code
	ErrorCode string `json:"errorCode,omitempty" example:"123-123" required:"false"`
	// Error message
	Message string `json:"message,omitempty" example:"Validation failed"`
	// Validation errors
	Details any `json:"details,omitempty" required:"false"`
}

// WriteError is a helper function for handling errors in HTTP handlers
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	errorCode := ""
	message := err.Error()
	var details any

	if statusCoder, ok := err.(StatusCodeProvider); ok {
		statusCode = statusCoder.StatusCode()
	}

	if errorProvider, ok := err.(ErrorCodeProvider); ok {
		errorCode = errorProvider.ErrorCode()
	}

	if msgProvider, ok := err.(PublicMessageProvider); ok {
		message = msgProvider.PublicMessage()
	}

	if detailProvider, ok := err.(DetailProvider); ok {
		details = detailProvider.Details()
	}

	logging.From(r.Context()).Error(err.Error(),
		"statusCode", statusCode,
		"error", err,
	)

	err = writeJSONError(w, newErrorResponse(r, statusCode, message, errorCode, details))
	if err != nil {
		HandleUnexpectedError(w)
		return
	}
}

// HandleUnexpectedError is a helper function for handling unexpected errors
func HandleUnexpectedError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
}

// writeJSONError writes a JSON error response to the response writer
func writeJSONError(w http.ResponseWriter, errorResponse *ErrorResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorResponse.Status)
	return json.NewEncoder(w).Encode(errorResponse)
}

// newErrorResponse creates a new ErrorResponse instance with the given status and message
func newErrorResponse(r *http.Request, status int, message string, errorCode string, details any) *ErrorResponse {
	// Safely get RequestID from context
	var requestID string
	if id := r.Context().Value(simbaContext.RequestIDKey); id != nil {
		if strID, ok := id.(string); ok {
			requestID = strID
		}
	}

	return &ErrorResponse{
		Timestamp: time.Now().UTC(),
		Status:    status,
		Error:     http.StatusText(status),
		Path:      r.URL.Path,
		Method:    r.Method,
		RequestID: requestID,
		ErrorCode: errorCode,
		Message:   message,
		Details:   details,
	}
}

// Predefined errors for common scenarios
var (
	ErrInvalidContentType = NewSimbaError(http.StatusBadRequest, "invalid content type", errors.New("invalid content type"))
	ErrInvalidRequest     = NewSimbaError(http.StatusUnprocessableEntity, "invalid request", errors.New("failed to decode request body"))
	ErrUnauthorized       = NewSimbaError(http.StatusUnauthorized, "unauthorized", errors.New("failed to authorize request"))
	ErrUnexpected         = NewSimbaError(http.StatusInternalServerError, "unexpected error", errors.New("unexpected error occurred"))
)
