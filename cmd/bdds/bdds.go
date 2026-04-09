// Package bdds implements CLI commands for managing TestRail BDD scenarios.
package bdds

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all BDD-related commands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	bddsCmd := &cobra.Command{
		Use:   "bdds",
		Short: "Manage BDD scenarios",
		Long: `Manage BDD (Behavior Driven Development) scenarios.

BDD scenarios describe system behavior in Given-When-Then format
using the Gherkin language. They are linked to test cases and
enable writing tests in a business-readable language.

Available operations:
  • get — retrieve a BDD scenario for a test case
  • add — add a BDD scenario to a test case`,
	}

	bddsCmd.AddCommand(newGetCmd(getClient))
	bddsCmd.AddCommand(newAddCmd(getClient))

	root.AddCommand(bddsCmd)
}
