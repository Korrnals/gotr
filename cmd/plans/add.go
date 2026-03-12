package plans

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'plans add'
// Эндпоинт: POST /add_plan/{project_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <project_id>",
		Short: "Создать новый тест-план",
		Long:  `Создаёт новый тест-план в указанном проекте.`,
		Example: `  # Создать план для спринта
  gotr plans add 1 --name="План спринта 1"

  # Создать план регрессии с описанием
  gotr plans add 1 --name="Регрессия" --description="Полный набор регрессионных тестов"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := flags.ValidateRequiredID(args, 0, "project_id")
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			req := data.AddPlanRequest{
				Name: name,
			}

			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetInt64("milestone-id"); v > 0 {
				req.MilestoneID = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans add")
				dr.PrintSimple("Create Plan", fmt.Sprintf("Project ID: %d, Name: %s", projectID, req.Name))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.AddPlan(ctx, projectID, &req)
			if err != nil {
				return fmt.Errorf("failed to create plan: %w", err)
			}

			ui.Successf(os.Stdout, "Plan created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без создания")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Название плана (обязательно)")
	cmd.Flags().String("description", "", "Описание плана")
	cmd.Flags().Int64("milestone-id", 0, "ID майлстона")

	return cmd
}
