package migrate

import "fmt"

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

func preparePlanForCmd(input *preparePlanInput) (Steps, DB, error) {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return nil, nil, err
	}

	driver, ok := GetDriver(cfg.Driver)
	if !ok {
		return nil, nil, fmt.Errorf("invalid DB driver: %s", cfg.Driver)
	}

	db, err := driver.Open(cfg.DataSource)
	if err != nil {
		return nil, nil, err
	}

	mdb, err := driver.NewMigrationDB(cfg.MigrationsTable)
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

	source, ok := GetMigrationSource("dir")
	if !ok {
		panic("can't get migration source")
	}
	migrations, err := source.MigrationEntries(input.ConfigFile, cfg.MigrationSource)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading migrations from source %q: %s", cfg.MigrationSource, err)
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
