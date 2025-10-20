package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaModels"
)

func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var traceID string

		requestSettings, ok := r.Context().Value(simbaContext.RequestSettingsKey).(*settings.Request)
		if ok && requestSettings.TraceIDMode == simbaModels.AcceptFromHeader {
			traceID = r.Header.Get(simbaContext.TraceIDHeader)
		}

		if traceID == "" {
			id, err := uuid.NewV7()
			if err != nil || id == uuid.Nil {
				traceID = uuid.NewString()
			} else {
				traceID = id.String()
			}
		}

		ctx := context.WithValue(r.Context(), simbaContext.TraceIDKey, traceID)
		w.Header().Set(simbaContext.TraceIDHeader, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
