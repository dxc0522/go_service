package codegen

import (
	"path"
)

func (g Generator) GenerateBizLogicBizLogic() error {
	context := struct {
		AppName    string
		AppPackage string
	}{
		g.AppName,
		g.AppPackage,
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
