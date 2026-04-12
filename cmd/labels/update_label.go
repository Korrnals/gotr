// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateLabelCmd creates the 'labels update-label' command.
// Endpoint: POST /update_label/{label_id}
func newUpdateLabelCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-label [label_id]",
		Short: "Update a label",
		Long: `Updates an existing label.

Requires label editing permissions in the project.
Maximum label name length is 20 characters.`,
		Example: `  # Update label name
  gotr labels update-label 123 --project 1 --title "New Label Name"

  # Output as JSON
  gotr labels update-label 123 --project 1 --title "Bug" -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var labelID int64
			var err error
			if len(args) > 0 {
				labelID, err = flags.ValidateRequiredID(args, 0, "label_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("label_id is required in non-interactive mode: gotr labels update-label [label_id]")
				}
				if _, ok := interactive.PrompterFromContext(cmd.Context()).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("label_id is required in non-interactive mode: gotr labels update-label [label_id]")
				}
				if labelID, err = resolveLabelIDInteractive(cmd.Context(), getClient(cmd)); err != nil {
					return err
				}
			}

			projectID, _ := cmd.Flags().GetInt64("project")
			title, _ := cmd.Flags().GetString("title")

			req := data.UpdateLabelRequest{
				ProjectID: projectID,
				Title:     title,
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("labels update-label")
				dr.PrintOperation(
					fmt.Sprintf("Update label %d", labelID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/update_label/%d", labelID),
					req,
				)
				return nil
			}

			client := getClient(cmd)
			ctx := cmd.Context()

			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Updating label",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Label, error) {
				return client.UpdateLabel(ctx, labelID, req)
			})
			if err != nil {
				return fmt.Errorf("failed to update label: %w", err)
			}

			_, err = output.Output(cmd, resp, "labels", "json")
			return err
		},
	}

	cmd.Flags().Int64P("project", "p", 0, "Project ID (required)")
	cmd.Flags().StringP("title", "t", "", "New label name (required, max 20 characters)")
	cmd.Flags().Bool("dry-run", false, "Show what would be done without updating")
	output.AddFlag(cmd)

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}
