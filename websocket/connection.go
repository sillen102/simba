package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

// WebSocketConnection represents an active WebSocket connection.
// It provides thread-safe methods for sending messages.
// The ID can be used to reference this connection in external systems
// (e.g., Redis, database) for multi-instance message routing.
type Connection struct {
	// ID is a unique identifier (UUID) for this connection.
	// Use this to track connections in external registries.
	ID string

	conn *websocket.Conn
}

// WriteText sends a text message to the client (thread-safe).
func (c *Connection) WriteText(ctx context.Context, msg string) error {
	return c.conn.Write(ctx, websocket.MessageText, []byte(msg))
}

// WriteBinary sends a binary message to the client (thread-safe).
func (c *Connection) WriteBinary(ctx context.Context, data []byte) error {
	return c.conn.Write(ctx, websocket.MessageBinary, data)
}

// WriteJSON marshals v to JSON and sends it as a text message (thread-safe).
func (c *Connection) WriteJSON(ctx context.Context, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return c.conn.Write(ctx, websocket.MessageText, data)
}

// Close closes the WebSocket connection.
func (c *Connection) Close() error {
	return c.conn.CloseNow()
}
