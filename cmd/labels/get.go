// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'labels get'
// Эндпоинт: GET /get_label/{label_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <label_id>",
		Short: "Получить информацию о метке",
		Long:  `Получает информацию о метке по её ID.`,
		Example: `  # Получить метку
  gotr labels get 123

  # Вывод в JSON
  gotr labels get 123 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			labelID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || labelID <= 0 {
				return fmt.Errorf("invalid label_id: %s", args[0])
			}

			client := getClient(cmd)
			resp, err := client.GetLabel(labelID)
			if err != nil {
				return fmt.Errorf("failed to get label: %w", err)
			}

			_, err = save.Output(cmd, resp, "labels", "json")
			return err
		},
	}
	save.AddFlag(cmd)
	return cmd
}
