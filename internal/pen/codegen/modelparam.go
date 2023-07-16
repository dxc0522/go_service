package codegen

import (
	"fmt"
	"path"

	"github.tesla.cn/itapp/lines/errorx"
	"github.tesla.cn/itapp/lines/filex"
)

func (g Generator) GenerateModelByParam() error {
	ops, err := OperationDefinitions(g.T)
	if err != nil {
		return errorx.Wrap(err, "error creating operation definitions")
	}
	for _, op := range ops {
		for _, typeDef := range op.TypeDefinitions {
			context := struct {
				TypeDefinition
				OperationId string
				PackageName string
			}{
				typeDef,
				op.OperationId,
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
	}

	return nil
}
