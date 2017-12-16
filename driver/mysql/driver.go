package mysql

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/pasztorpisti/migrate/core"
)

func init() {
	core.RegisterDriverFactory("mysql", driverFactory{})
}

type driverFactory struct{}

func (o driverFactory) NewDriver(params map[string]string) (core.Driver, error) {
	takeParam := func(key string) (string, bool) {
		val, ok := params[key]
		if ok {
			delete(params, key)
		}
		return val, ok
	}

	tableName, ok := takeParam("migrations_table")
	if !ok || tableName == "" {
		tableName = "migrations"
	}

	if len(params) != 0 {
		return nil, fmt.Errorf("unrecognised driver params: %q", params)
	}

	return &driver{
		TableName: tableName,
	}, nil
}

type driver struct {
	TableName string
}

func (*driver) Open(dataSourceName string) (core.ClosableDB, error) {
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
	return core.WrapDB(db), nil
}

func (o *driver) NewMigrationDB() (core.MigrationDB, error) {
	return newMigrationDB(o.TableName)
}
