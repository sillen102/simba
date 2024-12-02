# Simba

Simba is a lightweight, type-safe HTTP router framework for Go that makes building REST APIs simple and enjoyable. It provides strong type safety through generics and a clean, intuitive API for handling HTTP requests.

## Features

- **Type-safe routing** with Go generics
- **Built-in authentication** support
- **Zero-allocation URL parameters**
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
    router := simba.NewRouter()
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

router := simba.NewRouterWithAuth[User](authFunc)
router.GET("/users/:userId", simba.AuthenticatedHandlerFunc(getUser))
```

## URL Parameters

Handle URL parameters with type safety:

```go
type Params struct {
    UserID string `path:"userId"`
    Name   string `query:"name"`
    Age    int    `header:"age"`
}

func getUser(ctx context.Context, req *simba.Request[simba.NoBody, Params]) (*simba.Response, error) {
    userID := req.Params.UserID
    name := req.Params.Name
    age := req.Params.Age

    // ... handle the request
}

router.GET("/users/:userId", simba.HandlerFunc(getUser))
```

## Configuration

Customize router behavior with options:

```go
router := simba.NewRouter(simba.RouterOptions{
    RequestDisallowUnknownFields: true,
})
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.