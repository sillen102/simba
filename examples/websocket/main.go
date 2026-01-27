package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
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

// Simple WebSocket echo handler
func echoHandler(ctx context.Context, conn net.Conn, params Params) error {
	defer conn.Close()

	slog.Info("WebSocket connection established", "room", params.Room)

	for {
		// Read message from client
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			// Connection closed or error occurred
			slog.Info("WebSocket connection closed", "error", err)
			return err
		}

		slog.Info("Received message", "opcode", op, "message", string(msg))

		// Echo the message back to client
		err = wsutil.WriteServerMessage(conn, op, msg)
		if err != nil {
			slog.Error("Failed to write message", "error", err)
			return err
		}
	}
}

// Authenticated WebSocket chat handler
func chatHandler(ctx context.Context, conn net.Conn, params Params, auth AuthModel) error {
	defer conn.Close()

	slog.Info("Authenticated WebSocket connection established",
		"room", params.Room,
		"userID", auth.ID,
		"userName", auth.Name,
	)

	// Send welcome message
	welcomeMsg := fmt.Sprintf("Welcome to room %s, %s!", params.Room, auth.Name)
	err := wsutil.WriteServerText(conn, []byte(welcomeMsg))
	if err != nil {
		return err
	}

	for {
		// Read message from client
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			slog.Info("WebSocket connection closed", "error", err)
			return err
		}

		// Only handle text messages
		if op != ws.OpText {
			continue
		}

		slog.Info("Received chat message",
			"room", params.Room,
			"user", auth.Name,
			"message", string(msg),
		)

		// Echo with username prefix
		response := fmt.Sprintf("[%s]: %s", auth.Name, string(msg))
		err = wsutil.WriteServerText(conn, []byte(response))
		if err != nil {
			slog.Error("Failed to write message", "error", err)
			return err
		}
	}
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

func main() {
	// Set up logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	app := simba.Default()

	// Simple WebSocket echo endpoint
	// Usage: ws://localhost:8080/ws/echo/lobby
	app.Router.GET("/ws/echo/{room}", simba.WebSocketHandler(echoHandler))

	// Authenticated WebSocket chat endpoint
	// Usage: ws://localhost:8080/ws/chat/general
	// Requires: Authorization: Bearer valid-token
	bearerAuth := simba.BearerAuth(authHandler, simba.BearerAuthConfig{
		Name:        "BearerAuth",
		Format:      "JWT",
		Description: "Bearer token authentication",
	})
	app.Router.GET("/ws/chat/{room}", simba.AuthWebSocketHandler(chatHandler, bearerAuth))

	// WebSocket with query parameters
	app.Router.GET("/ws/query", simba.WebSocketHandler(func(ctx context.Context, conn net.Conn, params simbaModels.NoParams) error {
		defer conn.Close()
		slog.Info("WebSocket with query params connected")

		// Simple echo loop
		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return err
			}
			err = wsutil.WriteServerMessage(conn, op, msg)
			if err != nil {
				return err
			}
		}
	}))

	slog.Info("Starting server with WebSocket support on :8080")
	slog.Info("Echo endpoint: ws://localhost:8080/ws/echo/{room}")
	slog.Info("Chat endpoint: ws://localhost:8080/ws/chat/{room} (requires Bearer token)")
	slog.Info("Query endpoint: ws://localhost:8080/ws/query")

	app.Start()
}
