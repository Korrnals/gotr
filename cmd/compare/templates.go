package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newTemplatesCmd creates the 'compare templates' subcommand.
func newTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Сравнить шаблоны между проектами",
		Long: `Выполняет сравнение шаблонов кейсов между двумя проектами.

Примеры:
  # Сравнить шаблоны
  gotr compare templates --pid1 30 --pid2 31

  # Сохранить результат
  gotr compare templates --pid1 30 --pid2 31 --format json --save templates_diff.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			// Parse flags
			pid1, pid2, format, saveFlag, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			// Get project names
			project1Name, project2Name, err := GetProjectNames(cli, pid1, pid2)
			if err != nil {
				return err
			}

			// Compare templates
			result, err := compareTemplatesInternal(cli, pid1, pid2)
			if err != nil {
				return fmt.Errorf("ошибка сравнения шаблонов: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, saveFlag)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// templatesCmd — экспортированная команда
var templatesCmd = newTemplatesCmd()

// compareTemplatesInternal compares templates between two projects and returns the result.
func compareTemplatesInternal(cli client.ClientInterface, pid1, pid2 int64) (*CompareResult, error) {
	templates1, err := fetchTemplateItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения шаблонов проекта %d: %w", pid1, err)
	}

	templates2, err := fetchTemplateItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения шаблонов проекта %d: %w", pid2, err)
	}

	return compareItemInfos("templates", pid1, pid2, templates1, templates2), nil
}

// fetchTemplateItems fetches all templates for a project and returns them as ItemInfo slice.
func fetchTemplateItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	templates, err := cli.GetTemplates(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(templates))
	for _, t := range templates {
		items = append(items, ItemInfo{
			ID:   t.ID,
			Name: t.Name,
		})
	}

	return items, nil
}
