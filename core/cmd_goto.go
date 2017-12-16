package core

import (
	"errors"
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

	driverFactory, ok := GetDriverFactory(cfg.Driver)
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

	forwardMigrated := make([]bool, migrations.NumMigrations())
	for _, name := range forwardNames {
		index, ok := migrations.IndexForName(name)
		// We don't accept aliases as forward migrated names.
		// This is why we check for (name != input.Migrations.Name(index)).
		if !ok || name != migrations.Name(index) {
			return nil, nil, fmt.Errorf("can't find migration file for forward migrated item %q", name)
		}
		forwardMigrated[index] = true
	}

	if !cfg.AllowMigrationGaps {
		allowForwardMigrated := true
		for _, fm := range forwardMigrated {
			if fm {
				if !allowForwardMigrated {
					return nil, nil, errMigrationGap
				}
			} else {
				allowForwardMigrated = false
			}
		}
	}

	steps, err := Plan(&PlanInput{
		Migrations:      migrations,
		ForwardMigrated: forwardMigrated,
		Target:          input.MigrationID,
		MigrationDB:     mdb,
	})
	if err != nil {
		return nil, nil, err
	}

	if len(steps) == 0 {
		input.Output.Println("Nothing to migrate.")
	}

	return steps, db, nil
}

// errMigrationGap is an ugly error message.
var errMigrationGap = errors.New(`There are gaps between the migrations that have already been applied so the plan
and goto commands don't work because you don't have allow_migration_gaps=true
in your config. You can still use other commands (e.g.: status, hack)
or fix the DB and migrations manually if necessary.`)
