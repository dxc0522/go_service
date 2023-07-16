package codegen

import (
	"fmt"
	"path"

	"github.tesla.cn/itapp/lines/filex"
)

func (g Generator) GenerateModelByComponent() error {
	typeDefs, err := GenerateTypeDefinitions(g.T)
	if err != nil {
		return err
	}
	for _, typeDef := range typeDefs {
		context := struct {
			TypeDefinition
			PackageName string
		}{
			typeDef,
			"model",
		}
		if g.Options.GenerateClient {
			context.PackageName = SDKPackageName(g.AppName)
		}

		content, err := g.ExecuteTemplate(context)
		if err != nil {
			return err
		}
		target := path.Join(g.TargetDir, fmt.Sprintf(g.TargetFile, filex.GoFileName(typeDef.TypeName)))
		if err := g.writeFile(target, content); err != nil {
			return err
		}
	}
	return nil
}
