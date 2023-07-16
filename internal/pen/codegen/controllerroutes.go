package codegen

import (
	"path"

	"github.com/getkin/kin-openapi/openapi3"
)

func (g Generator) GenerateControllerRoutes() error {
	context := struct {
		Paths openapi3.Paths
	}{
		g.T.Paths,
	}

	content, err := g.ExecuteTemplate(context)
	if err != nil {
		return err
	}
	target := path.Join(g.TargetDir, g.TargetFile)
	return g.writeFile(target, content)
}
