// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"
	"strconv"

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
			attachmentID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || attachmentID <= 0 {
				return fmt.Errorf("invalid attachment_id: %s", args[0])
			}

			client := getClient(cmd)
			if err := client.DeleteAttachment(attachmentID); err != nil {
				return fmt.Errorf("failed to delete attachment: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Attachment %d deleted successfully\n", attachmentID)
			return nil
		},
	}
}
