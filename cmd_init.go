package migrate

type CmdInitInput struct {
	ConfigFile string
	DB         string
}

func CmdInit(input *CmdInitInput) error {
	cfg, err := loadAndValidateDBConfig(input.ConfigFile, input.DB)
	if err != nil {
		return err
	}

	driver, err := GetDriver(cfg.Driver)
	if err != nil {
		return err
	}

	db, err := driver.Open(cfg.DataSource)
	if err != nil {
		return err
	}

	mdb, err := driver.NewMigrationDB(cfg.MigrationsTable)
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
