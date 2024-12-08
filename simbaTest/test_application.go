package simbaTest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sillen102/simba"
)

// TestApplication represents a test application with additional testing utilities
type TestApplication[AuthModel any] struct {
	*simba.Application[AuthModel]
	Server *httptest.Server
	T      *testing.T
}

// New creates a new test application with the given settings
func New[AuthModel any](t *testing.T, authFunc simba.AuthFunc[AuthModel], settings ...simba.Settings) *TestApplication[AuthModel] {
	app := simba.NewAuthWith(authFunc, settings...)

	return &TestApplication[AuthModel]{
		Application: app,
		Server:      httptest.NewServer(app),
		T:           t,
	}
}

// Start starts the test server
func (a *TestApplication[AuthModel]) Start() {
	a.Application.Server.Addr = a.Server.URL[7:]
}

// Stop stops the test server
func (a *TestApplication[AuthModel]) Stop() {
	if a.Server != nil {
		a.Server.Close()
	}
}

// URL returns the base URL of the test server
func (a *TestApplication[AuthModel]) URL() string {
	if a.Server == nil {
		return ""
	}
	return a.Server.URL
}

// Client returns an HTTP client configured to work with the test server
func (a *TestApplication[AuthModel]) Client() *http.Client {
	if a.Server == nil {
		return nil
	}
	return a.Server.Client()
}

// RunTest runs a test function with a started test server and handles cleanup
func (a *TestApplication[AuthModel]) RunTest(fn func()) {
	a.Start()
	defer a.Stop()

	fn()
}
