package simbaTest

import "github.com/google/uuid"

type User struct {
	ID   int    `json:"id" description:"ID of the user" example:"1"`
	Name string `json:"name" description:"Name of the user" example:"John Doe"`
	Role string `json:"role" description:"Role of the user" example:"admin"`
}

type Params struct {
	ID      uuid.UUID `path:"id" description:"ID of the user" example:"1"`
	Gender  string    `query:"gender" description:"Gender of the user" example:"male"`
	TraceID string    `header:"X-TRACE-ID" description:"Request ID" example:"1234"`
	Active  bool      `query:"active" description:"Active status of the user" example:"true"`
	Page    int       `query:"page" description:"Page number" example:"1" default:"1"`
	Size    int64     `query:"size" description:"Page size" example:"10" default:"10"`
	Score   float64   `query:"score" description:"User score" example:"9.5" default:"10.0"`
}

type RequestBody struct {
	Name        string `json:"name" description:"Name of the user" example:"John Doe" validate:"required"`
	Age         int    `json:"age" description:"Age of the user" example:"30"`
	Description string `json:"description" description:"description of the user" example:"A test user"`
}

type ResponseBody struct {
	ID          uuid.UUID `json:"id" description:"ID of the user" example:"1"`
	Name        string    `json:"name" description:"Name of the user" example:"John Doe"`
	Age         int       `json:"age" description:"Age of the user" example:"30"`
	Description string    `json:"description" description:"description of the user" example:"A test user"`
}

// AuthRequestBody is a test struct for authenticated request body
type AuthRequestBody struct {
	Token string `json:"token" validate:"required"`
}
