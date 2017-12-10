package migrate

import "fmt"

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

	source, ok := GetMigrationSource("dir")
	if !ok {
		panic("can't get migration source")
	}
	migrations, err := source.MigrationEntries(input.ConfigFile, cfg.MigrationSource)
	if err != nil {
		return fmt.Errorf("error loading migrations from source %q: %s", cfg.MigrationSource, err)
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

	return nil
}
