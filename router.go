package simba

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/settings"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// Router is a simple Mux that wraps [http.ServeMux] and allows for middleware chaining
// and type information storage for routes.
type Router struct {
	Mux              *http.ServeMux
	middleware       []func(http.Handler) http.Handler
	openApiReflector openapi31.Reflector
}

// RouteInfo stores type information about a route
type RouteInfo struct {
	Method       string
	Path         string
	BodyType     reflect.Type
	ParamsType   reflect.Type
	AuthType     reflect.Type
	ResponseType reflect.Type
}

// Handler specifies the interface for a handler that can be registered with the [Router].
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	getTypes() (reflect.Type, reflect.Type, reflect.Type)
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
		openApiReflector: openapi31.Reflector{},
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
	var bodyType, paramsType, responseType reflect.Type
	bodyType, paramsType, responseType = getHandlerTypes(handler)
	r.addRoute(method, path, handler)
	r.generateRouteDocumentation(&RouteInfo{
		Method:       method,
		Path:         path,
		BodyType:     bodyType,
		ParamsType:   paramsType,
		ResponseType: responseType,
	}, handler)
}

func getHandlerTypes[H any](handler H) (bodyType, paramsType, responseType reflect.Type) {
	t := reflect.TypeOf(handler)

	// Check if handler implements Handler interface
	if h, ok := any(handler).(Handler); ok {
		return h.getTypes()
	}

	if t.Kind() != reflect.Func {
		panic("handler must be a function")
	}

	// Get *Request[B, P] from second parameter
	reqType := t.In(1).Elem()
	bodyType = reqType.Field(1).Type   // Request body is the second field
	paramsType = reqType.Field(2).Type // Params is the third field

	// Get *Response[R] from first return value
	respType := t.Out(0).Elem()
	responseType = respType.Field(2).Type // Response body is the third field

	slog.Debug("type information",
		"bodyType", bodyType,
		"paramsType", paramsType,
		"responseType", responseType,
	)

	return
}

func (r *Router) addRoute(method, path string, handler http.Handler) {
	r.Mux.Handle(fmt.Sprintf("%s %s", method, path), r.applyMiddleware(handler))
}

func (r *Router) generateRouteDocumentation(routeInfo *RouteInfo, handler any) {
	op, err := r.openApiReflector.NewOperationContext(routeInfo.Method, routeInfo.Path)
	if err != nil {
		panic(fmt.Errorf("failed to create operation context: %w", err))
	}

	// Add request body if it exists
	if routeInfo.BodyType != nil {
		op.AddReqStructure(reflect.New(routeInfo.BodyType).Interface(), func(cu *openapi.ContentUnit) {
			cu.ContentType = mimetypes.ApplicationJSON
			cu.Description = getFunctionComment(handler)
		})
	}

	// Add params separately if they exist
	if routeInfo.ParamsType != nil {
		op.AddReqStructure(reflect.New(routeInfo.ParamsType).Interface())
	}

	// Add response with 200 status code
	op.AddRespStructure(reflect.New(routeInfo.ResponseType).Interface(), func(cu *openapi.ContentUnit) {
		cu.HTTPStatus = http.StatusOK
	})

	err = r.openApiReflector.AddOperation(op)
	if err != nil {
		panic(fmt.Errorf("failed to add operation to openapi reflector: %w", err))
	}
}

func (r *Router) mountOpenApiEndpoint(path string) {
	r.Mux.Handle(fmt.Sprintf("GET %s", path), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		schema, err := r.openApiReflector.Spec.MarshalYAML()
		if err != nil {
			slog.Error("failed to generate API docs", "error", err)
			http.Error(w, "failed to generate API docs", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", mimetypes.ApplicationYAML)
		_, _ = w.Write(schema)
	}))
}
