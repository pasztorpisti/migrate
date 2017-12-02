package mysql

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/pasztorpisti/migrate"
)

func init() {
	migrate.RegisterDriver("mysql", driver{})
}

type driver struct{}

func (driver) Open(dataSourceName string) (migrate.DB, error) {
	cfg, err := mysql.ParseDSN(dataSourceName)
	if err != nil {
		return nil, err
	}
	cfg.MultiStatements = true
	cfg.ParseTime = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("error opening mysql connection: %s", err)
	}
	return migrate.WrapDB(db), nil
}

func (driver) NewMigrationDB(tableName string) (migrate.MigrationDB, error) {
	return newMigrationDB(tableName)
}
