package simba

import (
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

// Router holds the router and options
type Router[User any] struct {
	router   *httprouter.Router
	options  *RouterOptions
	authFunc AuthFunc[User]
}

// NewRouter returns a new Router
func NewRouter(opts ...RouterOptions) *Router[struct{}] {
	options := defaultRouterOptions()

	if len(opts) > 0 {
		options = mergeOptionsWithReflection(options, opts[0])
	}

	return &Router[struct{}]{
		router:  httprouter.New(),
		options: &options,
	}
}

// NewRouterWithAuth returns a new Router with ability to have authenticated routes using the provided authFunc to
// authenticate and retrieve the user
func NewRouterWithAuth[User any](authFunc AuthFunc[User], opts ...RouterOptions) *Router[User] {
	options := defaultRouterOptions()

	if len(opts) > 0 {
		options = mergeOptionsWithReflection(options, opts[0])
	}

	return &Router[User]{
		router:   httprouter.New(),
		options:  &options,
		authFunc: authFunc,
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

// AuthFunc is a function type for authenticating and retrieving a user from a request
type AuthFunc[User any] func(r *http.Request) (*User, error)

// ServeHTTP implements the http.Handler interface
func (s *Router[User]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// POST registers a handler for POST requests to the given pattern
func (s *Router[User]) POST(path string, handler http.Handler) {
	s.router.Handler(http.MethodPost, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// GET registers a handler for GET requests to the given pattern
func (s *Router[User]) GET(path string, handler http.Handler) {
	s.router.Handler(http.MethodGet, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// PUT registers a handler for PUT requests to the given pattern
func (s *Router[User]) PUT(path string, handler http.Handler) {
	s.router.Handler(http.MethodPut, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// DELETE registers a handler for DELETE requests to the given pattern
func (s *Router[User]) DELETE(path string, handler http.Handler) {
	s.router.Handler(http.MethodDelete, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// PATCH registers a handler for PATCH requests to the given pattern
func (s *Router[User]) PATCH(path string, handler http.Handler) {
	s.router.Handler(http.MethodPatch, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (s *Router[User]) OPTIONS(path string, handler http.Handler) {
	s.router.Handler(http.MethodOptions, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// HEAD registers a handler for HEAD requests to the given pattern
func (s *Router[User]) HEAD(path string, handler http.Handler) {
	s.router.Handler(http.MethodHead, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// CONNECT registers a handler for CONNECT requests to the given pattern
func (s *Router[User]) CONNECT(path string, handler http.Handler) {
	s.router.Handler(http.MethodConnect, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
}

// TRACE registers a handler for TRACE requests to the given pattern
func (s *Router[User]) TRACE(path string, handler http.Handler) {
	s.router.Handler(http.MethodTrace, path, injectAuthFunc(injectOptions(handler, s.options), s.authFunc))
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
