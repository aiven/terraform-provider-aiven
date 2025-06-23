package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "providertool",
	Short: "A helper CLI for the Aiven Terraform Provider development.",
	Long:  `This tool contains various helper scripts that assist in the development and maintenance of the Aiven Terraform Provider.`,

	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(changelogCmd)
	rootCmd.AddCommand(versionCmd)
}
