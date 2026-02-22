package compare

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newConfigurationsCmd creates the 'compare configurations' subcommand.
func newConfigurationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configurations",
		Short: "Сравнить конфигурации между проектами",
		Long: `Выполняет сравнение конфигураций (config groups и configs) между двумя проектами.

Примеры:
  # Сравнить конфигурации
  gotr compare configurations --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare configurations --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare configurations --pid1 30 --pid2 31 --save-to configs_diff.json
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

			// Compare configurations
			result, err := compareConfigurationsInternal(cli, pid1, pid2, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения конфигураций: %w", err)
			}

			elapsed := time.Since(startTime)

			// Print or save result
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print statistics
			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				PrintCompareStats("configurations", pid1, pid2,
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// configurationsCmd — экспортированная команда
var configurationsCmd = newConfigurationsCmd()

// compareConfigurationsInternal compares configurations between two projects and returns the result.
func compareConfigurationsInternal(cli client.ClientInterface, pid1, pid2 int64, pm *progress.Manager) (*CompareResult, error) {
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка конфигураций из проекта %d...", pid1))
	configs1, err := fetchConfigurationItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения конфигураций проекта %d: %w", pid1, err)
	}

	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка конфигураций из проекта %d...", pid2))
	configs2, err := fetchConfigurationItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения конфигураций проекта %d: %w", pid2, err)
	}

	return compareItemInfos("configurations", pid1, pid2, configs1, configs2), nil
}

// fetchConfigurationItems fetches all configurations for a project and returns them as ItemInfo slice.
// This includes both config groups and individual configs (formatted as "Group / Config").
func fetchConfigurationItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	groups, err := cli.GetConfigs(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0)

	for _, group := range groups {
		// Add the config group itself
		items = append(items, ItemInfo{
			ID:   group.ID,
			Name: group.Name,
		})

		// Add individual configs
		for _, config := range group.Configs {
			items = append(items, ItemInfo{
				ID:   config.ID,
				Name: fmt.Sprintf("%s / %s", group.Name, config.Name),
			})
		}
	}

	return items, nil
}
