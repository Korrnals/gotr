package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// outputResult выводит результат в зависимости от формата.
// Поддерживает флаг --save для сохранения в файл.
func outputResult(cmd *cobra.Command, data interface{}, start time.Time) error {
	filename, _ := cmd.Flags().GetString("save")
	if filename != "" {
		return saveToFile(cmd, data, filename)
	}

	return printJSON(cmd, data, start)
}

// saveToFile сохраняет данные в файл
func saveToFile(cmd *cobra.Command, data interface{}, filename string) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filename, jsonBytes, 0644); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Результат сохранён в %s\n", filename)
	return nil
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
