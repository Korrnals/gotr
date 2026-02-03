package cmd

import (
	"strings"

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

	Args: cobra.ExactArgs(1), // Требует ровно один аргумент
	Run: func(cmd *cobra.Command, args []string) {
		resource := strings.ToLower(args[0])

		// Читаем флаги, которые объявили ниже в init()
		jsonOutput, _ := cmd.Flags().GetBool("json")   // true, если --json
		shortOutput, _ := cmd.Flags().GetBool("short") // true, если --short

		// Красивый вывод в JSON
		if jsonOutput {
			getResourceEndpoints(resource, "json")
			return
		}
		// Короткий вывод (Method + URI)
		if shortOutput {
			getResourceEndpoints(resource, "short")
			return
		}
		// Полный, красивый вывод
		getResourceEndpoints(resource, "")
	},
}
