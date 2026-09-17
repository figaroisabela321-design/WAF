package docs

import _ "embed"

// OpenAPIYAML is the OpenAPI 3.0 specification.
//
//go:embed openapi.yaml
var OpenAPIYAML []byte
