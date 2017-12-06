package migrate

import "fmt"

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

	source, ok := GetMigrationSource("dir")
	if !ok {
		panic("can't get migration source")
	}
	migrations, err := source.MigrationEntries(input.ConfigFile, cfg.MigrationSource)
	if err != nil {
		return fmt.Errorf("error loading migrations from source %q: %s", cfg.MigrationSource, err)
	}

	_, err = migrations.New(input.Args)
	return err
}
