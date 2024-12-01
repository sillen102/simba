package simba

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Global router options variable
var options RouterOptions

// Router is a HTTP router
type Router struct {
	router *httprouter.Router
}

// RouterOptions are options for the router
type RouterOptions struct {
	// RequestDisallowUnknownFields will disallow unknown fields in the request body,
	// resulting in a 400 Bad Request response if a field is present that cannot be
	// mapped to the model struct.
	RequestDisallowUnknownFields bool
}

// NewRouter returns a new Router
func NewRouter(opts ...RouterOptions) *Router {
	if len(opts) > 0 {
		options = opts[0]
	}

	return &Router{
		router: httprouter.New(),
	}
}

// ServeHTTP implements the http.Handler interface
func (s *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// POST registers a handler for POST requests to the given pattern
func (s *Router) POST(path string, handler http.Handler) {
	s.router.Handler(http.MethodPost, path, handler)
}

// GET registers a handler for GET requests to the given pattern
func (s *Router) GET(path string, handler http.Handler) {
	s.router.Handler(http.MethodGet, path, handler)
}

// PUT registers a handler for PUT requests to the given pattern
func (s *Router) PUT(path string, handler http.Handler) {
	s.router.Handler(http.MethodPut, path, handler)
}

// DELETE registers a handler for DELETE requests to the given pattern
func (s *Router) DELETE(path string, handler http.Handler) {
	s.router.Handler(http.MethodDelete, path, handler)
}

// PATCH registers a handler for PATCH requests to the given pattern
func (s *Router) PATCH(path string, handler http.Handler) {
	s.router.Handler(http.MethodPatch, path, handler)
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (s *Router) OPTIONS(path string, handler http.Handler) {
	s.router.Handler(http.MethodOptions, path, handler)
}

// HEAD registers a handler for HEAD requests to the given pattern
func (s *Router) HEAD(path string, handler http.Handler) {
	s.router.Handler(http.MethodHead, path, handler)
}

// CONNECT registers a handler for CONNECT requests to the given pattern
func (s *Router) CONNECT(path string, handler http.Handler) {
	s.router.Handler(http.MethodConnect, path, handler)
}

// TRACE registers a handler for TRACE requests to the given pattern
func (s *Router) TRACE(path string, handler http.Handler) {
	s.router.Handler(http.MethodTrace, path, handler)
}
