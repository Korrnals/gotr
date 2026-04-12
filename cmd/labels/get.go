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

// newGetCmd creates the 'labels get' command.
// Endpoint: GET /get_label/{label_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [label_id]",
		Short: "Get label information",
		Long:  `Retrieves label information by its ID.`,
		Example: `  # Get a label
  gotr labels get 123

  # Output as JSON
  gotr labels get 123 -o json`,
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
					return fmt.Errorf("label_id is required in non-interactive mode: gotr labels get [label_id]")
				}
				if _, ok := interactive.PrompterFromContext(cmd.Context()).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("label_id is required in non-interactive mode: gotr labels get [label_id]")
				}
				if labelID, err = resolveLabelIDInteractive(cmd.Context(), getClient(cmd)); err != nil {
					return err
				}
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Loading label",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Label, error) {
				return client.GetLabel(ctx, labelID)
			})
			if err != nil {
				return fmt.Errorf("failed to get label: %w", err)
			}

			_, err = output.Output(cmd, resp, "labels", "json")
			return err
		},
	}
	output.AddFlag(cmd)
	return cmd
}
