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
	Driver       string
	DriverParams map[string]string
	DataSource   string

	MigrationSourceType   string
	MigrationSourceParams map[string]string
}

func (o *dbConfig) Validate() error {
	dsn, err := performSubstitution(o.DataSource)
	if err != nil {
		return fmt.Errorf("error substituting template parameters to db.data_source %q: %s", o.DataSource, err)
	}
	o.DataSource = dsn
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

	type section struct {
		DB              map[string]string `yaml:"db"`
		MigrationSource map[string]string `yaml:"migration_source"`
	}
	var cfg map[string]*section
	err = yaml.UnmarshalStrict(b, &cfg)
	if err != nil {
		return nil, err
	}

	section2DBConfig := func(s *section) (*dbConfig, error) {
		driver, ok := s.DB["driver"]
		if !ok {
			return nil, errors.New("missing db.driver field")
		}
		delete(s.DB, "driver")

		dsn, ok := s.DB["data_source"]
		if !ok {
			return nil, errors.New("missing db.data_source field")
		}
		delete(s.DB, "data_source")

		mst, ok := s.MigrationSource["type"]
		if !ok {
			mst = "dir"
		}
		delete(s.MigrationSource, "type")

		return &dbConfig{
			Driver:                driver,
			DriverParams:          s.DB,
			DataSource:            dsn,
			MigrationSourceType:   mst,
			MigrationSourceParams: s.MigrationSource,
		}, nil
	}

	res := make(map[string]*dbConfig, len(cfg))
	for name, sect := range cfg {
		dbc, err := section2DBConfig(sect)
		if err != nil {
			return nil, fmt.Errorf("error loading section %q from config: %s", name, err)
		}
		res[name] = dbc
	}
	return res, err
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
