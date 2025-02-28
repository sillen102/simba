package test

// RequestBody is a test struct for request body
type RequestBody struct {
	Test string `json:"test" validate:"required"`
}

// Params is a test struct for request params
type Params struct {
	Name   string  `header:"name" validate:"required"`
	ID     int     `path:"id" validate:"required"`
	Active bool    `query:"active" validate:"required"`
	Page   int     `query:"page" validate:"omitempty,min=0" default:"1"`
	Size   int64   `query:"size" validate:"omitempty,min=0" default:"10"`
	Score  float64 `query:"score" default:"10.0"`
}

// User is a test struct for user entity for authenticated routes
type User struct {
	ID   int
	Name string
	Role string
}

// AuthRequestBody is a test struct for authenticated request body
type AuthRequestBody struct {
	Token string `json:"token" validate:"required"`
}

type AuthParams struct {
	Token string `header:"Authorization" validate:"required"`
}
