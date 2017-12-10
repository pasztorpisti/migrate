package migrate

import "time"

type DriverFactory interface {
	NewDriver(params string) (Driver, error)
}

type Driver interface {
	Open(dataSourceName string) (DB, error)
	NewMigrationDB() (MigrationDB, error)
}

type MigrationDB interface {
	GetForwardMigrations(Querier) ([]*MigrationNameAndTime, error)
	CreateTableIfNotExists() (Step, error)
	ForwardMigrate(migrationName string) (Step, error)
	BackwardMigrate(migrationName string) (Step, error)
}

type MigrationNameAndTime struct {
	Name string
	Time time.Time
}

var GetDriver = driverRegistry.GetDriverFactory
var RegisterDriver = driverRegistry.RegisterDriverFactory

var driverRegistry = make(driverFactoryMap)

type driverFactoryMap map[string]DriverFactory

func (o driverFactoryMap) GetDriverFactory(name string) (d DriverFactory, ok bool) {
	d, ok = o[name]
	return
}

func (o driverFactoryMap) RegisterDriverFactory(name string, d DriverFactory) {
	_, ok := o[name]
	if ok {
		panic("duplicate database driver name: " + name)
	}
	if d == nil {
		panic("driver is nil")
	}
	o[name] = d
}
