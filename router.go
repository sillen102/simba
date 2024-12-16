package simba

import (
	"net/http"

	"github.com/sillen102/simba/settings"
)

// Router is a simple Mux that wraps [http.ServeMux] and allows for middleware chaining
type Router struct {
	Mux        *http.ServeMux
	middleware []func(http.Handler) http.Handler
}

// newRouter creates a new [Router] instance with the given logger (that is injected in each Request context) and [Config]
func newRouter(requestSettings settings.Request) *Router {
	return &Router{
		Mux: http.NewServeMux(),
		middleware: []func(http.Handler) http.Handler{
			closeRequestBody,
			func(next http.Handler) http.Handler {
				return injectRequestSettings(next, &requestSettings)
			},
		},
	}
}

// ServeHTTP implements the [http.Handler] interface for the [Router] type
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Mux.ServeHTTP(w, req)
}

// Use registers a middleware handler
func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware)
}

// Extend extends the middleware chain with another chain
func (r *Router) Extend(middleware []func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware...)
}

func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}

// Handle registers a standard lib handler for the given pattern
func (r *Router) Handle(pattern string, handler http.HandlerFunc) {
	r.Mux.Handle(pattern, r.applyMiddleware(handler))
}

// POST registers a handler for POST requests to the given pattern
func (r *Router) POST(path string, handler http.Handler) {
	r.Mux.Handle("POST "+path, r.applyMiddleware(handler))
}

// GET registers a handler for GET requests to the given pattern
func (r *Router) GET(path string, handler http.Handler) {
	r.Mux.Handle("GET "+path, r.applyMiddleware(handler))
}

// PUT registers a handler for PUT requests to the given pattern
func (r *Router) PUT(path string, handler http.Handler) {
	r.Mux.Handle("PUT "+path, r.applyMiddleware(handler))
}

// DELETE registers a handler for DELETE requests to the given pattern
func (r *Router) DELETE(path string, handler http.Handler) {
	r.Mux.Handle("DELETE "+path, r.applyMiddleware(handler))
}

// PATCH registers a handler for PATCH requests to the given pattern
func (r *Router) PATCH(path string, handler http.Handler) {
	r.Mux.Handle("PATCH "+path, r.applyMiddleware(handler))
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (r *Router) OPTIONS(path string, handler http.Handler) {
	r.Mux.Handle("OPTIONS "+path, r.applyMiddleware(handler))
}

// HEAD registers a handler for HEAD requests to the given pattern
func (r *Router) HEAD(path string, handler http.Handler) {
	r.Mux.Handle("HEAD "+path, r.applyMiddleware(handler))
}
