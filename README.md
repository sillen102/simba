# Simba

Simba is a type-safe HTTP router framework for Go that makes building REST APIs simple and enjoyable. It provides strong type safety through generics and a clean, intuitive API for handling HTTP requests.
It also automatically generates OpenAPI (v3.1) documentation for your API.

## Features

- **Type-safe routing** with Go generics
- **Built-in authentication** support
- **Middleware support**
- **Strong request/response typing**
- **Automatic OpenAPI documentation generation**

## Installation

```bash
go get -u github.com/sillen102/simba
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

func handler(ctx context.Context, req *simba.Request[RequestBody, simba.NoParams]) (*simba.Response[ResponseBody], error) {

    // Access the request body fields
    // req.Body.Age
    // req.Body.Name

    // Access the request cookies
    // req.Cookies

    // Access the request headers
    // req.Headers

    return &simba.Response[ResponseBody]{
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
    app.Router.POST("/users", simba.JsonHandler(handler))
    app.Start()
}
```

## Parameters

Handle parameters with type safety and validation support using go-playground validator:

```go
type Params struct {
    UserID    string `path:"userId"`
    Name      string `query:"name" validate:"required"`
    Age       int    `header:"age" validate:"required"`
    SessionID string `cookie:"session_id" validate:"required"`
    Page      int64  `query:"page" validate:"omitempty,min=0" default:"0"`
    Size      int64  `query:"size" validate:"omitempty,min=0" default:"10"`
}

func getUser(ctx context.Context, req *simba.Request[simba.NoBody, Params]) (*simba.Response[respBody], error) {
    userID := req.Params.UserID
    name := req.Params.Name
    age := req.Params.Age

    // ... handle the request
}

app.GET("/users/{userId}", simba.JsonHandler(getUser))
```

## Logging

Simba relies on slog to handle logging. If no logger is provided slog.Default will be used.
If you use the Default constructor an slog logger will be injected into the request context for all requests.
To access the injected logger, use the `logging.From` function the logging package.

Example:

```go
func handler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response[respBody], error) {
    logger := logging.From(ctx)
    logger.Info("handling request")
    // ... handle the request
}
```

## Configuration

Customize behavior with options functions:

```go
app := simba.New(
    settings.WithServerHost("localhost"),
    settings.WithServerPort(8080),
})
```

Or use default and change a single one or few of the settings:

```go
app := simba.Default()
app.Settings.Server.Port = 8080
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
  "requestId": "40ad8bb4-215a-4748-8a7f-9e236d988c5b",
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

// @BasicAuth "admin" "admin access only"
func authFunc(r *http.Request) (*User, error) {
    // Your authentication logic here, either be it a database lookup or any other authentication method
    return &User{ID: "123", Name: "John"}, nil
}

// This handler will only be called if the user is authenticated and the user is available as one of the function parameters
func getUser(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], user *User) (*simba.Response[respBody], error) {
    // ... handle the request
}

app := simba.Default()
app.GET("/users/{userId}", simba.AuthJsonHandler(getUser, authFunc))
```

In this example the `authFunc` function is called for every request to authenticate the user.
If the user is authenticated, the user is injected into the handler function.
If the authFunc returns an error, a 401 Unauthorized response is returned.

## OpenAPI Documentation Generation

Simba will automatically generate OpenAPI documentation for your APIs. By default, the OpenAPI documentation is available at `/openapi.yml`.
And the Scalar UI is available at `/docs`. You can change these paths by providing a custom path in the application settings.

```go
app := simba.Default()
app.Settings.Docs.OpenAPIFileType = mimetypes.ApplicationJSON
app.Settings.Docs.OpenAPIPath = "/swagger.json"
app.Settings.Docs.DocsPath = "/swagger-ui"
```

If you want to you can disable the OpenAPI documentation generation by setting the GenerateOpenAPIDocs to false in application settings.

```go
app := simba.Default()
app.Settings.Docs.GenerateOpenAPIDocs = false
```

You can also generate the OpenAPI documentation yaml file but not serve any UI (in case you want to use a different or customized UI)
by setting the MountDocsEndpoint to false in the application settings.

```go
app := simba.Default()
app.Settings.Docs.MountDocsEndpoint = false
```

### Customizing OpenAPI Documentation
By default, Simba will generate OpenAPI documentation based on the handler you have registered. It will use the package name to group the endpoints,
and the handler name to generate the operation id, summary for the endpoint and the comment you have on your handler to generate a description.
This makes it easy to generate OpenAPI documentation without any additional configuration.
Just organize your handlers in packages, name them well and add descriptive comments.

If you want greater control over the generated API documentation you can customize the OpenAPI documentation by providing tags in the comment of your handler.
Simba uses [openapi-go](https://github.com/swaggest/openapi-go) under the hood to generate the documentation.

```go
// @BasicAuth "admin" "admin access only"
func authFunc(r *http.Request) (*User, error) {
    // Your authentication logic here, either be it a database lookup or any other authentication method
    return &User{ID: "123", Name: "John"}, nil
}

type reqParams struct {
    ID       string `path:"id" example:"XXX-XXXXX"`
    Locale   string `query:"locale" pattern:"^[a-z]{2}-[A-Z]{2}$"`
    MyHeader string `header:"My-Header" required:"true"`
    MyCookie string `cookie:"My-Cookie" required:"true"`
}

type reqBody struct {
    Title  string `json:"string" example:"My Order"`
    Amount int    `json:"amount" example:"100" required:"true"`
    Items  []struct {
        Count uint   `json:"count" example:"2"`
        Name  string `json:"name" example:"Item 1"`
    } `json:"items"`
}

// @ID get-user
// @Tag Users
// @Summary Get user
// @Description Get a user by ID (can span across multiple lines)
// @Error 404 User not found
func getUser(ctx context.Context, req *simba.Request[reqBody, reqParams], user *User) (*simba.Response[respBody], error) {
    // ... handle the request
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

All dependencies are under their respective licenses, which can be found in their repositories via the go.mod file.
