# Simba

Simba is a lightweight, type-safe HTTP router framework for Go that makes building REST APIs simple and enjoyable. It provides strong type safety through generics and a clean, intuitive API for handling HTTP requests.

## Features

- **Type-safe routing** with Go generics
- **Built-in authentication** support
- **Middleware support**
- **Strong request/response typing**
- **High performance** through [httprouter](https://github.com/julienschmidt/httprouter)
- **Modular design** for easy extensibility

## Installation

```bash
go get github.com/sillen102/simba
```

## Quick Start

Here's a simple example showing how to create a basic HTTP server with Simba:

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

func handler(ctx context.Context, req *simba.Request[RequestBody, simba.NoParams]) (*simba.Response, error) {
    return &simba.Response{
        Body: ResponseBody{
            Message: fmt.Sprintf("Hello %s, you are %d years old", req.Body.Name, req.Body.Age),
        },
        Status: http.StatusOK,
    }, nil
}

func main() {
    router := simba.Default()
    router.POST("/users", simba.HandlerFunc(handler))
    http.ListenAndServe(":9999", router)
}
```

## Authentication

Simba provides built-in support for authentication:

```go
type User struct {
    ID   string
    Name string
}

func authFunc(r *http.Request) (*User, error) {
    // Your authentication logic here, either be it a database lookup or any other authentication method
    return &User{ID: "123", Name: "John"}, nil
}

// This handler will only be called if the user is authenticated and the user is available as one of the function parameters
func getUser(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], user *User) (*simba.Response, error) {
    // ... handle the request
}

router := simba.DefaultWithAuth[User](authFunc)
router.GET("/users/:userId", simba.AuthenticatedHandlerFunc(getUser))
```

## Parameters

Handle parameters with type safety and validation support using go-playground validator:

```go
type Params struct {
    UserID string `path:"userId"`
    Name   string `query:"name" validate:"required"`
    Age    int    `header:"age" validate:"required"`
    Page   int64  `query:"page" validate:"omitempty,min=0" default:"0"`
    Size   int64  `query:"size" validate:"omitempty,min=0" default:"10"`
}

func getUser(ctx context.Context, req *simba.Request[simba.NoBody, Params]) (*simba.Response, error) {
    userID := req.Params.UserID
    name := req.Params.Name
    age := req.Params.Age

    // ... handle the request
}

router.GET("/users/:userId", simba.HandlerFunc(getUser))
```

## Logging

Simba automatically injects a [zerolog](https://github.com/rs/zerolog) logger into every request's context. You can access this logger from any handler or middleware using the `logging` package:

```go
func handler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
    logger := logging.FromCtx(ctx)
    logger.Info().Msg("handling request")
    // ... handle the request
}
```

## Configuration

Customize router behavior with options:

```go
router := simba.NewRouter(simba.RouterOptions{
    RequestDisallowUnknownFields: true,
    RequestIdAcceptHeader:        true,
    LogRequestBody:               true,
    LogLevel:                     zerolog.DebugLevel,
    LogFormat:                    logging.JSONFormat,
    LogOutput:                    os.Stdout,
})

```

## Error Handling

Simba provides automatic error handling with standardized JSON responses. All errors are automatically wrapped and returned in a consistent format:

```json
{
  "timestamp": "2023-01-01T12:00:00Z",
  "status": 400,
  "error": "Bad Request",
  "path": "/api/resource",
  "method": "POST",
  "requestId": "01JE61MX24YEGF08E8G0RA0Y14",
  "message": "invalid request body",
  "validationErrors": [
    {
      "parameter": "email",
      "type": "body",
      "message": "'email' is not a valid email address"
    }
  ]
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

All dependencies are under their respective licenses, which can be found in their repositories via the go.mod file.
