package simba

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/sillen102/simba/settings"
)

// Router is a simple Mux that wraps [http.ServeMux] and allows for middleware chaining
// and type information storage for routes.
type Router struct {
	Mux        *http.ServeMux
	middleware []func(http.Handler) http.Handler
	routes     map[string]RouteInfo
}

// RouteInfo stores type information about a route
type RouteInfo struct {
	Method     string
	Path       string
	BodyType   reflect.Type
	ParamsType reflect.Type
	AuthType   reflect.Type
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
		routes: make(map[string]RouteInfo),
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

// POST registers a handler for POST requests to the given pattern
func (r *Router) POST(path string, handler http.Handler) {
	method := http.MethodPost
	r.addRoute(method, path, handler)
}

// GET registers a handler for GET requests to the given pattern
func (r *Router) GET(path string, handler http.Handler) {
	method := http.MethodGet
	r.addRoute(method, path, handler)
}

// PUT registers a handler for PUT requests to the given pattern
func (r *Router) PUT(path string, handler http.Handler) {
	method := http.MethodPut
	r.addRoute(method, path, handler)
}

// DELETE registers a handler for DELETE requests to the given pattern
func (r *Router) DELETE(path string, handler http.Handler) {
	method := http.MethodDelete
	r.addRoute(method, path, handler)
}

// PATCH registers a handler for PATCH requests to the given pattern
func (r *Router) PATCH(path string, handler http.Handler) {
	method := http.MethodPatch
	r.addRoute(method, path, handler)
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (r *Router) OPTIONS(path string, handler http.Handler) {
	method := http.MethodOptions
	r.addRoute(method, path, handler)
}

// HEAD registers a handler for HEAD requests to the given pattern
func (r *Router) HEAD(path string, handler http.Handler) {
	method := http.MethodHead
	r.addRoute(method, path, handler)
}

func (r *Router) addRoute(method, path string, handler http.Handler) {
	r.storeRouteInfo(method, path, handler)
	r.Mux.Handle(fmt.Sprintf("%s %s", method, path), r.applyMiddleware(handler))
}

// storeRouteInfo stores type information for a route
func (r *Router) storeRouteInfo(method, path string, handler any) {
	key := method + " " + path
	info := RouteInfo{
		Method: method,
		Path:   path,
	}

	t := reflect.TypeOf(handler)
	if t.Kind() == reflect.Func && t.NumIn() == 2 {
		reqType := t.In(1)
		if reqType.Kind() == reflect.Ptr {
			reqType = reqType.Elem()
		}
		if reqType.NumField() > 0 {
			info.BodyType = reqType.Field(0).Type
			info.ParamsType = reqType.Field(1).Type
		}
	}

	r.routes[key] = info
}
