package middleware

import (
	"net/http"

	"github.com/sillen102/simba/models"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaContext"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var traceID string

		// Check if OTEL trace ID exists in span context
		// This takes precedence when telemetry is enabled
		spanCtx := trace.SpanContextFromContext(r.Context())
		if spanCtx.IsValid() {
			traceID = spanCtx.TraceID().String()
		} else {
			// Fallback to existing Simba trace ID logic
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
