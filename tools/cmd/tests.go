package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aiven/terraform-provider-aiven/tools/test"
)

var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "CI/CD helper commands",
}

var discoverTestMatrixCmd = &cobra.Command{
	Use:   "matrix",
	Short: "Finds Go tests, partitions them, and generates a JSON matrix.",
	Long: `This command scans the './internal' directory for Go test files (*_test.go),
identifies the unique directories containing them, and partitions them into two groups: 'normal' and 'slow'.
The partitioning is based on a comma-separated list of test names provided via the --slow-tests-csv flag.
Optionally filter tests by service names using --filter flag.
The output is a JSON object suitable for use as a job matrix in GitHub Actions.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		slowTestsCSV, err := cmd.Flags().GetString("slow-tests-csv")
		if err != nil {
			return fmt.Errorf("could not retrieve slow-tests-csv flag: %w", err)
		}

		filterCSV, err := cmd.Flags().GetString("filter")
		if err != nil {
			return fmt.Errorf("could not retrieve filter flag: %w", err)
		}

		// the root directory to scan for tests is hardcoded to "./internal"
		matrix, err := test.GenerateMatrix("./internal", slowTestsCSV, filterCSV)
		if err != nil {
			return fmt.Errorf("failed to generate test matrix: %w", err)
		}

		jsonOutput, err := matrix.ToJSON()
		if err != nil {
			return err
		}

		// print the JSON to stdout
		fmt.Println(jsonOutput)

		return nil
	},
}

func init() {
	testsCmd.AddCommand(discoverTestMatrixCmd)
	discoverTestMatrixCmd.Flags().String("slow-tests-csv", "", "A comma-separated list of test names to be considered slow.")
	discoverTestMatrixCmd.Flags().String("filter", "", "Comma or space-separated list of service names to filter tests (e.g., 'kafka,pg,vpc' or 'kafka pg vpc').")
}
