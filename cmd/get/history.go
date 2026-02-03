package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// caseHistoryCmd — подкоманда для истории кейса
var caseHistoryCmd = &cobra.Command{
	Use:   "case-history <case-id>",
	Short: "Получить историю изменений кейса по ID кейса",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID кейса: %w", err)
		}

		history, err := client.GetHistoryForCase(id)
		if err != nil {
			return err
		}

		return handleOutput(command, history, start)
	},
}

// sharedStepHistoryCmd — подкоманда для истории shared step
var sharedStepHistoryCmd = &cobra.Command{
	Use:   "sharedstep-history <step-id>",
	Short: "Получить историю изменений shared step по ID шага",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID шага: %w", err)
		}

		history, err := client.GetSharedStepHistory(id)
		if err != nil {
			return err
		}

		return handleOutput(command, history, start)
	},
}
