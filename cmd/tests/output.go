package tests

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// outputResult выводит результат в зависимости от формата.
// Поддерживает флаг --save для сохранения в файл.
func outputResult(cmd *cobra.Command, data interface{}, start time.Time) error {
	_, err := save.Output(cmd, data, "tests", "json")
	return err
}

// printJSON выводит данные в формате JSON.
func printJSON(cmd *cobra.Command, data interface{}, start time.Time) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))
	return nil
}
