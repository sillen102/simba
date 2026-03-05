module github.com/sillen102/simba/websocket

go 1.26

replace github.com/sillen102/simba => ../

replace github.com/sillen102/simba/models => ../models

require (
	github.com/coder/websocket v1.8.14
	github.com/google/uuid v1.6.0
	github.com/sillen102/simba v0.29.5
	github.com/sillen102/simba/models v0.30.0-dev.1
)

require (
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.1 // indirect
	github.com/hashicorp/go-envparse v0.1.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/sillen102/config-loader v0.3.0 // indirect
	github.com/swaggest/jsonschema-go v0.3.79 // indirect
	github.com/swaggest/openapi-go v0.2.60 // indirect
	github.com/swaggest/refl v1.4.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
