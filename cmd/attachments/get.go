// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'attachments get'
// Эндпоинт: GET /get_attachment/{attachment_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <attachment_id>",
		Short: "Получить информацию о вложении",
		Long: `Получает детальную информацию о вложении по его ID.

Выводит: ID, имя файла, размер, MIME-тип, дату создания и привязку к ресурсам.`,
		Example: `  # Получить информацию о вложении
  gotr attachments get 12345

  # Вывод в JSON
  gotr attachments get 12345 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			attachmentID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || attachmentID <= 0 {
				return fmt.Errorf("invalid attachment_id: %s", args[0])
			}

			client := getClient(cmd)
			resp, err := client.GetAttachment(attachmentID)
			if err != nil {
				return fmt.Errorf("failed to get attachment: %w", err)
			}

			return output.Result(cmd, resp)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	return cmd
}
