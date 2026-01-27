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

// chatCallbacks returns authenticated WebSocket callbacks for the chat handler.
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

			// Notify others in the room (distributed broadcast)
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
			// With distributed registry, this goes to ALL instances
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

	// Create a DISTRIBUTED registry
	// This mock implementation simulates distributed behavior
	distributedRegistry := NewMockDistributedRegistry("instance-1")

	slog.Info("==============================================")
	slog.Info("WebSocket Distributed Example")
	slog.Info("==============================================")
	slog.Info("")
	slog.Info("This example demonstrates a CUSTOM CONNECTION REGISTRY")
	slog.Info("that simulates distributed deployment across multiple instances.")
	slog.Info("")
	slog.Info("Key features demonstrated:")
	slog.Info("  - Custom registry implementation")
	slog.Info("  - Simulated cross-instance broadcasting")
	slog.Info("  - Instance-local connection management")
	slog.Info("  - Group/room tracking across instances")
	slog.Info("")
	slog.Info("In production, you would replace MockDistributedRegistry with")
	slog.Info("a real implementation using Redis, Cassandra, or similar.")
	slog.Info("")

	// Authenticated chat handler with DISTRIBUTED broadcasting
	// Usage: ws://localhost:8080/ws/chat/{room}
	// Requires: Authorization: Bearer valid-token
	// Note: Using AuthWebSocketHandlerFuncWithRegistry for cleaner syntax with custom registry
	app.Router.GET("/ws/chat/{room}", simba.AuthWebSocketHandlerFuncWithRegistry(
		chatCallbacks,
		bearerAuth,
		distributedRegistry, // Custom distributed registry
	))

	// Simple WebSocket without params (also using distributed registry)
	// Usage: ws://localhost:8080/ws/simple
	// Note: Using WebSocketHandlerFuncWithRegistry for cleaner syntax with custom registry
	app.Router.GET("/ws/simple", simba.WebSocketHandlerFuncWithRegistry(
		simpleEchoCallbacks,
		distributedRegistry, // Custom distributed registry
	))

	slog.Info("Starting server with DISTRIBUTED WebSocket support on :8080")
	slog.Info("")
	slog.Info("Available endpoints:")
	slog.Info("  Authenticated chat: ws://localhost:8080/ws/chat/{room} (requires Bearer: valid-token)")
	slog.Info("  Simple echo: ws://localhost:8080/ws/simple")
	slog.Info("")
	slog.Info("Try connecting from multiple terminals to see distributed behavior!")

	app.Start()
}
