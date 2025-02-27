package simba

import "net/http"

// AuthFunc is a function type for authenticating a request.
type AuthFunc[AuthModel any] func(r *http.Request) (*AuthModel, error)
