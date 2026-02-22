package compare

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
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

  # Сохранить результат в файл по умолчанию
  gotr compare templates --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare templates --pid1 30 --pid2 31 --save-to templates_diff.json
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

			// Compare templates
			result, err := compareTemplatesInternal(cli, pid1, pid2, pm)
			if err != nil {
				return fmt.Errorf("ошибка сравнения шаблонов: %w", err)
			}

			elapsed := time.Since(startTime)

			// Print or save result
			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			// Print statistics
			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				PrintCompareStats("templates", pid1, pid2,
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed)
			}

			return nil
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// templatesCmd — экспортированная команда
var templatesCmd = newTemplatesCmd()

// compareTemplatesInternal compares templates between two projects and returns the result.
func compareTemplatesInternal(cli client.ClientInterface, pid1, pid2 int64, pm *progress.Manager) (*CompareResult, error) {
	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка шаблонов из проекта %d...", pid1))
	templates1, err := fetchTemplateItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения шаблонов проекта %d: %w", pid1, err)
	}

	progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Загрузка шаблонов из проекта %d...", pid2))
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
