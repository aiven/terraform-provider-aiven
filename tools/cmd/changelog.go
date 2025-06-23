package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aiven/terraform-provider-aiven/tools/changelog"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: fmt.Sprintf("Manages %s file operations", changelog.FileName),
}

var (
	changelogUpdateCmd = &cobra.Command{
		Use:   "update [version]",
		Short: "Adds a new version entry to the changelog",
		Args:  cobra.ExactArgs(1),
		Example: `  # Add a new entry for a version, using today's date automatically.
  providertool changelog update 4.5.0

  # Add a new entry with a specific release date.
  providertool changelog update 4.5.1 --date 2025-07-01`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			date, _ := cmd.Flags().GetString("date")
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}

			return changelog.Update(changelog.FileName, version, date)
		},
	}

	changelogExtractCmd = &cobra.Command{
		Use:   "extract [version]",
		Short: "Extracts the notes for a specific version",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return changelog.Extract(changelog.FileName, args[0])
		},
	}
)

func init() {
	changelogCmd.AddCommand(changelogUpdateCmd)
	changelogCmd.AddCommand(changelogExtractCmd)

	changelogUpdateCmd.Flags().String("date", "", "The release date in YYYY-MM-DD format (defaults to today)")
}
