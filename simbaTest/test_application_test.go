package simbaTest_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaTest"
	"gotest.tools/v3/assert"
)

func TestTestApplication(t *testing.T) {
	// Create a new test application
	app := simbaTest.NewWithAuth[struct{}](nil)

	// Add a test route
	app.Router.GET("/test", simba.JsonHandler(func(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
		return &simba.Response{Status: http.StatusOK, Body: []byte("test response")}, nil
	}))

	// Run test with the application
	app.RunTest(func() {
		// Create a test request
		resp, err := app.Client().Get(app.URL() + "/test")
		assert.NilError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestTestApplicationWithAuth(t *testing.T) {
	type User struct {
		ID string
	}

	// Mock auth function
	authFunc := func(r *http.Request) (*User, error) {
		return &User{ID: "test-user"}, nil
	}

	// Create a new test application with auth
	app := simbaTest.NewWithAuth(authFunc)

	// Add an authenticated test route
	app.Router.GET("/protected", simba.AuthJsonHandler(func(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams], user *User) (*simba.Response, error) {
		return &simba.Response{Status: http.StatusOK}, nil
	}))

	// Run test with the application
	app.RunTest(func() {
		// Create a test request
		resp, err := app.Client().Get(app.URL() + "/protected")
		assert.NilError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
