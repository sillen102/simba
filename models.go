package simba

import (
	"net/http"
)

// Request represents a HTTP request
type Request[RequestBody any, RequestParams any] struct {
	Cookies []*http.Cookie
	Body    RequestBody
	Params  RequestParams
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
