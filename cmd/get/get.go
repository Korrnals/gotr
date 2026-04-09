package get

import (
	"context"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an HTTP client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Cmd is the root command for GET requests to the TestRail API.
var Cmd = &cobra.Command{
	Use:   "get",
	Short: "GET requests to TestRail API",
	Long: `Performs GET requests to TestRail API.

Subcommands:
	case               - get a single case by case ID
	cases              - get project cases (requires project ID and suite ID)
	case-types         - get list of case types
	case-fields        - get list of case fields
	case-history       - get case change history by case ID

	project            - get a single project by project ID
	projects           - get all projects

	sharedstep         - get a single shared step by step ID
	sharedsteps        - get project shared steps (requires project ID)
	sharedstep-history - get shared step change history by step ID

	suite              - get a single test suite by suite ID
	suites             - get project test suites (requires project ID)

Examples:
	gotr get project 30
	gotr get projects

	gotr get case 12345
	gotr get cases 30 --suite-id 20069

	gotr get suite 20069
	gotr get suites 30
	
	gotr get sharedstep 45678
	gotr get sharedsteps 30
`,
}

var getClient GetClientFunc

// SetGetClientForTests overrides the getClient accessor for testing.
func SetGetClientForTests(fn GetClientFunc) {
	getClient = fn
}

// handleOutput delegates get-command rendering/output orchestration to internal/output.
func handleOutput(command *cobra.Command, data any, start time.Time) error {
	return output.OutputGetResult(command, data, start)
}

func runGetStatus[T any](command *cobra.Command, title string, fn func(context.Context) (T, error)) (T, error) {
	quiet, _ := command.Flags().GetBool("quiet")
	return ui.RunWithStatus(command.Context(), ui.StatusConfig{
		Title:  title,
		Writer: os.Stderr,
		Quiet:  quiet,
	}, fn)
}

// Register adds the get command and all its subcommands to rootCmd.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	getClient = clientFn
	rootCmd.AddCommand(Cmd)

	// Register subcommands
	Cmd.AddCommand(casesCmd)
	Cmd.AddCommand(caseCmd)
	Cmd.AddCommand(caseTypesCmd)
	Cmd.AddCommand(caseFieldsCmd)
	Cmd.AddCommand(caseHistoryCmd)
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(projectCmd)
	Cmd.AddCommand(sharedStepsCmd)
	Cmd.AddCommand(sharedStepCmd)
	Cmd.AddCommand(sharedStepHistoryCmd)
	Cmd.AddCommand(suitesCmd)
	Cmd.AddCommand(suiteCmd)

	// Local flags — scoped to get subcommands and their children
	for _, subCmd := range Cmd.Commands() {
		subCmd.Flags().StringP("type", "t", "json", "Output format: json, json-full, table")
		output.AddFlag(subCmd)
		subCmd.Flags().BoolP("quiet", "q", false, "Quiet mode")
		subCmd.Flags().BoolP("jq", "j", false, "Enable jq formatting (overrides config jq_format)")
		subCmd.Flags().String("jq-filter", "", "jq filter")
		subCmd.Flags().BoolP("body-only", "b", false, "Save response body only (without metadata)")
	}

	// Cases-specific flags are already defined in the newCasesCmd constructor
}
