package get

import (
	"context"
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newSharedStepsCmd creates the command for retrieving project shared steps.
func newSharedStepsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sharedsteps [project-id]",
		Short: "Получить shared steps проекта",
		Long: `Получить shared steps (общие шаги) проекта.

Если ID проекта не указан, будет предложено выбрать проект из списка.

Примеры:
	# Интерактивный выбор проекта
	gotr get sharedsteps

	# Явное указание проекта
	gotr get sharedsteps 30
	gotr get sharedsteps --project-id 30
`,
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			// Resolve project ID
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

			steps, err := runGetStatus(command, "Loading shared steps...", func(ctx context.Context) (any, error) {
				return cli.GetSharedSteps(ctx, projectID)
			})
			if err != nil {
				return err
			}

			return handleOutput(command, steps, start)
		},
	}

	cmd.Flags().String("project-id", "", "ID проекта (альтернатива позиционному аргументу)")

	return cmd
}

// newSharedStepCmd creates the command for retrieving a single shared step.
func newSharedStepCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "sharedstep [step-id]",
		Short: "Получить один shared step по ID шага",
		Long:  "Получает детальную информацию о конкретном shared step по его ID.",
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
					return fmt.Errorf("invalid step ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("step_id required: gotr get sharedstep [step-id]")
				}

				projectID, err := interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}

				steps, err := cli.GetSharedSteps(ctx, projectID)
				if err != nil {
					return fmt.Errorf("failed to get shared steps: %w", err)
				}
				if len(steps) == 0 {
					return fmt.Errorf("no shared steps found in project %d", projectID)
				}

				id, err = selectSharedStepID(ctx, steps)
				if err != nil {
					return err
				}
			}

			step, err := cli.GetSharedStep(ctx, id)
			if err != nil {
				return err
			}

			return handleOutput(command, step, start)
		},
	}
}

func selectSharedStepID(ctx context.Context, steps data.GetSharedStepsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(steps))
	for i, step := range steps {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, step.ID, step.Title))
	}

	idx, _, err := p.Select("Select shared step:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select shared step: %w", err)
	}

	return steps[idx].ID, nil
}

// sharedStepsCmd is the exported command registered with the root.
var sharedStepsCmd = newSharedStepsCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// sharedStepCmd is the exported command registered with the root.
var sharedStepCmd = newSharedStepCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
