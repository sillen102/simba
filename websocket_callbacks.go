package simba

import (
	"context"

	"github.com/gobwas/ws"
)

// WebSocketCallbacks defines the lifecycle callbacks for a WebSocket connection
// All callbacks are called sequentially for a single connection
// Multiple connections can have their callbacks called concurrently
type WebSocketCallbacks[Params any] struct {
	// OnConnect is called after the WebSocket upgrade succeeds (optional)
	// Use this to send welcome messages, initialize state, join groups, etc.
	// If an error is returned, the connection is closed and OnDisconnect is called
	OnConnect func(ctx context.Context, conn *WebSocketConnection, registry ConnectionRegistry, params Params) error

	// OnMessage is called for each incoming message from the client (required)
	// The messageType indicates if this is text (ws.OpText) or binary (ws.OpBinary)
	// If an error is returned, OnError is called (if provided), otherwise connection closes
	OnMessage func(ctx context.Context, conn *WebSocketConnection, registry ConnectionRegistry, messageType ws.OpCode, data []byte) error

	// OnDisconnect is called when the connection is closed (optional)
	// This is ALWAYS called, even if OnConnect or OnMessage returns an error
	// The err parameter contains the error that caused disconnection (nil for clean close)
	// This is guaranteed to run via defer, making it perfect for cleanup
	OnDisconnect func(ctx context.Context, params Params, err error)

	// OnError is called when an error occurs during OnConnect or OnMessage (optional)
	// Return true to continue processing messages, false to close the connection
	// If not provided, any error will close the connection
	OnError func(ctx context.Context, conn *WebSocketConnection, err error) bool
}

// AuthWebSocketCallbacks defines the lifecycle callbacks for an authenticated WebSocket connection
// All callbacks are called sequentially for a single connection
// Multiple connections can have their callbacks called concurrently
type AuthWebSocketCallbacks[Params, AuthModel any] struct {
	// OnConnect is called after the WebSocket upgrade succeeds (optional)
	// The auth parameter contains the authenticated user model
	// If an error is returned, the connection is closed and OnDisconnect is called
	OnConnect func(ctx context.Context, conn *WebSocketConnection, registry ConnectionRegistry, params Params, auth AuthModel) error

	// OnMessage is called for each incoming message from the client (required)
	// The messageType indicates if this is text (ws.OpText) or binary (ws.OpBinary)
	// If an error is returned, OnError is called (if provided), otherwise connection closes
	OnMessage func(ctx context.Context, conn *WebSocketConnection, registry ConnectionRegistry, messageType ws.OpCode, data []byte) error

	// OnDisconnect is called when the connection is closed (optional)
	// This is ALWAYS called, even if OnConnect or OnMessage returns an error
	// The err parameter contains the error that caused disconnection (nil for clean close)
	// This is guaranteed to run via defer, making it perfect for cleanup
	OnDisconnect func(ctx context.Context, params Params, auth AuthModel, err error)

	// OnError is called when an error occurs during OnConnect or OnMessage (optional)
	// Return true to continue processing messages, false to close the connection
	// If not provided, any error will close the connection
	OnError func(ctx context.Context, conn *WebSocketConnection, err error) bool
}
