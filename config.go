package migrate

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type dbConfig struct {
	Driver          string `yaml:"driver"`
	DataSource      string `yaml:"data_source"`
	MigrationsTable string `yaml:"migrations_table"`
	MigrationSource string `yaml:"migration_source"`
}

func (o *dbConfig) Validate(configFilename string) error {
	if o.Driver == "" {
		return errors.New("driver must be set")
	}
	if _, err := GetDriver(o.Driver); err != nil {
		return fmt.Errorf("%q is an invalid database driver", o.Driver)
	}

	if o.MigrationSource == "" {
		return errors.New("migration_source must be set")
	}
	if !filepath.IsAbs(o.MigrationSource) {
		if !filepath.IsAbs(configFilename) {
			absConfigFilename, err := filepath.Abs(configFilename)
			if err != nil {
				return fmt.Errorf("can't convert config file path (%q) to absolute: %s", configFilename, err)
			}
			configFilename = absConfigFilename
		}
		o.MigrationSource = filepath.Join(filepath.Dir(configFilename), o.MigrationSource)
	}

	if o.MigrationsTable == "" {
		o.MigrationsTable = "migrations"
	}
	return nil
}

func loadConfigFile(filename string) (map[string]*dbConfig, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg map[string]*dbConfig
	err = yaml.UnmarshalStrict(b, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, err
}

func loadAndValidateDBConfig(configFilename, db string) (*dbConfig, error) {
	cfg, err := loadConfigFile(configFilename)
	if err != nil {
		return nil, fmt.Errorf("error loading config file %q: %s", configFilename, err)
	}

	dbCfg, ok := cfg[db]
	if !ok {
		return nil, fmt.Errorf("DB %q isn't defined in config file %q", db, configFilename)
	}

	return dbCfg, dbCfg.Validate(configFilename)
}
