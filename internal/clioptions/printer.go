package clioptions

import (
	"os"

	"github.com/mia-platform/miactl/internal/printer"
)

type printerOptions struct {
	noWrapLines bool
}
type PrinterOption func(p *printerOptions)

func DisableWrapLines(noWrap bool) PrinterOption {
	return func(p *printerOptions) {
		p.noWrapLines = noWrap
	}
}

func (o *CLIOptions) Printer(options ...PrinterOption) printer.IPrinter {
	opts := &printerOptions{}
	for _, option := range options {
		option(opts)
	}

	return printer.NewTablePrinter(printer.TablePrinterOptions{
		WrapLinesDisabled: opts.noWrapLines,
	}).SetWriter(os.Stdout)
}
