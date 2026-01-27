package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gobwas/ws"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
)

// Params defines the WebSocket endpoint parameters
type Params struct {
	Room string `path:"room" validate:"required"`
}

// AuthModel represents the authenticated user
type AuthModel struct {
	ID   int
	Name string
}

// Simple bearer token auth handler for demonstration
func authHandler(ctx context.Context, token string) (AuthModel, error) {
	// In a real application, validate the token against a database or JWT
	if token == "valid-token" {
		return AuthModel{
			ID:   1,
			Name: "John Doe",
		}, nil
	}
	return AuthModel{}, fmt.Errorf("invalid token")
}

// echoCallbacks returns WebSocket callbacks for the echo handler.
// This demonstrates how to structure handlers in separate functions or files.
func echoCallbacks() simba.WebSocketCallbacks[Params] {
	return simba.WebSocketCallbacks[Params]{
		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, params Params) error {
			slog.Info("Connection established", "room", params.Room, "connID", conn.ID)
			// Join the room automatically
			registry.Join(conn.ID, params.Room)
			// Send welcome message
			return conn.WriteText(fmt.Sprintf("Welcome to %s!", params.Room))
		},

		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, msgType ws.OpCode, data []byte) error {
			params := conn.Params.(Params)
			slog.Info("Received message", "room", params.Room, "message", string(data))

			// Broadcast to everyone in the room
			return registry.BroadcastToGroup(params.Room, data)
		},

		OnDisconnect: func(ctx context.Context, params Params, err error) {
			slog.Info("Connection closed", "room", params.Room, "error", err)
		},

		OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
			slog.Error("Error occurred", "error", err)
			return false // Close connection on error
		},
	}
}

// chatCallbacks returns authenticated WebSocket callbacks for the chat handler.
// This demonstrates how to structure authenticated handlers in separate functions or files.
func chatCallbacks() simba.AuthWebSocketCallbacks[Params, AuthModel] {
	return simba.AuthWebSocketCallbacks[Params, AuthModel]{
		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, params Params, auth AuthModel) error {
			slog.Info("Authenticated user connected",
				"room", params.Room,
				"user", auth.Name,
				"connID", conn.ID,
			)

			// Join the room
			registry.Join(conn.ID, params.Room)

			// Send welcome to the user
			conn.WriteText(fmt.Sprintf("Welcome %s to %s!", auth.Name, params.Room))

			// Notify others in the room
			msg := fmt.Sprintf("%s joined the room", auth.Name)
			registry.BroadcastToGroup(params.Room, []byte(msg))

			return nil
		},

		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, msgType ws.OpCode, data []byte) error {
			params := conn.Params.(Params)
			auth := conn.Auth.(AuthModel)

			slog.Info("Chat message",
				"room", params.Room,
				"user", auth.Name,
				"message", string(data),
			)

			// Format message with username and broadcast to room
			message := fmt.Sprintf("[%s]: %s", auth.Name, string(data))
			return registry.BroadcastToGroupText(params.Room, message)
		},

		OnDisconnect: func(ctx context.Context, params Params, auth AuthModel, err error) {
			slog.Info("User disconnected",
				"room", params.Room,
				"user", auth.Name,
				"error", err,
			)
		},

		OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
			slog.Error("Chat error", "error", err)
			return false // Close on error
		},
	}
}

// simpleEchoCallbacks returns WebSocket callbacks for a simple echo handler without params.
// This demonstrates how to structure handlers without path parameters.
func simpleEchoCallbacks() simba.WebSocketCallbacks[simbaModels.NoParams] {
	return simba.WebSocketCallbacks[simbaModels.NoParams]{
		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, params simbaModels.NoParams) error {
			slog.Info("Simple connection", "connID", conn.ID)
			return conn.WriteText("Connected! Send me messages.")
		},

		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, registry simba.ConnectionRegistry, msgType ws.OpCode, data []byte) error {
			slog.Info("Simple message", "message", string(data))
			// Echo back with prefix
			return conn.WriteText(fmt.Sprintf("You said: %s", string(data)))
		},
	}
}

func main() {
	// Set up logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	app := simba.Default()

	// Bearer token auth handler for authenticated endpoints
	bearerAuth := simba.BearerAuth(authHandler, simba.BearerAuthConfig{
		Name:        "BearerAuth",
		Format:      "JWT",
		Description: "Bearer token authentication",
	})

	// Echo handler with room support and broadcasting
	// Usage: ws://localhost:8080/ws/echo/{room}
	app.Router.GET("/ws/echo/{room}", simba.WebSocketHandler(echoCallbacks()))

	// Authenticated chat handler with broadcasting
	// Usage: ws://localhost:8080/ws/chat/{room}
	// Requires: Authorization: Bearer valid-token
	app.Router.GET("/ws/chat/{room}", simba.AuthWebSocketHandler(chatCallbacks(), bearerAuth))

	// Simple WebSocket without params
	// Usage: ws://localhost:8080/ws/simple
	app.Router.GET("/ws/simple", simba.WebSocketHandler(simpleEchoCallbacks()))

	slog.Info("Starting server with WebSocket support on :8080")
	slog.Info("")
	slog.Info("Available endpoints:")
	slog.Info("  Echo with rooms: ws://localhost:8080/ws/echo/{room}")
	slog.Info("  Authenticated chat: ws://localhost:8080/ws/chat/{room} (requires Bearer token)")
	slog.Info("  Simple echo: ws://localhost:8080/ws/simple")

	app.Start()
}
