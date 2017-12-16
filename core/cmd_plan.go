package core

type CmdPlanInput struct {
	Output         Printer
	ConfigFile     string
	DB             string
	MigrationID    string
	PrintSQL       bool
	PrintSystemSQL bool
}

func CmdPlan(input *CmdPlanInput) error {
	steps, db, err := preparePlanForCmd(&preparePlanInput{
		Output:      input.Output,
		ConfigFile:  input.ConfigFile,
		DB:          input.DB,
		MigrationID: input.MigrationID,
	})
	if err != nil {
		return err
	}
	db.Close()

	steps.Print(PrintCtx{
		Output:         input.Output,
		PrintSQL:       input.PrintSQL || input.PrintSystemSQL,
		PrintSystemSQL: input.PrintSystemSQL,
	})
	return nil
}
