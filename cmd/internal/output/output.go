// Package output предоставляет общие функции для вывода результатов команд.
// Стандартизирует поведение флага -o (output) во всех CLI командах.
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Result выводит результат в зависимости от флага -o.
// Без флага -o: выводит JSON в stdout.
// С флагом -o: сохраняет JSON в указанный файл.
func Result(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации в JSON: %w", err)
	}

	if output != "" {
		// Сохраняем в файл
		if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
			return fmt.Errorf("ошибка записи файла %s: %w", output, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Результат сохранён в %s\n", output)
		return nil
	}

	// Выводим в stdout
	fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))
	return nil
}

// JSON выводит данные в JSON формате в stdout (без сохранения в файл).
func JSON(cmd *cobra.Command, data interface{}) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// List выводит список результатов с учетом флага -o.
// Делегирует вызов Result.
func List(cmd *cobra.Command, data interface{}) error {
	return Result(cmd, data)
}
