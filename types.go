package simba

import (
	"net/http"

	"github.com/uptrace/bunrouter"
)

// Handler is an interface that all handlers must implement
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r bunrouter.Request) error
}

// Request represents a HTTP request
type Request[RequestBody any] struct {
	Headers http.Header
	Cookies []*http.Cookie
	Body    RequestBody
}

// Response represents a HTTP response
type Response struct {
	Headers http.Header
	Cookies []*http.Cookie
	Body    any
	Status  int
}

// NoBody is an empty struct used to represent no body
type NoBody struct {
}

// NoParams is an empty struct used to represent no params
type NoParams struct {
}
