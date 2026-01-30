# Simba

Simba is a type-safe HTTP router framework for Go that makes building REST APIs simple and enjoyable. It provides strong type safety through generics, clean and intuitive APIs for request handling, and generates OpenAPI (v3.1) docs automatically.

> **Telemetry Notice:** Simba never instruments requests or metrics unless you provide a telemetry provider. Wire in any provider (like OpenTelemetry) using `SetTelemetryProvider()` as demonstrated below.

## Features

- **Type-safe routing** with Go generics
- **Built-in authentication** (API key, Basic, Bearer)
- **Middleware support**
- **Strong request/response typing**
- **Automatic OpenAPI docs generation**
- **WebSocket support**

---

## Telemetry & Observability (Advanced, Optional)
Simba is telemetry-provider agnostic. To emit traces/metrics, inject any telemetry provider you like:

```go
import (
    "context"
    "github.com/sillen102/simba"
    "github.com/sillen102/simba/telemetry"
    "github.com/sillen102/simba/telemetry/config"
)

func main() {
    app := simba.Default() // or simba.New(...)

    tcfg := &config.TelemetryConfig{
        Enabled:        true,
        ServiceName:    "my-service",
        ServiceVersion: "1.0.0",
        Tracing:        config.TracingConfig{Enabled: true, Exporter: "otlp"},
        Metrics:        config.MetricsConfig{Enabled: true, Exporter: "otlp"},
    }
    prov, err := telemetry.NewOtelTelemetryProvider(context.Background(), tcfg)
    if err != nil {
        panic("Failed to create telemetry provider: " + err.Error())
    }
    app.SetTelemetryProvider(prov)

    // Advanced: Create custom spans/counters by type asserting your provider
    // otel := prov.(*telemetry.OtelTelemetryProvider)
    // meter := otel.Provider().Meter("my.custom")
    // tracer := otel.Provider().Tracer("my.custom")

    app.Start()
}
```

See [`examples/telemetry`](./examples/telemetry) for advanced provider use, custom spans and metrics. If you do not set a provider, Simba disables all. 

---

## Installation

```bash
go get -u github.com/sillen102/simba
```

---

## Quick Start

```go
package main
import (
    "context"
    "fmt"
    "net/http"
    "github.com/sillen102/simba"
)

type RequestBody struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

type ResponseBody struct {
    Message string `json:"message"`
}

func handler(ctx context.Context, req *simba.Request[RequestBody, simba.NoParams]) (*simba.Response[ResponseBody], error) {
    return &simba.Response[ResponseBody]{
        Body: ResponseBody{
            Message: fmt.Sprintf("Hello %s, you are %d years old", req.Body.Name, req.Body.Age),
        },
        Status: http.StatusOK,
    }, nil
}

func main() {
    app := simba.Default()
    app.Router.POST("/users", simba.JsonHandler(handler))
    app.Start()
}
```

---

## Parameters

Simba provides deep parameter binding and validation from path, query, headers, and cookies. Default values (`default:"val"`) and validation (`validate:"..."`) are supported via `go-playground/validator` tags.

**Simple example:**
```go
type Params struct {
    UserID    string `path:"userId"`
    Name      string `query:"name" validate:"required"`
    Age       int    `header:"age" validate:"required"`
    SessionID string `cookie:"session_id" validate:"required"`
    Page      int64  `query:"page" default:"0"`
    Size      int64  `query:"size" default:"10"`
}

func getUser(ctx context.Context, req *simba.Request[simba.NoBody, Params]) (*simba.Response[ResponseBody], error) {
    // ... handler logic
}
app.GET("/users/{userId}", simba.JsonHandler(getUser))
```

**Advanced example:**
Handles custom types (`uuid.UUID`), booleans, floats, defaults, and OpenAPI multiline handler comments.
```go
import "github.com/google/uuid"
type Params struct {
    Name   string      `header:"name" validate:"required"`
    ID     uuid.UUID   `path:"id" validate:"required"`
    Active bool        `query:"active" validate:"required"`
    Page   int         `query:"page" validate:"omitempty,min=0" default:"1"`
    Size   int64       `query:"size" validate:"omitempty,min=0" default:"10"`
    Score  float64     `query:"score" default:"10.0"`
}

// Handler with OpenAPI docs
// This is a description of what the handler does.
// Can span multiple lines.
func handler(ctx context.Context, req *simba.Request[RequestBody, Params]) (*simba.Response[ResponseBody], error) {
    // req.Params.ID (uuid)
    // req.Params.Active (bool, parsed from query)
    // req.Params.Page (default/explicit)
    // ...
}
```

---

## No Body Responses & Status Codes
To return a 204 No Content response:
```go
func noBodyHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response[simba.NoBody], error) {
    return &simba.Response[simba.NoBody]{}, nil // No body = 204
}
```
If you omit `Status`, Simba uses 200 for non-empty, 204 for empty bodies.

---

## Response Headers & Cookies
Set headers/cookies in the response struct:
```go
return &simba.Response[ResponseBody]{
    Headers: map[string][]string{"My-Header": {"value"}},
    Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "val"}},
    Body: ...,
}
```

---

## WebSocket Support

Simba provides first-class generic WebSocket support (with middleware and optional authentication):

```go
import (
    "context"
    "github.com/sillen102/simba"
    "github.com/sillen102/simba/websocket"
    wsmw "github.com/sillen102/simba/websocket/middleware"
)

type User struct {
    ID   int
    Name string
}

// Simple bearer token authentication function
func authHandler(ctx context.Context, token string) (User, error) {
    if token == "valid-token" {
        return User{ID: 1, Name: "John Doe"}, nil
    }
    return User{}, fmt.Errorf("invalid token")
}

// Echo handler
func echoCallbacks() websocket.Callbacks[simba.NoParams] {
    return websocket.Callbacks[simba.NoParams]{
        OnConnect: func(ctx context.Context, conn *websocket.Connection, params simba.NoParams) error {
            return conn.WriteText("Connected! Send messages and I'll echo them.")
        },
        OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte) error {
            return conn.WriteText("Echo: " + string(data))
        },
    }
}

// Authenticated chat handler
func chatCallbacks() websocket.AuthCallbacks[simba.NoParams, User] {
    return websocket.AuthCallbacks[simba.NoParams, User]{
        OnConnect: func(ctx context.Context, conn *websocket.Connection, params simba.NoParams, user User) error {
            return conn.WriteText(fmt.Sprintf("Welcome %s!", user.Name))
        },
        OnMessage: func(ctx context.Context, conn *websocket.Connection, data []byte, user User) error {
            return conn.WriteText(fmt.Sprintf("[%s]: %s", user.Name, string(data)))
        },
    }
}

func main() {
    app := simba.Default()
    // Unauthenticated echo endpoint
    app.Router.GET("/ws/echo", websocket.Handler(
        echoCallbacks(),
        websocket.WithMiddleware(
            wsmw.TraceID(),
            wsmw.Logger(),
        ),
    ))
    // Authenticated chat endpoint
    bearer := simba.BearerAuth(authHandler, simba.BearerAuthConfig{
        Name: "BearerAuth", Format: "JWT", Description: "Bearer token",
    })
    app.Router.GET("/ws/chat", websocket.AuthHandler(
        chatCallbacks(),
        bearer,
        websocket.WithMiddleware(
            wsmw.TraceID(),
            wsmw.Logger(),
        ),
    ))
    app.Start()
}
```

---

## Logging

Simba uses `slog` for logging. With `simba.Default()`, a logger is injected into each request context. To use a custom logger:

```go
import (
    "log/slog"
    "github.com/sillen102/simba"
    "github.com/sillen102/simba/settings"
)
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
app := simba.Default(settings.WithLogger(logger))
```
Retrieve logger in handlers:
```go
import "github.com/sillen102/simba/logging"
func handler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response[ResponseBody], error) {
    logger := logging.From(ctx)
    logger.Info("handling request")
}
```

---

## Middleware

Register standard Go http.Handler compatible middleware:
```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        next.ServeHTTP(w, r)
    })
}
app.Router.Use(myMiddleware)
```

You can inject data into request headers in middleware and access via validated handler params:
```go
app.Router.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.Header.Set("X-Middleware", "123")
        next.ServeHTTP(w, r)
    })
})
```
And access:
```go
type Params struct { MiddlewareHeader string `header:"X-Middleware"` }
```

---

## Authentication (API Key, Basic, Bearer)

All Simba built-in authentication handlers use **pointer-based generics** for the user type:

### API Key Auth (Canonical Example)
```go
type User struct {
    ID   int
    Name string
    Role string
}

func authFunc(ctx context.Context, apiKey string) (*User, error) {
    if apiKey != "valid-key" {
        return nil, fmt.Errorf("invalid api key")
    }
    return &User{ID: 1, Name: "John Doe", Role: "admin"}, nil
}

authHandler := simba.APIKeyAuth[*User](authFunc, simba.APIKeyAuthConfig{
    Name: "admin", FieldName: "sessionid", In: openapi.InHeader, Description: "admin only",
})

func authenticatedHandler(ctx context.Context, req *simba.Request[simba.NoBody, struct { UserID int `path:"userId"` }], user *User) (*simba.Response[ResponseBody], error) {
    return &simba.Response[ResponseBody]{
        Body: ResponseBody{Message: fmt.Sprintf("Hello %s, you are an %s", user.Name, user.Role)},
    }, nil
}

app := simba.Default()
app.Router.GET("/users/{userId}", simba.AuthJsonHandler(authenticatedHandler, authHandler))
```

### Bearer Auth
```go
func bearerAuthFunc(ctx context.Context, token string) (*User, error) { /* ... */ }
bearer := simba.BearerAuth[*User](bearerAuthFunc, simba.BearerAuthConfig{
    Name: "bearer", Format: "jwt", Description: "token",
})
```

### Basic Auth
```go
func basicAuthFunc(ctx context.Context, username, password string) (*User, error) { /* ... */ }
basic := simba.BasicAuth[*User](basicAuthFunc, simba.BasicAuthConfig{
    Name: "basic", Description: "desc",
})
```

---

## OpenAPI Documentation

Simba automatically generates a full [OpenAPI](https://spec.openapis.org/) specification at `/openapi.yml` and serves interactive documentation at `/docs` using the Scalar UI. You can customize these endpoints or disable documentation entirely via settings:

```go
app.Settings.Docs.OpenAPIPath = "/openapi.yml"      // Change OpenAPI spec path (default)
app.Settings.Docs.DocsPath = "/docs"               // Change Scalar UI path (default)
app.Settings.Docs.GenerateOpenAPIDocs = false       // Disable OpenAPI generation entirely
```

### Customizing OpenAPI Documentation

Simba generates your OpenAPI specification using handler comments: 
- The package and function names define operationId and grouping.
- Use line comments above your handler function to add descriptions, summaries, and error information. Simba parses these for OpenAPI summaries/descriptions.

Example:
```go
// @ID get-user
// @Tag Users
// @Summary Get user
// @Description Get user by ID
// @Error 404 User not found
func getUser(...) {...}
```
For details, see [swaggest/openapi-go](https://github.com/swaggest/openapi-go). You do not need or use Swagger tags within Simba.

---

## License
MIT (see LICENSE file)

---

For advanced usage, see additional folders in [`examples/`](./examples/) for:
- Observability and advanced tracing/metrics with OpenTelemetry (`examples/telemetry`)
- WebSocket (`examples/websocket`)
- Advanced parameter handling, authentication, and middleware patterns
