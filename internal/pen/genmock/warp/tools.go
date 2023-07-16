package warp

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/smallnest/gen/dbmeta"
)

func InitializeDB(sqlType string, sqlConnStr string) (db *sql.DB, err error) {
	db, err = sql.Open(sqlType, sqlConnStr)
	if err != nil {
		return nil, fmt.Errorf("Error in open database: %v\n\n", err.Error())
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Error pinging database: %v\n\n", err.Error())
	}

	return
}

func MakeLoadTemplateFunc(opt *Option) func(filename string) (tpl *dbmeta.GenTemplate, err error) {

	// LoadTemplate return template from template dir, falling back to the embedded templates
	return func(filename string) (tpl *dbmeta.GenTemplate, err error) {
		baseName := filepath.Base(filename)
		// fmt.Printf("LoadTemplate: %s / %s\n", filename, baseName)

		if opt.TemplateDir != "" {
			fpath := filepath.Join(opt.TemplateDir, filename)
			var b []byte
			b, err = ioutil.ReadFile(fpath)
			if err == nil {

				absPath, err := filepath.Abs(fpath)
				if err != nil {
					absPath = fpath
				}
				// fmt.Printf("Loaded template from file: %s\n", fpath)
				tpl = &dbmeta.GenTemplate{Name: "file://" + absPath, Content: string(b)}
				return tpl, nil
			}
		}

		content, err := baseTemplates.FindString(baseName)
		if err != nil {
			return nil, fmt.Errorf("%s not found internally", baseName)
		}
		if opt.Verbose {
			fmt.Printf("Loaded template from app: %s\n", filename)
		}

		tpl = &dbmeta.GenTemplate{Name: "internal://" + filename, Content: content}
		return tpl, nil
	}
}

func (opt *Option) InitConfig(conf *dbmeta.Config) {
	opt.Outdir = "."

	// load fragments if specified
	if opt.FragmentsDir != "" {
		_ = conf.LoadFragments(opt.FragmentsDir)
	}
	// if packageName is not set we need to default it
	if opt.ModelPackageName == "" {
		opt.ModelPackageName = "model"
	}

	if opt.DaoPackageName == "" {
		opt.DaoPackageName = "dao"
	}
	if opt.ApiPackageName == "" {
		opt.ApiPackageName = "api"
	}

	conf.SQLType = opt.SqlType
	conf.SQLConnStr = opt.SqlConnStr
	conf.SQLDatabase = opt.SqlDatabase
	conf.ModelPackageName = opt.ModelPackageName
	conf.ModelFQPN = opt.Module + "/" + opt.ModelPackageName
	conf.AddGormAnnotation = opt.Gorm
	conf.UseGureguTypes = opt.Guregu
	conf.Overwrite = opt.Overwrite
	conf.OutDir = opt.Outdir

	// TemplateDir
	conf.Verbose = opt.Verbose

	conf.Module = opt.Module
	// conf.ModelPackageName = opt.ModelPackageName
	// conf.ModelFQPN = opt.Module + "/" + opt.ModelPackageName

	conf.DaoPackageName = opt.DaoPackageName
	conf.DaoFQPN = opt.Module + "/" + opt.DaoPackageName

	conf.APIPackageName = opt.ApiPackageName
	conf.APIFQPN = opt.Module + "/" + opt.ApiPackageName

	// conf.Swagger.Version = *swaggerVersion
	// conf.Swagger.BasePath = *swaggerBasePath
	conf.Swagger.Title = fmt.Sprintf("Sample CRUD api for %s db", opt.SqlDatabase)
	conf.Swagger.Description = fmt.Sprintf("Sample CRUD api for %s db", opt.SqlDatabase)
	if opt.ServerPort == 80 {
		conf.Swagger.Host = opt.ServerHost
	} else {
		conf.Swagger.Host = fmt.Sprintf("%s:%d", opt.ServerHost, opt.ServerPort)
	}
}

func loadDefaultDBMappings(conf *dbmeta.Config) error {
	var err error
	var content []byte
	content, err = baseTemplates.Find("mapping.json")
	if err != nil {
		return err
	}

	err = dbmeta.ProcessMappings("internal", content, conf.Verbose)
	if err != nil {
		return err
	}
	return nil
}

func (opt *Option) Generate(conf *dbmeta.Config) error {
	var err error
	modelDir := filepath.Join(opt.Outdir, opt.ModelPackageName)

	err = os.MkdirAll(opt.Outdir, 0777)
	if err != nil && !opt.Overwrite {
		return fmt.Errorf("unable to create outDir: %s error: %v\n", opt.Outdir, err)
	}

	err = os.MkdirAll(modelDir, 0777)
	if err != nil && !opt.Overwrite {
		return fmt.Errorf("unable to create modelDir: %s error: %v\n", modelDir, err)
	}

	var ModelTmpl *dbmeta.GenTemplate
	var ModelBaseTmpl *dbmeta.GenTemplate

	if ModelTmpl, err = opt.LoadTemplate("model.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}
	if ModelBaseTmpl, err = opt.LoadTemplate("model_base.go.tmpl"); err != nil {
		fmt.Print(au.Red(fmt.Sprintf("Error loading template %v\n", err)))
		return err
	}

	// generate go files for each table
	for tableName, tableInfo := range tableInfos {

		if len(tableInfo.Fields) == 0 {
			if opt.Verbose {
				fmt.Printf("[%d] Table: %s - No Fields Available\n", tableInfo.Index, tableName)
			}

			continue
		}

		modelInfo := conf.CreateContextForTableFile(tableInfo)
		modelInfo["UseGuregu"] = opt.Guregu

		modelFile := filepath.Join(modelDir, opt.CreateGoSrcFileName(tableName))
		err = conf.WriteTemplate(ModelTmpl, modelInfo, modelFile)
		if err != nil {
			return fmt.Errorf("Error writing file: %v\n", err)
		}
	}

	data := map[string]interface{}{}
	err = conf.WriteTemplate(ModelBaseTmpl, data, filepath.Join(modelDir, "model_base.go"))
	if err != nil {
		return fmt.Errorf("Error writing file: %v\n", err)
	}

	// ...

	return nil
}

// LoadTemplate return template from template dir, falling back to the embedded templates
func (opt *Option) LoadTemplate(filename string) (tpl *dbmeta.GenTemplate, err error) {
	baseName := filepath.Base(filename)
	// fmt.Printf("LoadTemplate: %s / %s\n", filename, baseName)

	if opt.TemplateDir != "" {
		fpath := filepath.Join(opt.TemplateDir, filename)
		var b []byte
		b, err = ioutil.ReadFile(fpath)
		if err == nil {

			absPath, err := filepath.Abs(fpath)
			if err != nil {
				absPath = fpath
			}
			// fmt.Printf("Loaded template from file: %s\n", fpath)
			tpl = &dbmeta.GenTemplate{Name: "file://" + absPath, Content: string(b)}
			return tpl, nil
		}
	}

	content, err := baseTemplates.FindString(baseName)
	if err != nil {
		return nil, fmt.Errorf("%s not found internally", baseName)
	}
	if opt.Verbose {
		fmt.Printf("Loaded template from app: %s\n", filename)
	}

	tpl = &dbmeta.GenTemplate{Name: "internal://" + filename, Content: content}
	return tpl, nil
}

// CreateGoSrcFileName ensures name doesnt clash with go naming conventions like _test.go
func (opt *Option) CreateGoSrcFileName(tableName string) string {
	name := dbmeta.Replace(opt.FileNaming, tableName)
	// name := inflection.Singular(tableName)

	if strings.HasSuffix(name, "_test") {
		name = name[0 : len(name)-5]
		name = name + "_tst"
	}
	return name + ".go"
}
