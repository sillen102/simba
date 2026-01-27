package main

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/sillen102/simba"
)

// MockDistributedRegistry simulates a distributed connection registry
// that could be backed by Redis, Cassandra, or similar in production.
//
// This implementation demonstrates the key concepts:
// - Local connection storage (actual WebSocket connections)
// - Simulated distributed state (groups, metadata)
// - Cross-instance broadcast simulation
//
// In production, replace with actual distributed storage and pub/sub.
type MockDistributedRegistry struct {
	instanceID string

	// Local storage for THIS instance's connections
	mu          sync.RWMutex
	connections map[string]*simba.WebSocketConnection

	// Simulated distributed state (in production, this would be Redis/Cassandra)
	distributedMu     sync.RWMutex
	distributedGroups map[string]map[string]InstanceConnection // group -> connID -> instance info
	connGroups        map[string]map[string]bool               // connID -> groups set
}

// InstanceConnection represents connection metadata stored in distributed state
type InstanceConnection struct {
	InstanceID string
	ConnID     string
}

// NewMockDistributedRegistry creates a new mock distributed registry
func NewMockDistributedRegistry(instanceID string) simba.ConnectionRegistryInternal {
	slog.Info("Creating MockDistributedRegistry", "instanceID", instanceID)
	return &MockDistributedRegistry{
		instanceID:        instanceID,
		connections:       make(map[string]*simba.WebSocketConnection),
		distributedGroups: make(map[string]map[string]InstanceConnection),
		connGroups:        make(map[string]map[string]bool),
	}
}

// ============================================================================
// Internal Methods - Called by Simba Framework
// ============================================================================

// AddConnection registers a new connection when WebSocket upgrade succeeds
func (r *MockDistributedRegistry) AddConnection(conn *simba.WebSocketConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store the actual connection locally (only this instance can write to it)
	r.connections[conn.ID] = conn

	slog.Debug("Connection added to registry",
		"instanceID", r.instanceID,
		"connID", conn.ID,
		"localCount", len(r.connections),
	)
}

// RemoveConnection unregisters a connection and removes it from all groups
func (r *MockDistributedRegistry) RemoveConnection(connID string) {
	r.mu.Lock()
	delete(r.connections, connID)
	localCount := len(r.connections)
	r.mu.Unlock()

	// Remove from distributed state
	r.distributedMu.Lock()
	defer r.distributedMu.Unlock()

	// Remove from all groups
	if groups, exists := r.connGroups[connID]; exists {
		for group := range groups {
			delete(r.distributedGroups[group], connID)
			if len(r.distributedGroups[group]) == 0 {
				delete(r.distributedGroups, group)
			}
		}
		delete(r.connGroups, connID)
	}

	slog.Debug("Connection removed from registry",
		"instanceID", r.instanceID,
		"connID", connID,
		"localCount", localCount,
	)
}

// ============================================================================
// ConnectionRegistry Interface - Public Methods
// ============================================================================

// Join adds a connection to a group (stored in distributed state)
func (r *MockDistributedRegistry) Join(connID, group string) error {
	r.distributedMu.Lock()
	defer r.distributedMu.Unlock()

	// Initialize group if needed
	if r.distributedGroups[group] == nil {
		r.distributedGroups[group] = make(map[string]InstanceConnection)
	}

	// Add to distributed group
	r.distributedGroups[group][connID] = InstanceConnection{
		InstanceID: r.instanceID,
		ConnID:     connID,
	}

	// Track groups per connection
	if r.connGroups[connID] == nil {
		r.connGroups[connID] = make(map[string]bool)
	}
	r.connGroups[connID][group] = true

	slog.Info("Connection joined group",
		"instanceID", r.instanceID,
		"connID", connID,
		"group", group,
		"groupSize", len(r.distributedGroups[group]),
	)

	return nil
}

// Leave removes a connection from a group
func (r *MockDistributedRegistry) Leave(connID, group string) error {
	r.distributedMu.Lock()
	defer r.distributedMu.Unlock()

	if r.distributedGroups[group] != nil {
		delete(r.distributedGroups[group], connID)
		if len(r.distributedGroups[group]) == 0 {
			delete(r.distributedGroups, group)
		}
	}

	if r.connGroups[connID] != nil {
		delete(r.connGroups[connID], group)
	}

	slog.Info("Connection left group",
		"instanceID", r.instanceID,
		"connID", connID,
		"group", group,
	)

	return nil
}

// LeaveAll removes a connection from all groups
func (r *MockDistributedRegistry) LeaveAll(connID string) error {
	r.distributedMu.Lock()
	defer r.distributedMu.Unlock()

	if groups, exists := r.connGroups[connID]; exists {
		for group := range groups {
			delete(r.distributedGroups[group], connID)
			if len(r.distributedGroups[group]) == 0 {
				delete(r.distributedGroups, group)
			}
		}
		delete(r.connGroups, connID)
	}

	return nil
}

// Groups returns all groups a connection belongs to
func (r *MockDistributedRegistry) Groups(connID string) []string {
	r.distributedMu.RLock()
	defer r.distributedMu.RUnlock()

	result := make([]string, 0)
	if groups, exists := r.connGroups[connID]; exists {
		for group := range groups {
			result = append(result, group)
		}
	}
	return result
}

// BroadcastToGroup sends a binary message to all connections in a group
// This demonstrates the KEY distributed concept: only write to LOCAL connections
func (r *MockDistributedRegistry) BroadcastToGroup(group string, data []byte) error {
	// Get all connection IDs in the group from distributed state
	r.distributedMu.RLock()
	instanceConnections := make(map[string]InstanceConnection)
	if groupMembers, exists := r.distributedGroups[group]; exists {
		for connID, instConn := range groupMembers {
			instanceConnections[connID] = instConn
		}
	}
	r.distributedMu.RUnlock()

	// Log the broadcast attempt
	slog.Info("Broadcasting to group",
		"instanceID", r.instanceID,
		"group", group,
		"totalInGroup", len(instanceConnections),
	)

	// Get LOCAL connections that are in this group
	r.mu.RLock()
	localConnections := make([]*simba.WebSocketConnection, 0)
	localCount := 0
	remoteCount := 0

	for connID, instConn := range instanceConnections {
		if instConn.InstanceID == r.instanceID {
			// This is OUR connection, we can write to it
			if conn, exists := r.connections[connID]; exists {
				localConnections = append(localConnections, conn)
				localCount++
			}
		} else {
			// This connection belongs to another instance
			// In production, you would use Redis pub/sub or similar
			remoteCount++
			slog.Debug("Skipping remote connection",
				"instanceID", r.instanceID,
				"remoteInstance", instConn.InstanceID,
				"connID", connID,
			)
		}
	}
	r.mu.RUnlock()

	// Write to LOCAL connections outside the lock
	var errs []error
	for _, conn := range localConnections {
		if err := conn.WriteBinary(data); err != nil {
			errs = append(errs, err)
			slog.Error("Failed to write to connection",
				"instanceID", r.instanceID,
				"connID", conn.ID,
				"error", err,
			)
		}
	}

	slog.Info("Broadcast complete",
		"instanceID", r.instanceID,
		"group", group,
		"localSent", localCount,
		"remotePending", remoteCount,
		"errors", len(errs),
	)

	if len(errs) > 0 {
		return fmt.Errorf("failed to broadcast to %d/%d local connections", len(errs), localCount)
	}
	return nil
}

// BroadcastToGroupText sends a text message to all connections in a group
func (r *MockDistributedRegistry) BroadcastToGroupText(group string, msg string) error {
	// Get all connection IDs in the group from distributed state
	r.distributedMu.RLock()
	instanceConnections := make(map[string]InstanceConnection)
	if groupMembers, exists := r.distributedGroups[group]; exists {
		for connID, instConn := range groupMembers {
			instanceConnections[connID] = instConn
		}
	}
	r.distributedMu.RUnlock()

	// Get LOCAL connections
	r.mu.RLock()
	localConnections := make([]*simba.WebSocketConnection, 0)
	for connID, instConn := range instanceConnections {
		if instConn.InstanceID == r.instanceID {
			if conn, exists := r.connections[connID]; exists {
				localConnections = append(localConnections, conn)
			}
		}
	}
	r.mu.RUnlock()

	// Write to LOCAL connections
	var errs []error
	for _, conn := range localConnections {
		if err := conn.WriteText(msg); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to broadcast to %d/%d connections", len(errs), len(localConnections))
	}
	return nil
}

// BroadcastToAll sends a binary message to all connections (across all instances)
func (r *MockDistributedRegistry) BroadcastToAll(data []byte) error {
	r.mu.RLock()
	connections := make([]*simba.WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		connections = append(connections, conn)
	}
	r.mu.RUnlock()

	slog.Info("Broadcasting to all LOCAL connections",
		"instanceID", r.instanceID,
		"count", len(connections),
	)

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

// BroadcastToAllText sends a text message to all connections
func (r *MockDistributedRegistry) BroadcastToAllText(msg string) error {
	r.mu.RLock()
	connections := make([]*simba.WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		connections = append(connections, conn)
	}
	r.mu.RUnlock()

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

// Get retrieves a connection by ID (only local connections)
func (r *MockDistributedRegistry) Get(id string) *simba.WebSocketConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connections[id]
}

// All returns all LOCAL connections
func (r *MockDistributedRegistry) All() []*simba.WebSocketConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*simba.WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		result = append(result, conn)
	}
	return result
}

// Filter returns LOCAL connections that match the predicate function
func (r *MockDistributedRegistry) Filter(fn func(*simba.WebSocketConnection) bool) []*simba.WebSocketConnection {
	r.mu.RLock()
	conns := make([]*simba.WebSocketConnection, 0, len(r.connections))
	for _, conn := range r.connections {
		conns = append(conns, conn)
	}
	r.mu.RUnlock()

	// Filter outside the lock
	result := make([]*simba.WebSocketConnection, 0)
	for _, conn := range conns {
		if fn(conn) {
			result = append(result, conn)
		}
	}
	return result
}

// Count returns the number of LOCAL connections
func (r *MockDistributedRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.connections)
}

// GroupCount returns the TOTAL number of connections in a group (across all instances)
func (r *MockDistributedRegistry) GroupCount(group string) int {
	r.distributedMu.RLock()
	defer r.distributedMu.RUnlock()
	return len(r.distributedGroups[group])
}
