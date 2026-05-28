package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/auth"
	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/models"
	"github.com/sillen102/simba/simbaContext"
	"github.com/sillen102/simba/websocket"
	"github.com/sillen102/simba/websocket/middleware"
)

// User represents an authenticated user
type User struct {
	ID   int
	Name string
}

// Simple bearer token auth handler for demonstration
func authHandler(ctx context.Context, token string) (User, error) {
	if token == "valid-token" {
		return User{
			ID:   1,
			Name: "John Doe",
		}, nil
	}
	return User{}, fmt.Errorf("invalid token")
}

// echoCallbacks returns WebSocket callbacks for a simple echo handler
func echoCallbacks() websocket.Callbacks[models.NoParams] {
	return websocket.Callbacks[models.NoParams]{
		OnConnect: func(ctx context.Context, conn *websocket.Connection, params models.NoParams) error {
			// Logger from middleware includes connectionID and traceID
			logger := logging.From(ctx)
			logger.Info("Connection established")

			// WriteText propagates context (with traceID, timeout, cancellation)
			return conn.WriteText(ctx, "Welcome! Send me messages and I'll echo them back.")
		},

		OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
			// Logger automatically includes fresh traceID for each message
			logger := logging.From(ctx)
			logger.Info("Received message", "message", string(data))

			// Show traceID and connectionID are available
			traceID := simbaContext.GetTraceID(ctx)
			connID, _ := ctx.Value(simbaContext.ConnectionIDKey).(string)

			// WriteText propagates context with traceID for distributed tracing
			return conn.WriteText(ctx, fmt.Sprintf("Echo: %s (traceID: %s, connID: %s)",
				string(data), traceID[:8], connID[:8]))
		},

		OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, err error) {
			logger := logging.From(ctx)
			logger.Info("Connection closed", "error", err)
		},

		OnError: func(ctx context.Context, conn *websocket.Connection, err error) bool {
			logger := logging.From(ctx)
			logger.Error("Error occurred", "error", err)
			return false // Close connection on error
		},
	}
}

// chatCallbacks demonstrates authenticated WebSocket
func chatCallbacks() websocket.AuthCallbacks[models.NoParams, User] {
	return websocket.AuthCallbacks[models.NoParams, User]{
		OnConnect: func(ctx context.Context, conn *websocket.Connection, params models.NoParams, user User) error {
			logger := logging.From(ctx)
			logger.Info("User connected", "user", user.Name)

			// WriteText propagates context for tracing and cancellation
			return conn.WriteText(ctx, fmt.Sprintf("Welcome %s!", user.Name))
		},

		OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, user User) error {
			logger := logging.From(ctx)
			logger.Info("Chat message", "user", user.Name, "message", string(data))

			// Echo back with user name, propagating context
			message := fmt.Sprintf("[%s]: %s", user.Name, string(data))
			return conn.WriteText(ctx, message)
		},

		OnDisconnect: func(ctx context.Context, connID string, params models.NoParams, user User, err error) {
			logger := logging.From(ctx)
			logger.Info("User disconnected", "user", user.Name, "error", err)
		},

		OnError: func(ctx context.Context, conn *websocket.Connection, err error) bool {
			logger := logging.From(ctx)
			logger.Error("Chat error", "error", err)
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
	bearerAuth := auth.BearerAuth(authHandler, auth.BearerAuthConfig{
		Name:        "BearerAuth",
		Format:      "JWT",
		Description: "Bearer token authentication",
	})

	// Simple echo endpoint (no authentication) with middleware
	// Middleware generates fresh traceID per message and injects logger with context
	// Usage: ws://localhost:8080/ws/echo
	app.Router.GET("/ws/echo", websocket.Handler(
		echoCallbacks,
		websocket.WithMiddleware(
			middleware.TraceID(), // Fresh traceID per callback
			middleware.Logger(),  // Logger with connectionID + traceID
		)))

	// Authenticated chat endpoint with middleware
	// Usage: ws://localhost:8080/ws/chat with header "Authorization: Bearer valid-token"
	app.Router.GET("/ws/chat", websocket.AuthHandler(
		chatCallbacks,
		bearerAuth,
		websocket.WithMiddleware(
			middleware.TraceID(),
			middleware.Logger(),
		)))

	slog.Info("Starting server with WebSocket support on :8080")
	slog.Info("")
	slog.Info("Available endpoints:")
	slog.Info("  Echo: ws://localhost:8080/ws/echo")
	slog.Info("  Chat: ws://localhost:8080/ws/chat (requires Bearer token: 'valid-token')")

	app.Start()
}
