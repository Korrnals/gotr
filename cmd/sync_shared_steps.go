package cmd

import (
	"fmt"
	"gotr/internal/migration"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var syncSharedStepsCmd = &cobra.Command{
	Use:   "shared-steps",
	Short: "Миграция общих шагов (shared steps)",

	RunE: func(cmd *cobra.Command, args []string) error {
		client := GetClient(cmd)

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		autoSaveFiltered, _ := cmd.Flags().GetBool("save-filtered")

		logDir := ".testrail"
		m, err := migration.NewMigration(client, srcProject, srcSuite, dstProject, 0, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		mainBar := pb.StartNew(6)
		mainBar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }}`)
		defer mainBar.Finish()

		mainBar.Increment()
		sourceSteps, targetSteps, err := m.FetchSharedStepsData()
		if err != nil {
			return err
		}

		mainBar.Increment()
		sourceCases, err := m.Client.GetCases(srcProject, srcSuite, 0)
		if err != nil {
			return err
		}
		caseIDsSet := make(map[int64]struct{})
		for _, c := range sourceCases {
			caseIDsSet[c.ID] = struct{}{}
		}

		mainBar.Increment()
		filtered, err := m.FilterSharedSteps(sourceSteps, targetSteps, caseIDsSet)
		if err != nil {
			return err
		}

		fmt.Printf("\nГотово к импорту: %d новых shared steps\n", len(filtered))

		if dryRun {
			fmt.Println("Dry-run: импорт не выполнен")
			return nil
		}

		if len(filtered) == 0 {
			fmt.Println("Нет новых shared steps")
			return nil
		}

		mainBar.Increment()
		if !autoApprove {
			fmt.Printf("Подтверждение импорта %d shared steps...\n", len(filtered))
			fmt.Print("Продолжить? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Println("Отменено")
				return nil
			}
		}

		mainBar.Increment()
		importBar := pb.StartNew(len(filtered))
		importBar.SetTemplateString(`Импорт: {{counters . }} {{bar . | green}} {{percent . }}`)
		defer importBar.Finish()

		err = m.ImportSharedSteps(filtered, false)
		if err != nil {
			return err
		}

		mainBar.Increment()
		if autoSaveMapping {
			m.ExportMapping(logDir)
		} else if len(m.Mapping()) > 0 {
			fmt.Print("\nСохранить mapping? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				m.ExportMapping(logDir)
			}
		}

		if autoSaveFiltered {
			// Сохранение filtered
		} else if len(filtered) > 0 {
			fmt.Print("\nСохранить filtered shared steps? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				// Сохранение
			}
		}

		return nil
	},
}

func init() {
	syncCmd.AddCommand(syncSharedStepsCmd)

	// Флаги как в cases + shared
	syncSharedStepsCmd.Flags().Int64("src-project", 0, "Source project ID")
	syncSharedStepsCmd.Flags().Int64("src-suite", 0, "Source suite ID")
	syncSharedStepsCmd.Flags().Int64("dst-project", 0, "Destination project ID")
	syncSharedStepsCmd.Flags().Int64("dst-suite", 0, "Destination suite ID")
	syncSharedStepsCmd.Flags().String("compare-field", "title", "Поле для дубликатов")
	syncSharedStepsCmd.Flags().Bool("dry-run", false, "Просмотр без импорта")
	syncSharedStepsCmd.Flags().BoolP("approve", "y", false, "Автоматическое подтверждение")
	syncSharedStepsCmd.Flags().BoolP("save-mapping", "m", false, "Автоматически сохранить mapping")
}
