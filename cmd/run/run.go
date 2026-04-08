package run

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for managing test runs.
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Manage test runs in TestRail",
	Long: `Commands for managing test runs in TestRail.

A test run is an instance of a test suite launched for test execution.

Subcommands:
	get     — get test run information by ID
	list    — get list of project test runs
	create  — create a new test run
	update  — update an existing test run
	close   — close a test run (complete)
	delete  — delete a test run

Examples:
	# Get test run information
	gotr run get 12345

	# Get project runs list
	gotr run list 30

	# Create a new test run
	gotr run create 30 --name "Smoke Tests v2.0" --suite-id 20069

	# Close a test run
	gotr run close 12345
`,
}

var clientAccessor *client.Accessor

// SetGetClientForTests sets getClient for tests.
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// getClientSafe safely calls getClient with a nil check.
func getClientSafe(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd.Context())
}

// Register registers the run command and all its subcommands.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Add subcommands
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(closeCmd)
	Cmd.AddCommand(deleteCmd)

	// Common flags for all subcommands
	for _, subCmd := range Cmd.Commands() {
		output.AddFlag(subCmd)
	}

	// Mark required flags for create (already defined in constructor)
	_ = createCmd.MarkFlagRequired("name")
}
