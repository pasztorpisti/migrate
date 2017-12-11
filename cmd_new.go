package migrate

import (
	"fmt"
	"path/filepath"
)

type CmdNewInput struct {
	ConfigFile string
	DB         string
	Args       []string
}

func CmdNew(input *CmdNewInput) error {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
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

	_, err = migrations.New(input.Args)
	return err
}
