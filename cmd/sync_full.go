package cmd

import (
	"fmt"
	"gotr/internal/migration"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var syncFullCmd = &cobra.Command{
	Use:   "full",
	Short: "Полная миграция (shared-steps + cases за один проход)",

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

		logDir := ".testrail"
		m, err := migration.NewMigration(client, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		mainBar := pb.StartNew(2) // 2 основных этапа: shared + cases
		mainBar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }}`)
		defer mainBar.Finish()

		mainBar.Increment() // Этап 1: shared-steps
		fmt.Println("Этап 1: Миграция shared steps...")
		if err := m.MigrateSharedSteps(dryRun || !autoApprove); err != nil { // если dry-run — без импорта
			return err
		}

		if dryRun {
			fmt.Println("Dry-run завершён")
			return nil
		}

		mainBar.Increment() // Этап 2: cases
		fmt.Println("Этап 2: Миграция cases...")
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
	syncCmd.AddCommand(syncFullCmd)

	// Флаги как в cases + shared
	syncFullCmd.Flags().Int64("src-project", 0, "Source project ID")
	syncFullCmd.Flags().Int64("src-suite", 0, "Source suite ID")
	syncFullCmd.Flags().Int64("dst-project", 0, "Destination project ID")
	syncFullCmd.Flags().Int64("dst-suite", 0, "Destination suite ID")
	syncFullCmd.Flags().String("compare-field", "title", "Поле для дубликатов")
	syncFullCmd.Flags().Bool("dry-run", false, "Просмотр без импорта")
	syncFullCmd.Flags().BoolP("approve", "y", false, "Автоматическое подтверждение")
	syncFullCmd.Flags().BoolP("save-mapping", "m", false, "Автоматически сохранить mapping")
}
