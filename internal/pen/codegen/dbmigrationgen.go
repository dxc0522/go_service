package codegen

import (
	"database/sql"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose"
	"github.com/sillydong/dbdiffer"
	dbdiff "github.com/sillydong/dbdiffer/mysql"

	"github.tesla.cn/itapp/lines/logx"
)

const versionControlTableName = "goose_db_version"

const (
	schemaDropSqlTmpl   = "DROP SCHEMA %s;"
	tableDropSqlTmpl    = "DROP TABLE %s;"
	schemaCreateSqlTmpl = "CREATE SCHEMA %s DEFAULT CHARACTER SET UTF8 COLLATE UTF8_GENERAL_CI;"
	schemaGrantSqlTmpl  = "GRANT SELECT, INSERT, UPDATE, DELETE, CREATE ON %s.* TO %s;"
	flushPrivilegesSql  = "FLUSH PRIVILEGES;"
)

type GooseTemplate struct {
	GooseUpStatement   string
	GooseDownStatement string
}

func (g Generator) GenerateMigrationFile() error {
	template, err := generateTemplate(g.DataSource, g.TablePrefix)
	if err != nil {
		return err
	}

	if template == nil {
		return nil
	}

	content, err := g.ExecuteTemplate(template)
	if err != nil {
		return err
	}

	target := path.Join(g.TargetDir, fmt.Sprintf(g.TargetFile, time.Now().Format("20060102150304")))
	if err := g.writeFile(target, content); err != nil {
		return err
	}

	return nil
}

func generateTemplate(sourceDsn string, tablePrefix string) (*GooseTemplate, error) {
	// parse source dsn
	sourceConfig, err := mysql.ParseDSN(sourceDsn)
	if sourceConfig == nil || err != nil {
		return nil, err
	}

	// verify source sourceDB connection string
	sourceDB, err := getDBConnection(sourceConfig)
	if err != nil {
		return nil, err
	}

	// create new schema
	targetConfig, err := createNewSchema(*sourceConfig, sourceDB)
	if err != nil {
		if e, ok := err.(*mysql.MySQLError); ok {
			if e.Number == 1410 && strings.Contains(e.Error(), "You are not allowed to create a user with GRANT") {
				logx.Error(`found mysql err : 1410 You are not allowed to create a user with GRANT 
if you are in localhost mysql,you can try followings and then try again

update user set host='%' where user='root';
flush privileges;
GRANT ALL ON *.* TO 'root'@'%';

`)
			}
		}
		return nil, err
	}

	targetDsn := targetConfig.FormatDSN()

	// set dialect
	if err = goose.SetDialect("mysql"); err != nil {
		return nil, err
	}

	targetDB, err := getDBConnection(targetConfig)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = dropSchema(targetConfig.DBName, sourceDB)
		sourceDB.Close()
		targetDB.Close()
	}()

	dirs := []string{
		"./db/migrations",
	}

	for _, dir := range dirs {
		// goose up new schema
		if err := goose.Up(targetDB, dir); err != nil {
			return nil, err
		}

		if err := dropTable(versionControlTableName, targetDB); err != nil {
			return nil, err
		}
	}

	// goose up
	gooseUpSql, err := diff(sourceDsn, targetDsn, tablePrefix)
	if err != nil {
		return nil, err
	}

	// goose down
	gooseDownSql, err := diff(targetDsn, sourceDsn, tablePrefix)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(gooseUpSql, "") && strings.EqualFold(gooseDownSql, "") {
		logx.Info("there are no different between source and target database")
		return nil, nil
	}

	file := GooseTemplate{
		GooseUpStatement:   gooseUpSql,
		GooseDownStatement: gooseDownSql,
	}

	return &file, nil
}

func getDBConnection(config *mysql.Config) (*sql.DB, error) {
	config.MultiStatements = true
	sourceSchema, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}
	if err := sourceSchema.Ping(); err != nil {
		return nil, err
	}
	return sourceSchema, nil
}

func createNewSchema(config mysql.Config, db *sql.DB) (*mysql.Config, error) {
	config.DBName = generateSchemaName()

	if err := createSchema(config.DBName, config.User, db); err != nil {
		return nil, err
	}
	return &config, nil
}

func dropSchema(schemaName string, db *sql.DB) error {
	schemaDropSql := fmt.Sprintf(schemaDropSqlTmpl, schemaName)
	return execSql(schemaDropSql, db)
}

func dropTable(tableName string, db *sql.DB) error {
	tableDropSql := fmt.Sprintf(tableDropSqlTmpl, tableName)
	return execSql(tableDropSql, db)
}

func createSchema(schemaName string, userName string, db *sql.DB) error {
	schemaCreateSql := fmt.Sprintf(schemaCreateSqlTmpl, schemaName)
	schemaGrantSql := fmt.Sprintf(schemaGrantSqlTmpl, schemaName, userName)
	return execSql(schemaCreateSql+schemaGrantSql+flushPrivilegesSql, db)
}

func execSql(sql string, db *sql.DB) error {
	if _, err := db.Exec(sql); err != nil {
		return err
	}
	return nil
}

func generateSchemaName() string {
	schemaPrefix := "MIGRATION_TEMP_SCHEMA_"
	unixInt := time.Now().Unix()
	return schemaPrefix + strconv.Itoa(int(unixInt))
}

func removeGooseVersionControl(result *dbdiffer.Result) {
	for idx, table := range result.Create {
		if strings.HasSuffix(table.Name, versionControlTableName) {
			result.Create = append(result.Create[:idx], result.Create[idx+1:]...)
			break
		}
	}

	for idx, table := range result.Drop {
		if strings.HasSuffix(table.Name, versionControlTableName) {
			result.Drop = append(result.Drop[:idx], result.Drop[idx+1:]...)
			break
		}
	}

}

func diff(sourceDsn string, targetDsn string, tablePrefix string) (sql string, err error) {
	differ, err := dbdiff.New(sourceDsn, targetDsn)
	if differ == nil || err != nil {
		return "", err
	}
	upDiff, err := differ.Diff(tablePrefix)
	if err != nil {
		return "", err
	}

	removeGooseVersionControl(upDiff)

	gooseString, err := differ.Generate(upDiff)
	if err != nil {
		return "", err
	}
	join := strings.Join(gooseString, "\n")
	return join, nil
}
