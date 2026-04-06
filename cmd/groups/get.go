package groups

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'groups get' command.
// Endpoint: GET /get_group/{group_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [group_id]",
		Short: "Получить группу по ID",
		Long: `Получает детальную информацию о группе пользователей по её ID.

Включает название группы и полный список пользователей,
входящих в эту группу с их ролями и контактной информацией.`,
		Example: `  # Получить информацию о группе
  gotr groups get 1

  # Сохранить в файл
  gotr groups get 5 -o group.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var groupID int64
			var err error
			if len(args) > 0 {
				groupID, err = flags.ValidateRequiredID(args, 0, "group_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("group_id required: gotr groups get [group_id]")
				}
				groupID, err = resolveGroupIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetGroup(ctx, groupID)
			if err != nil {
				return fmt.Errorf("failed to get group: %w", err)
			}

			return output.OutputResult(cmd, resp, "groups")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
