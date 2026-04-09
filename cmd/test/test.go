package test

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for managing tests.
var Cmd = &cobra.Command{
	Use:   "test",
	Short: "Manage tests in TestRail",
	Long: `Commands for retrieving and managing tests in TestRail.

A test is a specific instance of a test case within a test run.
Each test has a status (passed, failed, blocked, etc.) and can be
assigned to a specific user.

Subcommands:
	get     — get test information by ID
	list    — list tests in a run

Examples:
	# Get test information
	gotr test get 12345

	# List tests in a run
	gotr test list 100

	# Get only failed tests
	gotr test list 100 --status-id 5

	# Get tests assigned to a user
	gotr test list 100 --assigned-to 10
`,
}

// clientAccessor is the global accessor for obtaining a client.
var clientAccessor *client.Accessor

// getClientInterface returns the client as ClientInterface.
func getClientInterface(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd.Context())
}

// SetGetClientForTests sets the getClient function for testing.
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// Register registers the test command and all its subcommands.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Create and register subcommands using constructors.
	// Flags are defined inside the constructors.
	getCmd := newGetCmd(getClientInterface)
	listCmd := newListCmd(getClientInterface)

	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
}
