package core

import (
	"errors"
	"time"
)

type DriverFactory interface {
	NewDriver(params map[string]string) (Driver, error)
}

type Driver interface {
	Open(dataSourceName string) (ClosableDB, error)
	NewMigrationDB() (MigrationDB, error)
}

// ErrMigrationsTableAlreadyExists can be returned by the Step returned by
// MigrationDB.CreateTable. Detecting this condition in the MigrationDB
// implementation is optional. It is valid to return nil (no error) when
// the table already exists if implementing the check isn't possible.
var ErrMigrationsTableAlreadyExists = errors.New("the migrations table already exists")

type MigrationDB interface {
	GetForwardMigrations(Querier) ([]*MigrationNameAndTime, error)
	CreateTable() (Step, error)
	ForwardMigrate(migrationName string) (Step, error)
	BackwardMigrate(migrationName string) (Step, error)
}

type MigrationNameAndTime struct {
	Name string
	Time time.Time
}

func GetDriverFactory(name string) (d DriverFactory, ok bool) {
	return driverRegistry.GetDriverFactory(name)
}

func RegisterDriverFactory(name string, d DriverFactory) {
	driverRegistry.RegisterDriverFactory(name, d)
}

func SupportedDrivers() []string {
	return driverRegistry.SupportedDrivers()
}

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

func (o driverFactoryMap) SupportedDrivers() []string {
	a := make([]string, 0, len(o))
	for k := range o {
		a = append(a, k)
	}
	return a
}
