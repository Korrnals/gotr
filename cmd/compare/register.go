package compare

import (
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientInterfaceFunc is the function type used to obtain an API client.
type GetClientInterfaceFunc func(cmd *cobra.Command) client.ClientInterface

var getClient GetClientInterfaceFunc

// Cmd is the main compare command (populated in Register).
var Cmd *cobra.Command

// SetGetClientForTests sets the getClient function for tests.
func SetGetClientForTests(fn GetClientInterfaceFunc) {
	getClient = fn
}

// Register registers the compare command and all its subcommands.
func Register(rootCmd *cobra.Command, clientFn GetClientInterfaceFunc) {
	getClient = clientFn

	// Create main compare command
	Cmd = &cobra.Command{
		Use:   "compare",
		Short: "Compare data between projects",
		Long: `Compare resources between two projects.

Supported resources:
	cases          - compare test cases
	suites         - compare test suites
	sections       - compare sections
	sharedsteps    - compare shared steps
	runs           - compare test runs
	plans          - compare test plans
	milestones     - compare milestones
	datasets       - compare datasets
	groups         - compare groups
	labels         - compare labels
	templates      - compare templates
	configurations - compare configurations
	retry-failed-pages - selectively reload failed pages from a JSON report
	all            - compare all resources

Examples:
	gotr compare cases --pid1 30 --pid2 31
	gotr compare all --pid1 30 --pid2 31 --save
	gotr compare all --pid1 30 --pid2 31 --save-to result.json
`,
	}

	// Add persistent flags FIRST (before subcommands) for completion to work
	Cmd.PersistentFlags().StringP("pid1", "1", "", "First project ID (required)")
	Cmd.PersistentFlags().StringP("pid2", "2", "", "Second project ID (required)")
	Cmd.PersistentFlags().Bool("save", false, "Save result to file (default: ~/.gotr/exports/)")
	Cmd.PersistentFlags().String("save-to", "", "Save result to the specified file")
	Cmd.PersistentFlags().Int("rate-limit", -1, "API rate limit per minute. -1 = auto by profile/deployment, 0 = unlimited, >0 = fixed value.")
	Cmd.PersistentFlags().Int("parallel-suites", 10, "Maximum number of parallel suites")
	Cmd.PersistentFlags().Int("parallel-pages", 6, "Maximum number of parallel pages per suite")
	Cmd.PersistentFlags().Int("page-retries", 5, "Number of retries per page during initial loading")
	Cmd.PersistentFlags().Duration("timeout", 30*time.Minute, "Timeout for the compare operation")
	Cmd.PersistentFlags().Int("retry-attempts", 5, "Number of attempts for targeted auto-retry of failed pages")
	Cmd.PersistentFlags().Int("retry-workers", 12, "Number of parallel workers for failed page auto-retry")
	Cmd.PersistentFlags().Duration("retry-delay", 200*time.Millisecond, "Delay between attempts for one page during auto-retry")

	// Add all subcommands
	Cmd.AddCommand(casesCmd)
	Cmd.AddCommand(suitesCmd)
	Cmd.AddCommand(sectionsCmd)
	Cmd.AddCommand(sharedStepsCmd)
	Cmd.AddCommand(runsCmd)
	Cmd.AddCommand(plansCmd)
	Cmd.AddCommand(milestonesCmd)
	Cmd.AddCommand(datasetsCmd)
	Cmd.AddCommand(groupsCmd)
	Cmd.AddCommand(labelsCmd)
	Cmd.AddCommand(templatesCmd)
	Cmd.AddCommand(configurationsCmd)
	Cmd.AddCommand(retryFailedPagesCmd)
	Cmd.AddCommand(allCmd)

	rootCmd.AddCommand(Cmd)
}
