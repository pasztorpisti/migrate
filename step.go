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
	Printf PrintfFunc
}

type PrintCtx struct {
	Printf       PrintfFunc
	PrintSQL     bool
	PrintMetaSQL bool
}

type PrintfFunc func(format string, args ...interface{})

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
	ctx.Printf("%s\n", strings.TrimSpace(o.Query))
	if len(o.Args) != 0 {
		ctx.Printf("QueryArgs: %v\n", o.Args)
	}
	ctx.Printf("\n")
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
					ctx.Printf("Rollback error: %s\n", err)
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
		ctx.Printf("BEGIN;\n")
	}

	o.Steps.Print(ctx)

	if doPrint {
		ctx.Printf("COMMIT;\n\n")
	}
}

type StepTitle struct {
	Step
	Title string
}

func (o StepTitle) Execute(ctx ExecCtx) error {
	if o.Title != "" {
		ctx.Printf("%s\n", o.Title)
	}
	return o.Step.Execute(ctx)
}

func (o *StepTitle) Print(ctx PrintCtx) {
	if o.Title != "" {
		if ctx.PrintSQL {
			ctx.Printf("-- %s\n", o.Title)
		} else {
			ctx.Printf("%s\n", o.Title)
		}
	}

	o.Step.Print(ctx)
}
