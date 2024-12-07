package middleware

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/oklog/ulid"
	"github.com/sillen102/simba/logging"
)

type RequestIdMode int

const (
	AcceptFromHeader RequestIdMode = iota
	AlwaysGenerate
)

func (r RequestIdMode) String() string {
	switch r {
	case AcceptFromHeader:
		return "accept_from_header"
	case AlwaysGenerate:
		return "always_generate"
	default:
		return "accept_from_header"
	}
}

func (f *RequestIdMode) SetValue(s string) error {
	*f = ParseRequestId(s)
	return nil
}

func ParseRequestId(s string) RequestIdMode {
	switch string(s) {
	case "alwaysGenerate":
		return AlwaysGenerate
	default:
		return AcceptFromHeader
	}
}

type contextKey string

const (
	RequestIDKey    contextKey = "requestId"
	RequestIDHeader string     = "X-Request-Id"
)

// RequestIdConfig is the configuration for the request ID middleware
type RequestIdConfig struct {
	AcceptFromHeader bool
}

// RequestID middleware that adds a request ID to the context of the request
func (c *RequestIdConfig) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestID string

		if c.AcceptFromHeader {
			requestID = r.Header.Get(RequestIDHeader)
		}

		if requestID == "" {
			ms := ulid.Timestamp(time.Now())
			entropy := ulid.Monotonic(rand.Reader, 0)
			id, _ := ulid.New(ms, entropy)
			requestID = id.String()
		}

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Add request ID to logger in context
		logger := logging.Get(r.Context()).With().Str(string(RequestIDKey), requestID).Logger()
		ctx = logger.WithContext(ctx)

		// Set the request ID header
		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
