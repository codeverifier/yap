package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type CreateOptions struct {
	genericclioptions.IOStreams
}

func NewCreateOptions() *CreateOptions {
	o := &CreateOptions{
		IOStreams: genericclioptions.IOStreams{Out: os.Stdout, ErrOut: os.Stderr, In: os.Stdin},
	}
	return o
}

func (o *CreateOptions) Command() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create [cluster]",
		Short: "Create a cluster",
		Example: "  yap create cluster docker-desktop\n" +
			"  yap create cluster kind",
		Run: o.Run,
	}

	cmd.SetOut(o.Out)
	cmd.SetErr(o.ErrOut)
	cmd.AddCommand(NewCreateClusterOptions().Command())

	return cmd
}

func (o *CreateOptions) Run(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
	os.Exit(1)
}
