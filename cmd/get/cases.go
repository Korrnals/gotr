package get

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newCasesCmd creates the command for retrieving project test cases.
func newCasesCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cases [project-id]",
		Short: "Получить кейсы проекта",
		Long: `Получить кейсы проекта.

Если проект содержит несколько сьютов и --suite-id не указан, 
будет предложено выбрать сьют из списка.

Обязательные параметры:
	ID проекта — позиционный аргумент или флаг --project-id

Опциональные параметры:
	ID сюиты — флаг --suite-id (обязателен для проектов с multiple suites)
	ID секции — флаг --section-id
	Все сьюты — флаг --all-suites (получить кейсы из всех сьютов проекта)

Примеры:
	# Автоматический выбор сьюта (если один) или интерактивный выбор
	gotr get cases 30

	# Явное указание сьюта
	gotr get cases 30 --suite-id 20069

	# Получить кейсы из всех сьютов проекта
	gotr get cases 30 --all-suites

	# С фильтрацией по секции
	gotr get cases 30 --suite-id 20069 --section-id 100
`,
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			projectIDStr := ""
			if len(args) > 0 {
				projectIDStr = args[0]
			}
			if pid, _ := command.Flags().GetString("project-id"); pid != "" {
				projectIDStr = pid
			}
			var projectID int64
			var err error

			if projectIDStr == "" {
				// Interactive project selection
				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			} else {
				projectID, err = flags.ParseID(projectIDStr)
				if err != nil {
					return fmt.Errorf("invalid project_id: %w", err)
				}
			}

			sectionID, _ := command.Flags().GetInt64("section-id")
			allSuites, _ := command.Flags().GetBool("all-suites")
			suiteID, _ := command.Flags().GetInt64("suite-id")

			// Use the explicit suite-id if provided
			if suiteID != 0 {
				return fetchAndOutputCases(ctx, command, cli, projectID, suiteID, sectionID, start)
			}

			// Fetch suites for the project
			suites, err := cli.GetSuites(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to get suites for project %d: %w", projectID, err)
			}

			if len(suites) == 0 {
				return fmt.Errorf("no suites found in project %d", projectID)
			}

			// If --all-suites, collect cases from every suite
			if allSuites {
				return fetchCasesFromAllSuites(ctx, command, cli, projectID, suites, sectionID, start)
			}

			// Single suite — use it automatically
			if len(suites) == 1 {
				ui.Infof(os.Stdout, "Project has one suite (ID: %d), using automatically...", suites[0].ID)
				return fetchAndOutputCases(ctx, command, cli, projectID, suites[0].ID, sectionID, start)
			}

			// Multiple suites — interactive selection
			selectedSuiteID, err := interactive.SelectSuite(ctx, interactive.PrompterFromContext(ctx), suites, "")
			if err != nil {
				return err
			}

			return fetchAndOutputCases(ctx, command, cli, projectID, selectedSuiteID, sectionID, start)
		},
	}

	cmd.Flags().Int64P("suite-id", "s", 0, "ID тест-сюиты (если не указан — будет предложен выбор)")
	cmd.Flags().Int64("section-id", 0, "ID секции (опционально)")
	cmd.Flags().Bool("all-suites", false, "Получить кейсы из всех сьютов проекта")
	cmd.Flags().String("project-id", "", "ID проекта (альтернатива позиционному аргументу)")

	return cmd
}

// newCaseCmd creates the command for retrieving a single test case.
func newCaseCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "case [case-id]",
		Short: "Получить один кейс по ID кейса",
		Long:  "Получает детальную информацию о конкретном тест-кейсе по его ID.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			var id int64
			var err error
			if len(args) > 0 {
				id, err = flags.ParseID(args[0])
				if err != nil {
					return fmt.Errorf("invalid case ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("case_id required: gotr get case [case-id]")
				}

				projectID, err := interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}

				suites, err := cli.GetSuites(ctx, projectID)
				if err != nil {
					return fmt.Errorf("failed to get suites: %w", err)
				}
				if len(suites) == 0 {
					return fmt.Errorf("no suites found in project %d", projectID)
				}

				suiteID, err := interactive.SelectSuite(ctx, interactive.PrompterFromContext(ctx), suites, "")
				if err != nil {
					return err
				}

				cases, err := cli.GetCases(ctx, projectID, suiteID, 0)
				if err != nil {
					return fmt.Errorf("failed to get cases: %w", err)
				}
				if len(cases) == 0 {
					return fmt.Errorf("no cases found in suite %d", suiteID)
				}

				id, err = selectCaseID(ctx, cases)
				if err != nil {
					return err
				}
			}

			kase, err := cli.GetCase(ctx, id)
			if err != nil {
				return err
			}

			return handleOutput(command, kase, start)
		},
	}
}

// fetchAndOutputCases retrieves cases and outputs the result.
func fetchAndOutputCases(ctx context.Context, cmd *cobra.Command, cli client.ClientInterface, projectID, suiteID, sectionID int64, start time.Time) error {
	cases, err := cli.GetCases(ctx, projectID, suiteID, sectionID)
	if err != nil {
		return err
	}

	return handleOutput(cmd, cases, start)
}

// fetchCasesFromAllSuites retrieves cases from all suites in the project.
func fetchCasesFromAllSuites(ctx context.Context, cmd *cobra.Command, cli client.ClientInterface, projectID int64, suites data.GetSuitesResponse, sectionID int64, start time.Time) error {
	quiet, _ := cmd.Flags().GetBool("quiet")
	op := ui.NewOperation(ui.StatusConfig{
		Title:  fmt.Sprintf("Loading cases from %d suites...", len(suites)),
		Writer: os.Stderr,
		Quiet:  quiet,
	})
	defer op.Finish()
	task := op.AddTask("suites", len(suites))

	allCases := make(data.GetCasesResponse, 0)
	for _, suite := range suites {
		cases, err := cli.GetCases(ctx, projectID, suite.ID, sectionID)
		if err != nil {
			task.Error(err)
			task.Increment()
			continue // Skip suites that fail
		}
		allCases = append(allCases, cases...)
		task.Increment()
	}
	task.Finish()

	return handleOutput(cmd, allCases, start)
}

// casesCmd is the exported command registered with the root.
var casesCmd = newCasesCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// caseCmd is the exported command registered with the root.
var caseCmd = newCaseCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
