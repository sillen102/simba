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
			}
			traceID = id.String()
		}
		
		ctx := context.WithValue(r.Context(), simbaContext.TraceIDKey, traceID)
		w.Header().Set(simbaContext.TraceIDHeader, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetWithTraceID returns a context with a trace ID. If the context already has a trace ID, it is reused.
func GetWithTraceID(ctx context.Context) context.Context {
	var traceID string
	traceID, ok := ctx.Value(simbaContext.TraceIDKey).(string)
	if !ok {
		id, err := uuid.NewV7()
		if err != nil || id == uuid.Nil {
			traceID = uuid.NewString()
		}
		traceID = id.String()
	}
	return context.WithValue(ctx, simbaContext.TraceIDKey, traceID)
}
