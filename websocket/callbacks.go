package websocket

import (
	"context"
)

// Callbacks defines the lifecycle callbacks for a WebSocket connection.
//
// The framework handles protocol details (upgrade, framing, etc.).
// You handle application logic (authentication, routing, persistence).
type Callbacks[Params any] struct {
	// OnConnect is called after the WebSocket upgrade succeeds (optional).
	// Return an error to reject the connection.
	OnConnect func(ctx context.Context, conn *Connection, params Params) error

	// OnMessage is called for each incoming message from the client (required).
	// Return an error to trigger OnError (if provided) or close the connection.
	OnMessage func(ctx context.Context, conn *Connection, data []byte) error

	// OnDisconnect is called when the connection is closed (optional).
	// The connID is provided since the connection is already closed.
	// The err parameter contains the error that caused disconnection (nil for clean close).
	// This is guaranteed to run via defer, making it safe for cleanup.
	OnDisconnect func(ctx context.Context, connID string, params Params, err error)

	// OnError is called when an error occurs during OnMessage (optional).
	// Return true to continue processing messages, false to close the connection.
	// If not provided, any error will close the connection.
	OnError func(ctx context.Context, conn *Connection, err error) bool
}

// AuthCallbacks defines the lifecycle callbacks for an authenticated WebSocket connection.
//
// Same as WebSocketCallbacks but includes the authenticated user model in each callback.
type AuthCallbacks[Params, AuthModel any] struct {
	// OnConnect is called after the WebSocket upgrade succeeds (optional).
	// The auth parameter contains the authenticated user model.
	// Return an error to reject the connection.
	OnConnect func(ctx context.Context, conn *Connection, params Params, auth AuthModel) error

	// OnMessage is called for each incoming message from the client (required).
	// The auth parameter contains the authenticated user model.
	// Return an error to trigger OnError (if provided) or close the connection.
	OnMessage func(ctx context.Context, conn *Connection, data []byte, auth AuthModel) error

	// OnDisconnect is called when the connection is closed (optional).
	// The connID is provided since the connection is already closed.
	// The err parameter contains the error that caused disconnection (nil for clean close).
	// This is guaranteed to run via defer, making it safe for cleanup.
	OnDisconnect func(ctx context.Context, connID string, params Params, auth AuthModel, err error)

	// OnError is called when an error occurs during OnMessage (optional).
	// Return true to continue processing messages, false to close the connection.
	// If not provided, any error will close the connection.
	OnError func(ctx context.Context, conn *Connection, err error) bool
}
