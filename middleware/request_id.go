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

const (
	// RequestIDKey is the key used to store the request ID in the context
	RequestIDKey    = "requestId"
	RequestIDHeader = "X-Request-Id"
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
		logger := logging.Get().With(RequestIDKey, requestID)

		// Add both request ID and logger to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logging.LoggerKey, logger)

		w.Header().Set(RequestIDHeader, requestID)

		next(w, r.WithContext(ctx), ps)
	}
}
