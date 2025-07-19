package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/sillen102/simba/settings"
)

func CORS(cfg settings.Cors) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			if origin == "" {
				// If no Origin header is present, do not set CORS headers
				next.ServeHTTP(w, r)
				return
			}

			// Split allowed origins by comma and check if the request origin is in the list
			origins := strings.Split(cfg.AllowedOrigins, ",")

			for _, o := range origins {
				if o == origin || o == "*" {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", cfg.AllowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", cfg.AllowedHeaders)
				w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(cfg.AllowCredentials))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
