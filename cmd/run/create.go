package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newCreateCmd creates the create command for test runs (also used in tests).
func newCreateCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [project-id]",
		Short: "Создать новый test run",
		Long: `Создаёт новый test run в указанном проекте.

Test run создаётся на основе тест-сюиты (suite). Можно указать:
- название и описание
- milestone для привязки
- пользователя для назначения (assignedto_id)
- конкретные case_ids (если не нужны все кейсы сьюты)
- config_ids для конфигурационного тестирования

Примеры:
	# Создать run с минимальными параметрами
	gotr run create 30 --suite-id 20069 --name "Smoke Tests"

	# Создать run с описанием и назначением
	gotr run create 30 --suite-id 20069 --name "Regression" \\
		--description "Full regression suite" --assigned-to 5

	# Создать run только с определёнными кейсами
	gotr run create 30 --suite-id 20069 --name "Critical Path" \\
		--case-ids 123,456,789

	# Dry-run режим
	gotr run create 30 --suite-id 20069 --name "Test" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newRunServiceFromInterface(cli)
			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid project ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr run create [project-id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr run create [project-id]")
				}
				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			}

			// Collect parameters from flags
			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")
			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("suite-id is required in non-interactive mode: gotr run create [project-id] --suite-id <suite_id>")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("suite-id is required in non-interactive mode: gotr run create [project-id] --suite-id <suite_id>")
				}
				suiteID, err = interactive.SelectSuiteForProject(ctx, interactive.PrompterFromContext(ctx), cli, projectID, "")
				if err != nil {
					return err
				}
			}
			milestoneID, _ := cmd.Flags().GetInt64("milestone-id")
			assignedTo, _ := cmd.Flags().GetInt64("assigned-to")
			caseIDs, _ := cmd.Flags().GetInt64Slice("case-ids")
			configIDs, _ := cmd.Flags().GetInt64Slice("config-ids")
			includeAll, _ := cmd.Flags().GetBool("include-all")

			req := &data.AddRunRequest{
				Name:        name,
				Description: description,
				SuiteID:     suiteID,
				MilestoneID: milestoneID,
				AssignedTo:  assignedTo,
				CaseIDs:     caseIDs,
				ConfigIDs:   configIDs,
				IncludeAll:  includeAll,
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run create")
				dr.PrintOperation(
					fmt.Sprintf("Create Run in Project %d", projectID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_run/%d", projectID),
					req,
				)
				return nil
			}

			run, err := svc.Create(ctx, projectID, req)
			if err != nil {
				return fmt.Errorf("failed to create test run: %w", err)
			}

			svc.PrintSuccess(ctx, cmd, "Test run создан успешно (ID: %d):", run.ID)
			return svc.Output(ctx, cmd, run)
		},
	}

	cmd.Flags().Int64P("suite-id", "s", 0, "ID тест-сюиты (обязательный)")
	cmd.Flags().String("name", "", "Название test run (обязательный)")
	cmd.Flags().String("description", "", "Описание test run")
	cmd.Flags().Int64("milestone-id", 0, "ID milestone")
	cmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")
	cmd.Flags().Int64Slice("case-ids", nil, "Список ID кейсов для включения (через запятую)")
	cmd.Flags().Int64Slice("config-ids", nil, "Список ID конфигураций (через запятую)")
	cmd.Flags().Bool("include-all", true, "Включить все кейсы сьюты")
	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	cmd.MarkFlagRequired("name")

	return cmd
}

// createCmd is the exported command for registration.
var createCmd = newCreateCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
