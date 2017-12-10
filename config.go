package migrate

import (
	"errors"
	"fmt"
	"github.com/pasztorpisti/migrate/template"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type dbConfig struct {
	Driver          string `yaml:"driver"`
	DriverParams    string `yaml:"driver_params"`
	DataSource      string `yaml:"data_source"`
	MigrationSource string `yaml:"migration_source"`
}

func (o *dbConfig) Validate() error {
	if o.Driver == "" {
		return errors.New("driver must be set")
	}
	if _, ok := GetDriver(o.Driver); !ok {
		return fmt.Errorf("invalid DB driver: %s", o.Driver)
	}

	if o.MigrationSource == "" {
		return errors.New("migration_source must be set")
	}

	ms, err := performSubstitution(o.MigrationSource)
	if err != nil {
		return fmt.Errorf("error substituting template parameters to migration_source %q: %s", o.MigrationSource, err)
	}
	o.MigrationSource = ms

	dsn, err := performSubstitution(o.DataSource)
	if err != nil {
		return fmt.Errorf("error substituting template parameters to data_source %q: %s", o.DataSource, err)
	}
	o.DataSource = dsn

	dp, err := performSubstitution(o.DriverParams)
	if err != nil {
		return fmt.Errorf("error substituting template parameters to driver_params %q: %s", o.DriverParams, err)
	}
	o.DriverParams = dp

	return nil
}

func performSubstitution(s string) (string, error) {
	sections, err := template.Parse(s)
	if err != nil {
		return "", err
	}
	return template.Execute(&template.ExecuteInput{
		Sections:     sections,
		LookupVar:    os.LookupEnv,
		ExecCmd:      template.RemoveTrailingNewlines(template.ExecCmd),
		VarParamName: "env",
		CmdParamName: "cmd",
	})
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

	return dbCfg, dbCfg.Validate()
}
