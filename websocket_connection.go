package simba

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/gobwas/ws/wsutil"
)

// WebSocketConnection wraps a WebSocket connection with thread-safe operations
type WebSocketConnection struct {
	// ID is a unique identifier for this connection
	ID string

	// Params contains the parsed route/query/header parameters
	Params any

	// Auth contains the authentication model (nil if unauthenticated)
	Auth any

	conn   net.Conn
	ctx    context.Context
	cancel context.CancelFunc

	// writeMu protects concurrent writes to the connection
	writeMu sync.Mutex
}

// WriteText sends a text message to the client (thread-safe)
func (wc *WebSocketConnection) WriteText(msg string) error {
	wc.writeMu.Lock()
	defer wc.writeMu.Unlock()
	return wsutil.WriteServerText(wc.conn, []byte(msg))
}

// WriteBinary sends a binary message to the client (thread-safe)
func (wc *WebSocketConnection) WriteBinary(data []byte) error {
	wc.writeMu.Lock()
	defer wc.writeMu.Unlock()
	return wsutil.WriteServerBinary(wc.conn, data)
}

// WriteJSON marshals v to JSON and sends it as a text message (thread-safe)
func (wc *WebSocketConnection) WriteJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	wc.writeMu.Lock()
	defer wc.writeMu.Unlock()
	return wsutil.WriteServerText(wc.conn, data)
}

// Close closes the WebSocket connection (thread-safe)
func (wc *WebSocketConnection) Close() error {
	wc.cancel() // Cancel context to signal shutdown
	return wc.conn.Close()
}

// Context returns the connection's context
// The context is cancelled when the connection is closed
func (wc *WebSocketConnection) Context() context.Context {
	return wc.ctx
}

// ConnectionRegistry manages WebSocket connections and provides broadcasting capabilities
// All methods are thread-safe and can be called concurrently
type ConnectionRegistry interface {
	// Join adds a connection to a group (e.g., chat room, topic)
	Join(connID, group string) error

	// Leave removes a connection from a group
	Leave(connID, group string) error

	// LeaveAll removes a connection from all groups
	LeaveAll(connID string) error

	// Groups returns all groups a connection belongs to
	Groups(connID string) []string

	// BroadcastToGroup sends a binary message to all connections in a group
	BroadcastToGroup(group string, data []byte) error

	// BroadcastToGroupText sends a text message to all connections in a group
	BroadcastToGroupText(group string, msg string) error

	// BroadcastToAll sends a binary message to all connections
	BroadcastToAll(data []byte) error

	// BroadcastToAllText sends a text message to all connections
	BroadcastToAllText(msg string) error

	// Get retrieves a connection by ID
	Get(id string) *WebSocketConnection

	// All returns all active connections
	All() []*WebSocketConnection

	// Filter returns connections that match the predicate function
	Filter(fn func(*WebSocketConnection) bool) []*WebSocketConnection

	// Count returns the total number of active connections
	Count() int

	// GroupCount returns the number of connections in a group
	GroupCount(group string) int
}

// connectionRegistry is the internal implementation of ConnectionRegistry
type connectionRegistry struct {
	mu          sync.RWMutex
	connections map[string]*WebSocketConnection
	groups      map[string]map[string]bool // group -> set of connection IDs
}

// newConnectionRegistry creates a new connection registry
func newConnectionRegistry() *connectionRegistry {
	return &connectionRegistry{
		connections: make(map[string]*WebSocketConnection),
		groups:      make(map[string]map[string]bool),
	}
}

// add adds a connection to the registry (internal method)
func (r *connectionRegistry) add(conn *WebSocketConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[conn.ID] = conn
}

// remove removes a connection from the registry and all groups (internal method)
func (r *connectionRegistry) remove(connID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.connections, connID)

	// Remove from all groups
	for group := range r.groups {
		delete(r.groups[group], connID)
		if len(r.groups[group]) == 0 {
			delete(r.groups, group)
		}
	}
}

// Join implements ConnectionRegistry.Join
func (r *connectionRegistry) Join(connID, group string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.connections[connID]; !exists {
		return fmt.Errorf("connection %s not found", connID)
	}

	if r.groups[group] == nil {
		r.groups[group] = make(map[string]bool)
	}
	r.groups[group][connID] = true
	return nil
}

// Leave implements ConnectionRegistry.Leave
func (r *connectionRegistry) Leave(connID, group string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.groups[group] != nil {
		delete(r.groups[group], connID)
		if len(r.groups[group]) == 0 {
			delete(r.groups, group)
		}
	}
	return nil
}

// LeaveAll implements ConnectionRegistry.LeaveAll
func (r *connectionRegistry) LeaveAll(connID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for group := range r.groups {
		delete(r.groups[group], connID)
		if len(r.groups[group]) == 0 {
			delete(r.groups, group)
		}
	}
	return nil
}

// Groups implements ConnectionRegistry.Groups
func (r *connectionRegistry) Groups(connID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0)
	for group, members := range r.groups {
		if members[connID] {
			result = append(result, group)
		}
	}
	return result
}

// BroadcastToGroup implements ConnectionRegistry.BroadcastToGroup
func (r *connectionRegistry) BroadcastToGroup(group string, data []byte) error {
	// Snapshot connections while holding read lock
	r.mu.RLock()
	connIDs := make([]string, 0, len(r.groups[group]))
	for id := range r.groups[group] {
		connIDs = append(connIDs, id)
	}
	connections := make([]*WebSocketConnection, 0, len(connIDs))
	for _, id := range connIDs {
		if conn := r.connections[id]; conn != nil {
			connections = append(connections, conn)
		}
	}
	r.mu.RUnlock()

	// Write to connections outside the lock
	// Each connection has its own write mutex
	var errs []error
	for _, conn := range connections {
		if err := conn.WriteBinary(data); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to broadcast to %d/%d connections", len(errs), len(connections))
	}
	return nil
}

// BroadcastToGroupText implements ConnectionRegistry.BroadcastToGroupText
func (r *connectionRegistry) BroadcastToGroupText(group string, msg string) error {
	// Snapshot connections while holding read lock
	r.mu.RLock()
	connIDs := make([]string, 0, len(r.groups[group]))
	for id := range r.groups[group] {
		connIDs = append(connIDs, id)
	}
	connections := make([]*WebSocketConnection, 0, len(connIDs))
	for _, id := range connIDs {
		if conn := r.connections[id]; conn != nil {
			connections = append(connections, conn)
		}
	}
	r.mu.RUnlock()

	// Write to connections outside the lock
	var errs []error
	for _, conn := range connections {
		if err := conn.WriteText(msg); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to broadcast to %d/%d connections", len(errs), len(connections))
	}
	return nil
}

// BroadcastToAll implements ConnectionRegistry.BroadcastToAll
func (r *connectionRegistry) BroadcastToAll(data []byte) error {
	// Snapshot connections while holding read lock
	r.mu.RLock()
	connections := make([]*WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		connections = append(connections, conn)
	}
	r.mu.RUnlock()

	// Write to connections outside the lock
	var errs []error
	for _, conn := range connections {
		if err := conn.WriteBinary(data); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to broadcast to %d/%d connections", len(errs), len(connections))
	}
	return nil
}

// BroadcastToAllText implements ConnectionRegistry.BroadcastToAllText
func (r *connectionRegistry) BroadcastToAllText(msg string) error {
	// Snapshot connections while holding read lock
	r.mu.RLock()
	connections := make([]*WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		connections = append(connections, conn)
	}
	r.mu.RUnlock()

	// Write to connections outside the lock
	var errs []error
	for _, conn := range connections {
		if err := conn.WriteText(msg); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to broadcast to %d/%d connections", len(errs), len(connections))
	}
	return nil
}

// Get implements ConnectionRegistry.Get
func (r *connectionRegistry) Get(id string) *WebSocketConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connections[id]
}

// All implements ConnectionRegistry.All
func (r *connectionRegistry) All() []*WebSocketConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		result = append(result, conn)
	}
	return result
}

// Filter implements ConnectionRegistry.Filter
func (r *connectionRegistry) Filter(fn func(*WebSocketConnection) bool) []*WebSocketConnection {
	// Snapshot connections while holding read lock
	r.mu.RLock()
	conns := make([]*WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		conns = append(conns, conn)
	}
	r.mu.RUnlock()

	// Filter outside the lock
	result := make([]*WebSocketConnection, 0)
	for _, conn := range conns {
		if fn(conn) {
			result = append(result, conn)
		}
	}
	return result
}

// Count implements ConnectionRegistry.Count
func (r *connectionRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.connections)
}

// GroupCount implements ConnectionRegistry.GroupCount
func (r *connectionRegistry) GroupCount(group string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.groups[group])
}
