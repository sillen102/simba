package simba

import (
	"mime/multipart"
	"net/http"
)

// Request represents a HTTP Request
type Request[RequestBody any, RequestParams any] struct {
	Cookies []*http.Cookie
	Body    RequestBody
	Params  RequestParams
}

type MultipartRequest[RequestParams any] struct {
	Cookies []*http.Cookie
	Reader  *multipart.Reader
	Params  RequestParams
}

// Response represents a HTTP response
type Response[ResponseBody any] struct {
	Headers http.Header
	Cookies []*http.Cookie
	Body    ResponseBody
	Status  int
}

// NoBody is an empty struct used to represent no body
type NoBody struct {
}

// NoParams is an empty struct used to represent no params
type NoParams struct {
}
