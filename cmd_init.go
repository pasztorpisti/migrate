package migrate

type CmdInitInput struct {
	ConfigFile string
	Env        string
}

func CmdInit(input *CmdInitInput) error {
	env, err := loadAndValidateEnv(input.ConfigFile, input.Env)
	if err != nil {
		return err
	}

	driver, err := GetDriver(env.Driver)
	if err != nil {
		return err
	}

	db, err := driver.Open(env.DataSource)
	if err != nil {
		return err
	}

	mdb, err := driver.NewMigrationDB(env.MigrationsTable)
	if err != nil {
		return err
	}
	step, err := mdb.CreateTableIfNotExists()
	if err != nil {
		return err
	}

	return step.Execute(ExecCtx{
		DB:     db,
		Printf: func(format string, args ...interface{}) {},
	})
}
