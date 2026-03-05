package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/sillen102/simba/simbaContext"
)

// TraceIDFromOTel injects the current OTel trace ID into Simba request context when available.
func TraceIDFromOTel(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanCtx := trace.SpanContextFromContext(r.Context())
		if spanCtx.IsValid() {
			r = r.WithContext(simbaContext.WithTraceID(r.Context(), spanCtx.TraceID().String()))
		}

		next.ServeHTTP(w, r)
	})
}
