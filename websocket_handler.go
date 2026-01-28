package simba

import (
	"context"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// WebSocketCallbackHandlerFunc handles WebSocket connections with callbacks.
type WebSocketCallbackHandlerFunc[Params any] struct {
	callbacks WebSocketCallbacks[Params]
}

// WebSocketHandler creates a handler that uses callbacks for WebSocket lifecycle events.
//
// Example usage:
//
//	app.Router.GET("/ws/echo", simba.WebSocketHandler(simba.WebSocketCallbacks[Params]{
//		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, params Params) error {
//			// Register connection in your registry
//			registry.Add(conn)
//			return conn.WriteText("Welcome!")
//		},
//		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, msgType simba.MessageType, data []byte) error {
//			return conn.WriteText("Echo: " + string(data))
//		},
//		OnDisconnect: func(ctx context.Context, connID string, params Params, err error) {
//			// Clean up your registry
//			registry.Remove(connID)
//		},
//	}))
func WebSocketHandler[Params any](callbacks WebSocketCallbacks[Params]) Handler {
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	return &WebSocketCallbackHandlerFunc[Params]{
		callbacks: callbacks,
	}
}

// WebSocketHandlerFunc creates a handler from a function that returns WebSocket callbacks.
func WebSocketHandlerFunc[Params any](callbacksFunc func() WebSocketCallbacks[Params]) Handler {
	return WebSocketHandler(callbacksFunc())
}

// ServeHTTP implements the http.Handler interface.
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

	// Handle the connection synchronously - the HTTP server runs each
	// request in its own goroutine, so blocking here is correct
	h.handleConnection(ctx, conn, params)
}

// handleConnection manages the lifecycle of a WebSocket connection.
func (h *WebSocketCallbackHandlerFunc[Params]) handleConnection(ctx context.Context, rawConn net.Conn, params Params) {
	// Create connection wrapper with unique ID
	wsConn := &WebSocketConnection{
		ID:   uuid.New().String(),
		conn: rawConn,
	}

	// Always cleanup
	var handlerErr error
	defer func() {
		_ = rawConn.Close()
		if h.callbacks.OnDisconnect != nil {
			// Use background context for cleanup as connection context may be cancelled
			h.callbacks.OnDisconnect(context.Background(), wsConn.ID, params, handlerErr)
		}
	}()

	// Call OnConnect
	if h.callbacks.OnConnect != nil {
		if err := h.callbacks.OnConnect(ctx, wsConn, params); err != nil {
			handlerErr = err
			return
		}
	}

	// Message loop
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

			// Convert ws.OpCode to MessageType
			msgType := opCodeToMessageType(op)

			// Call OnMessage
			if err := h.callbacks.OnMessage(ctx, wsConn, msgType, msg); err != nil {
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

// AuthWebSocketCallbackHandlerFunc handles authenticated WebSocket connections with callbacks.
type AuthWebSocketCallbackHandlerFunc[Params, AuthModel any] struct {
	callbacks   AuthWebSocketCallbacks[Params, AuthModel]
	authHandler AuthHandler[AuthModel]
}

// AuthWebSocketHandler creates an authenticated handler that uses callbacks for WebSocket lifecycle events.
//
// Example usage:
//
//	bearerAuth := simba.BearerAuth(authFunc, simba.BearerAuthConfig{...})
//
//	app.Router.GET("/ws/chat", simba.AuthWebSocketHandler(
//		simba.AuthWebSocketCallbacks[Params, User]{
//			OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, params Params, user User) error {
//				registry.Add(user.ID, conn)
//				return conn.WriteText(fmt.Sprintf("Welcome %s!", user.Name))
//			},
//			OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, msgType simba.MessageType, data []byte, user User) error {
//				// Handle message with user context
//				return nil
//			},
//			OnDisconnect: func(ctx context.Context, connID string, params Params, user User, err error) {
//				registry.Remove(user.ID, connID)
//			},
//		},
//		bearerAuth,
//	))
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
	}
}

// AuthWebSocketHandlerFunc creates an authenticated handler from a function that returns WebSocket callbacks.
func AuthWebSocketHandlerFunc[Params, AuthModel any](
	callbacksFunc func() AuthWebSocketCallbacks[Params, AuthModel],
	authHandler AuthHandler[AuthModel],
) Handler {
	return AuthWebSocketHandler(callbacksFunc(), authHandler)
}

// ServeHTTP implements the http.Handler interface.
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

	// Handle the connection synchronously - the HTTP server runs each
	// request in its own goroutine, so blocking here is correct
	h.handleConnection(ctx, conn, params, authModel)
}

// handleConnection manages the lifecycle of an authenticated WebSocket connection.
func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) handleConnection(ctx context.Context, rawConn net.Conn, params Params, auth AuthModel) {
	// Create connection wrapper with unique ID
	wsConn := &WebSocketConnection{
		ID:   uuid.New().String(),
		conn: rawConn,
	}

	// Always cleanup
	var handlerErr error
	defer func() {
		_ = rawConn.Close()
		if h.callbacks.OnDisconnect != nil {
			// Use background context for cleanup as connection context may be cancelled
			h.callbacks.OnDisconnect(context.Background(), wsConn.ID, params, auth, handlerErr)
		}
	}()

	// Call OnConnect
	if h.callbacks.OnConnect != nil {
		if err := h.callbacks.OnConnect(ctx, wsConn, params, auth); err != nil {
			handlerErr = err
			return
		}
	}

	// Message loop
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

			// Convert ws.OpCode to MessageType
			msgType := opCodeToMessageType(op)

			// Call OnMessage
			if err := h.callbacks.OnMessage(ctx, wsConn, msgType, msg, auth); err != nil {
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

// opCodeToMessageType converts gobwas/ws OpCode to our MessageType.
func opCodeToMessageType(op ws.OpCode) MessageType {
	if op == ws.OpBinary {
		return MessageBinary
	}
	return MessageText
}
