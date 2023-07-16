package codegen

import (
	"path"
)

func (g Generator) GenerateConfigYaml() error {
	context := struct {
		AppName string
	}{
		g.AppName,
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
