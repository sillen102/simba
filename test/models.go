package test

// RequestBody is a test struct for request body
type RequestBody struct {
	Test string `json:"test" validate:"required"`
}

// Params is a test struct for request params
type Params struct {
	Name string `header:"name" validate:"required"`
	ID   int    `path:"id" validate:"required"`
	Page int64  `query:"page" validate:"required"`
	Size int64  `query:"size" validate:"required"`
}
