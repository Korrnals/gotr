// cmd/common/flags.go
// Общие функции для работы с флагами CLI
//
// DEPRECATED: This package is being reorganized. Please use the subpackages:
//   - github.com/Korrnals/gotr/cmd/common/flags/parse for ID parsing functions
//   - github.com/Korrnals/gotr/cmd/common/flags/get for flag retrieval functions
//   - github.com/Korrnals/gotr/cmd/common/flags/save for output saving functions
//
// This file re-exports functions from the subpackages for backward compatibility.
package common

import (
	"github.com/Korrnals/gotr/cmd/common/flags/get"
	"github.com/Korrnals/gotr/cmd/common/flags/parse"
	"github.com/spf13/cobra"
)

// ParseID парсит ID из строки
// Deprecated: Use parse.ID from github.com/Korrnals/gotr/cmd/common/flags/parse instead
func ParseID(s string) (int64, error) {
	return parse.ID(s)
}

// ParseIDFromArgs парсит ID из аргументов команды
// Deprecated: Use parse.IDFromArgs from github.com/Korrnals/gotr/cmd/common/flags/parse instead
func ParseIDFromArgs(args []string, index int) (int64, error) {
	return parse.IDFromArgs(args, index)
}

// GetFlagInt64 получает int64 флаг с проверкой
// Deprecated: Use get.FlagInt64 from github.com/Korrnals/gotr/cmd/common/flags/get instead
func GetFlagInt64(cmd *cobra.Command, name string) (int64, error) {
	return get.FlagInt64(cmd, name)
}

// GetFlagString получает string флаг
// Deprecated: Use get.FlagString from github.com/Korrnals/gotr/cmd/common/flags/get instead
func GetFlagString(cmd *cobra.Command, name string) string {
	return get.FlagString(cmd, name)
}

// GetFlagBool получает bool флаг
// Deprecated: Use get.FlagBool from github.com/Korrnals/gotr/cmd/common/flags/get instead
func GetFlagBool(cmd *cobra.Command, name string) bool {
	return get.FlagBool(cmd, name)
}

// ValidateRequiredID проверяет что ID указан (для не-интерактивных команд)
// Deprecated: Use parse.RequiredID from github.com/Korrnals/gotr/cmd/common/flags/parse instead
func ValidateRequiredID(args []string, index int, name string) (int64, error) {
	return parse.RequiredID(args, index, name)
}
