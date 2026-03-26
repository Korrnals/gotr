package cmd

import (
	"fmt"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// listCmd — основная субкоманда: gotr list <resource>
var listCmd = &cobra.Command{
	Use:   "list <resource>",
	Short: "Вывод списка доступных эндпоинтов TestRail API по ресурсу",
	Long: `Выводит список доступных эндпоинтов TestRail API v2 для указанного ресурса.

Примеры:
	gotr list projects          # эндпоинты для проектов
	gotr list cases             # эндпоинты для кейсов
	gotr list all               # все эндпоинты
	gotr list cases --json      # в формате JSON
	gotr list cases --short     # краткий вывод (Method URI)`,

	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resource := ""
		if len(args) > 0 {
			resource = strings.ToLower(args[0])
		} else {
			if !interactive.HasPrompterInContext(cmd.Context()) {
				return fmt.Errorf("resource required: gotr list <resource>")
			}
			p := interactive.PrompterFromContext(cmd.Context())
			idx, _, err := p.Select("Select resource:", ValidResources)
			if err != nil {
				return fmt.Errorf("failed to select resource: %w", err)
			}
			resource = strings.ToLower(ValidResources[idx])
		}

		// Читаем флаги, которые объявили ниже в init()
		jsonOutput, _ := cmd.Flags().GetBool("json")   // true, если --json
		shortOutput, _ := cmd.Flags().GetBool("short") // true, если --short

		// Красивый вывод в JSON
		if jsonOutput {
			getResourceEndpoints(resource, "json")
			return nil
		}
		// Короткий вывод (Method + URI)
		if shortOutput {
			getResourceEndpoints(resource, "short")
			return nil
		}
		// Полный, красивый вывод
		getResourceEndpoints(resource, "")
		return nil
	},
}
