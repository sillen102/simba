package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sillen102/simba"
)

type RequestBody struct {
	Age int `json:"age"`
}

type Params struct {
	Name   string `header:"name" validate:"required"`
	ID     int    `path:"id" validate:"required"`
	Active bool   `query:"active" validate:"required"`
	Page   int64  `query:"page" validate:"min=0"`
	Size   int64  `query:"size" validate:"min=0"`
}

type ResponseBody struct {
	Message string `json:"message"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Page    int64  `json:"page"`
	Size    int64  `json:"size"`
}

func handler(ctx context.Context, req *simba.Request[RequestBody, Params]) (*simba.Response, error) {
	return &simba.Response{
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are %d years old", req.Params.Name, req.Body.Age),
			ID:      req.Params.ID,
			Active:  req.Params.Active,
			Page:    req.Params.Page,
			Size:    req.Params.Size,
			Name:    req.Params.Name,
		},
	}, nil
}

func main() {
	router := simba.Default()
	router.POST("/params/:id", simba.HandlerFunc(handler))
	http.ListenAndServe(":9999", router)
}
