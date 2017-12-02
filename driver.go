package migrate

import "errors"

type Driver interface {
	Open(dataSourceName string) (DB, error)
	NewMigrationDB(tableName string) (MigrationDB, error)
}

var ErrDriverNotFound = errors.New("driver not found")

func GetDriver(name string) (Driver, error) {
	if d, ok := registry[name]; ok {
		return d, nil
	}
	return nil, ErrDriverNotFound
}

func RegisterDriver(name string, d Driver) {
	registry[name] = d
}

var registry = map[string]Driver{}
