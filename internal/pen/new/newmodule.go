package new

import (
	"path"
	"strings"
	"text/template"

	"github.tesla.cn/itapp/lines/constantx"
	"github.tesla.cn/itapp/lines/errorx"
	"github.com/go_service/internal/pen/pkg/options"
	"github.com/go_service/internal/pen/templates"
)

func GenerateNewModules(opts options.Options) error {
	//TemplateFunctions["opts"] = func() options.Options { return opts }
	t := template.New(opts.AppName) //.
	//Funcs(TemplateFunctions).Funcs(sprig.TxtFuncMap())
	gens := []GeneratorFunc{
		Generator{nil, t, opts,
			"Makefile", "Makefile", false}.GenerateMakefile,
		Generator{nil, t, opts,
			"ModuleName.yaml", "ModuleName.yaml", false}.GenerateOpenAPI,
		Generator{nil, t, opts,
			"pen.yaml", "pen.yaml", false}.GeneratePenYaml,
	}

	err := templates.Parse(t, templates.NewModuleDirName)
	if err != nil {
		return errorx.Wrap(err, "error parsing pen templates")
	}

	for _, g := range gens {
		if err := g(); err != nil {
			return errorx.WithMessage(err, "error when generating target files")
		}
	}
	return nil
}

func (g Generator) GenerateMakefile() error {
	context := struct {
		ModuleName string
		AppPackage string
	}{
		g.ModuleName,
		g.AppPackage,
	}

	content, err := g.ExecuteTemplate(context)
	if err != nil {
		return err
	}
	target := path.Join(g.TargetDir, g.Options.ModuleName, g.TargetFile)
	return g.writeFile(target, content)
}

func (g Generator) GenerateOpenAPI() error {
	context := struct {
		ModuleName string
	}{
		g.ModuleName,
	}

	content, err := g.ExecuteTemplate(context)
	if err != nil {
		return err
	}
	g.TargetFile = strings.ReplaceAll(g.TargetFile, "ModuleName", g.ModuleName)
	target := path.Join(g.TargetDir, g.Options.ModuleName, g.TargetFile)
	return g.writeFile(target, content)
}

func (g Generator) GeneratePenYaml() error {
	context := struct {
		ModuleName   string
		LinesVersion string
	}{
		g.ModuleName,
		constantx.LinesVersion,
	}

	content, err := g.ExecuteTemplate(context)
	if err != nil {
		return err
	}
	target := path.Join(g.TargetDir, g.Options.ModuleName, g.TargetFile)
	return g.writeFile(target, content)
}
