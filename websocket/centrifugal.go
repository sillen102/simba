package websocket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/centrifugal/centrifuge"
)

// CentrifugalSetupFunc allows callers to attach Centrifuge callbacks and
// additional node configuration before the node starts accepting connections.
type CentrifugalSetupFunc func(node *centrifuge.Node) error

// CentrifugalConfig configures the embedded Centrifuge node and websocket handler.
type CentrifugalConfig struct {
	Node      centrifuge.Config
	Websocket centrifuge.WebsocketConfig
	Setup     CentrifugalSetupFunc
}

// Centrifugal wraps a Centrifuge node together with its websocket HTTP handler.
// It implements http.Handler so it can be mounted directly in Simba via Router.HandleHTTP.
type Centrifugal struct {
	node    *centrifuge.Node
	handler http.Handler
}

// NewCentrifugal creates, configures, and starts a Centrifuge node backed by the
// in-memory broker/presence implementations, then exposes the websocket HTTP handler.
func NewCentrifugal(config CentrifugalConfig) (*Centrifugal, error) {
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

	return &Centrifugal{
		node:    node,
		handler: centrifuge.NewWebsocketHandler(node, config.Websocket),
	}, nil
}

// Node returns the underlying Centrifuge node so callers can publish,
// inspect presence, or attach additional handlers.
func (c *Centrifugal) Node() *centrifuge.Node {
	return c.node
}

// Handler returns the underlying HTTP handler used for websocket handshakes.
func (c *Centrifugal) Handler() http.Handler {
	return c.handler
}

// ServeHTTP proxies the request to the underlying Centrifuge websocket handler.
func (c *Centrifugal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.handler.ServeHTTP(w, r)
}

// Shutdown gracefully stops the underlying Centrifuge node.
func (c *Centrifugal) Shutdown(ctx context.Context) error {
	if c == nil || c.node == nil {
		return nil
	}
	return c.node.Shutdown(ctx)
}
