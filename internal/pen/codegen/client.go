package codegen

import (
	"fmt"
	"path"

	"github.tesla.cn/itapp/lines/errorx"
)

func (g Generator) GenerateClient() error {
	ops, err := OperationDefinitions(g.T)
	if err != nil {
		return errorx.Wrap(err, "error creating operation definitions")
	}

	content, err := g.ExecuteTemplate(ops)
	if err != nil {
		return err
	}
	packageName := SDKPackageName(g.AppName)
	content = []byte(fmt.Sprintf("package %s\n\n%s", packageName, content))
	target := path.Join(g.TargetDir, g.TargetFile)
	if err := g.writeFile(target, content); err != nil {
		return err
	}

	return nil
}
