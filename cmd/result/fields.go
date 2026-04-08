package result

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newFieldsCmd creates the 'result fields' command.
func newFieldsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "Get list of result fields",
		Long: `Gets the list of available fields for test results.

This command is useful for understanding the result data structure
and the available fields to fill in.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("result fields")
				dr.PrintOperation(
					"Get Result Fields",
					"GET",
					"/index.php?/api/v2/get_result_fields",
					nil,
				)
				return nil
			}

			fields, err := cli.GetResultFields(ctx)
			if err != nil {
				return fmt.Errorf("failed to get result fields: %w", err)
			}

			return output.OutputResult(cmd, fields, "results")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")

	return cmd
}

// saveToFile writes data to a file as formatted JSON.
func saveToFile(data interface{}, filename string) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON serialization error: %w", err)
	}
	return os.WriteFile(filename, jsonBytes, 0o644)
}

// printJSON prints data as formatted JSON to stdout.
func printJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON serialization error: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

// fieldsCmd is the exported command.
var fieldsCmd = newFieldsCmd(getClientSafe)
