package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/internal/client"
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
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			idStr := args[0]
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID кейса: %w", err)
			}

			history, err := cli.GetHistoryForCase(id)
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
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			idStr := args[0]
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID шага: %w", err)
			}

			history, err := cli.GetSharedStepHistory(id)
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
