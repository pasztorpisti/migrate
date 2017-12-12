package migrate

import "fmt"

type CmdInitInput struct {
	Output     Printer
	ConfigFile string
	DB         string
}

func CmdInit(input *CmdInitInput) error {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return err
	}

	driverFactory, ok := GetDriverFactory(cfg.Driver)
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
	step, err := mdb.CreateTable()
	if err != nil {
		return err
	}

	err = step.Execute(ExecCtx{
		DB:     db,
		Output: nullPrinter{},
	})

	switch err {
	case nil:
		input.Output.Println("Init success.")
	case ErrMigrationsTableAlreadyExists:
		input.Output.Println("Already initialised.")
		err = nil
	}

	return err
}
