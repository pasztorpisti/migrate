package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pasztorpisti/migrate"
)

func init() {
	migrate.RegisterDriverFactory("postgres", driverFactory{})
}

type driverFactory struct{}

func (o driverFactory) NewDriver(params map[string]string) (migrate.Driver, error) {
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

func (*driver) Open(dataSourceName string) (migrate.ClosableDB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening postgres connection: %s", err)
	}
	return migrate.WrapDB(db), nil
}

func (o *driver) NewMigrationDB() (migrate.MigrationDB, error) {
	return newMigrationDB(o.TableName)
}
