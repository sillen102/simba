package simba

import (
	"net/http"
)

// Handler represents a HTTP handler all handler types must implement this interface
type Handler[RequestBody any, Params any] interface {
	Handle(w http.ResponseWriter, r *http.Request)
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
