package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// compareCmd — команда для сравнения данных между проектами
var compareCmd = &cobra.Command{
	Use:   "compare <resource> [args...]",
	Short: "Сравнение данных между проектами",
	Long: `Сравнивает данные между двумя проектами по указанному полю.

Примеры:
	gotr compare cases --pid1 30 --pid2 31 --field title
	gotr compare cases --pid1 30 --pid2 31 --field priority_id
`,
	Args: cobra.MinimumNArgs(1), // resource обязателен
	RunE: func(cmd *cobra.Command, args []string) error {
		client := GetClient(cmd)

		resource := args[0]
		if resource != "cases" {
			return fmt.Errorf("пока поддерживается только 'cases'. Добавим позже другие ресурсы")
		}

		pid1Str, _ := cmd.Flags().GetString("pid1")
		pid1, err := strconv.ParseInt(pid1Str, 10, 64)
		if err != nil || pid1 <= 0 {
			return fmt.Errorf("укажите корректный pid1 (--pid1)")
		}

		pid2Str, _ := cmd.Flags().GetString("pid2")
		pid2, err := strconv.ParseInt(pid2Str, 10, 64)
		if err != nil || pid2 <= 0 {
			return fmt.Errorf("укажите корректный pid2 (--pid2)")
		}

		field, _ := cmd.Flags().GetString("field")
		if field == "" {
			field = "title" // default
		}

		diff, err := client.DiffCasesData(pid1, pid2, field)
		if err != nil {
			return fmt.Errorf("ошибка сравнения: %w", err)
		}

		// Вывод
		fmt.Printf("Сравнение проектов %d и %d по полю '%s':\n\n", pid1, pid2, field)

		fmt.Printf("Только в проекте %d:\n", pid1)
		if len(diff.OnlyInFirst) == 0 {
			fmt.Println("  - Нет уникальных кейсов")
		} else {
			for _, c := range diff.OnlyInFirst {
				fmt.Printf("  - %d: %s\n", c.ID, c.Title)
			}
		}

		fmt.Printf("\nТолько в проекте %d:\n", pid2)
		if len(diff.OnlyInSecond) == 0 {
			fmt.Println("  - Нет уникальных кейсов")
		} else {
			for _, c := range diff.OnlyInSecond {
				fmt.Printf("  - %d: %s\n", c.ID, c.Title)
			}
		}

		fmt.Printf("\nОтличаются по полю '%s':\n", field)
		if len(diff.DiffByField) == 0 {
			fmt.Println("  - Нет отличий")
		} else {
			for _, d := range diff.DiffByField {
				fmt.Printf("  - Кейс %d:\n", d.CaseID)
				fmt.Printf("    Проект %d: %s\n", pid1, d.First.Title)
				fmt.Printf("    Проект %d: %s\n", pid2, d.Second.Title)
			}
		}

		return nil
	},
}
