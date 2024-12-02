package middleware

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/oklog/ulid"
	"github.com/sillen102/simba/logging"
)

type contextKey string

const (
	RequestIDKey    contextKey = "requestId"
	RequestIDHeader string     = "X-Request-Id"
)

// RequestIdConfig is the configuration for the request ID middleware
type RequestIdConfig struct {
	AcceptFromHeader bool
}

// AddRequestID middleware that adds a request ID to the context of the request
func (c *RequestIdConfig) AddRequestID(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

		// Create a logger with the request ID
		logger := logging.Get().With().Str(string(RequestIDKey), requestID).Logger()

		// Add both request ID and logger to context
		ctx := logger.WithContext(context.WithValue(r.Context(), RequestIDKey, requestID))

		// Set the request ID header
		w.Header().Set(RequestIDHeader, requestID)

		next(w, r.WithContext(ctx), ps)
	}
}
