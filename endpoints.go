package simba

import (
	"net/http"
)

// addDefaultEndpoints adds the default endpoints to the Mux
func (a *Application[AuthModel]) addDefaultEndpoints() {
	a.Router.Mux.HandleFunc("GET /health", healthCheck)
}

// healthCheck is a simple health check endpoint
func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{\"status\":\"ok\"}"))
}
