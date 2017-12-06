package migrate

import "time"

type MigrationNameAndTime struct {
	Name string
	Time time.Time
}

type MigrationDB interface {
	GetForwardMigrations(Querier) ([]*MigrationNameAndTime, error)
	CreateTableIfNotExists() (Step, error)
	ForwardMigrate(migrationName string) (Step, error)
	BackwardMigrate(migrationName string) (Step, error)
}

type Driver interface {
	Open(dataSourceName string) (DB, error)
	NewMigrationDB(tableName string) (MigrationDB, error)
}

var GetDriver = driverRegistry.GetDriver
var RegisterDriver = driverRegistry.RegisterDriver

var driverRegistry = make(driverMap)

type driverMap map[string]Driver

func (o driverMap) GetDriver(name string) (d Driver, ok bool) {
	d, ok = o[name]
	return
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
