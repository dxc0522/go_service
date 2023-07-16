package main

import (
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/subchen/go-log"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go_service/internal/pen/codegen"
	"github.com/go_service/internal/pen/genmock/warp"
	"github.com/go_service/internal/pen/new"
	"github.com/go_service/internal/pen/pkg/options"
)

const (
	// command
	CommandVersion   = "version"
	CommandModule    = "module"
	CommandProject   = "project"
	CommandClient    = "client"
	CommandStructure = "structure"
	CommandMigration = "migration"
	CommandDbmodel   = "dbmodel"

	// option common
	OptionVersion     = "version"
	OptionTargetDir   = "target-dir"
	OptionAppName     = "app-name"
	OptionAppPackage  = "app-package"
	OptionDataSource  = "data-source"
	OptionTablePrefix = "table-prefix"

	// option dbmodel
	OptionSqlType   = "sqltype"
	OptionConnStr   = "connstr"
	OptionDatabase  = "database"
	OptionTables    = "tables"
	OptionModel     = "model"
	OptionGrom      = "gorm"
	OptionGuregu    = "guregu"
	OptionOverwrite = "overwrite"
	OptionOutdir    = "out"
)

var (
	// flag
	FlagOptionVersion = &cli.BoolFlag{
		Name:    OptionVersion,
		Aliases: []string{"v"},
		Value:   false,
		Usage:   "Show the version of pen",
	}
	FlagOptionTargetDir = &cli.StringFlag{
		Name:    OptionTargetDir,
		Aliases: []string{"o"},
		Value:   "",
		Usage:   " output dir,default current dir",
	}
	FlagOptionAppName = &cli.StringFlag{
		Name:  OptionAppName,
		Value: "",
		Usage: " app name,default calculate autoly ",
	}
	FlagOptionAppPackage = &cli.StringFlag{
		Name:  OptionAppPackage,
		Value: "",
		Usage: " app package name,default calculate autoly",
	}
	FlagOptionDataSource = &cli.StringFlag{
		Name:  OptionDataSource,
		Value: "",
		Usage: " data source DSN",
	}
	FlagOptionTablePrefix = &cli.StringFlag{
		Name:  OptionTablePrefix,
		Value: "",
		Usage: " table name prefix",
	}

	FlagOptionSqlType = &cli.StringFlag{
		Name:  OptionSqlType,
		Value: "",
		Usage: " db type,default mysql",
	}
	FlagOptionConnStr = &cli.StringFlag{
		Name:    OptionConnStr,
		Aliases: []string{"c"},
		Value:   "",
		Usage:   " db connent DSN",
	}
	FlagOptionDatabase = &cli.StringFlag{
		Name:    OptionDatabase,
		Aliases: []string{"d"},
		Value:   "",
		Usage:   " db databse name ",
	}
	FlagOptionTables = &cli.StringFlag{
		Name:    OptionTables,
		Aliases: []string{"t"},
		Value:   "",
		Usage:   " db tables name ,join with comma if multiple",
	}
	FlagOptionModel = &cli.StringFlag{
		Name:  OptionModel,
		Value: "",
		Usage: " output package name,,default dbmodel",
	}
	FlagOptionGrom = &cli.BoolFlag{
		Name:  OptionGrom,
		Value: true,
		Usage: " equal to gen --gorm",
	}
	FlagOptionGuregu = &cli.BoolFlag{
		Name:  OptionGuregu,
		Value: true,
		Usage: " equal to gen --guregu",
	}
	FlagOptionOverwrite = &cli.BoolFlag{
		Name:  OptionOverwrite,
		Value: true,
		Usage: " equal to gen --overwrite",
	}
	FlagOptionOutdir = &cli.StringFlag{
		Name:  OptionOutdir,
		Value: "",
		Usage: " output dir,default current dir ./",
	}
)

var (
	opts                  = options.Options{}
	PenYaml               = make(map[string]string)
	PenYamlKeySrcYaml     = "src-yaml"
	PenYamlKeyTargetDir   = OptionTargetDir
	PenYamlKeyAppName     = OptionAppName
	PenYamlKeyAppPackage  = OptionAppPackage
	PenYamlKeyDataSource  = OptionDataSource
	PenYamlKeyTablePrefix = OptionTablePrefix

	PenYamlKeySqlType   = OptionSqlType
	PenYamlKeyConnStr   = OptionConnStr
	PenYamlKeyDatabase  = OptionDatabase
	PenYamlKeyTables    = OptionTables
	PenYamlKeyModel     = OptionModel
	PenYamlKeyGorm      = OptionGrom
	PenYamlKeyGuregu    = OptionGuregu
	PenYamlKeyOverwrite = OptionOverwrite
	PenYamlKeyOutdir    = OptionOutdir
)

func main() {
	if err := Main(); err != nil {
		logx.Error(fmt.Sprintf("Main func error :\n%+v", err))
		return
	}
}

func loadPenYaml(penYamlFilePath string) error {
	if os.Stat(penYamlFilePath); os.IsNotExist(err) {
		return error.New(err)
	}
	if !existPenYaml {
		return nil
	}

	fp, err := os.Open(penYamlFilePath)
	if err != nil {
		return error.New(err, "open fail")
	} else {
		defer fp.Close()
	}
	bytes, err := ioutil.ReadAll(fp)
	if err != nil {
		return error.New(err, "read fail")
	}
	err = yaml.Unmarshal(bytes, &PenYaml)
	if err != nil {
		return error.New(err, "Unmarshal fail")
	}

	return nil
}

func Main() (err error) {
	err = loadPenYaml("pen.yaml")
	if err != nil {
		return
	}

	app := &cli.App{
		Name:  "pen",
		Usage: "pen is a code generate tool",
		Flags: []cli.Flag{
			FlagOptionVersion, FlagOptionAppName, FlagOptionAppPackage, FlagOptionTargetDir,
		},
		Action: func(c *cli.Context) error {
			if c.IsSet(FlagOptionVersion.Name) && c.Bool(FlagOptionVersion.Name) {
				fmt.Println("pen version", "version")
				return nil
			}

			yamlFilePath := getFirstArg(c)
			if yamlFilePath == "" {
				return error.New("must pass yaml file path")
			}

			var err error
			opts.AppName = mustAppName(c)
			opts.AppPackage = mustAppPackage(c)
			opts.TargetDir = mustTargetDir(c)
			opts.Swagger, err = loadSwaggerWithExtends(yamlFilePath)
			if err != nil {
				return error.New(err)
			}

			fmt.Println("generating app structure", "app name", opts.AppName)
			err = codegen.Generate(opts)
			if err != nil {
				return err
			}
			fmt.Println("generate app structure finished")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  CommandVersion,
				Usage: FlagOptionVersion.Usage,
				Action: func(c *cli.Context) error {
					fmt.Println("pen version", "version")
					return nil
				},
			},
			{
				Name:  CommandModule,
				Usage: "Followed by module name,generate module's dir and init file",
				Action: func(c *cli.Context) error {
					opts.ModuleName = getFirstArg(c)
					if opts.ModuleName == "" {
						return error.New("need pass module name ")
					}

					fmt.Println("generating module", "module name", opts.ModuleName)
					err := new.GenerateNewModules(opts)
					if err != nil {
						return err
					}
					fmt.Println("generate module finished")
					return nil
				},
			},
			{
				Name:  CommandProject,
				Usage: "Followed by project name,generate project",
				Action: func(c *cli.Context) error {
					return new.GenerateNewProject(opts)
				},
			},
			{
				Name:    CommandClient,
				Aliases: []string{"c", "sdk"},
				Usage:   "Followed with yaml file,generate client sdk",
				Flags:   []cli.Flag{FlagOptionAppName, FlagOptionTargetDir},
				Action: func(c *cli.Context) error {
					var err error
					opts.GenerateClient = true
					opts.AppName = mustAppName(c)
					opts.TargetDir = mustTargetDir(c)
					opts.Swagger, err = loadSwaggerWithExtends(mustYamlFilePath(c))
					if err != nil {
						return error.New(err)
					}

					fmt.Println("generating client sdk", "app name", opts.AppName)
					err = codegen.Generate(opts)
					if err != nil {
						return err
					}
					fmt.Println("generate client sdk finished")
					return nil
				},
			},
			{
				Name:    CommandStructure,
				Aliases: []string{"s", "logic"},
				Usage:   "Followed with yaml file,generate app structure",
				Flags:   []cli.Flag{FlagOptionAppName, FlagOptionAppPackage, FlagOptionTargetDir},
				Action: func(c *cli.Context) error {
					var err error
					opts.AppName = mustAppName(c)
					opts.AppPackage = mustAppPackage(c)
					opts.TargetDir = mustTargetDir(c)
					opts.Swagger, err = loadSwaggerWithExtends(mustYamlFilePath(c))
					if err != nil {
						return error.New(err)
					}

					fmt.Println("generating app structure", "app name", opts.AppName)
					err = codegen.Generate(opts)
					if err != nil {
						return err
					}
					fmt.Println("generate app structure finished")
					return nil
				},
			},
			{
				Name:    CommandMigration,
				Aliases: []string{"m"},
				Usage:   "Followed with yaml file,generate db migration file",
				Flags:   []cli.Flag{FlagOptionTargetDir, FlagOptionDataSource, FlagOptionTablePrefix},
				Action: func(c *cli.Context) error {
					var err error
					opts.GenerateMigration = true
					opts.TargetDir = mustTargetDir(c)
					opts.DataSource = mustDataSource(c)
					opts.TablePrefix = mustTablePrefix(c)

					fmt.Println("generating migration")
					err = codegen.Generate(opts)
					if err != nil {
						return err
					}
					fmt.Println("generate migration finished")
					return nil
				},
			},
			{
				Name:  CommandDbmodel,
				Usage: "Generate dbmodel",
				Flags: []cli.Flag{FlagOptionSqlType, FlagOptionConnStr, FlagOptionDatabase, FlagOptionTables, FlagOptionModel,
					FlagOptionGrom, FlagOptionGuregu, FlagOptionOverwrite, FlagOptionOutdir},
				Action: func(c *cli.Context) error {
					genOpt := warp.NewOption()
					genOpt.SqlType = mustSqlType(c)        // --sqltype mysql
					genOpt.SqlConnStr = mustConStr(c)      // --connstr root:root@tcp(127.0.0.1)/bjm...
					genOpt.SqlDatabase = mustDatabase(c)   // -d bjm
					genOpt.SqlTable = mustTables(c)        // -t user_user,user_trt
					genOpt.ModelPackageName = mustModel(c) // --model=dbmodel
					genOpt.Gorm = mustGrom(c)              // --gorm
					genOpt.Guregu = mustGuregu(c)          // --guregu
					genOpt.Overwrite = mustOverwrite(c)    // --overwrite
					genOpt.Outdir = mustOutdir(c)          // --out ./

					fmt.Println("generating dbmodel")
					err = warp.DoGenDBmodel(genOpt)
					if err != nil {
						return err
					}
					fmt.Println("generate dbmodel finished")
					return nil
				},
			},
		},
	}
	return app.Run(os.Args)
}

func getFirstArg(c *cli.Context) string {
	if c.NArg() > 0 {
		return c.Args().First()
	}

	return ""
}

func mustYamlFilePath(c *cli.Context) string {
	yamlFilePath := getFirstArg(c)
	if yamlFilePath == "" {
		yamlFilePath = PenYaml[PenYamlKeySrcYaml]
	}

	if yamlFilePath != "" {
		return yamlFilePath
	}

	panic("Please specify a path to a OpenAPI 3.0 spec file")
}

func mustAppName(c *cli.Context) string {
	res := c.String(FlagOptionAppName.Name)
	if res != "" {
		return res
	}
	if PenYaml[PenYamlKeyAppName] != "" {
		return PenYaml[PenYamlKeyAppName]
	}

	yamlFilePath := mustYamlFilePath(c)
	nameParts := strings.Split(filepath.Base(yamlFilePath), ".")
	return nameParts[0]
}

func mustTargetDir(c *cli.Context) string {
	res := c.String(FlagOptionTargetDir.Name)
	if res != "" {
		return res
	}
	if PenYaml[PenYamlKeyTargetDir] != "" {
		return PenYaml[PenYamlKeyTargetDir]
	}

	workDir, err := os.Getwd()
	if err != nil {
		panic(error.New(err))
	}
	return workDir
}

func mustAppPackage(c *cli.Context) string {
	res := c.String(FlagOptionAppPackage.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyAppPackage] != "" {
		return PenYaml[PenYamlKeyAppPackage]
	}

	res, err := calculatePackageNameByGomod()
	if err != nil {
		panic(err)
	}
	return res
}

func mustDataSource(c *cli.Context) string {
	res := c.String(FlagOptionDataSource.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyDataSource] != "" {
		return PenYaml[PenYamlKeyDataSource]
	}

	panic("lack" + FlagOptionDataSource.Name)
}

func mustTablePrefix(c *cli.Context) string {
	res := c.String(FlagOptionTablePrefix.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyTablePrefix] != "" {
		return PenYaml[PenYamlKeyTablePrefix]
	}

	// panic("lack" + FlagOptionTablePrefix.Name) // allow empty mean no filter
	return ""
}

func mustSqlType(c *cli.Context) string {
	res := c.String(FlagOptionSqlType.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeySqlType] != "" {
		return PenYaml[PenYamlKeySqlType]
	}

	return "mysql"
}

func mustConStr(c *cli.Context) string {
	res := c.String(FlagOptionConnStr.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyConnStr] != "" {
		return PenYaml[PenYamlKeyConnStr]
	}

	panic("lack " + FlagOptionConnStr.Name)
}

func mustDatabase(c *cli.Context) string {
	res := c.String(FlagOptionDatabase.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyDatabase] != "" {
		return PenYaml[PenYamlKeyDatabase]
	}

	panic("lack " + FlagOptionDatabase.Name)
}

func mustTables(c *cli.Context) string {
	res := c.String(FlagOptionTables.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyTables] != "" {
		return PenYaml[PenYamlKeyTables]
	}

	panic("lack " + FlagOptionTables.Name)
}

func mustModel(c *cli.Context) string {
	res := c.String(FlagOptionModel.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyModel] != "" {
		return PenYaml[PenYamlKeyModel]
	}

	return "dbmodel"
}

func mustGrom(c *cli.Context) bool {
	if c.IsSet(FlagOptionGrom.Name) {
		return c.Bool(FlagOptionGrom.Name)
	}

	if PenYaml[PenYamlKeyGorm] != "" {
		v, _ := strconv.ParseBool(PenYaml[PenYamlKeyGorm])
		return v
	}

	return true
}

func mustGuregu(c *cli.Context) bool {
	if c.IsSet(FlagOptionGuregu.Name) {
		return c.Bool(FlagOptionGuregu.Name)
	}

	if PenYaml[PenYamlKeyGuregu] != "" {
		v, _ := strconv.ParseBool(PenYaml[PenYamlKeyGuregu])
		return v
	}

	return true
}

func mustOverwrite(c *cli.Context) bool {
	if c.IsSet(FlagOptionOverwrite.Name) {
		return c.Bool(FlagOptionOverwrite.Name)
	}

	if PenYaml[PenYamlKeyOverwrite] != "" {
		v, _ := strconv.ParseBool(PenYaml[PenYamlKeyOverwrite])
		return v
	}

	return true
}

func mustOutdir(c *cli.Context) string {
	res := c.String(FlagOptionOutdir.Name)
	if res != "" {
		return res
	}

	if PenYaml[PenYamlKeyOutdir] != "" {
		return PenYaml[PenYamlKeyOutdir]
	}

	return "./"
}

func loadSwagger(filePath string) (swagger *openapi3.T, err error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadFromURI(u)
	} else {
		return loader.LoadFromFile(filePath)
	}
}

func loadSwaggerWithExtends(filePath string) (*openapi3.T, error) {
	res, err := loadSwagger(filePath)
	if err != nil {
		return nil, error.New(err)
	}

	extends, err := loadExtendsSwagger(res)
	if err != nil {
		return nil, err
	}

	for _, extend := range extends {
		if extend != nil {
			for k, v := range extend.Paths {
				if _, ok := res.Paths[k]; ok {
					return nil, error.New("duplicate path in extends openapi: %s", k)
				}
				v := v
				res.Paths[k] = v
			}

			for k, v := range extend.Components.Schemas {
				if _, ok := res.Components.Schemas[k]; ok {
					return nil, error.New("duplicate scheme in extends openapi: %s", k)
				}
				v := v
				res.Components.Schemas[k] = v
			}
		}
	}

	return res, nil

}

func loadExtendsSwagger(src *openapi3.T) (res []*openapi3.T, err error) {
	xExtends, ok := src.Extensions[codegen.XExtends]
	if !ok {
		return
	}

	jsonRawMessage, ok := xExtends.(json.RawMessage)
	if !ok {
		return
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(jsonRawMessage, &m)
	if err != nil {
		return nil, error.New(err)
	}

	xExtendsUrlObject, ok := m[codegen.XExtendsUrls]
	if !ok {
		return nil, nil
	}

	xExtendsUrls := make([]string, 0)
	{
		v, err := json.Marshal(xExtendsUrlObject)
		if err != nil {
			return nil, error.New(err)
		}
		err = json.Unmarshal(v, &xExtendsUrls)
		if err != nil {
			return nil, error.New(err)
		}
	}
	res = make([]*openapi3.T, len(xExtendsUrls))
	for _, xExtendsUrl := range xExtendsUrls {
		v, err := loadSwagger(xExtendsUrl)
		if err != nil {
			return nil, error.New(err)
		}
		res = append(res, v)
	}

	return res, nil

}
func calculatePackageNameByGomod() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", error.New(err)
	}
	RecursiveDir := pwd

	dirList := []string{RecursiveDir}
	for i := 0; i < 10; i++ { // support most 10 layers
		RecursiveDir = getTopDir(RecursiveDir)
		if RecursiveDir == "" {
			break
		}
		dirList = append(dirList, RecursiveDir)
	}

	gomodPath := ""
	for _, v := range dirList {

		if _, err := os.Stat(v + "/go.mod"); os.IsNotExist(err) {
			panic(err)
		}

		if exist {
			gomodBody, err := os.ReadFile(v + "/go.mod")
			if err != nil {
				panic(err)
			}
			gomodFile, err := modfile.Parse("go.mod", gomodBody, nil)
			if err != nil {
				panic(err)
			}
			gomodPath = gomodFile.Module.Mod.Path
			break
		}
	}
	ss := strings.Split(pwd, gomodPath)
	if len(ss) < 2 {
		return "", fmt.Errorf("path err ,pwd: %v, gomodPath :%v", pwd, gomodPath)
	}
	return gomodPath + ss[1], nil
}

func getTopDir(dir string) string {
	top := filepath.Dir(dir)
	if top == dir {
		return ""
	}
	return top
}
