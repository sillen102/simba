# WebSocket Distributed Example

This example demonstrates how to implement a **custom connection registry** for distributed WebSocket deployments using the Simba framework.

## What This Example Shows

This example includes:

1. **MockDistributedRegistry**: A simulated distributed registry that demonstrates:
   - Local connection storage (actual WebSocket connections)
   - Distributed state management (groups/rooms across instances)
   - Cross-instance broadcast patterns
   - Instance-aware connection tracking

2. **Key Concepts**:
   - Only write to LOCAL connections (each instance manages its own)
   - Store group membership in distributed state
   - Simulate what would be Redis/Cassandra in production
   - Proper separation of local vs distributed state

## Why Use a Custom Registry?

Use a custom registry when you need:

- **Multiple instances**: Run your app across multiple servers
- **Load balancing**: Distribute WebSocket connections
- **Broadcasting across instances**: Send messages to users on different servers
- **Persistent state**: Maintain connection info across deployments
- **Horizontal scaling**: Scale WebSocket connections independently

## Architecture

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│ Instance 1  │         │ Instance 2  │         │ Instance 3  │
│             │         │             │         │             │
│ Local Conns │         │ Local Conns │         │ Local Conns │
│  - conn-1   │         │  - conn-3   │         │  - conn-5   │
│  - conn-2   │         │  - conn-4   │         │  - conn-6   │
└──────┬──────┘         └──────┬──────┘         └──────┬──────┘
       │                       │                       │
       └───────────────────────┴───────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │  Distributed State  │
                    │    (Redis/etc)      │
                    │                     │
                    │ Groups:             │
                    │  room-1:            │
                    │    - conn-1 (inst1) │
                    │    - conn-3 (inst2) │
                    │    - conn-5 (inst3) │
                    └─────────────────────┘
```

## Running the Example

```bash
cd examples/websocket-distributed
go run .
```

The server will start on `http://localhost:8080`

## Testing the Example

### Terminal 1: Connect to chat
```bash
# Connect with authentication
wscat -c "ws://localhost:8080/ws/chat/lobby" -H "Authorization: Bearer valid-token"
```

### Terminal 2: Connect to same room
```bash
# Connect another user to the same room
wscat -c "ws://localhost:8080/ws/chat/lobby" -H "Authorization: Bearer valid-token"
```

### Terminal 3: Simple echo
```bash
# Simple echo without authentication
wscat -c "ws://localhost:8080/ws/simple"
```

When you send a message from Terminal 1, you'll see:
- Terminal 1: Sees the broadcast message
- Terminal 2: Sees the broadcast message (distributed to same room)
- Terminal 3: Only sees its own echo (different endpoint)

The console logs will show:
```
Broadcasting to group instanceID=instance-1 group=lobby totalInGroup=2
Broadcast complete instanceID=instance-1 group=lobby localSent=2 remotePending=0
```

## Adapting for Production

To use this in production with real distributed storage:

### 1. Replace Mock with Redis

```go
package myapp

import (
    "github.com/redis/go-redis/v9"
    "github.com/sillen102/simba"
)

type RedisConnectionRegistry struct {
    client     *redis.Client
    instanceID string
    localConns map[string]*simba.WebSocketConnection
    mu         sync.RWMutex
}

func NewRedisConnectionRegistry(client *redis.Client, instanceID string) simba.ConnectionRegistryInternal {
    return &RedisConnectionRegistry{
        client:     client,
        instanceID: instanceID,
        localConns: make(map[string]*simba.WebSocketConnection),
    }
}

// Implement all methods using Redis for distributed state
func (r *RedisConnectionRegistry) Join(connID, group string) error {
    // Add to Redis set
    return r.client.SAdd(ctx, "ws:group:"+group,
        fmt.Sprintf("%s:%s", r.instanceID, connID)).Err()
}

func (r *RedisConnectionRegistry) BroadcastToGroup(group string, data []byte) error {
    // Get all connections in group
    members, _ := r.client.SMembers(ctx, "ws:group:"+group).Result()

    // Only write to LOCAL connections
    for _, member := range members {
        parts := strings.Split(member, ":")
        if parts[0] == r.instanceID {
            connID := parts[1]
            if conn := r.localConns[connID]; conn != nil {
                conn.WriteBinary(data)
            }
        }
    }

    // For remote instances, use Redis pub/sub
    r.client.Publish(ctx, "ws:broadcast:"+group, data)
    return nil
}
```

### 2. Add Redis Pub/Sub for Cross-Instance Broadcasting

```go
func (r *RedisConnectionRegistry) listenForBroadcasts() {
    pubsub := r.client.Subscribe(ctx, "ws:broadcast:*")

    for msg := range pubsub.Channel() {
        // Parse group from channel name
        group := strings.TrimPrefix(msg.Channel, "ws:broadcast:")

        // Broadcast to LOCAL connections in this group
        r.mu.RLock()
        for _, conn := range r.localConns {
            groups := r.Groups(conn.ID)
            if contains(groups, group) {
                conn.WriteBinary([]byte(msg.Payload))
            }
        }
        r.mu.RUnlock()
    }
}
```

### 3. Use in Your Application

```go
func main() {
    // Create Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Create custom registry
    instanceID := os.Getenv("INSTANCE_ID") // or generate UUID
    registry := myapp.NewRedisConnectionRegistry(redisClient, instanceID)

    // Use with handlers
    app.Router.GET("/ws/chat/{room}", simba.AuthWebSocketHandler(
        chatCallbacks(),
        bearerAuth,
        registry, // Your custom distributed registry
    ))
}
```

## Key Implementation Points

### 1. Separate Local and Distributed State

**Local State** (this instance only):
- Actual `*WebSocketConnection` objects
- Can only write to connections in this map
- Fast, no network calls

**Distributed State** (shared across instances):
- Connection metadata (ID, params, auth)
- Group membership
- Stored in Redis/Cassandra/etc
- Requires network calls

### 2. Broadcast Pattern

```go
// 1. Query distributed state for group members
members := redis.SMembers("ws:group:" + group)

// 2. Filter for LOCAL connections
for member in members {
    if member.instanceID == myInstanceID {
        // Write to local connection
        localConns[member.connID].Write(data)
    } else {
        // Skip - other instance will handle it
        // OR use pub/sub to notify other instance
    }
}
```

### 3. Cleanup is Critical

Ensure connections are removed from distributed state:
- In `remove()` method (called automatically by framework)
- On instance shutdown (cleanup handler)
- With TTLs in Redis (fallback for crashes)

### 4. Handle Network Failures

```go
func (r *RedisRegistry) Join(connID, group string) error {
    err := r.client.SAdd(ctx, "ws:group:"+group, connID).Err()
    if err != nil {
        // Redis is down - what to do?
        // Option 1: Fall back to local-only mode
        // Option 2: Return error and close connection
        // Option 3: Queue for retry
        return fmt.Errorf("failed to join group: %w", err)
    }
    return nil
}
```

## Performance Considerations

1. **Local writes are fast**: No network overhead
2. **Distributed queries add latency**: Cache when possible
3. **Broadcasting scales**: Each instance only writes to its connections
4. **Group lookups**: Consider caching group membership
5. **Connection metadata**: Store minimal data in distributed state

## Testing Distributed Behavior

To truly test distributed behavior:

1. **Run multiple instances**:
   ```bash
   # Terminal 1
   INSTANCE_ID=inst1 PORT=8080 go run .

   # Terminal 2
   INSTANCE_ID=inst2 PORT=8081 go run .
   ```

2. **Use a load balancer**:
   ```bash
   # nginx or similar to distribute connections
   ```

3. **Connect to different instances**:
   ```bash
   # Connect to instance 1
   wscat -c "ws://localhost:8080/ws/chat/lobby" -H "Authorization: Bearer valid-token"

   # Connect to instance 2
   wscat -c "ws://localhost:8081/ws/chat/lobby" -H "Authorization: Bearer valid-token"
   ```

4. **Verify cross-instance messaging**: Messages should flow between instances

## Common Pitfalls

1. ❌ **Writing to remote connections**: Only write to local connections
2. ❌ **Forgetting cleanup**: Always remove from distributed state on disconnect
3. ❌ **Blocking on distributed ops**: Use timeouts, avoid holding locks
4. ❌ **Not handling Redis failures**: Have a fallback strategy
5. ❌ **Storing connection objects in Redis**: Only store metadata (ID, params)

## Further Reading

- [Redis Pub/Sub Documentation](https://redis.io/docs/manual/pubsub/)
- [Distributed Systems Patterns](https://martinfowler.com/articles/patterns-of-distributed-systems/)
- [WebSocket Scaling Strategies](https://www.nginx.com/blog/websocket-nginx/)

## Questions?

See the main WebSocket documentation at `.claude/plans/WEBSOCKET_IMPLEMENTATION.md` for more details on the registry interface and design considerations.
