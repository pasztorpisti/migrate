package migrate

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CmdNewInput struct {
	Printf      PrintfFunc
	ConfigFile  string
	Env         string
	Space       string
	NoExt       bool
	Description string
}

func CmdNew(input *CmdNewInput) error {
	env, err := loadAndValidateEnv(input.ConfigFile, input.Env)
	if err != nil {
		return err
	}

	entries, err := LoadMigrationsDir(env.MigrationSource)
	if err != nil {
		return fmt.Errorf("error loading migrations dir %q: %s", env.MigrationSource, err)
	}

	ids := make(map[int64]struct{}, len(entries))
	for _, e := range entries {
		ids[e.MigrationID.Number] = struct{}{}
	}

	id := time.Now().Unix()
	for {
		if _, ok := ids[id]; !ok {
			break
		}
		id++
	}

	filename := strconv.FormatInt(id, 10)
	if input.Description != "" {
		filename += " " + input.Description
	}
	if !input.NoExt {
		filename += ".sql"
	}
	filename = strings.Replace(filename, " ", input.Space, -1)
	path := filepath.Join(env.MigrationSource, filename)

	err = ioutil.WriteFile(path, []byte(migrationTemplate), 0644)
	if err != nil {
		return fmt.Errorf("error writing file %q", path)
	}

	input.Printf("Created %s\n", path)
	return nil
}

const migrationTemplate = `-- +migrate forward

-- TODO: Implement forward migration. (required)

-- +migrate backward

-- TODO: Implement backward migration. (optional)
-- As an alternative you can delete the whole '+migrate backward' directive
-- because implementing backward migration is optinoal.
`
