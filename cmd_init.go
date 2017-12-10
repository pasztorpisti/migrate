package migrate

import "fmt"

type CmdInitInput struct {
	ConfigFile string
	DB         string
}

func CmdInit(input *CmdInitInput) error {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return err
	}

	driverFactory, ok := GetDriver(cfg.Driver)
	if !ok {
		return fmt.Errorf("invalid DB driver: %s", cfg.Driver)
	}

	driver, err := driverFactory.NewDriver(cfg.DriverParams)
	if err != nil {
		return fmt.Errorf("error creating %q DB driver: %s", cfg.Driver, err)
	}

	db, err := driver.Open(cfg.DataSource)
	if err != nil {
		return err
	}
	defer db.Close()

	mdb, err := driver.NewMigrationDB()
	if err != nil {
		return err
	}
	step, err := mdb.CreateTableIfNotExists()
	if err != nil {
		return err
	}

	return step.Execute(ExecCtx{
		DB:     db,
		Output: nullPrinter{},
	})
}
