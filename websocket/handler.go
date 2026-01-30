package websocket

import (

	"github.com/sillen102/simba"
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

// Middleware wraps a context to enrich it before callback invocations.
// Middleware runs before each callback (OnConnect, OnMessage, OnDisconnect, OnError).
type Middleware func(ctx context.Context) context.Context

// HandlerOption is an option for configuring WebSocket handlers.
type HandlerOption interface {
	apply(handler any)
}

// middlewareOption implements HandlerOption for middleware.
type middlewareOption struct {
	middleware []Middleware
}

func (m middlewareOption) apply(handler any) {
	// Use interface-based approach to work with generics
	if v, ok := handler.(interface{ setMiddleware([]Middleware) }); ok {
		v.setMiddleware(m.middleware)
	}
}

// WithWebsocketMiddleware adds middleware to the WebSocket handler.
// Middleware runs before each callback invocation, allowing you to enrich the context.
func WithMiddleware(middleware ...Middleware) HandlerOption {
	return middlewareOption{middleware: middleware}
}

// WebSocketCallbackHandlerFunc handles WebSocket connections with callbacks.
type CallbackHandlerFunc[Params any] struct {
	callbacks  Callbacks[Params]
	middleware []Middleware
}

func (h *CallbackHandlerFunc[Params]) setMiddleware(middleware []Middleware) {
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
//	func echoCallbacks() simba.Callbacks[simbaModels.NoParams] {
//		return simba.Callbacks[simbaModels.NoParams]{
//			OnMessage: func(ctx context.Context, conn *simba.Connection, data []byte) error {
//				return conn.WriteText("Echo: " + string(data))
//			},
//		}
//	}
func Handler[Params any](callbacksFunc func() Callbacks[Params], options ...HandlerOption) simba.Handler {
	callbacks := callbacksFunc()
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	handler := &CallbackHandlerFunc[Params]{
		callbacks: callbacks,
	}

	// Apply options
	for _, opt := range options {
		opt.apply(handler)
	}

	return handler
}

// ServeHTTP implements the http.Handler interface.
func (h *CallbackHandlerFunc[Params]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate params before upgrading connection
	params, err := simba.ParseAndValidateParams[Params](r)
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
func (h *CallbackHandlerFunc[Params]) handleConnection(ctx context.Context, rawConn net.Conn, params Params) {
	// Create connection wrapper with unique ID
	wsConn := &Connection{
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
func (h *CallbackHandlerFunc[Params]) applyMiddleware(ctx context.Context) context.Context {
	for _, mw := range h.middleware {
		ctx = mw(ctx)
	}
	return ctx
}

func (h *CallbackHandlerFunc[Params]) GetRequestBody() any {
	return simbaModels.NoBody{}
}

func (h *CallbackHandlerFunc[Params]) GetResponseBody() any {
	return simbaModels.NoBody{}
}

func (h *CallbackHandlerFunc[Params]) GetParams() any {
	var p Params
	return p
}

func (h *CallbackHandlerFunc[Params]) GetAccepts() string {
	return ""
}

func (h *CallbackHandlerFunc[Params]) GetProduces() string {
	return ""
}

func (h *CallbackHandlerFunc[Params]) GetHandler() any {
	return h.callbacks
}

func (h *CallbackHandlerFunc[Params]) GetAuthModel() any {
	return nil
}

func (h *CallbackHandlerFunc[Params]) GetAuthHandler() any {
	return nil
}

// AuthWebSocketCallbackHandlerFunc handles authenticated WebSocket connections with callbacks.
type AuthCallbackHandlerFunc[Params, AuthModel any] struct {
	callbacks   AuthCallbacks[Params, AuthModel]
	authHandler simba.AuthHandler[AuthModel]
	middleware  []Middleware
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) setMiddleware(middleware []Middleware) {
	h.middleware = middleware
}

// AuthWebSocketHandler creates an authenticated handler that uses callbacks for WebSocket lifecycle events.
//
// Example usage:
//
//	bearerAuth := simba.BearerAuth(authFunc, simba.BearerAuthConfig{...})
//
//	app.Router.GET("/ws/chat", simba.AuthWebSocketHandler(
//		simba.AuthCallbacks[Params, User]{
//			OnConnect: func(ctx context.Context, conn *simba.Connection, params Params, user User) error {
//				registry.Add(user.ID, conn)
//				return conn.WriteText(fmt.Sprintf("Welcome %s!", user.Name))
//			},
//			OnMessage: func(ctx context.Context, conn *simba.Connection, msgType simba.MessageType, data []byte, user User) error {
//				// Handle message with user context
//				return nil
//			},
//			OnDisconnect: func(ctx context.Context, connID string, params Params, user User, err error) {
//				registry.Remove(user.ID, connID)
//			},
//		},
//		bearerAuth,
//	))
func AuthHandler[Params, AuthModel any](
	callbacksFunc func() AuthCallbacks[Params, AuthModel],
	authHandler simba.AuthHandler[AuthModel],
	options ...HandlerOption,
) simba.Handler {
	callbacks := callbacksFunc()
	if callbacks.OnMessage == nil {
		panic("OnMessage callback is required")
	}

	handler := &AuthCallbackHandlerFunc[Params, AuthModel]{
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
func (h *AuthCallbackHandlerFunc[Params, AuthModel]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Authenticate before upgrading connection
	authModel, err := simba.HandleAuthRequest[AuthModel](h.authHandler, r)
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
	params, err := simba.ParseAndValidateParams[Params](r)
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
func (h *AuthCallbackHandlerFunc[Params, AuthModel]) handleConnection(ctx context.Context, rawConn net.Conn, params Params, auth AuthModel) {
	// Create connection wrapper with unique ID
	wsConn := &Connection{
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
func (h *AuthCallbackHandlerFunc[Params, AuthModel]) applyMiddleware(ctx context.Context) context.Context {
	for _, mw := range h.middleware {
		ctx = mw(ctx)
	}
	return ctx
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetRequestBody() any {
	return simbaModels.NoBody{}
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetResponseBody() any {
	return simbaModels.NoBody{}
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetParams() any {
	var p Params
	return p
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetAccepts() string {
	return ""
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetProduces() string {
	return ""
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetHandler() any {
	return h.callbacks
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetAuthModel() any {
	var am AuthModel
	return am
}

func (h *AuthCallbackHandlerFunc[Params, AuthModel]) GetAuthHandler() any {
	return h.authHandler
}
