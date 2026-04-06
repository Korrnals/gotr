package get

import (
	"context"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an HTTP client.
type GetClientFunc func(cmd *cobra.Command) *client.HTTPClient

// Cmd is the root command for GET requests to the TestRail API.
var Cmd = &cobra.Command{
	Use:   "get",
	Short: "GET-запросы к TestRail API",
	Long: `Выполняет GET-запросы к TestRail API.

Подкоманды:
	case               - получить один кейс по ID кейса
	cases              - получить кейсы проекта (требует ID проекта и ID сюиты)
	case-types         - получить список типов кейсов
	case-fields        - получить список полей кейсов
	case-history       - получить историю изменений кейса по ID кейса

	project            - получить один проект по ID проекта
	projects           - получить все projects

	sharedstep         - получить один shared step по ID шага
	sharedsteps        - получить shared steps проекта (требует ID проекта)
	sharedstep-history - получить историю изменений shared step по ID шага

	suite              - получить одну тест-сюиту по ID сюиты
	suites             - получить тест-сюиты проекта (требует ID проекта)

Примеры:
	gotr get project 30
	gotr get projects

	gotr get case 12345
	gotr get cases 30 --suite-id 20069

	gotr get suite 20069
	gotr get suites 30
	
	gotr get sharedstep 45678
	gotr get sharedsteps 30
`,
}

var getClient GetClientFunc

// SetGetClientForTests overrides the getClient accessor for testing.
func SetGetClientForTests(fn GetClientFunc) {
	getClient = fn
}

// handleOutput delegates get-command rendering/output orchestration to internal/output.
func handleOutput(command *cobra.Command, data any, start time.Time) error {
	return output.OutputGetResult(command, data, start)
}

func runGetStatus[T any](command *cobra.Command, title string, fn func(context.Context) (T, error)) (T, error) {
	quiet, _ := command.Flags().GetBool("quiet")
	return ui.RunWithStatus(command.Context(), ui.StatusConfig{
		Title:  title,
		Writer: os.Stderr,
		Quiet:  quiet,
	}, fn)
}

// Register adds the get command and all its subcommands to rootCmd.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	getClient = clientFn
	rootCmd.AddCommand(Cmd)

	// Register subcommands
	Cmd.AddCommand(casesCmd)
	Cmd.AddCommand(caseCmd)
	Cmd.AddCommand(caseTypesCmd)
	Cmd.AddCommand(caseFieldsCmd)
	Cmd.AddCommand(caseHistoryCmd)
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(projectCmd)
	Cmd.AddCommand(sharedStepsCmd)
	Cmd.AddCommand(sharedStepCmd)
	Cmd.AddCommand(sharedStepHistoryCmd)
	Cmd.AddCommand(suitesCmd)
	Cmd.AddCommand(suiteCmd)

	// Local flags — scoped to get subcommands and their children
	for _, subCmd := range Cmd.Commands() {
		subCmd.Flags().StringP("type", "t", "json", "Формат вывода: json, json-full, table")
		output.AddFlag(subCmd)
		subCmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
		subCmd.Flags().BoolP("jq", "j", false, "Включить jq-форматирование (переопределяет конфиг jq_format)")
		subCmd.Flags().String("jq-filter", "", "jq-фильтр")
		subCmd.Flags().BoolP("body-only", "b", false, "Сохранить только тело ответа (без метаданных)")
	}

	// Cases-specific flags are already defined in the newCasesCmd constructor
}
