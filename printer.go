package migrate

import (
	"fmt"
	"io"
)

type Printer interface {
	Print(a ...interface{})
	Println(a ...interface{})
	Printf(format string, a ...interface{})
}

func NewPrinter(w io.Writer) Printer {
	return &printer{w: w}
}

type printer struct {
	w io.Writer
}

func (o *printer) Print(a ...interface{}) {
	fmt.Fprint(o.w, a...)
}

func (o *printer) Println(a ...interface{}) {
	fmt.Fprintln(o.w, a...)
}

func (o *printer) Printf(format string, a ...interface{}) {
	fmt.Fprintf(o.w, format, a...)
}

type nullPrinter struct{}

func (nullPrinter) Print(a ...interface{})                 {}
func (nullPrinter) Println(a ...interface{})               {}
func (nullPrinter) Printf(format string, a ...interface{}) {}
