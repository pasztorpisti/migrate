package mysql

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/pasztorpisti/migrate"
	"net/url"
)

func init() {
	migrate.RegisterDriver("mysql", driverFactory{})
}

type driverFactory struct{}

func (o driverFactory) NewDriver(params string) (migrate.Driver, error) {
	values, err := url.ParseQuery(params)
	if err != nil {
		return nil, fmt.Errorf("can't parse driver parameters %q: %s", params, err)
	}

	tableName := values.Get("migrations_table")
	if tableName == "" {
		tableName = "migrations"
	}

	return &driver{
		TableName: tableName,
	}, nil
}

type driver struct {
	TableName string
}

func (*driver) Open(dataSourceName string) (migrate.DB, error) {
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

func (o *driver) NewMigrationDB() (migrate.MigrationDB, error) {
	return newMigrationDB(o.TableName)
}
