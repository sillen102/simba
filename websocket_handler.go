package simba

import (
	"context"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// WebSocketHandlerFunc is a function type for handling WebSocket connections with params
type WebSocketHandlerFunc[Params any] func(ctx context.Context, conn net.Conn, params Params) error

// AuthenticatedWebSocketHandlerFunc is a function type for handling authenticated WebSocket connections
type AuthenticatedWebSocketHandlerFunc[Params, AuthModel any] struct {
	handler     func(ctx context.Context, conn net.Conn, params Params, authModel AuthModel) error
	authHandler AuthHandler[AuthModel]
}

// WebSocketHandler handles a WebSocket connection with params.
//
//	Example usage:
//
// Define a Request params struct:
//
//	type Params struct {
//		Room   string `path:"room" validate:"required"`
//		Token  string `query:"token"`
//		UserID string `header:"X-User-ID"`
//	}
//
// Define a handler function:
//
//	func(ctx context.Context, conn net.Conn, params Params) error {
//		defer conn.Close()
//
//		// Access the params
//		room := params.Room
//		token := params.Token
//
//		// Read/write WebSocket messages
//		for {
//			msg, op, err := wsutil.ReadClientData(conn)
//			if err != nil {
//				return err
//			}
//
//			// Process message and send response
//			err = wsutil.WriteServerMessage(conn, op, msg)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
// Register the handler:
//
//	Mux.GET("/ws/{room}", simba.WebSocketHandler(handler))
func WebSocketHandler[Params any](h WebSocketHandlerFunc[Params]) Handler {
	return h
}

// ServeHTTP implements the http.Handler interface for WebSocketHandlerFunc
func (h WebSocketHandlerFunc[Params]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Call the user handler with the WebSocket connection
	// The handler is responsible for managing the connection lifecycle
	if err := h(ctx, conn, params); err != nil {
		// Connection is already upgraded, so we can't write HTTP errors
		// Just close the connection
		_ = conn.Close()
	}
}

func (h WebSocketHandlerFunc[Params]) getRequestBody() any {
	return simbaModels.NoBody{}
}

func (h WebSocketHandlerFunc[Params]) getResponseBody() any {
	return simbaModels.NoBody{}
}

func (h WebSocketHandlerFunc[Params]) getParams() any {
	var p Params
	return p
}

func (h WebSocketHandlerFunc[Params]) getAccepts() string {
	// WebSocket upgrade doesn't require specific content type
	return ""
}

func (h WebSocketHandlerFunc[Params]) getProduces() string {
	// WebSocket produces binary/text frames, not HTTP responses
	return ""
}

func (h WebSocketHandlerFunc[Params]) getHandler() any {
	return h
}

func (h WebSocketHandlerFunc[Params]) getAuthModel() any {
	return nil
}

func (h WebSocketHandlerFunc[Params]) getAuthHandler() any {
	return nil
}

// AuthWebSocketHandler handles an authenticated WebSocket connection with params.
//
//	Example usage:
//
// Define a Request params struct:
//
//	type Params struct {
//		Room string `path:"room" validate:"required"`
//	}
//
// Define an auth model struct:
//
//	type AuthModel struct {
//		ID   int
//		Name string
//		Role string
//	}
//
// Define a handler function:
//
//	func(ctx context.Context, conn net.Conn, params Params, authModel AuthModel) error {
//		defer conn.Close()
//
//		// Access the params and auth model
//		room := params.Room
//		userID := authModel.ID
//
//		// Read/write WebSocket messages
//		for {
//			msg, op, err := wsutil.ReadClientData(conn)
//			if err != nil {
//				return err
//			}
//
//			// Process message and send response
//			err = wsutil.WriteServerMessage(conn, op, msg)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
// Register the handler:
//
//	Mux.GET("/ws/{room}", simba.AuthWebSocketHandler(handler, authHandler))
func AuthWebSocketHandler[Params, AuthModel any](
	handler func(ctx context.Context, conn net.Conn, params Params, authModel AuthModel) error,
	authHandler AuthHandler[AuthModel],
) Handler {
	return AuthenticatedWebSocketHandlerFunc[Params, AuthModel]{
		handler:     handler,
		authHandler: authHandler,
	}
}

// ServeHTTP implements the http.Handler interface for AuthenticatedWebSocketHandlerFunc
func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Call the user handler with the WebSocket connection
	// The handler is responsible for managing the connection lifecycle
	if err = h.handler(ctx, conn, params, authModel); err != nil {
		// Connection is already upgraded, so we can't write HTTP errors
		// Just close the connection
		_ = conn.Close()
	}
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getRequestBody() any {
	return simbaModels.NoBody{}
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getResponseBody() any {
	return simbaModels.NoBody{}
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getParams() any {
	var p Params
	return p
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getAccepts() string {
	// WebSocket upgrade doesn't require specific content type
	return ""
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getProduces() string {
	// WebSocket produces binary/text frames, not HTTP responses
	return ""
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getHandler() any {
	return h.handler
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getAuthModel() any {
	var am AuthModel
	return am
}

func (h AuthenticatedWebSocketHandlerFunc[Params, AuthModel]) getAuthHandler() any {
	return h.authHandler
}
