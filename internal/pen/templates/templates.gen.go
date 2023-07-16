package templates

import (
	"embed"
	"fmt"
	"text/template"

	"github.tesla.cn/itapp/lines/errorx"
)

const (
	TmplDirName       = "tmpls"
	NewProjectDirName = "newproject"
	NewModuleDirName  = "newmodule"
)

//go:embed tmpls/*
var tmplFiles embed.FS

//go:embed newproject/*
var newProjectFiles embed.FS

//go:embed newmodule/*
var newModuleFiles embed.FS

var (
	dirMapping = map[string]embed.FS{
		TmplDirName:       tmplFiles,
		NewProjectDirName: newProjectFiles,
		NewModuleDirName:  newModuleFiles,
	}
)

// Parse parses declared templates.
func Parse(t *template.Template, dirName string) error {
	embedFiles := dirMapping[dirName]
	files, err := embedFiles.ReadDir(dirName)
	if err != nil {
		return errorx.WithStack(err)
	}
	templates := make(map[string]string)
	for _, file := range files {
		content, err := embedFiles.ReadFile(fmt.Sprintf("%s/%s", dirName, file.Name()))
		if err != nil {
			return errorx.WithStack(err)
		}
		templates[file.Name()] = string(content)
	}

	for name, s := range templates {
		if _, err := t.New(name).Parse(s); err != nil {
			return err
		}
	}
	return nil
}
