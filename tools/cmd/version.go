package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aiven/terraform-provider-aiven/tools/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manages versioning and git tag validation",
}

var verifyVersionCmd = &cobra.Command{
	Use:   "verify [version]",
	Short: "Verifies a new version against git tags and prints it as x.y.z",
	Long: "Verifies that the new version is a sequential semantic version bump (patch, minor, or major) from the latest git tag, " +
		"and prints it without the `v` prefix. Without a version, uses the next minor of the latest git tag.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var input string
		if len(args) > 0 {
			input = args[0]
		}

		v, err := version.Verify(version.RealGitRunner{}, input)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), v)

		return err
	},
}

func init() {
	versionCmd.AddCommand(verifyVersionCmd)
}
