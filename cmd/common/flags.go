// cmd/common/flags.go
// Общие функции для работы с флагами CLI
package common

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ParseID парсит ID из строки
func ParseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseIDFromArgs парсит ID из аргументов команды
func ParseIDFromArgs(args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("отсутствует аргумент с ID на позиции %d", index)
	}
	return ParseID(args[index])
}

// GetFlagInt64 получает int64 флаг с проверкой
func GetFlagInt64(cmd *cobra.Command, name string) (int64, error) {
	return cmd.Flags().GetInt64(name)
}

// GetFlagString получает string флаг
func GetFlagString(cmd *cobra.Command, name string) string {
	s, _ := cmd.Flags().GetString(name)
	return s
}

// GetFlagBool получает bool флаг
func GetFlagBool(cmd *cobra.Command, name string) bool {
	b, _ := cmd.Flags().GetBool(name)
	return b
}

// ValidateRequiredID проверяет что ID указан (для не-интерактивных команд)
func ValidateRequiredID(args []string, index int, name string) (int64, error) {
	if len(args) <= index {
		return 0, fmt.Errorf("необходимо указать %s", name)
	}
	id, err := ParseID(args[index])
	if err != nil {
		return 0, fmt.Errorf("некорректный %s: %w", name, err)
	}
	return id, nil
}
