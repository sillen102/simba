package simba

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// addDefaultEndpoints adds the default endpoints to the router
func (a *Application[AuthModel]) addDefaultEndpoints() {
	a.Router.router.GET("/health", healthCheck)
}

// healthCheck is a simple health check endpoint
func healthCheck(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{\"status\":\"ok\"}"))
}
