// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду 'attachments delete'
// Эндпоинт: POST /delete_attachment/{attachment_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <attachment_id>",
		Short: "Удалить вложение",
		Long: `Удаляет вложение по его ID.

⚠️ Внимание: удаление необратимо.`,
		Example: `  # Удалить вложение
  gotr attachments delete 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			attachmentID, err := flags.ValidateRequiredID(args, 0, "attachment_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			if err := client.DeleteAttachment(ctx, attachmentID); err != nil {
				return fmt.Errorf("failed to delete attachment: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Attachment %d deleted successfully\n", attachmentID)
			return nil
		},
	}
}
