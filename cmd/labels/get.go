// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
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
			labelID, err := flags.ValidateRequiredID(args, 0, "label_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			resp, err := client.GetLabel(ctx, labelID)
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
