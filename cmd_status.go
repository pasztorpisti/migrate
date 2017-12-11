package migrate

import (
	"fmt"
	"path/filepath"
)

type CmdStatusInput struct {
	Output     Printer
	ConfigFile string
	DB         string
}

func CmdStatus(input *CmdStatusInput) error {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return err
	}

	driverFactory, ok := GetDriver(cfg.Driver)
	if !ok {
		return fmt.Errorf("invalid DB driver: %s", cfg.Driver)
	}

	driver, err := driverFactory.NewDriver(cfg.DriverParams)
	if err != nil {
		return fmt.Errorf("error creating %q DB driver: %s", cfg.Driver, err)
	}

	db, err := driver.Open(cfg.DataSource)
	if err != nil {
		return err
	}
	defer db.Close()

	mdb, err := driver.NewMigrationDB()
	if err != nil {
		return err
	}

	sourceFactory, ok := GetMigrationSourceFactory(cfg.MigrationSourceType)
	if !ok {
		return fmt.Errorf("unknown migration_source type in config: %s", cfg.MigrationSourceType)
	}
	source, err := sourceFactory.NewMigrationSource(filepath.Dir(input.ConfigFile), cfg.MigrationSourceParams)
	if err != nil {
		return fmt.Errorf("error creating migration source: %s", err)
	}
	migrations, err := source.MigrationEntries()
	if err != nil {
		return fmt.Errorf("error loading migrations from source %q: %s", cfg.MigrationSourceParams, err)
	}

	forwardMigrations, err := mdb.GetForwardMigrations(db)
	if err != nil {
		return err
	}

	var invalidNames []string
	forwardMap := make(map[string]struct{}, len(forwardMigrations))
	for _, m := range forwardMigrations {
		_, ok := migrations.IndexForName(m.Name)
		if !ok {
			invalidNames = append(invalidNames, m.Name)
		} else {
			forwardMap[m.Name] = struct{}{}
		}
	}

	checkbox := func(checked bool) string {
		if checked {
			return "[X]"
		}
		return "[ ]"
	}

	numMigrations := migrations.NumMigrations()
	for i := 0; i < numMigrations; i++ {
		name := migrations.Name(i)
		_, ok := forwardMap[name]
		input.Output.Printf("%s %s\n", checkbox(ok), name)
	}

	for _, name := range invalidNames {
		input.Output.Printf("!!! Invalid name in migrations table: %s\n", name)
	}

	if numMigrations == 0 && len(invalidNames) == 0 {
		input.Output.Println("There are no migrations.")
	}

	return nil
}
