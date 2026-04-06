// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'attachments delete' command.
// Endpoint: POST /delete_attachment/{attachment_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [attachment_id]",
		Short: "Удалить вложение",
		Long: `Удаляет вложение по его ID.

⚠️ Внимание: удаление необратимо.`,
		Example: `  # Удалить вложение
  gotr attachments delete 12345`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
			ctx := cmd.Context()

			var attachmentID int64
			var err error
			if len(args) > 0 {
				attachmentID, err = flags.ValidateRequiredID(args, 0, "attachment_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("attachment_id required: gotr attachments delete [attachment_id]")
				}

				attachmentID, err = resolveAttachmentIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("attachments delete")
				dr.PrintOperation(
					fmt.Sprintf("Delete attachment %d", attachmentID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/delete_attachment/%d", attachmentID),
					nil,
				)
				return nil
			}

			if err := client.DeleteAttachment(ctx, attachmentID); err != nil {
				return fmt.Errorf("failed to delete attachment: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Attachment %d deleted successfully\n", attachmentID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
