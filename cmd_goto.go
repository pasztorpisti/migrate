package migrate

import (
	"fmt"
	"path/filepath"
)

type CmdGotoInput struct {
	Output      Printer
	ConfigFile  string
	DB          string
	MigrationID string
	Quiet       bool
}

func CmdGoto(input *CmdGotoInput) error {
	steps, db, err := preparePlanForCmd(&preparePlanInput{
		Output:      input.Output,
		ConfigFile:  input.ConfigFile,
		DB:          input.DB,
		MigrationID: input.MigrationID,
	})
	if err != nil {
		return err
	}
	defer db.Close()

	execCtx := ExecCtx{
		DB:     db,
		Output: input.Output,
	}
	if input.Quiet {
		execCtx.Output = nullPrinter{}
	}
	return steps.Execute(execCtx)
}

type preparePlanInput struct {
	Output      Printer
	ConfigFile  string
	DB          string
	MigrationID string
}

func preparePlanForCmd(input *preparePlanInput) (_ Steps, _ ClosableDB, retErr error) {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return nil, nil, err
	}

	driverFactory, ok := GetDriver(cfg.Driver)
	if !ok {
		return nil, nil, fmt.Errorf("invalid DB driver: %s", cfg.Driver)
	}

	driver, err := driverFactory.NewDriver(cfg.DriverParams)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating %q DB driver: %s", cfg.Driver, err)
	}

	db, err := driver.Open(cfg.DataSource)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if retErr != nil {
			db.Close()
		}
	}()

	mdb, err := driver.NewMigrationDB()
	if err != nil {
		return nil, nil, err
	}
	forwardMigrations, err := mdb.GetForwardMigrations(db)
	if err != nil {
		return nil, nil, err
	}
	forwardNames := make([]string, len(forwardMigrations))
	for i, m := range forwardMigrations {
		forwardNames[i] = m.Name
	}

	sourceFactory, ok := GetMigrationSourceFactory(cfg.MigrationSourceType)
	if !ok {
		return nil, nil, fmt.Errorf("unknown migration_source type in config: %s", cfg.MigrationSourceType)
	}
	source, err := sourceFactory.NewMigrationSource(filepath.Dir(input.ConfigFile), cfg.MigrationSourceParams)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating migration source: %s", err)
	}
	migrations, err := source.MigrationEntries()
	if err != nil {
		return nil, nil, fmt.Errorf("error loading migrations: %s", err)
	}

	steps, err := Plan(&PlanInput{
		Migrations:           migrations,
		ForwardMigratedNames: forwardNames,
		Target:               input.MigrationID,
		MigrationDB:          mdb,
	})
	if err != nil {
		return nil, nil, err
	}

	if len(steps) == 0 {
		input.Output.Println("Nothing to migrate.")
	}

	return steps, db, nil
}
