package sync

import (
	"context"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for migration.
var Cmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize TestRail data between projects",
	Long: `Parent command for migrating data between TestRail projects.

Supports interactive mode for selecting projects and suites.
If parameters are not specified, a selection from a list will be offered.

Subcommands:
	• shared-steps — migrate shared steps (generates mapping)
	• cases        — migrate cases (requires mapping)
	• full         — full migration (shared-steps + cases in one pass)
	• suites       — migrate suites between projects
	• sections     — migrate sections between suites

Logs and mapping are saved in the directory: .testrail (log files are in .testrail/logs/)

Examples:
	# Interactive mode (select all parameters)
	gotr sync full
	gotr sync cases

	# Using flags
	gotr sync full --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve --save-mapping
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping shared_steps_mapping.json --dry-run
	gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 31 --approve --output shared_steps_mapping.json
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

// testClientKey is the context key for mock client in tests.
var testClientKey = &struct{}{}

// SetTestClient sets the mock client for tests.
func SetTestClient(cmd *cobra.Command, mockClient client.ClientInterface) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, testClientKey, mockClient))
}

// getClientSafe safely calls getClient with a nil check.
// Fallback: gets client from context (for tests).
func getClientSafe(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor != nil {
		if c := clientAccessor.GetClientSafe(cmd.Context()); c != nil {
			return c
		}
	}
	// Fallback for old tests — get from context by old key
	if ctx := cmd.Context(); ctx != nil {
		if v := ctx.Value(testHTTPClientKey); v != nil {
			if c, ok := v.(client.ClientInterface); ok {
				return c
			}
		}
	}
	return nil
}

// getClientInterface safely returns ClientInterface (for tests with MockClient).
func getClientInterface(cmd *cobra.Command) client.ClientInterface {
	// First check the new key for mock clients
	if v := cmd.Context().Value(testClientKey); v != nil {
		if c, ok := v.(client.ClientInterface); ok {
			return c
		}
	}
	// Fallback: use regular getClientSafe
	return getClientSafe(cmd)
}

// Register registers the sync command and all its subcommands.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	Cmd.AddCommand(sharedStepsCmd)
	Cmd.AddCommand(casesCmd)
	Cmd.AddCommand(fullCmd)
	Cmd.AddCommand(suitesCmd)
	Cmd.AddCommand(sectionsCmd)

	// Flags for sync cases
	casesCmd.Flags().Int64("src-project", 0, "Source project ID (copy from)")
	casesCmd.Flags().Int64("src-suite", 0, "Source suite ID")
	casesCmd.Flags().Int64("dst-project", 0, "Destination project ID (copy to)")
	casesCmd.Flags().Int64("dst-suite", 0, "Destination suite ID")
	casesCmd.Flags().String("compare-field", "title", "Field for duplicate detection")
	casesCmd.Flags().String("mapping-file", "", "Mapping file for shared_step_id replacement")
	casesCmd.Flags().Bool("dry-run", false, "Preview without importing")
	casesCmd.Flags().String("output", "", "Additional JSON file with results")

	// Flags for sync shared-steps
	sharedStepsCmd.Flags().Int64("src-project", 0, "Source project ID")
	sharedStepsCmd.Flags().Int64("src-suite", 0, "Source suite ID")
	sharedStepsCmd.Flags().Int64("dst-project", 0, "Destination project ID")
	sharedStepsCmd.Flags().String("compare-field", "title", "Field for duplicate detection")
	sharedStepsCmd.Flags().Bool("approve", false, "Auto-approve confirmation")
	sharedStepsCmd.Flags().Bool("save-mapping", false, "Save mapping automatically")
	sharedStepsCmd.Flags().Bool("save-filtered", false, "Save filtered list automatically")
	sharedStepsCmd.Flags().Bool("dry-run", false, "Preview without importing")

	// Flags for sync sections
	sectionsCmd.Flags().Int64("src-project", 0, "Source project ID")
	sectionsCmd.Flags().Int64("src-suite", 0, "Source suite ID")
	sectionsCmd.Flags().Int64("dst-project", 0, "Destination project ID")
	sectionsCmd.Flags().Int64("dst-suite", 0, "Destination suite ID")
	sectionsCmd.Flags().String("compare-field", "title", "Field for duplicate detection")
	sectionsCmd.Flags().Bool("approve", false, "Auto-approve confirmation")
	sectionsCmd.Flags().Bool("dry-run", false, "Preview without importing")
	sectionsCmd.Flags().Bool("save-mapping", false, "Save mapping automatically")

	// Flags for sync full
	fullCmd.Flags().Int64("src-project", 0, "Source project ID")
	fullCmd.Flags().Int64("src-suite", 0, "Source suite ID")
	fullCmd.Flags().Int64("dst-project", 0, "Destination project ID")
	fullCmd.Flags().Int64("dst-suite", 0, "Destination suite ID")
	fullCmd.Flags().String("compare-field", "title", "Field for duplicate detection")
	fullCmd.Flags().Bool("approve", false, "Auto-approve confirmation")
	fullCmd.Flags().Bool("save-mapping", false, "Save mapping automatically")
	fullCmd.Flags().Bool("save-filtered", false, "Save filtered list automatically")
	fullCmd.Flags().Bool("dry-run", false, "Preview without importing")
}
