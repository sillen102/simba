package simba

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sillen102/simba/mimetypes"
	"github.com/sillen102/simba/settings"
	"github.com/sillen102/simba/simbaOpenapi"
	"github.com/sillen102/simba/simbaOpenapi/openapiModels"
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
	getHandler() any
	getAuthModel() any
	getAuthHandler() any
}

// Router is a simple Mux that wraps [http.ServeMux] and allows for middleware chaining
// and type information storage for routes.
type Router struct {
	Mux                  *http.ServeMux
	middleware           []func(http.Handler) http.Handler
	docsSettings         settings.Docs
	routes               []openapiModels.RouteInfo
	openApiReflector     *openapi31.Reflector
	schema               []byte
	docsEndpointsMounted bool
}

// newRouter creates a new [Router] instance with the given logger (that is injected in each Request context) and [Simba]
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
		openApiReflector: func() *openapi31.Reflector {
			if docsSettings.GenerateOpenAPIDocs {
				reflector, err := simbaOpenapi.GetReflector()
				if err != nil {
					slog.Error("failed to create OpenAPI reflector", "error", err)
					return nil
				}
				return reflector
			}
			return nil
		}(),
		schema:               nil,
		docsEndpointsMounted: false,
	}
}

// ServeHTTP implements the [http.Handler] interface for the [Router] type
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.docsSettings.GenerateOpenAPIDocs && r.docsSettings.MountDocsEndpoint {
		r.mountDocs(r.docsSettings.OpenAPIFilePath)
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

func (r *Router) addRoute(method, path string, handler http.Handler) {
	r.Mux.Handle(fmt.Sprintf("%s %s", method, path), r.applyMiddleware(handler))
}

func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}

func (r *Router) mountDocs(path string) {
	if r.docsEndpointsMounted || !r.docsSettings.GenerateOpenAPIDocs {
		return
	}

	r.Mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, path), r.openAPIDocsHandler())

	if r.docsSettings.MountDocsEndpoint {
		r.Mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, r.docsSettings.DocsPath), simbaOpenapi.ScalarDocsHandler(simbaOpenapi.DocsParams{
			OpenAPIFileType: r.docsSettings.OpenAPIFileType,
			OpenAPIPath:     r.docsSettings.OpenAPIFilePath,
			DocsPath:        r.docsSettings.DocsPath,
			ServiceName:     r.docsSettings.ServiceName,
		}))
	}

	r.docsEndpointsMounted = true
}

func (r *Router) openAPIDocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if r.schema == nil {

			for _, route := range r.routes {
				simbaOpenapi.GenerateRouteDocumentation(r.openApiReflector, &route, route.Handler)
			}

			var err error
			if r.docsSettings.OpenAPIFileType == mimetypes.ApplicationJSON {
				r.schema, err = r.openApiReflector.Spec.MarshalJSON()
			} else {
				r.schema, err = r.openApiReflector.Spec.MarshalYAML()
			}
			if err != nil {
				errMessage := "failed to generate API docs"
				slog.Error(errMessage, "error", err)
				http.Error(w, errMessage, http.StatusInternalServerError)
				return
			}

			// Clean up routes and reflector to free up memory
			r.routes = nil
			r.openApiReflector = nil
		}

		w.Header().Set("Content-Type", mimetypes.ApplicationYAML)
		_, _ = w.Write(r.schema)
	}
}
