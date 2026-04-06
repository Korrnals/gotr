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
		Short: "Получить список полей результатов",
		Long: `Получает список доступных полей для результатов тестов.

Эта команда полезна для понимания структуры данных результатов
и доступных полей для заполнения.`,
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

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// saveToFile writes data to a file as formatted JSON.
func saveToFile(data interface{}, filename string) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON serialization error: %w", err)
	}
	return os.WriteFile(filename, jsonBytes, 0644)
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
var fieldsCmd = newFieldsCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
