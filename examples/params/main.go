package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sillen102/simba"
	"github.com/sillen102/simba/settings"
)

type RequestBody struct {
	Age  int    `json:"age"`
	Text string `json:"text"`
}

type Params struct {
	Name   string    `header:"name" validate:"required"`
	ID     uuid.UUID `path:"id" validate:"required"`
	Active bool      `query:"active" validate:"required"`
	Page   int       `query:"page" validate:"omitempty,min=0" default:"1"`
	Size   int64     `query:"size" validate:"omitempty,min=0" default:"10"`
	Score  float64   `query:"score" default:"10.0"`
}

type ResponseBody struct {
	Message string    `json:"message"`
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Active  bool      `json:"active"`
	Page    int       `json:"page"`
	Size    int64     `json:"size"`
	Score   float64   `json:"score"`
}

// handler
// This is a description of what the handler does.
//
// It can span across multiple lines like this.
func handler(ctx context.Context, req *simba.Request[RequestBody, Params]) (*simba.Response[ResponseBody], error) {

	// Access the request body and params fields
	// req.Body.Age
	// req.Body.Text
	// req.Params.Name
	// req.Params.ID
	// req.Params.Page
	// req.Params.Size

	// Access the request cookies
	// req.Cookies

	// Access the request headers
	// req.Headers

	return &simba.Response[ResponseBody]{
		Body: ResponseBody{
			Message: fmt.Sprintf("Hello %s, you are %d years old", req.Params.Name, req.Body.Age),
			ID:      req.Params.ID,
			Active:  req.Params.Active,
			Page:    req.Params.Page,
			Size:    req.Params.Size,
			Name:    req.Params.Name,
			Score:   req.Params.Score,
		},
	}, nil
}

func main() {
	app := simba.Default(settings.Config{
		Server: settings.Server{
			Port: 9999,
		},
	})
	app.Router.POST("/params/{id}", simba.JsonHandler(handler))
	app.Start()
}
