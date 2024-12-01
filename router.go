package simba

import (
	"net/http"

	"github.com/uptrace/bunrouter"
)

// Router is a HTTP router
type Router struct {
	router *bunrouter.Router
}

// RouterOptions are options for the router
type RouterOptions struct {
	// RequestDisallowUnknownFields will disallow unknown fields in the request body
	RequestDisallowUnknownFields bool
}

var options RouterOptions

// NewRouter returns a new Router
func NewRouter(opts ...RouterOptions) *Router {
	if len(opts) > 0 {
		options = opts[0]
	}

	return &Router{
		router: bunrouter.New(),
	}
}

func (s *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// POST registers a handler for POST requests to the given pattern
func (s *Router) POST(path string, handler Handler) {
	s.router.POST(path, handler.ServeHTTP)
}

// GET registers a handler for GET requests to the given pattern
func (s *Router) GET(path string, handler Handler) {
	s.router.GET(path, handler.ServeHTTP)
}

// PUT registers a handler for PUT requests to the given pattern
func (s *Router) PUT(path string, handler Handler) {
	s.router.PUT(path, handler.ServeHTTP)
}

// DELETE registers a handler for DELETE requests to the given pattern
func (s *Router) DELETE(path string, handler Handler) {
	s.router.DELETE(path, handler.ServeHTTP)
}

// PATCH registers a handler for PATCH requests to the given pattern
func (s *Router) PATCH(path string, handler Handler) {
	s.router.PATCH(path, handler.ServeHTTP)
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (s *Router) OPTIONS(path string, handler Handler) {
	s.router.OPTIONS(path, handler.ServeHTTP)
}

// HEAD registers a handler for HEAD requests to the given pattern
func (s *Router) HEAD(path string, handler Handler) {
	s.router.HEAD(path, handler.ServeHTTP)
}
