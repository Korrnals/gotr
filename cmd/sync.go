package cmd

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Синхронизация данных TestRail между проектами",
	Long: `Родительская команда для миграции. Используется для перемещения сущностей между проектами/suites.

Подкоманды:
	• shared-steps — миграция общих шагов (генерирует mapping)
	• cases        — миграция кейсов (требует mapping)
	• full         — полная миграция (shared-steps + cases за один проход)
	• suites       — миграция suites между проектами
	• sections     — миграция sections между сюитами

Логи и mapping сохраняются в директории: .testrail (лог-файлы находятся в .testrail/logs/)

Примеры:
	gotr sync full --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve --save-mapping
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping shared_steps_mapping.json --dry-run
	gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 31 --approve --output shared_steps_mapping.json
`,

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.AddCommand(syncSharedStepsCmd)
	syncCmd.AddCommand(syncCasesCmd)
	syncCmd.AddCommand(syncFullCmd)
	syncCmd.AddCommand(syncSuitesCmd)
	syncCmd.AddCommand(syncSectionsCmd)
}
