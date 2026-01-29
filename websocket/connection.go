package websocket

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/gobwas/ws/wsutil"
)

// WebSocketConnection represents an active WebSocket connection.
// It provides thread-safe methods for sending messages.
// The ID can be used to reference this connection in external systems
// (e.g., Redis, database) for multi-instance message routing.
type Connection struct {
	// ID is a unique identifier (UUID) for this connection.
	// Use this to track connections in external registries.
	ID string

	conn net.Conn
	mu   sync.Mutex
}

// WriteText sends a text message to the client (thread-safe).
func (c *Connection) WriteText(msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return wsutil.WriteServerText(c.conn, []byte(msg))
}

// WriteBinary sends a binary message to the client (thread-safe).
func (c *Connection) WriteBinary(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return wsutil.WriteServerBinary(c.conn, data)
}

// WriteJSON marshals v to JSON and sends it as a text message (thread-safe).
func (c *Connection) WriteJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return wsutil.WriteServerText(c.conn, data)
}

// Close closes the WebSocket connection.
func (c *Connection) Close() error {
	return c.conn.Close()
}
