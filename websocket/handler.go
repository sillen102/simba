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

// AuthenticatedConnectingHandler is a Centrifuge connecting handler that also
// receives the authenticated Simba auth model.
type AuthenticatedConnectingHandler[AuthModel any] func(
	ctx context.Context,
	event centrifuge.ConnectEvent,
	authModel *AuthModel,
) (centrifuge.ConnectReply, error)

// OnConnecting is a typed identity helper to make config literals read naturally.
func OnConnecting[H any](handler H) H {
	return handler
}

// OnConnect is a typed identity helper to make config literals read naturally.
func OnConnect(handler centrifuge.ConnectHandler) centrifuge.ConnectHandler {
	return handler
}

// Config configures the embedded Centrifuge node and websocket handler.
type Config struct {
	Node         centrifuge.Config
	Websocket    centrifuge.WebsocketConfig
	OnConnecting centrifuge.ConnectingHandler
	OnConnect    centrifuge.ConnectHandler
	Setup        SetupFunc
}

// Handler creates a websocket handler and panics if the configuration is invalid.
func (config Config) Handler() *Handler {
	handler, err := New(config)
	if err != nil {
		panic(err)
	}
	return handler
}

// AuthenticatedConfig configures an authenticated websocket handler.
type AuthenticatedConfig[AuthModel any] struct {
	Node         centrifuge.Config
	Websocket    centrifuge.WebsocketConfig
	OnConnecting AuthenticatedConnectingHandler[AuthModel]
	OnConnect    centrifuge.ConnectHandler
	Setup        SetupFunc
}

// Handler creates an authenticated websocket handler and panics if the configuration is invalid.
func (config AuthenticatedConfig[AuthModel]) Handler() *AuthenticatedHandler[AuthModel] {
	handler, err := NewAuthenticatedHandler(config)
	if err != nil {
		panic(err)
	}
	return handler
}

// Handler wraps a Centrifuge node together with its websocket HTTP handler.
// It implements http.Handler so it can be mounted directly in Simba via Router.HandleHTTP.
type Handler struct {
	node    *centrifuge.Node
	handler http.Handler

	authModel   any
	authHandler any
}

// AuthenticatedHandler is a websocket handler that carries the auth model type
// so AuthHandler can infer it without an explicit type argument.
type AuthenticatedHandler[AuthModel any] struct {
	*Handler
}

// AuthenticatedHandlerBuilder allows route registration to accept a constructor
// function and keep websocket setup inline, similar to Simba's other auth helpers.
type AuthenticatedHandlerBuilder[AuthModel any] func() *AuthenticatedHandler[AuthModel]

// New creates, configures, and starts a Centrifuge node backed by the
// in-memory broker/presence implementations, then exposes the websocket HTTP handler.
func New(config Config) (*Handler, error) {
	return newHandler(config, nil, nil, nil)
}

// NewAuthenticatedHandler creates a websocket handler whose OnConnecting callback
// expects an authenticated Simba auth model to be available in context.
// Apply handshake authentication by wrapping the handler with NewAuthenticated(...)
// before mounting it in a Simba route.
func NewAuthenticatedHandler[AuthModel any](config AuthenticatedConfig[AuthModel]) (*AuthenticatedHandler[AuthModel], error) {
	var authModel AuthModel

	baseConfig := Config{
		Node:      config.Node,
		Websocket: config.Websocket,
		OnConnect: config.OnConnect,
		Setup:     config.Setup,
	}

	if config.OnConnecting != nil {
		baseConfig.OnConnecting = func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
			model, ok := AuthModelFromContext[AuthModel](ctx)
			if !ok {
				return centrifuge.ConnectReply{}, centrifuge.ErrorUnauthorized
			}
			return config.OnConnecting(ctx, event, &model)
		}
	}

	handler, err := newHandler(baseConfig, authModel, nil, nil)
	if err != nil {
		return nil, err
	}

	return &AuthenticatedHandler[AuthModel]{Handler: handler}, nil
}

// NewAuthenticated wraps a websocket handler with a Simba auth handler so that
// the HTTP upgrade request is authenticated before the WebSocket handshake.
func NewAuthenticated[AuthModel any](handler *AuthenticatedHandler[AuthModel], auth any) *Handler {
	var authModel AuthModel
	authFn, ok := getAuthHandlerFunc[AuthModel](auth)

	wrapped := *handler.Handler
	wrapped.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ok {
			http.Error(w, "invalid websocket auth handler", http.StatusInternalServerError)
			return
		}

		model, err := authFn(r)
		if err != nil {
			writeAuthError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), authModelContextKey{}, model)
		handler.Handler.handler.ServeHTTP(w, r.WithContext(ctx))
	})
	wrapped.authModel = authModel
	wrapped.authHandler = auth

	return &wrapped
}

// AuthHandler wraps an authenticated websocket handler factory with handshake auth.
func AuthHandler[AuthModel any](handler AuthenticatedHandlerBuilder[AuthModel], auth any) *Handler {
	return NewAuthenticated(handler(), auth)
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

	if config.OnConnecting != nil {
		node.OnConnecting(config.OnConnecting)
	}
	if config.OnConnect != nil {
		node.OnConnect(config.OnConnect)
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

// Publish sends data into a Centrifuge channel using the underlying node.
func (h *Handler) Publish(channel string, data []byte, opts ...centrifuge.PublishOption) (centrifuge.PublishResult, error) {
	if h == nil || h.node == nil {
		return centrifuge.PublishResult{}, errors.New("websocket handler is not initialized")
	}
	return h.node.Publish(channel, data, opts...)
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

func getAuthHandlerFunc[AuthModel any](authHandler any) (func(*http.Request) (AuthModel, error), bool) {
	method := reflect.ValueOf(authHandler).MethodByName("GetHandler")
	if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return nil, false
	}

	handlerFunc := method.Call(nil)[0]
	if handlerFunc.Kind() != reflect.Func {
		return nil, false
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
	}, true
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
