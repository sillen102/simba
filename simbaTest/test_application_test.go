package simbaTest_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaModels"
	"github.com/sillen102/simba/simbaTest"
	"github.com/sillen102/simba/simbaTestAssert"
)

func TestNew(t *testing.T) {
	// Create a new test application
	app := simbaTest.New()

	// Add a test route
	app.Router.GET("/test", simba.JsonHandler(func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[[]byte], error) {
		return &simbaModels.Response[[]byte]{Status: http.StatusOK, Body: []byte("test response")}, nil
	}))

	// Run test with the application
	app.RunTest(func() {
		// Create a test request
		resp, err := app.Client().Get(app.URL() + "/test")
		simbaTestAssert.NoError(t, err)
		simbaTestAssert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDefault(t *testing.T) {
	// Create a new test application with default settings
	app := simbaTest.Default()

	// Add a test route
	app.Router.GET("/test", simba.JsonHandler(func(ctx context.Context, req *simbaModels.Request[simbaModels.NoBody, simbaModels.NoParams]) (*simbaModels.Response[[]byte], error) {
		return &simbaModels.Response[[]byte]{Status: http.StatusOK, Body: []byte("test response")}, nil
	}))

	// Run test with the application
	app.RunTest(func() {
		// Create a test request
		resp, err := app.Client().Get(app.URL() + "/test")
		simbaTestAssert.NoError(t, err)
		simbaTestAssert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
