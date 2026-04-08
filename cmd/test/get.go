package test

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// newGetCmd creates the command for retrieving test information.
func newGetCmd(getClient func(cmd *cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [test-id]",
		Short: "Get test information",
		Long: `Retrieves detailed information about a test by its ID.

Examples:
	# Get a test by ID
	gotr test get 12345

	# Get and save to file
	gotr test get 12345 -o test.json
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := getClient(cmd)
			ctx := cmd.Context()
			if httpClient == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := service.NewTestService(httpClient)

			var testID int64
			var err error

			if len(args) > 0 {
				testID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid test ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr test get [test_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr test get [test_id]")
				}
				testID, err = resolveTestIDInteractive(ctx, httpClient)
				if err != nil {
					return err
				}
			}

			test, err := svc.Get(ctx, testID)
			if err != nil {
				return fmt.Errorf("failed to get test: %w", err)
			}

			// Check if output should be saved to file
			saveFlag, _ := cmd.Flags().GetBool("save")
			if saveFlag {
				filepath, err := output.Output(cmd, test, "test", "json")
				if err != nil {
					return fmt.Errorf("save error: %w", err)
				}
				if filepath != "" {
					output.PrintSuccess(cmd, "Test saved to %s", filepath)
				}
				return nil
			}

			return output.OutputResultWithFlags(cmd, test)
		},
	}

	output.AddFlag(cmd)

	return cmd
}
