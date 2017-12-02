package postgres

import (
	_ "github.com/lib/pq"
	"github.com/pasztorpisti/migrate"
	"database/sql"
	"fmt"
)

func init() {
	migrate.RegisterDriver("postgres", driver{})
}

type driver struct{}

func (driver) Open(dataSourceName string) (migrate.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening postgres connection: %s", err)
	}
	return migrate.WrapDB(db), nil
}

func (driver) NewMigrationDB(tableName string) (migrate.MigrationDB, error) {
	return newMigrationDB(tableName)
}
