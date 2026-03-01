package websocket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/centrifugal/centrifuge"
)

// SetupFunc allows callers to attach Centrifuge callbacks and
// additional node configuration before the node starts accepting connections.
type SetupFunc func(node *centrifuge.Node) error

// Config configures the embedded Centrifuge node and websocket handler.
type Config struct {
	Node      centrifuge.Config
	Websocket centrifuge.WebsocketConfig
	Setup     SetupFunc
}

// Handler wraps a Centrifuge node together with its websocket HTTP handler.
// It implements http.Handler so it can be mounted directly in Simba via Router.HandleHTTP.
type Handler struct {
	node    *centrifuge.Node
	handler http.Handler
}

// New creates, configures, and starts a Centrifuge node backed by the
// in-memory broker/presence implementations, then exposes the websocket HTTP handler.
func New(config Config) (*Handler, error) {
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

	return &Handler{
		node:    node,
		handler: centrifuge.NewWebsocketHandler(node, config.Websocket),
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
	return nil
}

// GetAuthHandler returns nil because auth is handled inside Centrifuge callbacks.
func (h *Handler) GetAuthHandler() any {
	return nil
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
