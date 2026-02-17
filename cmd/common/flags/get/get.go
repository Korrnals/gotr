// Package get provides utilities for retrieving flag values from cobra commands
package get

import (
	"github.com/spf13/cobra"
)

// FlagInt64 retrieves an int64 flag value from the command
func FlagInt64(cmd *cobra.Command, name string) (int64, error) {
	return cmd.Flags().GetInt64(name)
}

// FlagString retrieves a string flag value from the command
// Returns empty string if the flag is not set or has no value
func FlagString(cmd *cobra.Command, name string) string {
	s, _ := cmd.Flags().GetString(name)
	return s
}

// FlagBool retrieves a boolean flag value from the command
// Returns false if the flag is not set
func FlagBool(cmd *cobra.Command, name string) bool {
	b, _ := cmd.Flags().GetBool(name)
	return b
}
