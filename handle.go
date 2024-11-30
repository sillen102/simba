package simba

import "net/http"

// Handle handles a request and returns a response.
// The request body will be decoded into the given type. If the type is NoBody, the body will be ignored.
// The params will be decoded into the given type. If the type is NoParams, the params will be ignored.
func Handle[ReqBody, Params any, H Handler[ReqBody, Params]](h H) http.HandlerFunc {
	return h.Handle
}
