// Package parse provides utilities for parsing IDs from strings and command arguments
package parse

import (
	"fmt"
	"strconv"
)

// ID parses an ID from a string representation
func ID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// IDFromArgs parses an ID from command arguments at the specified index
func IDFromArgs(args []string, index int) (int64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("отсутствует аргумент с ID на позиции %d", index)
	}
	return ID(args[index])
}

// RequiredID validates that an ID is provided at the specified argument index
// Returns the parsed ID or an error if the ID is missing or invalid
func RequiredID(args []string, index int, name string) (int64, error) {
	if len(args) <= index {
		return 0, fmt.Errorf("необходимо указать %s", name)
	}
	id, err := ID(args[index])
	if err != nil {
		return 0, fmt.Errorf("некорректный %s: %w", name, err)
	}
	return id, nil
}
