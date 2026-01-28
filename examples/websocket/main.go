package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
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

// ConnectionRegistry demonstrates how to manage connections externally.
// In a real multi-instance setup, you would use Redis or similar.
type ConnectionRegistry struct {
	mu    sync.RWMutex
	conns map[string]*simba.WebSocketConnection // connID -> connection
	users map[int][]string                      // userID -> []connID
}

func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		conns: make(map[string]*simba.WebSocketConnection),
		users: make(map[int][]string),
	}
}

func (r *ConnectionRegistry) Add(userID int, conn *simba.WebSocketConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[conn.ID] = conn
	r.users[userID] = append(r.users[userID], conn.ID)
	slog.Info("Connection registered", "connID", conn.ID, "userID", userID)
}

func (r *ConnectionRegistry) Remove(userID int, connID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, connID)

	// Remove from user's connection list
	connIDs := r.users[userID]
	for i, id := range connIDs {
		if id == connID {
			r.users[userID] = append(connIDs[:i], connIDs[i+1:]...)
			break
		}
	}
	if len(r.users[userID]) == 0 {
		delete(r.users, userID)
	}
	slog.Info("Connection unregistered", "connID", connID, "userID", userID)
}

func (r *ConnectionRegistry) Broadcast(message string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, conn := range r.conns {
		conn.WriteText(message)
	}
}

func (r *ConnectionRegistry) SendToUser(userID int, message string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, connID := range r.users[userID] {
		if conn, ok := r.conns[connID]; ok {
			conn.WriteText(message)
		}
	}
}

func (r *ConnectionRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.conns)
}

// echoCallbacks returns WebSocket callbacks for a simple echo handler
func echoCallbacks() simba.WebSocketCallbacks[simbaModels.NoParams] {
	return simba.WebSocketCallbacks[simbaModels.NoParams]{
		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, params simbaModels.NoParams) error {
			slog.Info("Connection established", "connID", conn.ID)
			return conn.WriteText("Welcome! Send me messages and I'll echo them back.")
		},

		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, msgType simba.MessageType, data []byte) error {
			slog.Info("Received message", "connID", conn.ID, "message", string(data))
			return conn.WriteText(fmt.Sprintf("Echo: %s", string(data)))
		},

		OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, err error) {
			slog.Info("Connection closed", "connID", connID, "error", err)
		},

		OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
			slog.Error("Error occurred", "connID", conn.ID, "error", err)
			return false // Close connection on error
		},
	}
}

// chatCallbacks demonstrates authenticated WebSocket with external registry
func chatCallbacks(registry *ConnectionRegistry) simba.AuthWebSocketCallbacks[simbaModels.NoParams, User] {
	return simba.AuthWebSocketCallbacks[simbaModels.NoParams, User]{
		OnConnect: func(ctx context.Context, conn *simba.WebSocketConnection, params simbaModels.NoParams, user User) error {
			// Register connection in external registry
			registry.Add(user.ID, conn)

			// Send welcome to the user
			conn.WriteText(fmt.Sprintf("Welcome %s! (connID: %s)", user.Name, conn.ID))

			// Notify all other connections
			registry.Broadcast(fmt.Sprintf("%s joined the chat", user.Name))

			return nil
		},

		OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, msgType simba.MessageType, data []byte, user User) error {
			slog.Info("Chat message", "user", user.Name, "message", string(data))

			// Broadcast to all connections
			message := fmt.Sprintf("[%s]: %s", user.Name, string(data))
			registry.Broadcast(message)

			return nil
		},

		OnDisconnect: func(ctx context.Context, connID string, params simbaModels.NoParams, user User, err error) {
			// Clean up external registry
			registry.Remove(user.ID, connID)

			slog.Info("User disconnected", "user", user.Name, "connID", connID, "error", err)
		},

		OnError: func(ctx context.Context, conn *simba.WebSocketConnection, err error) bool {
			slog.Error("Chat error", "connID", conn.ID, "error", err)
			return false // Close on error
		},
	}
}

func main() {
	// Set up logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	app := simba.Default()

	// Create connection registry (in production, use Redis or similar)
	registry := NewConnectionRegistry()

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
	// Usage: ws://localhost:8080/ws/chat with header "Authorization: Bearer valid-token"
	app.Router.GET("/ws/chat", simba.AuthWebSocketHandler(chatCallbacks(registry), bearerAuth))

	slog.Info("Starting server with WebSocket support on :8080")
	slog.Info("")
	slog.Info("Available endpoints:")
	slog.Info("  Echo: ws://localhost:8080/ws/echo")
	slog.Info("  Chat: ws://localhost:8080/ws/chat (requires Bearer token)")
	slog.Info("")
	slog.Info("The chat example demonstrates external connection registry pattern.")
	slog.Info("In production, replace ConnectionRegistry with Redis for multi-instance support.")

	app.Start()
}
