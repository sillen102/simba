package simba

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/sillen102/simba/middleware"
)

// Router is a simple router that wraps [httprouter.Router] and allows for middleware chaining
type Router struct {
	router     *httprouter.Router
	middleware alice.Chain
}

// newRouter creates a new [Router] instance with the given logger (that is injected in each Request context) and [Settings]
func newRouter(requestSettings RequestSettings) *Router {
	return &Router{
		router: httprouter.New(),
		middleware: alice.New().
			Append(injectLogger).
			Append(middleware.PanicRecovery).
			Append(closeRequestBody).
			Append(func(next http.Handler) http.Handler {
				return injectRequestSettings(next, requestSettings)
			}),
	}
}

// ServeHTTP implements the [http.Handler] interface for the [Router] type
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// Use registers a middleware handler
func (r *Router) Use(middleware func(next http.Handler) http.Handler) {
	r.middleware = r.middleware.Append(middleware)
}

// Extend extends the middleware chain with another chain
func (r *Router) Extend(middleware alice.Chain) {
	r.middleware = r.middleware.Extend(middleware)
}

// POST registers a handler for POST requests to the given pattern
func (r *Router) POST(path string, handler http.Handler) {
	r.router.Handler(http.MethodPost, path, r.middleware.Then(handler))
}

// GET registers a handler for GET requests to the given pattern
func (r *Router) GET(path string, handler http.Handler) {
	r.router.Handler(http.MethodGet, path, r.middleware.Then(handler))
}

// PUT registers a handler for PUT requests to the given pattern
func (r *Router) PUT(path string, handler http.Handler) {
	r.router.Handler(http.MethodPut, path, r.middleware.Then(handler))
}

// DELETE registers a handler for DELETE requests to the given pattern
func (r *Router) DELETE(path string, handler http.Handler) {
	r.router.Handler(http.MethodDelete, path, r.middleware.Then(handler))
}

// PATCH registers a handler for PATCH requests to the given pattern
func (r *Router) PATCH(path string, handler http.Handler) {
	r.router.Handler(http.MethodPatch, path, r.middleware.Then(handler))
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (r *Router) OPTIONS(path string, handler http.Handler) {
	r.router.Handler(http.MethodOptions, path, r.middleware.Then(handler))
}

// HEAD registers a handler for HEAD requests to the given pattern
func (r *Router) HEAD(path string, handler http.Handler) {
	r.router.Handler(http.MethodHead, path, r.middleware.Then(handler))
}
