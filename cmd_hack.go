package migrate

import (
	"errors"
	"fmt"
)

type CmdHackInput struct {
	Output      Printer
	ConfigFile  string
	DB          string
	Forward     bool
	MigrationID string

	Force    bool
	UserOnly bool
	MetaOnly bool
}

func CmdHack(input *CmdHackInput) error {
	if input.UserOnly && input.MetaOnly {
		return errors.New("the UserOnly and MetaOnly parameters are exclusive")
	}

	if input.MigrationID == Initial || input.MigrationID == Latest {
		return fmt.Errorf("hack doesn't accept %q or %q as the migration ID", Initial, Latest)
	}

	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return err
	}

	driver, ok := GetDriver(cfg.Driver)
	if !ok {
		return fmt.Errorf("invalid DB driver: %s", cfg.Driver)
	}

	db, err := driver.Open(cfg.DataSource)
	if err != nil {
		return err
	}

	mdb, err := driver.NewMigrationDB(cfg.MigrationsTable)
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
	forwardMap := make(map[string]struct{}, len(forwardMigrations))
	for _, m := range forwardMigrations {
		forwardMap[m.Name] = struct{}{}
	}

	// Preparing for the worst: when some of the IDs exist only in
	// the migrations table but not in migration files and vice versa.

	index, hasMigration := migrations.IndexForName(input.MigrationID)

	name := input.MigrationID
	if hasMigration {
		name = migrations.Name(index)
	}
	_, hasForwardID := forwardMap[name]

	if !hasMigration && !hasForwardID {
		return errors.New("invalid migration ID")
	}

	// Dealing with user tables.

	var userStep Step
	if !input.MetaOnly {
		if !input.Force {
			if input.Forward == hasForwardID {
				input.Output.Println("Nothing to do according to the migrations table.")
				input.Output.Println("Use -force if you want to ignore the migrations table.")
				return nil
			}
		}
		if !hasMigration {
			return fmt.Errorf("there is no migrations file for %q", name)
		}
		forward, backward, err := migrations.Steps(index)
		if err != nil {
			return fmt.Errorf("error loading migration file %q: %s", name, err)
		}

		if input.Forward {
			userStep = forward
		} else {
			userStep = backward
			if userStep == nil {
				return fmt.Errorf("migration %q doesn't have backward step", name)
			}
		}
	}

	// Dealing with the migrations table.

	var metaStep Step
	if !input.UserOnly {
		if !input.Force {
			if input.Forward == hasForwardID {
				input.Output.Println("Nothing to do according to the migrations table.")
				input.Output.Println("Use -force if you want to ignore the migrations table.")
				return nil
			}
		}

		if input.Forward {
			metaStep, err = mdb.ForwardMigrate(name)
		} else {
			metaStep, err = mdb.BackwardMigrate(name)
		}
		if err != nil {
			return err
		}
	}

	var step Step
	switch {
	case userStep == nil:
		step = metaStep
	case metaStep == nil:
		step = userStep
	default:
		step = TransactionIfAllowed{Steps{
			userStep,
			metaStep,
		}}
	}

	return step.Execute(ExecCtx{
		DB:     db,
		Output: input.Output,
	})
}
