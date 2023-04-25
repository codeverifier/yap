package cmd

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"

	fmtprinters "github.com/pseudonator/yap/pkg/internal/printers"
)

func toPrinter(flags *genericclioptions.PrintFlags) (printers.ResourcePrinter, error) {
	p, err := flags.ToPrinter()
	if err != nil {
		return nil, err
	}
	namePrinter, ok := p.(*printers.NamePrinter)
	if ok {
		return &fmtprinters.NamePrinter{
			ShortOutput: namePrinter.ShortOutput,
			Operation:   namePrinter.Operation,
		}, nil
	}
	return p, nil
}
