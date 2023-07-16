package codegen

import (
	"path"

	"github.tesla.cn/itapp/lines/errorx"
)

func (g Generator) GenerateBizLogicInterface() error {
	ops, err := OperationDefinitions(g.T)
	if err != nil {
		return errorx.Wrap(err, "error creating operation definitions")
	}

	context := struct {
		AppName              string
		AppPackage           string
		OperationDefinitions []OperationDefinition
	}{
		g.AppName,
		g.AppPackage,
		ops,
	}

	content, err := g.ExecuteTemplate(context)
	if err != nil {
		return err
	}
	target := path.Join(g.TargetDir, g.TargetFile)
	if err := g.writeFile(target, content); err != nil {
		return err
	}

	return nil
}
