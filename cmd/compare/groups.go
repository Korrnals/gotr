package compare

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newGroupsCmd creates the 'compare groups' subcommand.
func newGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Сравнить группы между проектами",
		Long: `Выполняет сравнение групп пользователей между двумя проектами.

Примеры:
  # Сравнить группы
  gotr compare groups --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare groups --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare groups --pid1 30 --pid2 31 --save-to groups_diff.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			// Parse flags
			pid1, pid2, format, savePath, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Create progress manager
			pm := progress.NewManager()

			// Start timer
			startTime := time.Now()

			// Compare groups
			result, err := compareGroupsInternal(cli, pid1, pid2, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения групп: %w", err)
			}

			elapsed := time.Since(startTime)

			// Print or save result
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print statistics
			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				PrintCompareStats("groups", pid1, pid2,
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// groupsCmd — экспортированная команда
var groupsCmd = newGroupsCmd()

// compareGroupsInternal compares groups between two projects and returns the result.
func compareGroupsInternal(cli client.ClientInterface, pid1, pid2 int64, pm *progress.Manager) (*CompareResult, error) {
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка групп из проекта %d...", pid1))
	groups1, err := fetchGroupItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения групп проекта %d: %w", pid1, err)
	}

	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка групп из проекта %d...", pid2))
	groups2, err := fetchGroupItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения групп проекта %d: %w", pid2, err)
	}

	return compareItemInfos("groups", pid1, pid2, groups1, groups2), nil
}

// fetchGroupItems fetches all groups for a project and returns them as ItemInfo slice.
func fetchGroupItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	groups, err := cli.GetGroups(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(groups))
	for _, g := range groups {
		items = append(items, ItemInfo{
			ID:   g.ID,
			Name: g.Name,
		})
	}

	return items, nil
}
