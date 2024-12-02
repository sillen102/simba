package test

// RequestBody is a test struct for request body
type RequestBody struct {
	Test string `json:"test" validate:"required"`
}

// Params is a test struct for request params
type Params struct {
	Name   string `header:"name" validate:"required"`
	ID     int    `path:"id" validate:"required"`
	Active bool   `query:"active" validate:"required"`
	Page   int64  `query:"page" validate:"min=0"`
	Size   int64  `query:"size" validate:"min=0"`
}

// User is a test struct for user entity for authenticated routes
type User struct {
	ID   int
	Name string
	Role string
}
