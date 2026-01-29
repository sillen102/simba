# WebSocket Modularization Plan

## Executive Summary

This plan describes how to extract WebSocket functionality from the main `simba` package into a separate, optional Go module `github.com/sillen102/simba-websocket`. This approach eliminates the need for build flags while maintaining clean separation and ease of use.

**Recommended Approach**: **Option 1 - Separate Module with One-Way Dependency**

This is the cleanest, most idiomatic Go solution that aligns with standard Go practices.

---

## Analysis of Current Implementation

### Current WebSocket Files (in `simba` package)
1. **`websocket_handler.go`** (419 lines) - Main handler implementations
2. **`websocket_callbacks.go`** (60 lines) - Callback type definitions
3. **`websocket_connection.go`** (65 lines) - Connection wrapper
4. **`middleware/websocket_logger.go`** - WebSocket-specific middleware
5. **`middleware/websocket_trace_id.go`** - WebSocket-specific middleware

### Current Dependencies
- **External**: `github.com/gobwas/ws`, `github.com/google/uuid`
- **Internal Simba packages**: 
  - `simbaContext` (context keys)
  - `simbaErrors` (error handling)
  - `simbaModels` (NoParams, NoBody)
  - Core `simba` package (Handler interface, AuthHandler, params parsing)

### Key Integration Points
1. **Handler Interface Implementation** - WebSocket handlers implement `simba.Handler`
2. **Router Registration** - Uses `Router.GET()`, etc.
3. **Authentication** - Reuses `simba.AuthHandler[AuthModel]` interface
4. **Parameter Parsing** - Uses internal `parseAndValidateParams[Params]()`
5. **Context Keys** - Uses `simbaContext.{ConnectionIDKey, TraceIDKey, LoggerKey}`

---

## Recommended Solution: Separate Module with One-Way Dependency

### Architecture Overview

```
┌─────────────────────────────────────┐
│  github.com/sillen102/simba         │
│  (Core Framework)                    │
│  - Handler interface                 │
│  - Router                            │
│  - Auth system                       │
│  - Params parsing                    │
│  - Context keys                      │
│  - Error handling                    │
└─────────────────────────────────────┘
              ▲
              │ imports
              │
┌─────────────────────────────────────┐
│  github.com/sillen102/simba-websocket│
│  (WebSocket Extension)               │
│  - WebSocket handlers                │
│  - WebSocket callbacks               │
│  - WebSocket connection              │
│  - WebSocket middleware              │
└─────────────────────────────────────┘
```

### Why This Approach?

1. **Standard Go Practice** - Extension modules importing core modules is idiomatic
2. **No Build Flags Required** - Users just `import` when needed
3. **Clean Separation** - WebSocket code completely separate
4. **No Circular Dependencies** - One-way dependency is simple
5. **Backward Compatible** - Existing code can be migrated gradually
6. **Easy to Use** - Natural import experience

### Module Structure

#### New Module: `github.com/sillen102/simba-websocket`

```
simba-websocket/
├── go.mod                           # module github.com/sillen102/simba-websocket
├── go.sum
├── README.md
├── handler.go                       # WebSocketHandler, AuthWebSocketHandler
├── callbacks.go                     # WebSocketCallbacks, AuthWebSocketCallbacks
├── connection.go                    # WebSocketConnection
├── middleware.go                    # WebSocket middleware option types
├── handler_test.go
├── callbacks_test.go
├── connection_test.go
└── middleware/
    ├── logger.go                    # WebSocketLogger middleware
    ├── trace_id.go                  # WebSocketTraceID middleware
    ├── logger_test.go
    └── trace_id_test.go
```

---

## Detailed Design

### 1. Package Structure

#### `handler.go`
```go
package websocket

import (
    "context"
    "net"
    "net/http"
    
    "github.com/gobwas/ws"
    "github.com/gobwas/ws/wsutil"
    "github.com/google/uuid"
    "github.com/sillen102/simba"
    "github.com/sillen102/simba/simbaContext"
    "github.com/sillen102/simba/simbaErrors"
    "github.com/sillen102/simba/simbaModels"
)

// Middleware wraps a context to enrich it before callback invocations.
type Middleware func(ctx context.Context) context.Context

// HandlerOption is an option for configuring WebSocket handlers.
type HandlerOption interface {
    apply(handler any)
}

// Handler creates a handler that uses callbacks for WebSocket lifecycle events.
// Returns a simba.Handler that can be registered with Router.GET(), etc.
func Handler[Params any](
    callbacksFunc func() Callbacks[Params], 
    options ...HandlerOption,
) simba.Handler {
    // Implementation from websocket_handler.go
}

// AuthHandler creates an authenticated handler for WebSocket connections.
func AuthHandler[Params, AuthModel any](
    callbacksFunc func() AuthCallbacks[Params, AuthModel],
    authHandler simba.AuthHandler[AuthModel],
    options ...HandlerOption,
) simba.Handler {
    // Implementation from websocket_handler.go
}

// WithMiddleware adds middleware to the WebSocket handler.
func WithMiddleware(middleware ...Middleware) HandlerOption {
    // Implementation
}
```

#### `callbacks.go`
```go
package websocket

import "context"

// Callbacks defines the lifecycle callbacks for a WebSocket connection.
type Callbacks[Params any] struct {
    OnConnect    func(ctx context.Context, conn *Connection, params Params) error
    OnMessage    func(ctx context.Context, conn *Connection, data []byte) error
    OnDisconnect func(ctx context.Context, connID string, params Params, err error)
    OnError      func(ctx context.Context, conn *Connection, err error) bool
}

// AuthCallbacks defines the lifecycle callbacks for an authenticated WebSocket connection.
type AuthCallbacks[Params, AuthModel any] struct {
    OnConnect    func(ctx context.Context, conn *Connection, params Params, auth AuthModel) error
    OnMessage    func(ctx context.Context, conn *Connection, data []byte, auth AuthModel) error
    OnDisconnect func(ctx context.Context, connID string, params Params, auth AuthModel, err error)
    OnError      func(ctx context.Context, conn *Connection, err error) bool
}
```

#### `connection.go`
```go
package websocket

import (
    "encoding/json"
    "net"
    "sync"
    
    "github.com/gobwas/ws/wsutil"
)

// Connection represents an active WebSocket connection.
type Connection struct {
    ID   string
    conn net.Conn
    mu   sync.Mutex
}

func (c *Connection) WriteText(msg string) error { /* ... */ }
func (c *Connection) WriteBinary(data []byte) error { /* ... */ }
func (c *Connection) WriteJSON(v any) error { /* ... */ }
func (c *Connection) Close() error { /* ... */ }
```

#### `middleware/logger.go`
```go
package middleware

import (
    "context"
    "log/slog"
    
    "github.com/sillen102/simba/simbaContext"
    websocket "github.com/sillen102/simba-websocket"
)

// Logger injects a logger with connectionID and traceID into the context.
func Logger() websocket.Middleware {
    return func(ctx context.Context) context.Context {
        // Implementation from middleware/websocket_logger.go
    }
}
```

#### `middleware/trace_id.go`
```go
package middleware

import (
    "context"
    
    "github.com/google/uuid"
    "github.com/sillen102/simba/simbaContext"
    websocket "github.com/sillen102/simba-websocket"
)

// TraceID generates a fresh trace ID for each WebSocket callback invocation.
func TraceID() websocket.Middleware {
    return func(ctx context.Context) context.Context {
        // Implementation from middleware/websocket_trace_id.go
    }
}
```

### 2. User Experience

#### Before (Current)
```go
import "github.com/sillen102/simba"

app.Router.GET("/ws/echo", simba.WebSocketHandler(
    echoCallbacks,
    simba.WithWebsocketMiddleware(
        middleware.WebSocketTraceID(),
        middleware.WebSocketLogger(),
    ),
))
```

#### After (Separate Module)
```go
import (
    "github.com/sillen102/simba"
    websocket "github.com/sillen102/simba-websocket"
    wsmiddleware "github.com/sillen102/simba-websocket/middleware"
)

app.Router.GET("/ws/echo", websocket.Handler(
    echoCallbacks,
    websocket.WithMiddleware(
        wsmiddleware.TraceID(),
        wsmiddleware.Logger(),
    ),
))
```

### 3. Callback Definition Changes

#### Before
```go
func echoCallbacks() simba.WebSocketCallbacks[simbaModels.NoParams] {
    return simba.WebSocketCallbacks[simbaModels.NoParams]{
        OnMessage: func(ctx context.Context, conn *simba.WebSocketConnection, data []byte) error {
            return conn.WriteText("Echo: " + string(data))
        },
    }
}
```

#### After
```go
func echoCallbacks() websocket.Callbacks[simbaModels.NoParams] {
    return websocket.Callbacks[simbaModels.NoParams]{
        OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
            return conn.WriteText("Echo: " + string(data))
        },
    }
}
```

### 4. Authentication Integration

The separate module reuses simba's authentication system:

```go
import (
    "github.com/sillen102/simba"
    websocket "github.com/sillen102/simba-websocket"
)

bearerAuth := simba.BearerAuth(authFunc, simba.BearerAuthConfig{...})

app.Router.GET("/ws/chat", websocket.AuthHandler(
    chatCallbacks,
    bearerAuth,  // simba.AuthHandler[AuthModel]
))
```

### 5. go.mod Files

#### Main Simba `go.mod` (remains unchanged)
```go
module github.com/sillen102/simba

go 1.25

require (
    github.com/go-playground/validator/v10 v10.30.1
    github.com/google/uuid v1.6.0
    // ... other dependencies, NO websocket deps
)
```

#### New `simba-websocket/go.mod`
```go
module github.com/sillen102/simba-websocket

go 1.25

require (
    github.com/sillen102/simba v1.x.x  // Import main framework
    github.com/gobwas/ws v1.4.0
    github.com/google/uuid v1.6.0
)
```

---

## Migration Strategy

### Phase 1: Create Separate Repository/Module

1. Create new repository: `github.com/sillen102/simba-websocket`
2. Copy WebSocket files with package rename
3. Update imports to use `github.com/sillen102/simba`
4. Add go.mod with simba dependency
5. Run tests to ensure everything works
6. Publish v0.1.0 of simba-websocket

### Phase 2: Update Main Simba Repository

1. Add deprecation notices to existing WebSocket functions in simba
2. Update documentation to point to new module
3. Keep old code for 1-2 major versions for backward compatibility
4. Update examples to use new module

### Phase 3: Deprecation Timeline

**Version N (Current)**
- WebSocket code exists in simba
- Add deprecation warnings in documentation

**Version N+1**
- Mark WebSocket types as deprecated with Go doc comments
- Point users to simba-websocket module
- Keep functionality working

**Version N+2**
- Remove WebSocket code from simba
- Users must use simba-websocket module

### Backward Compatibility Bridge (Optional)

For smoother migration, simba could re-export websocket types:

```go
// websocket.go in main simba package
package simba

import websocket "github.com/sillen102/simba-websocket"

// Deprecated: Use github.com/sillen102/simba-websocket.Handler instead.
func WebSocketHandler[Params any](
    callbacksFunc func() websocket.Callbacks[Params],
    options ...websocket.HandlerOption,
) Handler {
    return websocket.Handler(callbacksFunc, options...)
}

// Deprecated: Use github.com/sillen102/simba-websocket.Callbacks instead.
type WebSocketCallbacks[Params any] = websocket.Callbacks[Params]

// Deprecated: Use github.com/sillen102/simba-websocket.Connection instead.
type WebSocketConnection = websocket.Connection
```

This allows existing code to continue working while users migrate.

---

## Files to Move

### From `simba` → `simba-websocket`

1. **`websocket_handler.go`** → `handler.go`
   - Rename package to `websocket`
   - Update `WebSocketHandler` → `Handler`
   - Update `AuthWebSocketHandler` → `AuthHandler`
   - Change `WebSocketMiddleware` → `Middleware`
   - Update imports to use `github.com/sillen102/simba`

2. **`websocket_callbacks.go`** → `callbacks.go`
   - Rename package to `websocket`
   - Update `WebSocketCallbacks` → `Callbacks`
   - Update `AuthWebSocketCallbacks` → `AuthCallbacks`

3. **`websocket_connection.go`** → `connection.go`
   - Rename package to `websocket`
   - Update `WebSocketConnection` → `Connection`

4. **`middleware/websocket_logger.go`** → `middleware/logger.go`
   - Rename function `WebSocketLogger()` → `Logger()`
   - Return type: `websocket.Middleware` (not simba function)

5. **`middleware/websocket_trace_id.go`** → `middleware/trace_id.go`
   - Rename function `WebSocketTraceID()` → `TraceID()`
   - Return type: `websocket.Middleware`

6. **Test files**
   - Move and update all WebSocket-related tests

### Keep in `simba` (No Changes)

- `simbaContext/keys.go` - Context keys used by both
- `simbaErrors/errors.go` - Error handling used by both
- `simbaModels/models.go` - NoParams, NoBody used by both
- `handler.go` - Handler interface
- `router.go` - Router
- `auth_handler.go` - Auth system
- `params.go` - Parameter parsing

---

## Code Changes Detail

### 1. Handler Implementation Changes

#### Internal Handler Structs

**Before** (in simba package):
```go
type WebSocketCallbackHandlerFunc[Params any] struct {
    callbacks  WebSocketCallbacks[Params]
    middleware []WebSocketMiddleware
}
```

**After** (in websocket package):
```go
type callbackHandlerFunc[Params any] struct {
    callbacks  Callbacks[Params]
    middleware []Middleware
}
```

Note: Make struct private (lowercase) as it's internal implementation.

#### Handler Interface Methods

The handlers must implement `simba.Handler` interface:

```go
func (h *callbackHandlerFunc[Params]) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (h *callbackHandlerFunc[Params]) getRequestBody() any
func (h *callbackHandlerFunc[Params]) getParams() any
func (h *callbackHandlerFunc[Params]) getResponseBody() any
func (h *callbackHandlerFunc[Params]) getAccepts() string
func (h *callbackHandlerFunc[Params]) getProduces() string
func (h *callbackHandlerFunc[Params]) getHandler() any
func (h *callbackHandlerFunc[Params]) getAuthModel() any
func (h *callbackHandlerFunc[Params]) getAuthHandler() any
```

These will now import and reference simba types:
```go
import "github.com/sillen102/simba/simbaModels"

func (h *callbackHandlerFunc[Params]) getRequestBody() any {
    return simbaModels.NoBody{}
}
```

### 2. Access to Internal Functions

WebSocket handlers need access to:

1. **`parseAndValidateParams[Params](r *http.Request)`** - Currently unexported
2. **`handleAuthRequest[AuthModel](...)`** - Currently unexported

#### Solution Options:

**Option A: Export These Functions** (Recommended)
In main simba package, export these utilities:

```go
// params.go
func ParseAndValidateParams[Params any](r *http.Request) (Params, error) {
    return parseAndValidateParams[Params](r)
}

// auth_handler.go  
func HandleAuthRequest[AuthModel any](
    authHandler AuthHandler[AuthModel],
    r *http.Request,
) (AuthModel, error) {
    return handleAuthRequest[AuthModel](authHandler, r)
}
```

Then websocket package uses:
```go
params, err := simba.ParseAndValidateParams[Params](r)
authModel, err := simba.HandleAuthRequest[AuthModel](h.authHandler, r)
```

**Option B: Duplicate Logic** - Not recommended, violates DRY

**Option C: Internal Package** - Could create `simba/internal/params` but adds complexity

**Recommendation: Option A** - Clean, follows Go conventions, useful for other extensions too.

### 3. Middleware Return Type

**Before** (in simba):
```go
type WebSocketMiddleware func(ctx context.Context) context.Context

func WebSocketTraceID() func(context.Context) context.Context {
    return func(ctx context.Context) context.Context {
        // ...
    }
}
```

**After** (in simba-websocket):
```go
package websocket

type Middleware func(ctx context.Context) context.Context

package middleware

import websocket "github.com/sillen102/simba-websocket"

func TraceID() websocket.Middleware {
    return func(ctx context.Context) context.Context {
        // ...
    }
}
```

---

## Testing Strategy

### Unit Tests
- All existing tests move to simba-websocket
- Test handler implementation of simba.Handler interface
- Test middleware application
- Test auth integration

### Integration Tests
Create a test in simba-websocket that:
1. Creates a simba.Application
2. Registers WebSocket handlers
3. Tests full request/response cycle
4. Validates authentication works
5. Validates parameter parsing works

Example:
```go
// handler_integration_test.go
package websocket_test

import (
    "testing"
    "github.com/sillen102/simba"
    websocket "github.com/sillen102/simba-websocket"
)

func TestIntegrationWithSimbaRouter(t *testing.T) {
    app := simba.New()
    
    handler := websocket.Handler(testCallbacks)
    app.Router.GET("/ws/test", handler)
    
    // Test connection upgrade and message handling
}
```

### Compatibility Tests
In simba repository, add test that imports simba-websocket to ensure API stability.

---

## Documentation Updates

### 1. Main Simba README.md

Add section:
```markdown
## Extensions

### WebSocket Support
WebSocket functionality is available as a separate module:

```bash
go get github.com/sillen102/simba-websocket
```

See [simba-websocket documentation](https://github.com/sillen102/simba-websocket) for details.
```

### 2. New simba-websocket README.md

```markdown
# Simba WebSocket

WebSocket support for the Simba Go framework.

## Installation

```bash
go get github.com/sillen102/simba-websocket
```

## Quick Start

[Include examples showing Handler, AuthHandler, middleware usage]

## Features
- Callback-based WebSocket handling
- Authentication support via simba.AuthHandler
- Parameter parsing
- Built-in middleware (TraceID, Logger)
- Thread-safe connection management

## API Documentation
[Link to pkg.go.dev]
```

### 3. Migration Guide

Create `MIGRATION.md` in simba-websocket:

```markdown
# Migrating from simba WebSocket to simba-websocket

## Import Changes
```go
// Old
import "github.com/sillen102/simba"

// New
import (
    "github.com/sillen102/simba"
    websocket "github.com/sillen102/simba-websocket"
    wsmiddleware "github.com/sillen102/simba-websocket/middleware"
)
```

## Type Changes
- `simba.WebSocketHandler` → `websocket.Handler`
- `simba.WebSocketCallbacks` → `websocket.Callbacks`
- `simba.WebSocketConnection` → `websocket.Connection`
- `middleware.WebSocketTraceID()` → `wsmiddleware.TraceID()`
- `middleware.WebSocketLogger()` → `wsmiddleware.Logger()`
```

---

## Alternative Approaches Considered

### Option 2: Interface Extraction

Create `simba-core` package with shared interfaces:

```
simba-core/          # Shared interfaces
  ├── handler.go     # Handler interface
  └── auth.go        # AuthHandler interface

simba/               # Main framework
  └── imports simba-core

simba-websocket/     # WebSocket extension
  └── imports simba-core
```

**Pros:**
- No circular dependency
- Both packages depend on abstraction

**Cons:**
- Adds complexity with 3 packages
- Users must import simba-core for type definitions
- More maintenance burden
- Violates YAGNI - we don't need this complexity yet

**Verdict:** Over-engineered for current needs. Option 1 is simpler.

### Option 3: Plugin Pattern with Reflection

Use runtime registration:

```go
package simba

var websocketPlugin WebSocketPlugin

func RegisterWebSocketPlugin(p WebSocketPlugin) {
    websocketPlugin = p
}
```

**Cons:**
- Loss of type safety
- Runtime errors instead of compile-time
- Performance overhead
- Complex to implement and maintain

**Verdict:** Not idiomatic Go. Rejected.

### Option 4: Keep in Main Module with Build Tags

```go
//go:build websocket
package simba

// WebSocket code here
```

**Cons:**
- Requires users to use build flags
- Framework shouldn't require build configuration
- Confusing developer experience
- Not what user requested

**Verdict:** User explicitly wants to avoid build flags. Rejected.

---

## Benefits of Recommended Approach

### For Framework Maintainers
1. **Cleaner Core** - Simba core is lighter, faster to build
2. **Separate Versioning** - WebSocket can evolve independently
3. **Focused PRs** - Changes isolated to relevant repos
4. **Easier Testing** - Each module tested independently
5. **Clear Dependencies** - Dependency graph is explicit

### For Users
1. **Optional Import** - Only import what you need
2. **No Build Flags** - Standard Go import
3. **Smaller Binaries** - If not using WebSocket, not included
4. **Clear Documentation** - WebSocket docs in separate repo
5. **Familiar Pattern** - Same as net/http and gorilla/websocket relationship

### For Ecosystem
1. **Extensibility Model** - Pattern for future extensions (gRPC, GraphQL, etc.)
2. **Community Contributions** - Easier to contribute to specific area
3. **Third-Party Extensions** - Others can create similar modules

---

## Risks and Mitigations

### Risk 1: Breaking Changes in Core
**Risk:** Changes to simba.Handler interface break simba-websocket

**Mitigation:**
- Version simba-websocket against specific simba versions
- Use semver strictly
- Maintain compatibility matrix in docs
- Add integration tests that catch breaks

### Risk 2: Duplication of Common Code
**Risk:** Some utility functions might need duplication

**Mitigation:**
- Export necessary utilities from simba (parseParams, etc.)
- Keep simba utilities general-purpose
- Document which functions are extension-friendly

### Risk 3: User Confusion
**Risk:** Users don't know about separate module

**Mitigation:**
- Clear documentation in main README
- Migration guide
- Code examples in both repos
- Deprecation warnings with links

### Risk 4: Maintenance Burden
**Risk:** Two repos to maintain

**Mitigation:**
- Automated testing between modules
- Keep WebSocket module stable (fewer changes needed)
- Good initial implementation reduces future maintenance

---

## Implementation Checklist

### Phase 1: Setup (Week 1)
- [ ] Create `simba-websocket` repository
- [ ] Setup go.mod with simba dependency
- [ ] Setup CI/CD pipeline
- [ ] Setup test infrastructure

### Phase 2: Code Migration (Week 1-2)
- [ ] Copy websocket_handler.go → handler.go
- [ ] Copy websocket_callbacks.go → callbacks.go
- [ ] Copy websocket_connection.go → connection.go
- [ ] Copy middleware files
- [ ] Update package names
- [ ] Update type names
- [ ] Update imports

### Phase 3: Core Changes (Week 2)
- [ ] Export ParseAndValidateParams in simba
- [ ] Export HandleAuthRequest in simba
- [ ] Add integration tests
- [ ] Verify all tests pass

### Phase 4: Documentation (Week 2-3)
- [ ] Write simba-websocket README
- [ ] Write migration guide
- [ ] Update simba README
- [ ] Add code examples
- [ ] Update godoc comments

### Phase 5: Release (Week 3)
- [ ] Tag simba-websocket v0.1.0
- [ ] Update simba examples to use new module
- [ ] Add deprecation notices to simba
- [ ] Announce in release notes

### Phase 6: Deprecation (Future)
- [ ] Version N+1: Mark old code deprecated
- [ ] Version N+2: Remove old WebSocket code

---

## Success Criteria

1. ✅ Users can import simba without WebSocket dependencies
2. ✅ Users can add WebSocket by importing separate module
3. ✅ No build flags or special configuration required
4. ✅ All existing WebSocket functionality preserved
5. ✅ Authentication integration works seamlessly
6. ✅ Parameter parsing works seamlessly
7. ✅ Middleware works as before
8. ✅ Clear migration path for existing users
9. ✅ Comprehensive documentation
10. ✅ All tests passing

---

## Example Usage Scenarios

### Scenario 1: New Project with WebSockets
```bash
go get github.com/sillen102/simba
go get github.com/sillen102/simba-websocket
```

```go
import (
    "github.com/sillen102/simba"
    websocket "github.com/sillen102/simba-websocket"
)

app := simba.Default()
app.Router.GET("/ws", websocket.Handler(callbacks))
app.Start()
```

### Scenario 2: Existing Project Adding WebSockets
```bash
# Already have simba
go get github.com/sillen102/simba-websocket
```

Add WebSocket endpoint to existing app.

### Scenario 3: REST-Only Project
```bash
go get github.com/sillen102/simba
# No WebSocket dependency needed!
```

Smaller dependency tree, faster builds.

---

## Conclusion

The recommended approach (separate module with one-way dependency) provides:

1. **Clean separation** - WebSocket code completely separate
2. **No build flags** - Standard Go import mechanism
3. **Backward compatible** - Migration path provided
4. **Idiomatic Go** - Follows standard patterns
5. **User-friendly** - Simple import, familiar API
6. **Maintainable** - Clear module boundaries
7. **Extensible** - Model for future extensions

This is the best balance of simplicity, usability, and maintainability for the Simba framework.
