package simba

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/settings"
	"github.com/swaggest/openapi-go/openapi31"
)

// Handler specifies the interface for a handler that can be registered with the [Router].
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	getRequestBody() any
	getParams() any
	getResponseBody() any
	getAccepts() string
	getProduces() string
	getAuthModel() any
	getAuthFunc() any
}

// Router is a simple Mux that wraps [http.ServeMux] and allows for middleware chaining
// and type information storage for routes.
type Router struct {
	Mux                 *http.ServeMux
	middleware          []func(http.Handler) http.Handler
	openApiReflector    *openapi31.Reflector
	schema              []byte
	docsEndpointMounted bool
}

// routeInfo stores type information about a route
type routeInfo struct {
	method    string
	path      string
	accepts   string
	produces  string
	reqBody   any
	params    any
	respBody  any
	authModel any
	authFunc  any
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
		openApiReflector:    &openapi31.Reflector{},
		schema:              nil,
		docsEndpointMounted: false,
	}
}

// ServeHTTP implements the [http.Handler] interface for the [Router] type
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mountOpenApiEndpoint("/openapi")
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

// POST registers a handler for POST requests to the given pattern
func (r *Router) POST(path string, handler Handler) {
	method := http.MethodPost
	r.Handle(method, path, handler)
}

// GET registers a handler for GET requests to the given pattern
func (r *Router) GET(path string, handler Handler) {
	method := http.MethodGet
	r.Handle(method, path, handler)
}

// PUT registers a handler for PUT requests to the given pattern
func (r *Router) PUT(path string, handler Handler) {
	method := http.MethodPut
	r.Handle(method, path, handler)
}

// DELETE registers a handler for DELETE requests to the given pattern
func (r *Router) DELETE(path string, handler Handler) {
	method := http.MethodDelete
	r.Handle(method, path, handler)
}

// PATCH registers a handler for PATCH requests to the given pattern
func (r *Router) PATCH(path string, handler Handler) {
	method := http.MethodPatch
	r.Handle(method, path, handler)
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (r *Router) OPTIONS(path string, handler Handler) {
	method := http.MethodOptions
	r.Handle(method, path, handler)
}

// HEAD registers a handler for HEAD requests to the given pattern
func (r *Router) HEAD(path string, handler Handler) {
	method := http.MethodHead
	r.Handle(method, path, handler)
}

func (r *Router) Handle(method, path string, handler Handler) {
	r.addRoute(method, path, handler)
	generateRouteDocumentation(r.openApiReflector, &routeInfo{
		method:    method,
		path:      path,
		accepts:   handler.getAccepts(),
		produces:  handler.getProduces(),
		reqBody:   handler.getRequestBody(),
		params:    handler.getParams(),
		respBody:  handler.getResponseBody(),
		authModel: handler.getAuthModel(),
		authFunc:  handler.getAuthFunc(),
	}, handler)
}

func (r *Router) addRoute(method, path string, handler http.Handler) {
	r.Mux.Handle(fmt.Sprintf("%s %s", method, path), r.applyMiddleware(handler))
}

func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}

func (r *Router) mountOpenApiEndpoint(path string) {
	if !r.docsEndpointMounted {
		r.Mux.Handle(fmt.Sprintf("GET %s", path), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if r.schema == nil {
				var err error
				r.schema, err = r.openApiReflector.Spec.MarshalYAML()
				if err != nil {
					errMessage := "failed to generate API docs"
					slog.Error(errMessage, "error", err)
					http.Error(w, errMessage, http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("Content-Type", mimetypes.ApplicationYAML)
			_, _ = w.Write(r.schema)
		}))
		r.docsEndpointMounted = true
	}
}
