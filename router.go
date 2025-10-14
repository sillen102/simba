package simba

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaOpenapi"
	"github.com/sillen102/simba/simbaOpenapi/openapiModels"
)

// Handler specifies the interface for a handler that can be registered with the [Router].
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	getRequestBody() any
	getParams() any
	getResponseBody() any
	getAccepts() string
	getProduces() string
	getHandler() any
	getAuthModel() any
	getAuthHandler() any
}

type openApiGenerator interface {
	GenerateDocumentation(ctx context.Context, title string, version string, routeInfos []openapiModels.RouteInfo) ([]byte, error)
}

// Router is a simple Mux that wraps [http.ServeMux] and allows for middleware chaining
// and type information storage for routes.
type Router struct {
	Mux                    *http.ServeMux
	middleware             []func(http.Handler) http.Handler
	docsSettings           settings.Docs
	routes                 []openapiModels.RouteInfo
	schema                 []byte
	openAPIEndpointMounted bool
	docsEndpointsMounted   bool
	openAPIGenerator       openApiGenerator
}

// GenerateOpenAPIDocumentation generates the OpenAPI documentation for the routes mounted in the router
// if enabled in [settings.Docs]
func (r *Router) GenerateOpenAPIDocumentation(ctx context.Context, title, version string) error {
	if r.docsSettings.GenerateOpenAPIDocs {
		var err error
		r.schema, err = r.openAPIGenerator.GenerateDocumentation(ctx, title, version, r.routes)
		if err != nil {
			return fmt.Errorf("failed to generate OpenAPI documentation: %w", err)
		}

		// Clear routes and generator reference after successful generation to free up memory
		r.routes = nil
		r.openAPIGenerator = nil
	}

	return nil
}

func newRouter(requestSettings settings.Request, docsSettings settings.Docs) *Router {
	return &Router{
		Mux: http.NewServeMux(),
		middleware: []func(http.Handler) http.Handler{
			closeRequestBody,
			func(next http.Handler) http.Handler {
				return injectRequestSettings(next, &requestSettings)
			},
		},
		docsSettings: docsSettings,
		routes: func() []openapiModels.RouteInfo {
			if docsSettings.GenerateOpenAPIDocs {
				return make([]openapiModels.RouteInfo, 0, 100)
			}
			return nil
		}(),
		schema:                 nil,
		openAPIEndpointMounted: false,
		docsEndpointsMounted:   false,
		openAPIGenerator:       simbaOpenapi.NewOpenAPIGenerator(),
	}
}

// ServeHTTP implements the [http.Handler] interface for the [Router] type
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.docsSettings.GenerateOpenAPIDocs {
		r.mountOpenAPIEndpoint()
	}

	if r.docsSettings.MountDocsUIEndpoint {
		r.mountDocsUIEndpoint()
	}

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

// POSTWithMiddleware registers a handler for POST requests to the given pattern wrapped with a middleware function
func (r *Router) POSTWithMiddleware(path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	method := http.MethodPost
	r.WithMiddleware(method, path, handler, middleware...)
}

// GET registers a handler for GET requests to the given pattern
func (r *Router) GET(path string, handler Handler) {
	method := http.MethodGet
	r.Handle(method, path, handler)
}

// GETWithMiddleware registers a handler for GET requests to the given pattern wrapped with a middleware function
func (r *Router) GETWithMiddleware(path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	method := http.MethodGet
	r.WithMiddleware(method, path, handler, middleware...)
}

// PUT registers a handler for PUT requests to the given pattern
func (r *Router) PUT(path string, handler Handler) {
	method := http.MethodPut
	r.Handle(method, path, handler)
}

// PUTWithMiddleware registers a handler for PUT requests to the given pattern wrapped with a middleware function
func (r *Router) PUTWithMiddleware(path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	method := http.MethodPut
	r.WithMiddleware(method, path, handler, middleware...)
}

// DELETE registers a handler for DELETE requests to the given pattern
func (r *Router) DELETE(path string, handler Handler) {
	method := http.MethodDelete
	r.Handle(method, path, handler)
}

// DELETEWithMiddleware registers a handler for DELETE requests to the given pattern wrapped with a middleware function
func (r *Router) DELETEWithMiddleware(path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	method := http.MethodDelete
	r.WithMiddleware(method, path, handler, middleware...)
}

// PATCH registers a handler for PATCH requests to the given pattern
func (r *Router) PATCH(path string, handler Handler) {
	method := http.MethodPatch
	r.Handle(method, path, handler)
}

// PATCHWithMiddleware registers a handler for PATCH requests to the given pattern wrapped with a middleware function
func (r *Router) PATCHWithMiddleware(path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	method := http.MethodPatch
	r.WithMiddleware(method, path, handler, middleware...)
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (r *Router) OPTIONS(path string, handler Handler) {
	method := http.MethodOptions
	r.Handle(method, path, handler)
}

// OPTIONSWithMiddleware registers a handler for OPTIONS requests to the given pattern wrapped with a middleware function
func (r *Router) OPTIONSWithMiddleware(path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	method := http.MethodOptions
	r.WithMiddleware(method, path, handler, middleware...)
}

// HEAD registers a handler for HEAD requests to the given pattern
func (r *Router) HEAD(path string, handler Handler) {
	method := http.MethodHead
	r.Handle(method, path, handler)
}

// HEADWithMiddleware registers a handler for HEAD requests to the given pattern wrapped with a middleware function
func (r *Router) HEADWithMiddleware(path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	method := http.MethodHead
	r.WithMiddleware(method, path, handler, middleware...)
}

// WithMiddleware registers a handler for the given method and pattern wrapped with a middleware function
func (r *Router) WithMiddleware(method, path string, handler Handler, middleware ...func(http.Handler) http.Handler) {
	h := handlerToHTTPHandler(handler)
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	r.addRoute(method, path, h)
	r.addRouteToDocs(method, path, handler)
}

func handlerToHTTPHandler(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// Handle registers a handler for the given method and pattern
func (r *Router) Handle(method, path string, handler Handler) {
	r.addRoute(method, path, handler)
	r.addRouteToDocs(method, path, handler)
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

func (r *Router) addRouteToDocs(method string, path string, handler Handler) {
	if r.docsSettings.GenerateOpenAPIDocs {
		r.routes = append(r.routes, openapiModels.RouteInfo{
			Method:      method,
			Path:        path,
			Accepts:     handler.getAccepts(),
			Produces:    handler.getProduces(),
			ReqBody:     handler.getRequestBody(),
			Params:      handler.getParams(),
			RespBody:    handler.getResponseBody(),
			Handler:     handler.getHandler(),
			AuthModel:   handler.getAuthModel(),
			AuthHandler: handler.getAuthHandler(),
		})
	}
}

func (r *Router) mountDocsUIEndpoint() {
	if r.docsEndpointsMounted {
		return
	}

	if r.docsSettings.MountDocsUIEndpoint {
		r.Mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, r.docsSettings.DocsUIPath), simbaOpenapi.ScalarDocsHandler(simbaOpenapi.DocsParams{
			OpenAPIPath: r.docsSettings.OpenAPIFilePath,
			DocsPath:    r.docsSettings.DocsUIPath,
			ServiceName: r.docsSettings.ServiceName,
		}))
	}

	r.docsEndpointsMounted = true
}

func (r *Router) mountOpenAPIEndpoint() {
	if r.openAPIEndpointMounted {
		return
	}

	r.Mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, r.docsSettings.OpenAPIFilePath), r.openAPIDocsHandler())

	r.openAPIEndpointMounted = true
}

func (r *Router) openAPIDocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", mimetypes.ApplicationJSON)
		_, _ = w.Write(r.schema)
	}
}
