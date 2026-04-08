// Package tests implements CLI commands for managing TestRail tests.
package tests

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type that returns a client instance.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all test-related subcommands on the given root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	testsCmd := &cobra.Command{
		Use:   "tests",
		Short: "Manage tests",
		Long: `Manage tests — results of test case executions
in test runs.

A test represents a specific execution of a test case within
a test run with a specific status and result.

Available operations:
  • update — update a test (status, comment, time)`,
	}

	testsCmd.AddCommand(newUpdateCmd(getClient))
	testsCmd.AddCommand(newGetCmd(getClient))
	testsCmd.AddCommand(newListCmd(getClient))

	root.AddCommand(testsCmd)
}
