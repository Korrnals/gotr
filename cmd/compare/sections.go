package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newSectionsCmd creates the 'compare sections' subcommand.
func newSectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sections",
		Short: "Сравнить секции между проектами",
		Long: `Выполняет сравнение секций (разделов) между двумя проектами.

Примеры:
  # Сравнить секции
  gotr compare sections --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare sections --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare sections --pid1 30 --pid2 31 --save-to sections_diff.json
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

			// Compare sections
			result, err := compareSectionsInternal(cli, pid1, pid2)
			if err != nil {
				return fmt.Errorf("ошибка сравнения секций: %w", err)
			}

			// Print or save result
			return PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath)
		},
	}

	// Add flags
	addCommonFlags(cmd)

	return cmd
}

// sectionsCmd — экспортированная команда
var sectionsCmd = newSectionsCmd()

// compareSectionsInternal compares sections between two projects and returns the result.
func compareSectionsInternal(cli client.ClientInterface, pid1, pid2 int64) (*CompareResult, error) {
	sections1, err := fetchSectionItems(cli, pid1)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения секций проекта %d: %w", pid1, err)
	}

	sections2, err := fetchSectionItems(cli, pid2)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения секций проекта %d: %w", pid2, err)
	}

	return compareItemInfos("sections", pid1, pid2, sections1, sections2), nil
}

// fetchSectionItems fetches all sections for a project and returns them as ItemInfo slice.
func fetchSectionItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	// Get all suites for the project
	suites, err := cli.GetSuites(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0)
	sectionIDs := make(map[int64]bool) // Track unique section IDs

	// Fetch sections from all suites
	for _, suite := range suites {
		sections, err := cli.GetSections(projectID, suite.ID)
		if err != nil {
			continue // Skip suites that fail
		}

		for _, s := range sections {
			if !sectionIDs[s.ID] {
				sectionIDs[s.ID] = true
				name := s.Name
				if suite.Name != "" {
					name = fmt.Sprintf("%s / %s", suite.Name, s.Name)
				}
				items = append(items, ItemInfo{
					ID:   s.ID,
					Name: name,
				})
			}
		}
	}

	// If no suites or no sections found, try without suite
	if len(items) == 0 {
		sections, err := cli.GetSections(projectID, 0)
		if err != nil {
			return nil, err
		}

		for _, s := range sections {
			items = append(items, ItemInfo{
				ID:   s.ID,
				Name: s.Name,
			})
		}
	}

	return items, nil
}
