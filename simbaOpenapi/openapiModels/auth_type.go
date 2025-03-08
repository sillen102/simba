package openapiModels

type AuthType int

const (
	AuthTypeBasic AuthType = iota
	AuthTypeAPIKey
	AuthTypeBearer
)
