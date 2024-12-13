package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sillen102/simba/enums"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestID string

		requestSettings, ok := r.Context().Value(simbaContext.RequestSettingsKey).(*settings.Request)
		if ok && requestSettings.RequestIdMode == enums.AcceptFromHeader {
			requestID = r.Header.Get(simbaContext.RequestIDHeader)
		}

		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), simbaContext.RequestIDKey, requestID)
		w.Header().Set(simbaContext.RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
