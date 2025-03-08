package simbaTest

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
)

// TestApplication represents a test application with additional testing utilities
type TestApplication struct {
	*simba.Application
	Server *httptest.Server
}

// New creates a new test application with the given settings
func New(opts ...settings.Option) *TestApplication {
	app := simba.New(opts...)

	return &TestApplication{
		Application: app,
		Server:      httptest.NewServer(app.Router),
	}
}

// Default creates a new test application with default settings
func Default(opts ...settings.Option) *TestApplication {
	app := simba.Default(opts...)

	return &TestApplication{
		Application: app,
		Server:      httptest.NewServer(app.Router),
	}
}

// Start starts the test server
func (a *TestApplication) Start() {
	_ = a.Router.GenerateOpenAPIDocumentation(context.Background())
	a.Application.Server.Addr = a.Server.URL[7:]
}

// Stop stops the test server
func (a *TestApplication) Stop() {
	if a.Server != nil {
		a.Server.Close()
	}
}

// URL returns the base URL of the test server
func (a *TestApplication) URL() string {
	if a.Server == nil {
		return ""
	}
	return a.Server.URL
}

// Client returns an HTTP client configured to work with the test server
func (a *TestApplication) Client() *http.Client {
	if a.Server == nil {
		return nil
	}
	return a.Server.Client()
}

// RunTest runs a test function with a started test server and handles cleanup
func (a *TestApplication) RunTest(fn func()) {
	a.Start()
	defer a.Stop()

	fn()
}
