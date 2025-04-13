package simbaOpenapi

import (
	"fmt"
	"net/http"
)

// ScalarDocsHandler returns a handler that serves the API documentation using the Scalar API Reference component.
func ScalarDocsHandler(params DocsParams) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `
			<!doctype html>
			<html>
			  <head>
				<title>%s - API Reference</title>
				<meta charset="utf-8" />
				<meta
				  name="viewport"
				  content="width=device-width, initial-scale=1" />
			  </head>
			  <body>
				<script
				  id="api-reference"
				  type="%s"
				  data-url="%s"
				  data-proxy-url="https://proxy.scalar.com"></script>
				<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
			  </body>
			</html>`, params.ServiceName, params.OpenAPIFileType, params.OpenAPIPath,
		)
	}
}
