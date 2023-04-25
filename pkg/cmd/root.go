package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "yap [command]",
		Short: "One provisioner to rule them all",
		Example: "  yap get clusters\n" +
			"  yap apply -f my-cluster.yaml",
	}

	rootCmd.AddCommand(NewCreateOptions().Command())
	rootCmd.AddCommand(NewGetOptions().Command())
	rootCmd.AddCommand(NewApplyOptions().Command())
	rootCmd.AddCommand(NewDeleteOptions().Command())

	return rootCmd
}
