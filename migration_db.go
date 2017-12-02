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
