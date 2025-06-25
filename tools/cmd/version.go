package cmd

import (
	"github.com/spf13/cobra"

	"github.com/aiven/terraform-provider-aiven/tools/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manages versioning and git tag validation",
}

var verifyVersionCmd = &cobra.Command{
	Use:   "verify [version]",
	Short: "Verifies a new version against git tags",
	Long:  "Verifies that the new version is a sequential semantic version bump (patch, minor, or major) from the latest git tag.",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		v := args[0]
		gitRunner := version.RealGitRunner{}
		return version.Verify(gitRunner, v)
	},
}

func init() {
	versionCmd.AddCommand(verifyVersionCmd)
}
