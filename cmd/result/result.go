package result

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an HTTP client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for managing test results.
var Cmd = &cobra.Command{
	Use:   "result",
	Short: "Manage test results in TestRail",
	Long: `Commands for adding and retrieving test results in TestRail.

Test result is the outcome of an individual test execution (passed, failed, blocked, etc.)

Subcommands:
	list       — get results for a test run (with interactive selection)
	get        — get results for a test
	get-case   — get results for a case in a run
	add        — add a result for a test
	add-case   — add a result for a case in a run
	add-bulk   — bulk add results

Examples:
	# Get results with interactive run selection
	gotr result list

	# Get results for a specific run
	gotr result list 12345

	# Get test results
	gotr result get 12345

	# Add a passed result
	gotr result add 12345 --status-id 1 --comment "Test passed successfully"

	# Add a failed result with a defect
	gotr result add 12345 --status-id 5 --comment "Found bug" --defects "BUG-123"
`,
}

var clientAccessor *client.Accessor

// SetGetClientForTests overrides the client accessor for testing.
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// getClientSafe safely calls getClient with a nil guard.
func getClientSafe(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd.Context())
}

// Register adds the result command and all its subcommands to rootCmd.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Register subcommands
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(getCaseCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(addCaseCmd)
	Cmd.AddCommand(addBulkCmd)
	Cmd.AddCommand(fieldsCmd)

	// Shared flags for all subcommands
	for _, subCmd := range Cmd.Commands() {
		output.AddFlag(subCmd)
	}

	// Mark required flags (already defined in constructors)
	_ = addCmd.MarkFlagRequired("status-id")
	_ = addCaseCmd.MarkFlagRequired("case-id")
	_ = addCaseCmd.MarkFlagRequired("status-id")
	_ = addBulkCmd.MarkFlagRequired("results-file")
}
