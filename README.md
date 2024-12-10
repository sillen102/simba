# Simba

Simba is a lightweight, type-safe HTTP router framework for Go that makes building REST APIs simple and enjoyable. It provides strong type safety through generics and a clean, intuitive API for handling HTTP requests.

## Features

- **Type-safe routing** with Go generics
- **Built-in authentication** support
- **Middleware support**
- **Strong request/response typing**

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

    // Access the request body fields
    // req.Body.Age
    // req.Body.Name

    // Access the request cookies
    // req.Cookies

    // Access the request headers
    // req.Headers

    return &simba.Response{
        Headers: map[string][]string{"My-Header": {"header-value"}},
        Cookies: []*http.Cookie{{Name: "My-Cookie", Value: "cookie-value"}},
        Body: ResponseBody{
            Message: fmt.Sprintf("Hello %s, you are %d years old", req.Body.Name, req.Body.Age),
        },
        Status: http.StatusOK, // Can be omitted, defaults to 200 if there's a body, 204 if there's no body
    }, nil
}

func main() {
    // Using simba.Default() will use the default options for logging and request validation,
    // add default middleware like panic recovery and request id and add some endpoints like /health
    //
    // If you wish to build up your own router without any default middleware etc., use simba.New()
    app := simba.Default()
	app.Router.POST("/users", simba.HandlerFunc(handler))
	app.Start(context.Background())
}
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

app.GET("/users/{userId}", simba.HandlerFunc(getUser))
```

## Logging

Simba automatically injects an slog logger into the request's context. To access the logger, use the `logging.From` function in Simba's logging package:

```go
func handler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
    logger := logging.From(ctx)
    logger.Info("handling request")
    // ... handle the request
}
```

## Configuration

Customize behavior with settings:

```go
app := simba.New(simba.Settings{
    Server: simba.ServerSettings{
        Host: "localhost",
        Port: 9999,
    },
    Request: simba.RequestSettings{
        AllowUnknownFields: enums.Allow,
        LogRequestBody: enums.Enabled,
        RequestIdMode: enums.AcceptFromHeader,
    },
    Logging: logging.Config{
        Level: slog.LevelInfo,
        Format: logging.JsonFormat,
        Output: os.Stdout,
    },
})

```

## Error Handling

Simba provides automatic error handling with standardized JSON responses. All errors are automatically wrapped and returned in a consistent format:

```json
{
  "timestamp": "2024-12-04T20:28:33.852965Z",
  "status": 400,
  "error": "Bad Request",
  "path": "/api/resource",
  "method": "POST",
  "requestId": "01JE61MX24YEGF08E8G0RA0Y14",
  "message": "request validation failed, 1 validation error",
  "validationErrors": [
    {
      "parameter": "email",
      "type": "body",
      "message": "'notanemail' is not a valid email address"
    }
  ]
}
```

## Middleware

Simba supports middleware. Simply create a function that takes a handler and returns a handler and register it with the `Use` method on the router:

```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.Header.Set("X-Middleware", "123") // Here we simply add a header to every request
        next.ServeHTTP(w, r) // And the proceed to the next handler
    })
}

app.Router.Use(myMiddleware)
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

app := simba.DefaultAuthWith(authFunc)
app.GET("/users/{userId}", simba.AuthHandlerFunc(getUser))
```


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

All dependencies are under their respective licenses, which can be found in their repositories via the go.mod file.
