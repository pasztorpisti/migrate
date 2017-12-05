package migrate

import "errors"

type Driver interface {
	Open(dataSourceName string) (DB, error)
	NewMigrationDB(tableName string) (MigrationDB, error)
}

var ErrDriverNotFound = errors.New("driver not found")

func GetDriver(name string) (Driver, error) {
	return driverRegistry.GetDriver(name)
}

func RegisterDriver(name string, d Driver) {
	driverRegistry.RegisterDriver(name, d)
}

var driverRegistry = make(driverMap)

type driverMap map[string]Driver

func (o driverMap) GetDriver(name string) (Driver, error) {
	if d, ok := o[name]; ok {
		return d, nil
	}
	return nil, ErrDriverNotFound
}

func (o driverMap) RegisterDriver(name string, d Driver) {
	_, ok := o[name]
	if ok {
		panic("duplicate database driver name: " + name)
	}
	if d == nil {
		panic("driver is nil")
	}
	o[name] = d
}
