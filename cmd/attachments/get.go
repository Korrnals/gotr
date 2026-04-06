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

// newGetCmd creates the 'attachments get' command.
// Endpoint: GET /get_attachment/{attachment_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [attachment_id]",
		Short: "Получить информацию о вложении",
		Long: `Получает детальную информацию о вложении по его ID.

Выводит: ID, имя файла, размер, MIME-тип, дату создания и привязку к ресурсам.`,
		Example: `  # Получить информацию о вложении
  gotr attachments get 12345

  # Вывод в JSON
  gotr attachments get 12345 -o json`,
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
					return fmt.Errorf("attachment_id required: gotr attachments get [attachment_id]")
				}

				attachmentID, err = resolveAttachmentIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			resp, err := client.GetAttachment(ctx, attachmentID)
			if err != nil {
				return fmt.Errorf("failed to get attachment: %w", err)
			}

			_, err = output.Output(cmd, resp, "attachments", "json")
			return err
		},
	}
	output.AddFlag(cmd)
	return cmd
}
