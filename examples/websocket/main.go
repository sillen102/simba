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

// AuthModel represents the authenticated user
type AuthModel struct {
	ID   int
	Name string
}

// Simple bearer token auth handler for demonstration
func authHandler(ctx context.Context, token string) (AuthModel, error) {
	if token == "valid-token" {
		return AuthModel{
			ID:   1,
			Name: "John Doe",
		}, nil
	}
	return AuthModel{}, fmt.Errorf("invalid token")
}

// echoCallbacks returns WebSocket callbacks for a simple echo handler
func echoCallbacks() simba.WebSocketCallbacks[simbaModels.NoParams] {
	return simba.WebSocketCallbacks[simbaModels.NoParams]{
		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams) error {
			slog.Info("Connection established", "connID", conn.ID, "totalConns", len(connections))
			return conn.WriteText("Welcome! Send me messages and I'll echo them back.")
		},

		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte) error {
			slog.Info("Received message", "message", string(data))
			// Echo back to sender
			return conn.WriteText(fmt.Sprintf("Echo: %s", string(data)))
		},

		OnDisconnect: func(ctx context.Context, params simbaModels.NoParams, err error) {
			slog.Info("Connection closed", "error", err)
		},

		OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
			slog.Error("Error occurred", "error", err)
			return false // Close connection on error
		},
	}
}

// broadcastCallbacks demonstrates authenticated WebSocket with broadcasting
func broadcastCallbacks() simba.AuthWebSocketCallbacks[simbaModels.NoParams, AuthModel] {
	return simba.AuthWebSocketCallbacks[simbaModels.NoParams, AuthModel]{
		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, params simbaModels.NoParams, auth AuthModel) error {
			slog.Info("Authenticated user connected",
				"user", auth.Name,
				"connID", conn.ID,
			)

			// Send welcome to the user
			conn.WriteText(fmt.Sprintf("Welcome %s!", auth.Name))

			// Notify all other connections
			msg := fmt.Sprintf("%s joined the chat", auth.Name)
			for _, c := range connections {
				if c.ID != conn.ID {
					c.WriteText(msg)
				}
			}

			return nil
		},

		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, connections map[string]*simba.WebSocketConnection, msgType ws.OpCode, data []byte, auth AuthModel) error {
			slog.Info("Chat message",
				"user", auth.Name,
				"message", string(data),
			)

			// Format message with username and broadcast to all
			message := fmt.Sprintf("[%s]: %s", auth.Name, string(data))
			for _, c := range connections {
				c.WriteText(message)
			}
			return nil
		},

		OnDisconnect: func(ctx context.Context, params simbaModels.NoParams, auth AuthModel, err error) {
			slog.Info("User disconnected",
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

func main() {
	// Set up logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	app := simba.Default()

	// Bearer token auth handler for authenticated endpoints
	bearerAuth := simba.BearerAuth(authHandler, simba.BearerAuthConfig{
		Name:        "BearerAuth",
		Format:      "JWT",
		Description: "Bearer token authentication",
	})

	// Simple echo endpoint (no authentication)
	// Usage: ws://localhost:8080/ws/echo
	app.Router.GET("/ws/echo", simba.WebSocketHandlerFunc(echoCallbacks))

	// Broadcast chat endpoint (requires authentication)
	// Usage: ws://localhost:8080/ws/broadcast with header "Authorization: Bearer valid-token"
	app.Router.GET("/ws/broadcast", simba.AuthWebSocketHandlerFunc(broadcastCallbacks, bearerAuth))

	slog.Info("Starting server with WebSocket support on :8080")
	slog.Info("")
	slog.Info("Available endpoints:")
	slog.Info("  Echo: ws://localhost:8080/ws/echo")
	slog.Info("  Broadcast chat: ws://localhost:8080/ws/broadcast (requires Bearer token)")
	slog.Info("")

	app.Start()
}
