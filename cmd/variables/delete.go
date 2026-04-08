package variables

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'variables delete' command.
// Endpoint: POST /delete_variable/{variable_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [variable_id]",
		Short: "Delete a variable",
		Long: `Deletes a variable from a dataset.

⚠️ Warning: deletion cannot be undone! All values of this variable
will be permanently deleted. Make sure the variable is not used
in active test cases.

Use --dry-run to preview before deleting.`,
		Example: `  # Delete a variable
  gotr variables delete 789

  # Preview before deleting
  gotr variables delete 789 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var variableID int64
			if len(args) > 0 {
				var err error
				variableID, err = flags.ValidateRequiredID(args, 0, "variable_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("variable_id is required in non-interactive mode: gotr variables delete [variable_id]")
				}
				var err error
				variableID, err = resolveVariableIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("variables delete")
				dr.PrintSimple("Delete Variable", fmt.Sprintf("Variable ID: %d", variableID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteVariable(ctx, variableID); err != nil {
				return fmt.Errorf("failed to delete variable: %w", err)
			}

			ui.Successf(os.Stdout, "Variable %d deleted", variableID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")

	return cmd
}
