package warp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/jimsmart/schema"
	"github.com/logrusorgru/aurora"
	"github.com/smallnest/gen/dbmeta"
)

var (
	baseTemplates *packr.Box
	tableInfos    map[string]*dbmeta.ModelInfo
	au            aurora.Aurora
)

type Option struct {
	SqlType          string   // --sqltype mysql
	SqlConnStr       string   // --connstr xxx
	SqlDatabase      string   // -d bjm
	SqlTable         string   // -t user_user,user_trt
	_SqlTable        []string // _SqlTable= strings.Split(SqlTable,",")
	ModelPackageName string   // --model=dbmodel
	Gorm             bool     // --gorm
	Guregu           bool     // --guregu
	Overwrite        bool     // --overwrite
	Outdir           string   // --out ./

	Module      string // --module
	TemplateDir string // --templateDir
	Verbose     bool   // --verbose

	// auto
	ServerScheme   string // --scheme
	FragmentsDir   string // --fragmentsDir
	DaoPackageName string // --dao
	ApiPackageName string // --api
	ServerHost     string // --host
	ServerPort     int    // --port
	FileNaming     string // --file_naming
}

func NewOption() *Option {
	return &Option{
		ServerScheme:   "http",
		FragmentsDir:   "",
		DaoPackageName: "dao",
		ApiPackageName: "api",
		ServerHost:     "localhost",
		ServerPort:     8080,
		FileNaming:     "{{.}}",
	}
}

func Init() {
	au = aurora.NewAurora(true)
	dbmeta.InitColorOutput(au)
	baseTemplates = packr.New("gen", "./../template")
}

func DoGenDBmodel(opt *Option) (err error) {
	Init()
	defer func() {
		if err != nil {
			fmt.Print(au.Red(err.Error()))
		}
	}()

	if opt.SqlConnStr == "" || opt.SqlConnStr == "nil" {
		return errors.New("sql connection string is required! Add it with --connstr=s\n\n")
	}

	if opt.SqlDatabase == "" || opt.SqlDatabase == "nil" {
		return errors.New("Database can not be null\n\n")
	}

	db, err := InitializeDB(opt.SqlType, opt.SqlConnStr)
	if err != nil {
		return err
	}
	defer db.Close()

	if opt.SqlTable != "" {
		opt._SqlTable = strings.Split(opt.SqlTable, ",")
	} else {
		schemaTables, err := schema.TableNames(db)
		if err != nil {
			return fmt.Errorf("Error in fetching tables information from %s information schema from %s\n", opt.SqlType, opt.SqlConnStr)
		}
		for _, st := range schemaTables {
			opt._SqlTable = append(opt._SqlTable, st[1]) // s[0] == sqlDatabase
		}
	}

	conf := dbmeta.NewConfig(MakeLoadTemplateFunc(opt))
	opt.InitConfig(conf)

	err = loadDefaultDBMappings(conf)
	if err != nil {
		return fmt.Errorf("Error processing default mapping file error: %v\n", err)
	}

	tableInfos = dbmeta.LoadTableInfo(db, opt._SqlTable, nil, conf)
	if len(tableInfos) == 0 {
		return fmt.Errorf("No tables loaded\n")
	}

	fmt.Printf("Generating code for the following tables (%d)\n", len(tableInfos))
	i := 0
	for tableName := range tableInfos {
		fmt.Printf("[%d] %s\n", i, tableName)
		i++
	}

	conf.TableInfos = tableInfos
	conf.ContextMap["tableInfos"] = tableInfos

	err = opt.Generate(conf)
	if err != nil {
		return fmt.Errorf("Error in executing generate %v\n", err)
	}

	return nil
}
