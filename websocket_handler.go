package simba

import (
	"context"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// WebSocketMiddleware wraps a context to enrich it before callback invocations.
// Middleware runs before each callback (OnConnect, OnMessage, OnDisconnect, OnError).
type WebSocketMiddleware func(ctx context.Context) context.Context

// WebSocketHandlerOption is an option for configuring WebSocket handlers.
type WebSocketHandlerOption interface {
	apply(handler any)
}

// middlewareOption implements WebSocketHandlerOption for middleware.
type middlewareOption struct {
	middleware []WebSocketMiddleware
}

func (m middlewareOption) apply(handler any) {
	// Use interface-based approach to work with generics
	if v, ok := handler.(interface{ setMiddleware([]WebSocketMiddleware) }); ok {
		v.setMiddleware(m.middleware)
	}
}

// WithMiddleware adds middleware to the WebSocket handler.
// Middleware runs before each callback invocation, allowing you to enrich the context.
func WithMiddleware(middleware ...WebSocketMiddleware) WebSocketHandlerOption {
	return middlewareOption{middleware: middleware}
}

// WebSocketCallbackHandlerFunc handles WebSocket connections with callbacks.
type WebSocketCallbackHandlerFunc[Params any] struct {
	callbacks  WebSocketCallbacks[Params]
	middleware []WebSocketMiddleware
}

func (h *WebSocketCallbackHandlerFunc[Params]) setMiddleware(middleware []WebSocketMiddleware) {
	h.middleware = middleware
}

// WebSocketHandler creates a handler that uses callbacks for WebSocket lifecycle events.
//
// Example usage:
//
//	app.Router.GET("/ws/echo", simba.WebSocketHandler(
//		echoCallbacks,
//		simba.WithMiddleware(
//			wsMiddleware.TraceID(),
//			wsMiddleware.Logger(),
//		),
//	))
//
// Where echoCallbacks is a function:
//
//	func echoCallbacks() simba.WebSocketCallbacks[simbaModels.NoParams] {
//		return simba.WebSocketCallbacks[simbaModels.NoParams]{
//			OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, data []byte) error {
//				return conn.WriteText("Echo: " + string(data))
//			},
//		}
//	}
func WebSocketHandler[Params any](callbacksFunc func() WebSocketCallbacks[Params], options ...WebSocketHandlerOption) Handler {
	callbacks := callbacksFunc()
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	handler := &WebSocketCallbackHandlerFunc[Params]{
		callbacks: callbacks,
	}

	// Apply options
	for _, opt := range options {
		opt.apply(handler)
	}

	return handler
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

	// Add connectionID to context (persistent for entire connection)
	ctx = context.WithValue(ctx, simbaContext.ConnectionIDKey, wsConn.ID)

	// Always cleanup
	var handlerErr error
	defer func() {
		_ = rawConn.Close()
		if h.callbacks.OnDisconnect != nil {
			// Use background context for cleanup as connection context may be cancelled
			// Apply middleware for OnDisconnect
			disconnectCtx := h.applyMiddleware(context.Background())
			disconnectCtx = context.WithValue(disconnectCtx, simbaContext.ConnectionIDKey, wsConn.ID)
			h.callbacks.OnDisconnect(disconnectCtx, wsConn.ID, params, handlerErr)
		}
	}()

	// Call OnConnect with middleware
	if h.callbacks.OnConnect != nil {
		connectCtx := h.applyMiddleware(ctx)
		if err := h.callbacks.OnConnect(connectCtx, wsConn, params); err != nil {
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
			msg, _, err := wsutil.ReadClientData(rawConn)
			if err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil {
					errorCtx := h.applyMiddleware(ctx)
					if h.callbacks.OnError(errorCtx, wsConn, err) {
						continue
					}
				}
				handlerErr = err
				return
			}

			// Call OnMessage with middleware (fresh context per message)
			messageCtx := h.applyMiddleware(ctx)
			if err := h.callbacks.OnMessage(messageCtx, wsConn, msg); err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil {
					errorCtx := h.applyMiddleware(ctx)
					if h.callbacks.OnError(errorCtx, wsConn, err) {
						continue
					}
				}
				handlerErr = err
				return
			}
		}
	}
}

// applyMiddleware applies the middleware chain to the context.
func (h *WebSocketCallbackHandlerFunc[Params]) applyMiddleware(ctx context.Context) context.Context {
	for _, mw := range h.middleware {
		ctx = mw(ctx)
	}
	return ctx
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
	middleware  []WebSocketMiddleware
}

func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) setMiddleware(middleware []WebSocketMiddleware) {
	h.middleware = middleware
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
	callbacksFunc func() AuthWebSocketCallbacks[Params, AuthModel],
	authHandler AuthHandler[AuthModel],
	options ...WebSocketHandlerOption,
) Handler {
	callbacks := callbacksFunc()
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	handler := &AuthWebSocketCallbackHandlerFunc[Params, AuthModel]{
		callbacks:   callbacks,
		authHandler: authHandler,
	}

	// Apply options
	for _, opt := range options {
		opt.apply(handler)
	}

	return handler
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

	// Add connectionID to context (persistent for entire connection)
	ctx = context.WithValue(ctx, simbaContext.ConnectionIDKey, wsConn.ID)

	// Always cleanup
	var handlerErr error
	defer func() {
		_ = rawConn.Close()
		if h.callbacks.OnDisconnect != nil {
			// Use background context for cleanup as connection context may be cancelled
			// Apply middleware for OnDisconnect
			disconnectCtx := h.applyMiddleware(context.Background())
			disconnectCtx = context.WithValue(disconnectCtx, simbaContext.ConnectionIDKey, wsConn.ID)
			h.callbacks.OnDisconnect(disconnectCtx, wsConn.ID, params, auth, handlerErr)
		}
	}()

	// Call OnConnect with middleware
	if h.callbacks.OnConnect != nil {
		connectCtx := h.applyMiddleware(ctx)
		if err := h.callbacks.OnConnect(connectCtx, wsConn, params, auth); err != nil {
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
			msg, _, err := wsutil.ReadClientData(rawConn)
			if err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil {
					errorCtx := h.applyMiddleware(ctx)
					if h.callbacks.OnError(errorCtx, wsConn, err) {
						continue
					}
				}
				handlerErr = err
				return
			}

			// Call OnMessage with middleware (fresh context per message)
			messageCtx := h.applyMiddleware(ctx)
			if err := h.callbacks.OnMessage(messageCtx, wsConn, msg, auth); err != nil {
				// Check if OnError wants to continue
				if h.callbacks.OnError != nil {
					errorCtx := h.applyMiddleware(ctx)
					if h.callbacks.OnError(errorCtx, wsConn, err) {
						continue
					}
				}
				handlerErr = err
				return
			}
		}
	}
}

// applyMiddleware applies the middleware chain to the context.
func (h *AuthWebSocketCallbackHandlerFunc[Params, AuthModel]) applyMiddleware(ctx context.Context) context.Context {
	for _, mw := range h.middleware {
		ctx = mw(ctx)
	}
	return ctx
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
