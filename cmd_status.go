package migrate

import (
	"fmt"
	"sort"
)

type CmdStatusInput struct {
	Printf     PrintfFunc
	ConfigFile string
	Env        string
}

func CmdStatus(input *CmdStatusInput) error {
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

	var invalidNames []string
	forwardIDMap := make(map[int64]MigrationID, len(forwardMigrations))
	for _, m := range forwardMigrations {
		var id MigrationID
		if err := id.SetName(m.Name); err != nil {
			invalidNames = append(invalidNames, m.Name)
		} else {
			forwardIDMap[id.Number] = id
		}
	}
	sort.Strings(invalidNames)

	forwardIDs := make([]MigrationID, 0, len(forwardIDMap))
	for _, id := range forwardIDMap {
		forwardIDs = append(forwardIDs, id)
	}
	sort.Slice(forwardIDs, func(i, j int) bool {
		return forwardIDs[i].Number < forwardIDs[j].Number
	})

	entries, err := LoadMigrationsDir(env.MigrationSource)
	if err != nil {
		return fmt.Errorf("error loading migrations dir %q: %s", env.MigrationSource, err)
	}
	entryMap := make(map[int64]*MigrationDirEntry, len(entries))
	for _, e := range entries {
		entryMap[e.MigrationID.Number] = e
	}

	checkbox := func(checked bool) string {
		if checked {
			return "[X]"
		}
		return "[ ]"
	}

	for _, e := range entries {
		_, ok := forwardIDMap[e.MigrationID.Number]
		input.Printf("%s %s\n", checkbox(ok), e.MigrationID.Name)
	}

	for _, id := range forwardIDs {
		_, ok := entryMap[id.Number]
		if !ok {
			input.Printf("!!! Exists only in migrations table: %s\n", id.Name)
		}
	}

	for _, name := range invalidNames {
		input.Printf("!!! Invalid name in migrations table: %s\n", name)
	}

	return nil
}
