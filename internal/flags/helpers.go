// Package flags provides helper functions for CLI flag handling.
package flags

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ParseID parses ID from string.
func ParseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseIDFromArgs parses ID from command arguments.
func ParseIDFromArgs(args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("отсутствует аргумент с ID на позиции %d", index)
	}
	return ParseID(args[index])
}

// ValidateRequiredID validates that ID is provided (for non-interactive commands).
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

// GetFlagInt64 gets int64 flag with error handling.
func GetFlagInt64(cmd *cobra.Command, name string) (int64, error) {
	return cmd.Flags().GetInt64(name)
}

// GetFlagString gets string flag.
func GetFlagString(cmd *cobra.Command, name string) string {
	s, _ := cmd.Flags().GetString(name)
	return s
}

// GetFlagBool gets bool flag.
func GetFlagBool(cmd *cobra.Command, name string) bool {
	b, _ := cmd.Flags().GetBool(name)
	return b
}
