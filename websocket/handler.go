package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/centrifugal/centrifuge"
)

type authModelContextKey struct{}

// SetupFunc allows callers to attach Centrifuge callbacks and
// additional node configuration before the node starts accepting connections.
type SetupFunc func(node *centrifuge.Node) error

// Config configures the embedded Centrifuge node and websocket handler.
type Config struct {
	Node      centrifuge.Config
	Websocket centrifuge.WebsocketConfig
	Setup     SetupFunc
}

// AuthenticatedConfig configures an authenticated websocket handler.
// Auth must be a Simba auth handler (for example simba.BearerAuth(...)).
type AuthenticatedConfig[AuthModel any] struct {
	Config
	Auth any
}

// Handler wraps a Centrifuge node together with its websocket HTTP handler.
// It implements http.Handler so it can be mounted directly in Simba via Router.HandleHTTP.
type Handler struct {
	node    *centrifuge.Node
	handler http.Handler

	authModel   any
	authHandler any
}

// New creates, configures, and starts a Centrifuge node backed by the
// in-memory broker/presence implementations, then exposes the websocket HTTP handler.
func New(config Config) (*Handler, error) {
	return newHandler(config, nil, nil, nil)
}

// NewAuthenticated creates a websocket handler that authenticates the HTTP handshake
// using a Simba auth handler before the connection is upgraded.
func NewAuthenticated[AuthModel any](config AuthenticatedConfig[AuthModel]) (*Handler, error) {
	if config.Auth == nil {
		return nil, errors.New("auth handler is required")
	}

	authFn, err := getAuthHandlerFunc[AuthModel](config.Auth)
	if err != nil {
		return nil, err
	}

	var authModel AuthModel

	return newHandler(config.Config, authModel, config.Auth, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			model, err := authFn(r)
			if err != nil {
				writeAuthError(w, r, err)
				return
			}

			ctx := context.WithValue(r.Context(), authModelContextKey{}, model)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
}

func newHandler(
	config Config,
	authModel any,
	authHandler any,
	wrapper func(http.Handler) http.Handler,
) (*Handler, error) {
	node, err := centrifuge.New(config.Node)
	if err != nil {
		return nil, fmt.Errorf("create centrifuge node: %w", err)
	}

	if config.Setup != nil {
		if err = config.Setup(node); err != nil {
			return nil, fmt.Errorf("setup centrifuge node: %w", err)
		}
	}

	if err = node.Run(); err != nil {
		return nil, fmt.Errorf("run centrifuge node: %w", err)
	}

	httpHandler := http.Handler(centrifuge.NewWebsocketHandler(node, config.Websocket))
	if wrapper != nil {
		httpHandler = wrapper(httpHandler)
	}

	return &Handler{
		node:        node,
		handler:     httpHandler,
		authModel:   authModel,
		authHandler: authHandler,
	}, nil
}

// Node returns the underlying Centrifuge node so callers can publish,
// inspect presence, or attach additional handlers.
func (h *Handler) Node() *centrifuge.Node {
	return h.node
}

// HTTPHandler returns the underlying HTTP handler used for websocket handshakes.
func (h *Handler) HTTPHandler() http.Handler {
	return h.handler
}

// ServeHTTP proxies the request to the underlying Centrifuge websocket handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

// GetRequestBody returns nil because WebSocket upgrade routes do not use Simba request decoding.
func (h *Handler) GetRequestBody() any {
	return nil
}

// GetParams returns nil because WebSocket upgrade routes currently do not expose typed Simba params.
func (h *Handler) GetParams() any {
	return nil
}

// GetResponseBody returns nil because WebSocket routes do not produce a typed HTTP response body.
func (h *Handler) GetResponseBody() any {
	return nil
}

// GetAccepts returns an empty content type because WebSocket upgrade routes are not regular REST handlers.
func (h *Handler) GetAccepts() string {
	return ""
}

// GetProduces returns an empty content type because WebSocket upgrade routes are not regular REST handlers.
func (h *Handler) GetProduces() string {
	return ""
}

// GetHandler returns the underlying HTTP handler.
func (h *Handler) GetHandler() any {
	return h.handler
}

// GetAuthModel returns nil because auth is handled inside Centrifuge callbacks.
func (h *Handler) GetAuthModel() any {
	return h.authModel
}

// GetAuthHandler returns the auth handler used for handshake authentication.
func (h *Handler) GetAuthHandler() any {
	return h.authHandler
}

// ShouldDocument disables OpenAPI generation for WebSocket upgrade routes.
func (h *Handler) ShouldDocument() bool {
	return false
}

// Shutdown gracefully stops the underlying Centrifuge node.
func (h *Handler) Shutdown(ctx context.Context) error {
	if h == nil || h.node == nil {
		return nil
	}
	return h.node.Shutdown(ctx)
}

// AuthModelFromContext returns the authenticated Simba auth model stored on the
// websocket connection context when using NewAuthenticated.
func AuthModelFromContext[AuthModel any](ctx context.Context) (AuthModel, bool) {
	authModel, ok := ctx.Value(authModelContextKey{}).(AuthModel)
	return authModel, ok
}

func getAuthHandlerFunc[AuthModel any](authHandler any) (func(*http.Request) (AuthModel, error), error) {
	method := reflect.ValueOf(authHandler).MethodByName("GetHandler")
	if !method.IsValid() {
		return nil, errors.New("auth handler must expose GetHandler()")
	}

	if method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return nil, errors.New("auth handler GetHandler() has an unsupported signature")
	}

	handlerFunc := method.Call(nil)[0]
	if handlerFunc.Kind() != reflect.Func {
		return nil, errors.New("auth handler GetHandler() must return a function")
	}

	authModelType := reflect.TypeFor[AuthModel]()

	return func(r *http.Request) (AuthModel, error) {
		var zero AuthModel

		if handlerFunc.Type().NumIn() != 1 || handlerFunc.Type().NumOut() != 2 {
			return zero, errors.New("auth handler function has an unsupported signature")
		}

		out := handlerFunc.Call([]reflect.Value{reflect.ValueOf(r)})

		if !out[0].Type().AssignableTo(authModelType) {
			return zero, fmt.Errorf("auth handler returned %s, expected %s", out[0].Type(), authModelType)
		}

		model := out[0].Interface().(AuthModel)

		if !out[1].IsNil() {
			err, ok := out[1].Interface().(error)
			if !ok {
				return zero, errors.New("auth handler returned a non-error value")
			}
			return zero, err
		}

		return model, nil
	}, nil
}

func writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusUnauthorized
	message := http.StatusText(statusCode)
	var details any

	type statusCodeProvider interface {
		StatusCode() int
	}
	type publicMessageProvider interface {
		PublicMessage() string
	}
	type detailProvider interface {
		Details() any
	}

	if provider, ok := err.(statusCodeProvider); ok {
		statusCode = provider.StatusCode()
	}
	if provider, ok := err.(publicMessageProvider); ok {
		message = provider.PublicMessage()
	}
	if provider, ok := err.(detailProvider); ok {
		details = provider.Details()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"timestamp": time.Now().UTC(),
		"status":    statusCode,
		"error":     http.StatusText(statusCode),
		"path":      r.URL.Path,
		"method":    r.Method,
		"message":   message,
		"details":   details,
	})
}
