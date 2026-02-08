package cmd

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// deleteCmd — команда для удаления ресурсов
var deleteCmd = &cobra.Command{
	Use:   "delete <endpoint> <id>",
	Short: "Удалить ресурс (DELETE/POST-запрос)",
	Long: `Удаляет существующий объект в TestRail.

Поддерживаемые эндпоинты:
  project <id>       Удалить проект
  suite <id>         Удалить сьют
  section <id>       Удалить секцию
  case <id>          Удалить тест-кейс
  run <id>           Удалить тест-ран
  shared-step <id>   Удалить shared step
  milestone <id>     Удалить milestone
  plan <id>          Удалить test plan

Примеры:
  gotr delete project 1
  gotr delete case 12345
  gotr delete run 1000

Dry-run mode:
  gotr delete case 12345 --dry-run  # Show what would be deleted`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	deleteCmd.Flags().Bool("soft", false, "Мягкое удаление (где поддерживается)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("необходимо указать endpoint и id: gotr delete <endpoint> <id>")
	}

	endpoint := args[0]
	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("неверный ID: %v", err)
	}

	// Проверяем dry-run режим
	isDryRun, _ := cmd.Flags().GetBool("dry-run")
	if isDryRun {
		dr := dryrun.New("delete " + endpoint)
		return runDeleteDryRun(dr, endpoint, id)
	}

	// Получаем клиент
	cli := GetClientInterface(cmd)

	// Маршрутизация по endpoint
	switch endpoint {
	case "project":
		return cli.DeleteProject(id)
	case "suite":
		return cli.DeleteSuite(id)
	case "section":
		return cli.DeleteSection(id)
	case "case":
		return cli.DeleteCase(id)
	case "run":
		return cli.DeleteRun(id)
	case "shared-step":
		// Для shared step есть специальный флаг keep_in_cases
		return cli.DeleteSharedStep(id, 0)
	default:
		return fmt.Errorf("неподдерживаемый endpoint: %s", endpoint)
	}
}

// runDeleteDryRun выполняет dry-run для delete команды
func runDeleteDryRun(dr *dryrun.Printer, endpoint string, id int64) error {
	var method, url string

	switch endpoint {
	case "project":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_project/%d", id)
	case "suite":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_suite/%d", id)
	case "section":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_section/%d", id)
	case "case":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_case/%d", id)
	case "run":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_run/%d", id)
	case "shared-step":
		method = "POST"
		url = fmt.Sprintf("/index.php?/api/v2/delete_shared_step/%d", id)
	default:
		return fmt.Errorf("неподдерживаемый endpoint для dry-run: %s", endpoint)
	}

	dr.PrintOperation(
		fmt.Sprintf("Delete %s %d", endpoint, id),
		method,
		url,
		nil,
	)
	return nil
}
