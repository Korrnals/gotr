package milestones

import (
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// outputResult выводит результат в зависимости от формата
func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := save.Output(cmd, data, "milestones", "json")
	return err
}

// outputList выводит список результатов в зависимости от формата
func outputList(cmd *cobra.Command, data interface{}) error {
	_, err := save.Output(cmd, data, "milestones", "json")
	return err
}
