package simba

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/gobwas/ws/wsutil"
)

// WebSocketConnection wraps a WebSocket connection with thread-safe write operations.
type WebSocketConnection struct {
	// ID is a unique identifier for this connection
	ID string

	// Params contains the parsed route/query/header parameters
	Params any

	conn   net.Conn
	ctx    context.Context
	cancel context.CancelFunc

	// writeMu protects concurrent writes to the connection
	writeMu sync.Mutex
}

// WriteText sends a text message to the client (thread-safe)
func (wc *WebSocketConnection) WriteText(msg string) error {
	wc.writeMu.Lock()
	defer wc.writeMu.Unlock()
	return wsutil.WriteServerText(wc.conn, []byte(msg))
}

// WriteBinary sends a binary message to the client (thread-safe)
func (wc *WebSocketConnection) WriteBinary(data []byte) error {
	wc.writeMu.Lock()
	defer wc.writeMu.Unlock()
	return wsutil.WriteServerBinary(wc.conn, data)
}

// WriteJSON marshals v to JSON and sends it as a text message (thread-safe)
func (wc *WebSocketConnection) WriteJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	wc.writeMu.Lock()
	defer wc.writeMu.Unlock()
	return wsutil.WriteServerText(wc.conn, data)
}

// Close closes the WebSocket connection (thread-safe)
func (wc *WebSocketConnection) Close() error {
	wc.cancel() // Cancel context to signal shutdown
	return wc.conn.Close()
}

// Context returns the connection's context
// The context is cancelled when the connection is closed
func (wc *WebSocketConnection) Context() context.Context {
	return wc.ctx
}
