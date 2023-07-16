package codegen

import (
	"fmt"
	"path"

	"github.tesla.cn/itapp/lines/errorx"
	"github.tesla.cn/itapp/lines/filex"
)

func (g Generator) GenerateController() error {
	ops, err := OperationDefinitions(g.T)
	if err != nil {
		return errorx.Wrap(err, "error creating operation definitions")
	}

	for _, operationDefinition := range ops {
		context := struct {
			AppName    string
			AppPackage string
			OperationDefinition
		}{
			g.AppName,
			g.AppPackage,
			operationDefinition,
		}

		content, err := g.ExecuteTemplate(context)
		if err != nil {
			return err
		}
		target := path.Join(g.TargetDir, fmt.Sprintf(g.TargetFile, filex.GoFileName(operationDefinition.OperationId)))
		if err := g.writeFile(target, content); err != nil {
			return err
		}
	}
	return nil
}
