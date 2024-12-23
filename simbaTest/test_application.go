package simbaTest

import (
	"net/http"
	"net/http/httptest"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
)

// TestApplication represents a test application with additional testing utilities
type TestApplication[AuthModel any] struct {
	*simba.Application[AuthModel]
	Server *httptest.Server
}

// New creates a new test application with the given settings
func New(settings ...settings.Config) *TestApplication[struct{}] {
	return NewWithAuth[struct{}](nil, settings...)
}

// NewWithAuth creates a new test application with the given settings and with authentication
func NewWithAuth[AuthModel any](authFunc simba.AuthFunc[AuthModel], settings ...settings.Config) *TestApplication[AuthModel] {
	app := simba.NewAuthWith(authFunc, settings...)

	return &TestApplication[AuthModel]{
		Application: app,
		Server:      httptest.NewServer(app.Router),
	}
}

// Default creates a new test application with default settings
func Default(settings ...settings.Config) *TestApplication[struct{}] {
	return DefaultWithAuth[struct{}](nil, settings...)
}

// DefaultWithAuth creates a new test application with default settings and with authentication
func DefaultWithAuth[AuthModel any](authFunc simba.AuthFunc[AuthModel], settings ...settings.Config) *TestApplication[AuthModel] {
	app := simba.DefaultAuthWith(authFunc, settings...)

	return &TestApplication[AuthModel]{
		Application: app,
		Server:      httptest.NewServer(app.Router),
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
