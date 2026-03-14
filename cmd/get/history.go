package get

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/spf13/cobra"
)

// newCaseHistoryCmd создаёт команду для получения истории кейса
func newCaseHistoryCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "case-history <case-id>",
		Short: "Получить историю изменений кейса по ID кейса",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			idStr := args[0]
			id, err := flags.ParseID(idStr)
			if err != nil {
				return fmt.Errorf("invalid case ID: %w", err)
			}

			history, err := cli.GetHistoryForCase(ctx, id)
			if err != nil {
				return err
			}

			return handleOutput(command, history, start)
		},
	}
}

// newSharedStepHistoryCmd создаёт команду для получения истории shared step
func newSharedStepHistoryCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "sharedstep-history <step-id>",
		Short: "Получить историю изменений shared step по ID шага",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			idStr := args[0]
			id, err := flags.ParseID(idStr)
			if err != nil {
				return fmt.Errorf("invalid step ID: %w", err)
			}

			history, err := cli.GetSharedStepHistory(ctx, id)
			if err != nil {
				return err
			}

			return handleOutput(command, history, start)
		},
	}
}

// caseHistoryCmd — экспортированная команда для регистрации
var caseHistoryCmd = newCaseHistoryCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// sharedStepHistoryCmd — экспортированная команда для регистрации
var sharedStepHistoryCmd = newSharedStepHistoryCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
