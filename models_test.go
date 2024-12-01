package simba_test

// RequestBody is a test struct for request body
type RequestBody struct {
	Test string `json:"test" validate:"required"`
}

// Params is a test struct for request params
type Params struct {
	Page int64  `query:"page" validate:"required"`
	Size int64  `query:"size" validate:"required"`
	ID   string `path:"id" validate:"required"`
	Name string `query:"name" validate:"required"`
}
