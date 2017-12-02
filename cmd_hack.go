package migrate

import (
	"errors"
	"fmt"
	"strconv"
)

type CmdHackInput struct {
	Printf      PrintfFunc
	ConfigFile  string
	Env         string
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

	env, err := loadAndValidateEnv(input.ConfigFile, input.Env)
	if err != nil {
		return err
	}

	driver, err := GetDriver(env.Driver)
	if err != nil {
		return err
	}

	db, err := driver.Open(env.DataSource)
	if err != nil {
		return err
	}

	mdb, err := driver.NewMigrationDB(env.MigrationsTable)
	if err != nil {
		return err
	}

	forwardMigrations, err := mdb.GetForwardMigrations(db)
	if err != nil {
		return err
	}

	forwardIDMap := make(map[string]string, len(forwardMigrations))
	for _, m := range forwardMigrations {
		forwardIDMap[m.Name] = m.Name
		var id MigrationID
		if err := id.SetName(m.Name); err == nil {
			numericID := strconv.FormatInt(id.Number, 10)
			forwardIDMap[numericID] = m.Name
		}
	}

	entries, err := LoadMigrationsDir(env.MigrationSource)
	if err != nil {
		return fmt.Errorf("error loading migrations dir %q: %s", env.MigrationSource, err)
	}
	entryMap := make(map[string]*MigrationDirEntry, len(entries))
	for _, e := range entries {
		numericID := strconv.FormatInt(e.MigrationID.Number, 10)
		entryMap[numericID] = e
		entryMap[e.MigrationID.Name] = e
	}

	// Preparing for the worst: when some of the IDs exist only in
	// the migrations table but not in migration files and vice versa.

	entry, hasEntry := entryMap[input.MigrationID]
	forwardName, hasForwardID := forwardIDMap[input.MigrationID]

	if !hasEntry && !hasForwardID {
		return errors.New("invalid migration ID")
	}

	// Dealing with user tables.

	var userStep Step
	if !input.MetaOnly {
		if !input.Force {
			if input.Forward == hasForwardID {
				input.Printf("Nothing to do according to the migrations table.\n")
				input.Printf("Use -force if you want to ignore the migrations table.\n")
				return nil
			}
		}
		if !hasEntry {
			return fmt.Errorf("there are no backward/forward migrations defined for %v", input.MigrationID)
		}
		migration, err := LoadMigrationFile(entry.Filepath)
		if err != nil {
			return fmt.Errorf("error loading migration file %q: %s", entry.Filepath, err)
		}
		if input.Forward {
			userStep = migration.Forward
		} else {
			userStep = migration.Backward
			if userStep == nil {
				return fmt.Errorf("migration %q doesn't have backward migration", migration.Name)
			}
		}
	}

	// Dealing with the migrations table.

	var metaStep Step
	if !input.UserOnly {
		if !input.Force {
			if input.Forward == hasForwardID {
				input.Printf("Nothing to do according to the migrations table.\n")
				input.Printf("Use -force if you want to ignore the migrations table.\n")
				return nil
			}
		}

		var name string
		if hasForwardID {
			name = forwardName
		} else if hasEntry {
			name = entry.MigrationID.Name
		} else {
			panic("this should never happen")
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
		Printf: input.Printf,
	})
}
