package simbaOpenapi

type DocsParams struct {
	OpenAPIFileType string `exhaustruct:"optional"`
	OpenAPIPath     string
	DocsPath        string
	ServiceName     string
}
