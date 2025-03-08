package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/sillen102/simba/logger"
)

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logger.From(r.Context()).Error("recovered from panic",
					"error", fmt.Sprint(err),
					"stacktrace", string(stack),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
