package simba

import "net/http"

// POST registers a handler for POST requests to the given pattern
func (s *Router) POST(uri string, handler http.Handler) {
	s.Handle("POST "+uri, handler)
}

// GET registers a handler for GET requests to the given pattern
func (s *Router) GET(uri string, handler http.Handler) {
	s.Handle("GET "+uri, handler)
}

// PUT registers a handler for PUT requests to the given pattern
func (s *Router) PUT(uri string, handler http.Handler) {
	s.Handle("PUT "+uri, handler)
}

// DELETE registers a handler for DELETE requests to the given pattern
func (s *Router) DELETE(uri string, handler http.Handler) {
	s.Handle("DELETE "+uri, handler)
}

// PATCH registers a handler for PATCH requests to the given pattern
func (s *Router) PATCH(uri string, handler http.Handler) {
	s.Handle("PATCH "+uri, handler)
}

// OPTIONS registers a handler for OPTIONS requests to the given pattern
func (s *Router) OPTIONS(uri string, handler http.Handler) {
	s.Handle("OPTIONS "+uri, handler)
}
