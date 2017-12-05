package migrate

import (
	"fmt"
	"strings"
)

type Step interface {
	Execute(ExecCtx) error
	AllowsTransaction() bool
	Print(PrintCtx)
}

type ExecCtx struct {
	DB     DB
	Output Printer
}

type PrintCtx struct {
	Output       Printer
	PrintSQL     bool
	PrintMetaSQL bool
}

type SQLExecStep struct {
	Query         string
	Args          []interface{}
	NoTransaction bool
	IsMeta        bool
}

func (o *SQLExecStep) Execute(ctx ExecCtx) error {
	_, err := ctx.DB.Exec(o.Query, o.Args...)
	return err
}

func (o *SQLExecStep) AllowsTransaction() bool {
	return !o.NoTransaction
}

func (o *SQLExecStep) Print(ctx PrintCtx) {
	if !ctx.PrintSQL || o.Query == "" {
		return
	}
	if o.IsMeta && !ctx.PrintMetaSQL {
		return
	}
	ctx.Output.Println(strings.TrimSpace(o.Query))
	if len(o.Args) != 0 {
		ctx.Output.Println("QueryArgs:", o.Args)
	}
	ctx.Output.Println()
}

type Steps []Step

func (o Steps) Execute(ctx ExecCtx) error {
	for _, step := range o {
		if err := step.Execute(ctx); err != nil {
			// TODO: return error with context
			return err
		}
	}
	return nil
}

func (o Steps) AllowsTransaction() bool {
	for _, step := range o {
		if !step.AllowsTransaction() {
			return false
		}
	}
	return true
}

func (o Steps) Print(ctx PrintCtx) {
	for _, step := range o {
		step.Print(ctx)
	}
}

type TransactionIfAllowed struct {
	Steps
}

func (o TransactionIfAllowed) Execute(ctx ExecCtx) (retErr error) {
	if len(o.Steps) == 0 {
		return nil
	}

	if o.AllowsTransaction() {
		tx, err := ctx.DB.BeginTX()
		if err != nil {
			return err
		}
		ctx.DB = tx
		defer func() {
			if p := recover(); p != nil {
				retErr = fmt.Errorf("%v", p)
			}
			if retErr != nil {
				err := tx.Rollback()
				if err != nil {
					ctx.Output.Println("Rollback error:", err)
				}
				return
			}
			retErr = tx.Commit()
		}()
	}

	return o.Steps.Execute(ctx)
}

func (o TransactionIfAllowed) Print(ctx PrintCtx) {
	if len(o.Steps) == 0 {
		return
	}

	doPrint := ctx.PrintMetaSQL && o.AllowsTransaction()
	if doPrint {
		ctx.Output.Println("BEGIN;")
	}

	o.Steps.Print(ctx)

	if doPrint {
		ctx.Output.Print("COMMIT;\n\n")
	}
}

type StepTitleAndResult struct {
	Step
	Title string
}

func (o StepTitleAndResult) Execute(ctx ExecCtx) error {
	if o.Title != "" {
		ctx.Output.Print(o.Title + " ... ")
	}

	err := o.Step.Execute(ctx)

	if o.Title != "" {
		if err != nil {
			ctx.Output.Println("FAILED")
		} else {
			ctx.Output.Println("OK")
		}
	}
	return err
}

func (o *StepTitleAndResult) Print(ctx PrintCtx) {
	if o.Title != "" {
		if ctx.PrintSQL {
			ctx.Output.Println("-- " + o.Title)
		} else {
			ctx.Output.Println(o.Title)
		}
	}

	o.Step.Print(ctx)
}
