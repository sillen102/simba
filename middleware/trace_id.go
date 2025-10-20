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
		var ctx context.Context

		requestSettings, ok := r.Context().Value(simbaContext.RequestSettingsKey).(*settings.Request)
		if ok && requestSettings.TraceIDMode == simbaModels.AcceptFromHeader {
			ctx = getWithNewTraceIDOrDefaultIfPresent(r.Context(), r.Header.Get(simbaContext.TraceIDHeader))
		} else {
			ctx = GetWithTraceID(r.Context())
		}

		w.Header().Set(simbaContext.TraceIDHeader, ctx.Value(simbaContext.TraceIDKey).(string))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetWithTraceID returns a context with a trace ID. If the context already has a trace ID, it is reused.
func GetWithTraceID(ctx context.Context) context.Context {
	var traceID string
	traceID, ok := ctx.Value(simbaContext.TraceIDKey).(string)
	if !ok || traceID == "" {
		id, err := uuid.NewV7()
		if err != nil || id == uuid.Nil {
			traceID = uuid.NewString()
		} else {
			traceID = id.String()
		}
	}
	return context.WithValue(ctx, simbaContext.TraceIDKey, traceID)
}

// GetTraceID retrieves the trace ID from the context. If no trace ID is present, it returns an empty string.
func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(simbaContext.TraceIDKey).(string)
	if !ok {
		return ""
	}
	return traceID
}

func getWithNewTraceIDOrDefaultIfPresent(ctx context.Context, defaultTraceID string) context.Context {
	if defaultTraceID == "" {
		return GetWithTraceID(ctx)
	}
	return context.WithValue(ctx, simbaContext.TraceIDKey, defaultTraceID)
}
