package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/sillen102/simba"
)

type RequestBody struct {
	Age int `json:"age"`
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
	zerolog.Ctx(context.Background()).Info().Msg("Listening on http://localhost:9999")
	http.ListenAndServe(":9999", router)
}
