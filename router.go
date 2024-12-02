package simba

import (
	"context"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

// Router is a HTTP router
type Router struct {
	router  *httprouter.Router
	options *RouterOptions
}

// NewRouter returns a new Router
func NewRouter(opts ...RouterOptions) *Router {
	options := defaultRouterOptions()

	if len(opts) > 0 {
		options = mergeOptionsWithReflection(options, opts[0])
	}

	return &Router{
		router:  httprouter.New(),
		options: &options,
	}
}

// RouterOptions are options for the router
type RouterOptions struct {
	// RequestDisallowUnknownFields will disallow unknown fields in the request body,
	// resulting in a 400 Bad Request response if a field is present that cannot be
	// mapped to the model struct.
	RequestDisallowUnknownFields *bool
}

// defaultRouterOptions creates a RouterOptions with default values
func defaultRouterOptions() RouterOptions {
	defaultTrue := true
	return RouterOptions{
		RequestDisallowUnknownFields: &defaultTrue,
	}
}

type AuthValidator[User any] func(ctx context.Context, accessToken string) (*User, error)

// ServeHTTP implements the http.Handler interface
func (s *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// POST registers a handler for POST requests to the given pattern
func (s *Router) POST(path string, handler http.Handler) {
	s.router.Handler(http.MethodPost, path, injectConfiguration(handler, s.options))
}

// GET registers a handler for GET requests to the given pattern
func (s *Router) GET(path string, handler http.Handler) {
	s.router.Handler(http.MethodGet, path, injectConfiguration(handler, s.options))
}

// PUT registers a handler for PUT requests to the given pattern
func (s *Router) PUT(path string, handler http.Handler) {
	s.router.Handler(http.MethodPut, path, injectConfiguration(handler, s.options))
}

// DELETE registers a handler for DELETE requests to the given pattern
func (s *Router) DELETE(path string, handler http.Handler) {
	s.router.Handler(http.MethodDelete, path, injectConfiguration(handler, s.options))
}

// PATCH registers a handler for PATCH requests to the given pattern
func (s *Router) PATCH(path string, handler http.Handler) {
	s.router.Handler(http.MethodPatch, path, injectConfiguration(handler, s.options))
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (s *Router) OPTIONS(path string, handler http.Handler) {
	s.router.Handler(http.MethodOptions, path, injectConfiguration(handler, s.options))
}

// HEAD registers a handler for HEAD requests to the given pattern
func (s *Router) HEAD(path string, handler http.Handler) {
	s.router.Handler(http.MethodHead, path, injectConfiguration(handler, s.options))
}

// CONNECT registers a handler for CONNECT requests to the given pattern
func (s *Router) CONNECT(path string, handler http.Handler) {
	s.router.Handler(http.MethodConnect, path, injectConfiguration(handler, s.options))
}

// TRACE registers a handler for TRACE requests to the given pattern
func (s *Router) TRACE(path string, handler http.Handler) {
	s.router.Handler(http.MethodTrace, path, injectConfiguration(handler, s.options))
}

// mergeOptionsWithReflection uses reflection to merge non-zero fields from provided into default options
func mergeOptionsWithReflection(defaultOpts, providedOpts RouterOptions) RouterOptions {
	defaultVal := reflect.ValueOf(&defaultOpts).Elem()
	providedVal := reflect.ValueOf(providedOpts)

	for i := 0; i < providedVal.NumField(); i++ {
		providedField := providedVal.Field(i)
		defaultField := defaultVal.Field(i)

		defaultField.Set(providedField)
	}
	return defaultOpts
}
