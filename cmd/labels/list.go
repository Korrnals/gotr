// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'labels list' command.
// Endpoint: GET /get_labels/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "List project labels",
		Long:  `Retrieves a list of all labels for the specified project.`,
		Example: `  # List project labels
  gotr labels list 123

  # Output as JSON
  gotr labels list 123 -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr labels list [project_id]")
				}
				if _, ok := interactive.PrompterFromContext(cmd.Context()).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr labels list [project_id]")
				}
				if projectID, err = resolveProjectIDInteractive(cmd.Context(), getClient(cmd)); err != nil {
					return err
				}
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			quiet, _ := cmd.Flags().GetBool("quiet")
			labels, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Loading labels",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (data.GetLabelsResponse, error) {
				return client.GetLabels(ctx, projectID)
			})
			if err != nil {
				return fmt.Errorf("failed to list labels: %w", err)
			}

			saveFlag, _ := cmd.Flags().GetBool("save")
			if saveFlag {
				_, err := output.Output(cmd, labels, "labels", "json")
				return err
			}

			if len(labels) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No labels found")
				return nil
			}

			t := ui.NewTable(cmd)
			t.AppendHeader(table.Row{"ID", "NAME"})
			for _, l := range labels {
				t.AppendRow(table.Row{l.ID, l.Name})
			}
			ui.Table(cmd, t)
			return nil
		},
	}
	output.AddFlag(cmd)
	return cmd
}

// Verify interface compliance
var _ interface {
	GetLabels(ctx context.Context, projectID int64) (data.GetLabelsResponse, error)
} = (*client.MockClient)(nil)
