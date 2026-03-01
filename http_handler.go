package simba

import "net/http"

type httpHandlerAdapter struct {
	handler http.Handler
}

// HTTPHandler adapts a plain http.Handler into a Simba handler.
// This is useful when mounting protocol-specific handlers, such as WebSocket upgrades,
// behind the Simba router.
func HTTPHandler(handler http.Handler) Handler {
	return httpHandlerAdapter{handler: handler}
}

func (h httpHandlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h httpHandlerAdapter) GetRequestBody() any {
	return nil
}

func (h httpHandlerAdapter) GetParams() any {
	return nil
}

func (h httpHandlerAdapter) GetResponseBody() any {
	return nil
}

func (h httpHandlerAdapter) GetAccepts() string {
	return ""
}

func (h httpHandlerAdapter) GetProduces() string {
	return ""
}

func (h httpHandlerAdapter) GetHandler() any {
	return h.handler
}

func (h httpHandlerAdapter) GetAuthModel() any {
	return nil
}

func (h httpHandlerAdapter) GetAuthHandler() any {
	return nil
}

func (h httpHandlerAdapter) ShouldDocument() bool {
	return false
}
