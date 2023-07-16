package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/getkin/kin-openapi/openapi3"

	"github.tesla.cn/itapp/lines/errorx"
	"github.tesla.cn/itapp/lines/filex"
	"github.tesla.cn/itapp/lines/logx"
	"github.com/go_service/internal/pen/pkg"
	"github.com/go_service/internal/pen/pkg/options"
	"github.com/go_service/internal/pen/templates"
)

const (
	penFileSuffix      = ".pen.go"
	modifiedFileSuffix = ".go"
)

type (
	GeneratorFunc func() error
	Generator     struct {
		*openapi3.T
		*template.Template
		options.Options
		TemplateName string
		TargetFile   string
		PenFile      bool // auto regenerated file
	}
)

func SDKPackageName(appName string) string {
	return fmt.Sprintf("%ssdk", filex.GoFileName(appName))
}

func clearDirPenFiles(dir string) error {
	if exits, err := filex.Exists(dir); err != nil {
		return err
	} else if !exits {
		return nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return errorx.WithStack(err)
	}
	defer func() {
		_ = d.Close()
	}()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return errorx.WithStack(err)
	}
	for _, name := range names {
		if strings.HasSuffix(name, penFileSuffix) {
			err = os.Remove(filepath.Join(dir, name))
			if err != nil {
				return errorx.WithStack(err)
			}
		}
	}
	return nil
}

func clearClientPenFiles(opts options.Options) error {
	return clearDirPenFiles(path.Join(opts.TargetDir, SDKPackageName(opts.AppName)))
}

func clearPenFiles(opts options.Options) error {
	dirNames := []string{"bizlogic", "cmd", "controller", "model"}
	for _, dirName := range dirNames {
		if err := clearDirPenFiles(path.Join(opts.TargetDir, dirName)); err != nil {
			return err
		}
	}
	return nil
}

func getMigrationGenerators(opts options.Options, t *template.Template) []GeneratorFunc {
	generators := []GeneratorFunc{
		Generator{nil, t, opts,
			"goose", "db/migrations/%s_feature.sql", false}.GenerateMigrationFile,
	}
	return generators
}

func getClientGenerators(swagger *openapi3.T, opts options.Options, t *template.Template) []GeneratorFunc {
	packageName := SDKPackageName(opts.AppName)
	generators := []GeneratorFunc{
		Generator{swagger, t, opts,
			"client", fmt.Sprintf("%s/sdk", packageName), true}.GenerateClient,
		Generator{swagger, t, opts,
			"modelcomponent", fmt.Sprintf("%s/", packageName) + "%s", true}.GenerateModelByComponent,
		Generator{swagger, t, opts,
			"modelparam", fmt.Sprintf("%s/", packageName) + "%s", true}.GenerateModelByParam,
		Generator{swagger, t, opts,
			"modelrequestbodies", fmt.Sprintf("%s/", packageName) + "%s", true}.GenerateModelByRequestBodies,
	}
	return generators
}

func getGenerators(swagger *openapi3.T, opts options.Options, t *template.Template) []GeneratorFunc {
	generators := []GeneratorFunc{
		Generator{swagger, t, opts,
			"configyaml", "config/app.local.yaml", false}.GenerateConfigYaml,
		Generator{swagger, t, opts,
			"modelcomponent", "model/%s", true}.GenerateModelByComponent,
		Generator{swagger, t, opts,
			"modelparam", "model/%s", true}.GenerateModelByParam,
		Generator{swagger, t, opts,
			"modelrequestbodies", "model/%s", true}.GenerateModelByRequestBodies,
		Generator{swagger, t, opts,
			"init", "init%s.go", false}.GenerateInit,
		Generator{swagger, t, opts,
			"bizlogicinterface", "bizlogic/interface", true}.GenerateBizLogicInterface,
		Generator{swagger, t, opts,
			"bizlogicbizlogic", "bizlogic/bizlogic.go", false}.GenerateBizLogicBizLogic,
		Generator{swagger, t, opts,
			"bizlogic", "bizlogic/%s.go", false}.GenerateBizLogic,
		Generator{swagger, t, opts,
			"controller", "controller/%s", true}.GenerateController,
		Generator{swagger, t, opts,
			"controllerroutes", "controller/controller", true}.GenerateControllerRoutes,
		Generator{swagger, t, opts,
			"cmd", "cmd/%s", true}.GenerateCmd,
	}
	return generators
}

func Generate(opts options.Options) error {
	TemplateFunctions["opts"] = func() options.Options { return opts }
	t := template.New(opts.AppName).Funcs(TemplateFunctions).Funcs(sprig.TxtFuncMap())

	err := templates.Parse(t, templates.TmplDirName)
	if err != nil {
		return errorx.Wrap(err, "error parsing pen templates")
	}
	var gens []GeneratorFunc

	if opts.GenerateMigration {
		gens = getMigrationGenerators(opts, t)
	} else if opts.GenerateClient {
		if err := clearClientPenFiles(opts); err != nil {
			return err
		}
		gens = getClientGenerators(opts.Swagger, opts, t)
	} else {
		if err := clearPenFiles(opts); err != nil {
			return err
		}
		gens = getGenerators(opts.Swagger, opts, t)
	}

	for _, g := range gens {
		if err := g(); err != nil {
			return errorx.WithMessage(err, "error when generating target files")
		}
	}
	return nil
}

func (g Generator) GenerateCmd() error {
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
	target := path.Join(g.TargetDir, fmt.Sprintf(g.TargetFile, filex.GoFileName(g.AppName)))
	return g.writeFile(target, content)
}

func (g Generator) GenerateInit() error {
	context := struct {
		AppPackage string
		AppName    string
	}{
		g.AppPackage,
		g.AppName,
	}

	content, err := g.ExecuteTemplate(context)
	if err != nil {
		return err
	}
	target := path.Join(g.TargetDir, fmt.Sprintf(g.TargetFile, filex.GoFileName(g.AppName)))
	return g.writeFile(target, content)
}

func (g Generator) ExecuteTemplate(context interface{}) ([]byte, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err := g.Template.ExecuteTemplate(w, g.TemplateName+pkg.TemplateSuffix, context)
	if err != nil {
		return nil, errorx.Wrap(err, "error when generating "+g.TemplateName)
	}
	err = w.Flush()
	if err != nil {
		return nil, errorx.Wrap(err, "error when generating "+g.TemplateName)
	}
	return buf.Bytes(), nil
}

func (g Generator) writeFile(target string, toBeFormat []byte) error {
	result := toBeFormat
	if !strings.HasSuffix(target, ".yaml") && !strings.HasSuffix(target, ".sql") {
		var err error
		content := []byte(pkg.ImportFormat(string(toBeFormat)))
		newContent, err := format.Source(content)
		if err != nil {
			return errorx.WithStack(err)
		}
		result = newContent
	}
	filex.EnsureDir(filepath.Dir(target))
	if g.PenFile {
		return g.overridePenFile(target, result)
	}
	if exists, err := filex.Exists(target); err != nil {
		return err
	} else if exists {
		logx.Info("skip non pen file", "file", target)
		return nil
	}
	return errorx.WithStack(ioutil.WriteFile(target, result, 0644))
}

func (g Generator) overridePenFile(target string, content []byte) error {
	if exists, err := filex.Exists(target + modifiedFileSuffix); err != nil {
		return err
	} else if exists {
		logx.Info("skip modified file", "file", target+modifiedFileSuffix)
		return nil
	}
	return errorx.WithStack(ioutil.WriteFile(target+penFileSuffix, content, 0644))
}
