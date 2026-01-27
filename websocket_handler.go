package simba

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// WebSocketCallbackHandlerFunc is a function type for handling WebSocket connections with callbacks
type WebSocketCallbackHandlerFunc[Params any] struct {
	callbacks   WebSocketCallbacks[Params]
	connections map[string]*WebSocketConnection
	mu          sync.RWMutex
}

// WebSocketHandler creates a handler that uses callbacks for WebSocket lifecycle events.
//
// Example usage:
//
//	func chatCallbacks() simba.WebSocketCallbacks[Params] {
//		return simba.WebSocketCallbacks[Params]{
//			OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*WebSocketConnection, params Params) error {
//				return conn.WriteText("Welcome!")
//			},
//			OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*WebSocketConnection, msgType ws.OpCode, data []byte) error {
//				// Broadcast to all connections
//				for _, c := range connections {
//					c.WriteText(string(data))
//				}
//				return nil
//			},
//		}
//	}
//
//	app.Router.GET("/ws/chat", simba.WebSocketHandler(chatCallbacks))
func WebSocketHandler[Params any](callbacks WebSocketCallbacks[Params]) Handler {
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	return &WebSocketCallbackHandlerFunc[Params]{
		callbacks:   callbacks,
		connections: make(map[string]*WebSocketConnection),
	}
}

// WebSocketHandlerFunc creates a handler from a function that returns WebSocket callbacks.
func WebSocketHandlerFunc[Params any](callbacksFunc func() WebSocketCallbacks[Params]) Handler {
	return WebSocketHandler(callbacksFunc())
}

// ServeHTTP implements the http.Handler interface for WebSocketCallbackHandlerFunc
func (h *WebSocketCallbackHandlerFunc[Params]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
func (h *WebSocketCallbackHandlerFunc[Params]) handleConnection(ctx context.Context, rawConn net.Conn, params Params, auth any) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create connection wrapper
	wsConn := &WebSocketConnection{
		ID:     generateConnID(),
		Params: params,
		conn:   rawConn,
		ctx:    ctx,
		cancel: cancel,
	}

	// Track connection
	h.mu.Lock()
	h.connections[wsConn.ID] = wsConn
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.connections, wsConn.ID)
		h.mu.Unlock()
	}()

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
		h.mu.RLock()
		connectionsSnapshot := h.connections
		h.mu.RUnlock()

		if err := h.callbacks.OnConnect(ctx, wsConn, connectionsSnapshot, params); err != nil {
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
			h.mu.RLock()
			connectionsSnapshot := h.connections
			h.mu.RUnlock()

			if err := h.callbacks.OnMessage(ctx, wsConn, connectionsSnapshot, op, msg); err != nil {
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

func (h *WebSocketCallbackHandlerFunc[Params]) getRequestBody() any {
	return simbaModels.NoBody{}
}

func (h *WebSocketCallbackHandlerFunc[Params]) getResponseBody() any {
	return simbaModels.NoBody{}
}

func (h *WebSocketCallbackHandlerFunc[Params]) getParams() any {
	var p Params
	return p
}

func (h *WebSocketCallbackHandlerFunc[Params]) getAccepts() string {
	return ""
}

func (h *WebSocketCallbackHandlerFunc[Params]) getProduces() string {
	return ""
}

func (h *WebSocketCallbackHandlerFunc[Params]) getHandler() any {
	return h.callbacks
}

func (h *WebSocketCallbackHandlerFunc[Params]) getAuthModel() any {
	return nil
}

func (h *WebSocketCallbackHandlerFunc[Params]) getAuthHandler() any {
	return nil
}

// AuthWebSocketCallbackHandlerFunc is a function type for handling authenticated WebSocket connections with callbacks
type AuthWebSocketCallbackHandlerFunc[Params, AuthModel any] struct {
	callbacks   AuthWebSocketCallbacks[Params, AuthModel]
	authHandler AuthHandler[AuthModel]
	connections map[string]*WebSocketConnection
	mu          sync.RWMutex
}

// AuthWebSocketHandler creates an authenticated handler that uses callbacks for WebSocket lifecycle events.
//
// Example usage:
//
//	func chatCallbacks() simba.AuthWebSocketCallbacks[Params, AuthModel] {
//		return simba.AuthWebSocketCallbacks[Params, AuthModel]{
//			OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*WebSocketConnection, params Params, auth AuthModel) error {
//				return conn.WriteText(fmt.Sprintf("Welcome %s!", auth.Name))
//			},
//			OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*WebSocketConnection, msgType ws.OpCode, data []byte, auth AuthModel) error {
//				message := fmt.Sprintf("[%s]: %s", auth.Name, string(data))
//				for _, c := range connections {
//					c.WriteText(message)
//				}
//				return nil
//			},
//		}
//	}
//
//	bearerAuth := simba.BearerAuth(authFunc, simba.BearerAuthConfig{...})
//	app.Router.GET("/ws/chat", simba.AuthWebSocketHandler(chatCallbacks, bearerAuth))
func AuthWebSocketHandler[Params, AuthModel any](
	callbacks AuthWebSocketCallbacks[Params, AuthModel],
	authHandler AuthHandler[AuthModel],
) Handler {
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	return &AuthWebSocketCallbackHandlerFunc[Params, AuthModel]{
		callbacks:   callbacks,
		authHandler: authHandler,
		connections: make(map[string]*WebSocketConnection),
	}
}

// AuthWebSocketHandlerFunc creates an authenticated handler from a function that returns WebSocket callbacks.
func AuthWebSocketHandlerFunc[Params, AuthModel any](
	callbacksFunc func() AuthWebSocketCallbacks[Params, AuthModel],
	authHandler AuthHandler[AuthModel],
) Handler {
	return AuthWebSocketHandler(callbacksFunc(), authHandler)
}

// ServeHTTP implements the http.Handler interface for AuthWebSocketCallbackHandlerFunc
func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) handleConnection(ctx context.Context, rawConn net.Conn, params Params, auth AuthModel) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create connection wrapper
	wsConn := &WebSocketConnection{
		ID:     generateConnID(),
		Params: params,
		conn:   rawConn,
		ctx:    ctx,
		cancel: cancel,
	}

	// Track connection
	h.mu.Lock()
	h.connections[wsConn.ID] = wsConn
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.connections, wsConn.ID)
		h.mu.Unlock()
	}()

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
		h.mu.RLock()
		connectionsSnapshot := h.connections
		h.mu.RUnlock()

		if err := h.callbacks.OnConnect(ctx, wsConn, connectionsSnapshot, params, auth); err != nil {
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
			h.mu.RLock()
			connectionsSnapshot := h.connections
			h.mu.RUnlock()

			if err := h.callbacks.OnMessage(ctx, wsConn, connectionsSnapshot, op, msg, auth); err != nil {
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

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getRequestBody() any {
	return simbaModels.NoBody{}
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getResponseBody() any {
	return simbaModels.NoBody{}
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getParams() any {
	var p Params
	return p
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getAccepts() string {
	return ""
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getProduces() string {
	return ""
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getHandler() any {
	return h.callbacks
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getAuthModel() any {
	var am AuthModel
	return am
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) getAuthHandler() any {
	return h.authHandler
}

// generateConnID generates a unique connection ID
func generateConnID() string {
	// Use a simple counter for now - could be UUID or more sophisticated
	// This is internal and doesn't need to be cryptographically secure
	return fmt.Sprintf("conn-%d", time.Now().UnixNano())
}
