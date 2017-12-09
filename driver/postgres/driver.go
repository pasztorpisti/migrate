package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pasztorpisti/migrate"
	"net/url"
)

func init() {
	migrate.RegisterDriver("postgres", driverFactory{})
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
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening postgres connection: %s", err)
	}
	return migrate.WrapDB(db), nil
}

func (o *driver) NewMigrationDB() (migrate.MigrationDB, error) {
	return newMigrationDB(o.TableName)
}
