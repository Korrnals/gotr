package get

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// newCaseHistoryCmd creates the command for retrieving case change history.
func newCaseHistoryCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "case-history [case-id]",
		Short: "Get case change history by case ID",
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
					return fmt.Errorf("case_id required: gotr get case-history [case-id]")
				}

				projectID, err := interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return fmt.Errorf("newCaseHistoryCmd.func: %w", err)
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
					return fmt.Errorf("newCaseHistoryCmd.func: %w", err)
				}

				cases, err := cli.GetCases(ctx, projectID, suiteID, 0)
				if err != nil {
					return fmt.Errorf("failed to get cases: %w", err)
				}
				if len(cases) == 0 {
					return fmt.Errorf("no cases found in suite %d", suiteID)
				}

				id, err = interactive.SelectCase(ctx, interactive.PrompterFromContext(ctx), cases, "")
				if err != nil {
					return fmt.Errorf("newCaseHistoryCmd.func: %w", err)
				}
			}

			history, err := cli.GetHistoryForCase(ctx, id)
			if err != nil {
				return fmt.Errorf("newCaseHistoryCmd.func: %w", err)
			}

			return handleOutput(command, history, start)
		},
	}
}

// newSharedStepHistoryCmd creates the command for retrieving shared step change history.
func newSharedStepHistoryCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "sharedstep-history [step-id]",
		Short: "Get shared step change history by step ID",
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
					return fmt.Errorf("step_id required: gotr get sharedstep-history [step-id]")
				}

				projectID, err := interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return fmt.Errorf("newSharedStepHistoryCmd.func: %w", err)
				}

				steps, err := cli.GetSharedSteps(ctx, projectID)
				if err != nil {
					return fmt.Errorf("failed to get shared steps: %w", err)
				}
				if len(steps) == 0 {
					return fmt.Errorf("no shared steps found in project %d", projectID)
				}

				id, err = interactive.SelectSharedStep(ctx, interactive.PrompterFromContext(ctx), steps, "")
				if err != nil {
					return fmt.Errorf("newSharedStepHistoryCmd.func: %w", err)
				}
			}

			history, err := cli.GetSharedStepHistory(ctx, id)
			if err != nil {
				return fmt.Errorf("newSharedStepHistoryCmd.func: %w", err)
			}

			return handleOutput(command, history, start)
		},
	}
}

// caseHistoryCmd is the exported command registered with the root.
var caseHistoryCmd = newCaseHistoryCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// sharedStepHistoryCmd is the exported command registered with the root.
var sharedStepHistoryCmd = newSharedStepHistoryCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
