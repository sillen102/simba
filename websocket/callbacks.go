package websocket

import (
	"context"
)

// Callbacks defines lifecycle callbacks for a WebSocket connection.
type Callbacks[Params any] struct {
	// OnConnect is called after WebSocket upgrade succeeds.
	// Return an error to reject the connection.
	OnConnect func(ctx context.Context, conn *Connection, params Params) error

	// OnMessage is called for each incoming message (required).
	// Return an error to trigger OnError or close the connection.
	OnMessage func(ctx context.Context, conn *Connection, data []byte) error

	// OnDisconnect is called when the connection closes.
	// The err parameter is nil for clean close, non-nil otherwise.
	// Guaranteed to run via defer, making it safe for cleanup.
	OnDisconnect func(ctx context.Context, connID string, params Params, err error)

	// OnError is called when OnMessage returns an error.
	// Return true to continue, false to close the connection.
	// If not provided, errors close the connection.
	OnError func(ctx context.Context, conn *Connection, err error) bool
}

// AuthCallbacks defines lifecycle callbacks for an authenticated WebSocket connection.
// Same as Callbacks but includes the authenticated user model in each callback.
type AuthCallbacks[Params, AuthModel any] struct {
	// OnConnect is called after WebSocket upgrade succeeds.
	// Return an error to reject the connection.
	OnConnect func(ctx context.Context, conn *Connection, params Params, auth AuthModel) error

	// OnMessage is called for each incoming message (required).
	// Return an error to trigger OnError or close the connection.
	OnMessage func(ctx context.Context, conn *Connection, data []byte, auth AuthModel) error

	// OnDisconnect is called when the connection closes.
	// The err parameter is nil for clean close, non-nil otherwise.
	// Guaranteed to run via defer, making it safe for cleanup.
	OnDisconnect func(ctx context.Context, connID string, params Params, auth AuthModel, err error)

	// OnError is called when OnMessage returns an error.
	// Return true to continue, false to close the connection.
	// If not provided, errors close the connection.
	OnError func(ctx context.Context, conn *Connection, err error) bool
}
