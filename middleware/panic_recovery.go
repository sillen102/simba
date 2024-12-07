package middleware

import (
	"fmt"
	"net/http"

	"github.com/sillen102/simba/logging"
)

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logging.Get(r.Context()).Error().Msg(fmt.Sprintf("Recovered from panic: %v", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
