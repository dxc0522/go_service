package options

import "github.com/getkin/kin-openapi/openapi3"

type (
	Options struct {
		Swagger           *openapi3.T
		AppName           string
		ModuleName        string
		AppPackage        string
		TargetDir         string
		DataSource        string
		TablePrefix       string
		GenerateClient    bool
		GenerateMigration bool
	}
)
