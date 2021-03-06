package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pasztorpisti/migrate/core"
)

func init() {
	core.RegisterDriverFactory("postgres", driverFactory{})
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
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening postgres connection: %s", err)
	}
	return core.WrapDB(db), nil
}

func (o *driver) NewMigrationDB() (core.MigrationDB, error) {
	return newMigrationDB(o.TableName)
}
