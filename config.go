package migrate

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type environment struct {
	Driver          string `yaml:"driver"`
	DataSource      string `yaml:"data_source"`
	MigrationsTable string `yaml:"migrations_table"`
	// MigrationSource might support many different kinds of sources in the future.
	// Currently it supports only directory sources with the "dir://" prefix.
	MigrationSource string `yaml:"migration_source"`
}

const dirSourcePrefix = "dir://"

func (o *environment) Validate(configFilename string) error {
	if o.Driver == "" {
		return errors.New("driver must be set")
	}
	if _, err := GetDriver(o.Driver); err != nil {
		return fmt.Errorf("%q is an invalid database driver", o.Driver)
	}

	if o.MigrationSource == "" {
		return errors.New("migration_source must be set")
	}
	if !strings.HasPrefix(o.MigrationSource, dirSourcePrefix) {
		return fmt.Errorf("invalid migration source (currently we support only %s): %s", dirSourcePrefix, o.MigrationSource)
	}
	o.MigrationSource = strings.TrimPrefix(o.MigrationSource, dirSourcePrefix)
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

func loadConfigFile(filename string) (map[string]*environment, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg map[string]*environment
	err = yaml.UnmarshalStrict(b, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, err
}

func loadAndValidateEnv(configFilename, envName string) (*environment, error) {
	cfg, err := loadConfigFile(configFilename)
	if err != nil {
		return nil, fmt.Errorf("error loading config file %q: %s", configFilename, err)
	}

	env, ok := cfg[envName]
	if !ok {
		return nil, fmt.Errorf("env %q isn't defined in config file %q", envName, configFilename)
	}

	return env, env.Validate(configFilename)
}
