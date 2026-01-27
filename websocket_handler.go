package simba

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// WebSocketCallbackHandlerFunc is a function type for handling WebSocket connections with callbacks
type WebSocketCallbackHandlerFunc[Params any] struct {
	callbacks WebSocketCallbacks[Params]
	registry  *connectionRegistry
}

// WebSocketHandler creates a handler that uses callbacks for WebSocket lifecycle events.
// The callbacks can be defined in a separate function for better organization.
//
// Example usage:
//
//	// Define callbacks in a separate function
//	func chatCallbacks() simba.WebSocketCallbacks[Params] {
//		return simba.WebSocketCallbacks[Params]{
//			OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, params Params) error {
//				registry.Join(conn.ID, params.Room)
//				return conn.WriteText("Welcome!")
//			},
//			OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, msgType ws.OpCode, data []byte) error {
//				params := conn.Params.(Params)
//				return registry.BroadcastToGroup(params.Room, data)
//			},
//			OnDisconnect: func(ctx context.Context, params Params, err error) {
//				// Cleanup logic here
//			},
//		}
//	}
//
//	// Register the handler
//	app.Router.GET("/ws/chat/{room}", simba.WebSocketHandler(chatCallbacks()))
func WebSocketHandler[Params any](callbacks WebSocketCallbacks[Params]) Handler {
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	return WebSocketCallbackHandlerFunc[Params]{
		callbacks: callbacks,
		registry:  newConnectionRegistry(),
	}
}

// ServeHTTP implements the http.Handler interface for WebSocketCallbackHandlerFunc
func (h WebSocketCallbackHandlerFunc[Params]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate params before upgrading connection
	params, err := parseAndValidateParams[Params](r)
	if err != nil {
		simbaErrors.WriteError(w, r, err)
		return
	}

	// Upgrade the HTTP connection to WebSocket
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		simbaErrors.WriteError(w, r, simbaErrors.NewSimbaError(
			http.StatusBadRequest,
			"failed to upgrade connection",
			err,
		))
		return
	}

	// Handle the connection
	h.handleConnection(ctx, conn, params, nil)
}

// handleConnection manages the lifecycle of a WebSocket connection
func (h WebSocketCallbackHandlerFunc[Params]) handleConnection(ctx context.Context, rawConn net.Conn, params Params, auth any) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create connection wrapper
	wsConn := &WebSocketConnection{
		ID:     generateConnID(),
		conn:   rawConn,
		ctx:    ctx,
		cancel: cancel,
		Params: params,
		Auth:   auth,
	}

	// Track connection
	h.registry.add(wsConn)
	defer h.registry.remove(wsConn.ID)

	// Always cleanup
	var handlerErr error
	defer func() {
		_ = rawConn.Close()
		if h.callbacks.OnDisconnect != nil {
			// Use background context for cleanup as connection context is cancelled
			h.callbacks.OnDisconnect(context.Background(), params, handlerErr)
		}
	}()

	// Call OnConnect
	if h.callbacks.OnConnect != nil {
		if err := h.callbacks.OnConnect(ctx, wsConn, h.registry, params); err != nil {
			handlerErr = err
			return
		}
	}

	// Message loop - process messages sequentially for this connection
	for {
		select {
		case <-ctx.Done():
			handlerErr = ctx.Err()
			return
		default:
			// Read message from client
			msg, op, err := wsutil.ReadClientData(rawConn)
			if err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil && h.callbacks.OnError(ctx, wsConn, err) {
					continue
				}
				handlerErr = err
				return
			}

			// Call OnMessage
			if err := h.callbacks.OnMessage(ctx, wsConn, h.registry, op, msg); err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil && h.callbacks.OnError(ctx, wsConn, err) {
					continue
				}
				handlerErr = err
				return
			}
		}
	}
}

func (h WebSocketCallbackHandlerFunc[Params]) getRequestBody() any {
	return simbaModels.NoBody{}
}

func (h WebSocketCallbackHandlerFunc[Params]) getResponseBody() any {
	return simbaModels.NoBody{}
}

func (h WebSocketCallbackHandlerFunc[Params]) getParams() any {
	var p Params
	return p
}

func (h WebSocketCallbackHandlerFunc[Params]) getAccepts() string {
	return ""
}

func (h WebSocketCallbackHandlerFunc[Params]) getProduces() string {
	return ""
}

func (h WebSocketCallbackHandlerFunc[Params]) getHandler() any {
	return h.callbacks
}

func (h WebSocketCallbackHandlerFunc[Params]) getAuthModel() any {
	return nil
}

func (h WebSocketCallbackHandlerFunc[Params]) getAuthHandler() any {
	return nil
}

// AuthWebSocketCallbackHandlerFunc is a function type for handling authenticated WebSocket connections with callbacks
type AuthWebSocketCallbackHandlerFunc[Params, AuthModel any] struct {
	callbacks   AuthWebSocketCallbacks[Params, AuthModel]
	authHandler AuthHandler[AuthModel]
	registry    *connectionRegistry
}

// AuthWebSocketHandler creates an authenticated handler that uses callbacks for WebSocket lifecycle events.
// The callbacks can be defined in a separate function for better organization.
//
// Example usage:
//
//	// Define callbacks in a separate function
//	func chatCallbacks() simba.AuthWebSocketCallbacks[Params, AuthModel] {
//		return simba.AuthWebSocketCallbacks[Params, AuthModel]{
//			OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, params Params, auth AuthModel) error {
//				registry.Join(conn.ID, params.Room)
//				return conn.WriteText(fmt.Sprintf("Welcome %s!", auth.Name))
//			},
//			OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, msgType ws.OpCode, data []byte) error {
//				params := conn.Params.(Params)
//				auth := conn.Auth.(AuthModel)
//				message := fmt.Sprintf("[%s]: %s", auth.Name, string(data))
//				return registry.BroadcastToGroup(params.Room, []byte(message))
//			},
//		}
//	}
//
//	// Register the handler
//	bearerAuth := simba.BearerAuth(authFunc, simba.BearerAuthConfig{...})
//	app.Router.GET("/ws/chat/{room}", simba.AuthWebSocketHandler(chatCallbacks(), bearerAuth))
func AuthWebSocketHandler[Params, AuthModel any](
	callbacks AuthWebSocketCallbacks[Params, AuthModel],
	authHandler AuthHandler[AuthModel],
) Handler {
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	return AuthWebSocketCallbackHandlerFunc[Params, AuthModel]{
		callbacks:   callbacks,
		authHandler: authHandler,
		registry:    newConnectionRegistry(),
	}
}

// ServeHTTP implements the http.Handler interface for AuthWebSocketCallbackHandlerFunc
func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Authenticate before upgrading connection
	authModel, err := handleAuthRequest[AuthModel](h.authHandler, r)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if statusCoder, ok := err.(simbaErrors.StatusCodeProvider); ok {
			statusCode = statusCoder.StatusCode()
		}

		errorMessage := "unauthorized"
		if msgProvider, ok := err.(simbaErrors.PublicMessageProvider); ok {
			errorMessage = msgProvider.PublicMessage()
		}

		simbaErrors.WriteError(w, r, simbaErrors.NewSimbaError(statusCode, errorMessage, err))
		return
	}

	// Parse and validate params before upgrading connection
	params, err := parseAndValidateParams[Params](r)
	if err != nil {
		simbaErrors.WriteError(w, r, err)
		return
	}

	// Upgrade the HTTP connection to WebSocket
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		simbaErrors.WriteError(w, r, simbaErrors.NewSimbaError(
			http.StatusBadRequest,
			"failed to upgrade connection",
			err,
		))
		return
	}

	// Handle the connection
	h.handleConnection(ctx, conn, params, authModel)
}

// handleConnection manages the lifecycle of an authenticated WebSocket connection
func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) handleConnection(ctx context.Context, rawConn net.Conn, params Params, auth AuthModel) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create connection wrapper
	wsConn := &WebSocketConnection{
		ID:     generateConnID(),
		conn:   rawConn,
		ctx:    ctx,
		cancel: cancel,
		Params: params,
		Auth:   auth,
	}

	// Track connection
	h.registry.add(wsConn)
	defer h.registry.remove(wsConn.ID)

	// Always cleanup
	var handlerErr error
	defer func() {
		_ = rawConn.Close()
		if h.callbacks.OnDisconnect != nil {
			// Use background context for cleanup as connection context is cancelled
			h.callbacks.OnDisconnect(context.Background(), params, auth, handlerErr)
		}
	}()

	// Call OnConnect
	if h.callbacks.OnConnect != nil {
		if err := h.callbacks.OnConnect(ctx, wsConn, h.registry, params, auth); err != nil {
			handlerErr = err
			return
		}
	}

	// Message loop - process messages sequentially for this connection
	for {
		select {
		case <-ctx.Done():
			handlerErr = ctx.Err()
			return
		default:
			// Read message from client
			msg, op, err := wsutil.ReadClientData(rawConn)
			if err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil && h.callbacks.OnError(ctx, wsConn, err) {
					continue
				}
				handlerErr = err
				return
			}

			// Call OnMessage
			if err := h.callbacks.OnMessage(ctx, wsConn, h.registry, op, msg); err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil && h.callbacks.OnError(ctx, wsConn, err) {
					continue
				}
				handlerErr = err
				return
			}
		}
	}
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getRequestBody() any {
	return simbaModels.NoBody{}
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getResponseBody() any {
	return simbaModels.NoBody{}
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getParams() any {
	var p Params
	return p
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getAccepts() string {
	return ""
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getProduces() string {
	return ""
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getHandler() any {
	return h.callbacks
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getAuthModel() any {
	var am AuthModel
	return am
}

func (h AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getAuthHandler() any {
	return h.authHandler
}

// generateConnID generates a unique connection ID
func generateConnID() string {
	// Use a simple counter for now - could be UUID or more sophisticated
	// This is internal and doesn't need to be cryptographically secure
	return fmt.Sprintf("conn-%d", time.Now().UnixNano())
}
