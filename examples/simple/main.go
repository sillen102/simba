package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/logging"
)

type RequestBody struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

func handler(ctx context.Context, req *simba.Request[RequestBody, simba.NoParams]) (*simba.Response, error) {
	return &simba.Response{
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are %d years old", req.Body.Name, req.Body.Age),
		},
		Status: http.StatusOK, // We can omit this and it will default to 200 OK if the body is not nil and there is no error
	}, nil
}

func noBodyHandler(ctx context.Context, req *simba.Request[simba.NoBody, simba.NoParams]) (*simba.Response, error) {
	return &simba.Response{}, nil // Returns 204 since there is no body in the response
}

func main() {
	router := simba.Default()
	router.POST("/users", simba.HandlerFunc(handler))
	router.GET("/no-body", simba.HandlerFunc(noBodyHandler))
	logging.GetDefault().Info().Msg("Listening on http://localhost:9999")
	http.ListenAndServe(":9999", router)
}
