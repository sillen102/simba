package openapiModels

// RouteInfo stores type information about a route
type RouteInfo struct {
	Method      string
	Path        string
	Accepts     string
	Produces    string
	ReqBody     any
	Params      any
	RespBody    any
	Handler     any
	AuthModel   any
	AuthHandler any
}
