package simba

import "net/http"

// Router is a HTTP router
type Router struct {
	*http.ServeMux
}

// NewRouter returns a new Router
func NewRouter() *Router {
	return &Router{
		ServeMux: http.NewServeMux(),
	}
}
