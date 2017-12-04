package migrate

import "fmt"

type CmdPlanInput struct {
	Printf       PrintfFunc
	ConfigFile   string
	DB           string
	MigrationID  string
	PrintSQL     bool
	PrintMetaSQL bool
}

func CmdPlan(input *CmdPlanInput) error {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return err
	}

	driver, err := GetDriver(cfg.Driver)
	if err != nil {
		return err
	}

	db, err := driver.Open(cfg.DataSource)
	if err != nil {
		return err
	}

	mdb, err := driver.NewMigrationDB(cfg.MigrationsTable)
	if err != nil {
		return err
	}
	forwardMigrations, err := mdb.GetForwardMigrations(db)
	if err != nil {
		return err
	}
	forwardNames := make([]string, len(forwardMigrations))
	for i, m := range forwardMigrations {
		forwardNames[i] = m.Name
	}

	entries, err := LoadMigrationsDir(cfg.MigrationSource)
	if err != nil {
		return fmt.Errorf("error loading migrations dir %q: %s", cfg.MigrationSource, err)
	}
	migrations, err := LoadMigrationFiles(entries)
	if err != nil {
		return err
	}

	steps, err := Plan(&PlanInput{
		Migrations:           migrations,
		ForwardMigratedNames: forwardNames,
		Target:               input.MigrationID,
		MigrationDB:          mdb,
	})
	if err != nil {
		return err
	}

	if len(steps) == 0 {
		input.Printf("Nothing to migrate.\n")
		return nil
	}

	steps.Print(PrintCtx{
		Printf:       input.Printf,
		PrintSQL:     input.PrintSQL || input.PrintMetaSQL,
		PrintMetaSQL: input.PrintMetaSQL,
	})
	return nil
}
