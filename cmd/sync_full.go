package cmd

import (
	"fmt"
	"gotr/internal/utils"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var syncFullCmd = &cobra.Command{
	Use:   "full",
	Short: "Полная миграция (shared-steps + cases за один проход)",
	Long: `Выполняет полную миграцию: сначала переносит shared steps (формирует mapping), затем переносит cases.

Процесс:
	1) Перенос shared steps (Fetch → Filter → Import)
	2) Перенос cases (Fetch → Filter → Import)
	3) Сохранение mapping (если --save-mapping)

Пример:
	gotr sync full --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve --save-mapping
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		client := GetClient(cmd)

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		logDir := utils.LogDir()
		m, err := newMigration(client, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		mainBar := pb.StartNew(2) // 2 основных этапа: shared + cases
		mainBar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }}`)
		defer mainBar.Finish()

		// Шаг 1) Миграция shared steps (Fetch → Filter → Import)
		mainBar.Increment()
		fmt.Println("Шаг 1/2: Миграция shared steps...")
		if err := m.MigrateSharedSteps(dryRun || !autoApprove); err != nil { // если dry-run — без импорта
			return err
		}

		if dryRun {
			fmt.Println("Dry-run завершён")
			return nil
		}

		// Шаг 2) Миграция cases (Fetch → Filter → Import)
		mainBar.Increment()
		fmt.Println("Шаг 2/2: Миграция cases...")
		if err := m.MigrateCases(dryRun); err != nil {
			return err
		}

		if autoSaveMapping {
			m.ExportMapping(logDir)
		}

		fmt.Println("Полная миграция завершена!")
		return nil
	},
}

func init() {
	addSyncFlags(syncFullCmd)
	syncCmd.AddCommand(syncFullCmd)
}
