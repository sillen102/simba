package middleware

import (
	"net/http"

	"github.com/sillen102/simba/models"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"

	"github.com/google/uuid"
)

func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := simbaContext.GetTraceID(r.Context())

		if traceID == "" {
			requestSettings, ok := r.Context().Value(simbaContext.RequestSettingsKey).(*settings.Request)
			if ok && requestSettings.TraceIDMode == models.AcceptFromHeader {
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
		}

		ctx := simbaContext.WithTraceID(r.Context(), traceID)
		w.Header().Set(simbaContext.TraceIDHeader, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
