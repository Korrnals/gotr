package result

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd creates the 'result add' command.
// Endpoint: POST /add_result/{test_id}
func newAddCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [test-id]",
		Short: "Add a result for a test",
		Long: `Adds an execution result for the specified test ID.

Result statuses (standard):
	1 — Passed
	2 — Blocked
	3 — Untested
	4 — Retest
	5 — Failed

You can specify: comment, elapsed time, software version,
defects (comma-separated), and user assignment.

Examples:
	# Successfully passed test
	gotr result add 12345 --status-id 1 --comment "All checks passed"

	# Failed test with a defect
	gotr result add 12345 --status-id 5 --comment "Bug found" --defects "BUG-123"

	# With elapsed time and version
	gotr result add 12345 --status-id 1 --elapsed "2m 30s" --version "v2.0.1"

	# Reassign to another user
	gotr result add 12345 --status-id 2 --assigned-to 10 \\
		--comment "Need re-test by another engineer"

	# Dry-run mode
	gotr result add 12345 --status-id 1 --comment "Test" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newResultServiceFromInterface(cli)
			var testID int64
			var err error
			if len(args) > 0 {
				testID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid test ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("test_id required in non-interactive mode: gotr result add [test-id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("test_id required in non-interactive mode: gotr result add [test-id]")
				}
				runID, err := resolveResultRunID(ctx, cli)
				if err != nil {
					return err
				}
				testID, err = selectTestIDForRun(ctx, cli, runID)
				if err != nil {
					return err
				}
			}

			req, err := buildAddResultRequest(cmd)
			if err != nil {
				return err
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("result add")
				dr.PrintOperation(
					fmt.Sprintf("Add Result for Test %d", testID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_result/%d", testID),
					req,
				)
				return nil
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			result, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Adding result",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Result, error) {
				return svc.AddForTest(ctx, testID, req)
			})
			if err != nil {
				return fmt.Errorf("failed to add result: %w", err)
			}

			output.PrintSuccess(cmd, "Result added successfully:")
			return output.OutputResultWithFlags(cmd, result)
		},
	}

	cmd.Flags().Int64("status-id", 0, "Result status ID (required)")
	cmd.Flags().String("comment", "", "Comment for the result")
	cmd.Flags().String("version", "", "Software version")
	cmd.Flags().String("elapsed", "", "Elapsed time (e.g. '1m 30s')")
	cmd.Flags().String("defects", "", "Defect IDs (comma-separated)")
	cmd.Flags().Int64("assigned-to", 0, "User ID for assignment")
	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")
	_ = cmd.MarkFlagRequired("status-id")

	return cmd
}

// newAddCaseCmd creates the 'result add-case' command.
// Endpoint: POST /add_result_for_case/{run_id}/{case_id}
func newAddCaseCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-case [run-id]",
		Short: "Add a result for a case in a run",
		Long: `Adds an execution result for the specified case in a test run.

Unlike 'add': here run_id and case_id are specified instead of test_id.
TestRail automatically finds the corresponding test in the run.

Examples:
	# Add a result for case 98765 in run 12345
	gotr result add-case 12345 --case-id 98765 --status-id 1 \\
		--comment "Smoke test passed"

	# Specify a defect and elapsed time
	gotr result add-case 12345 --case-id 98765 --status-id 5 \\
		--defects "JIRA-456" --elapsed "5m"

	# Dry-run mode
	gotr result add-case 12345 --case-id 98765 --status-id 1 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newResultServiceFromInterface(cli)
			var runID int64
			var err error
			if len(args) > 0 {
				runID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid run ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("run_id required in non-interactive mode: gotr result add-case [run-id] --case-id <case_id>")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("run_id required in non-interactive mode: gotr result add-case [run-id] --case-id <case_id>")
				}
				runID, err = resolveResultRunID(ctx, cli)
				if err != nil {
					return err
				}
			}

			caseID, _ := cmd.Flags().GetInt64("case-id")
			req, err := buildAddResultRequest(cmd)
			if err != nil {
				return err
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("result add-case")
				dr.PrintOperation(
					fmt.Sprintf("Add Result for Case %d in Run %d", caseID, runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_result_for_case/%d/%d", runID, caseID),
					req,
				)
				return nil
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			result, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Adding result",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Result, error) {
				return svc.AddForCase(ctx, runID, caseID, req)
			})
			if err != nil {
				return fmt.Errorf("failed to add result: %w", err)
			}

			output.PrintSuccess(cmd, "Result added successfully:")
			return output.OutputResultWithFlags(cmd, result)
		},
	}

	cmd.Flags().Int64("case-id", 0, "Test case ID (required)")
	cmd.Flags().Int64("status-id", 0, "Result status ID (required)")
	cmd.Flags().String("comment", "", "Comment for the result")
	cmd.Flags().String("version", "", "Software version")
	cmd.Flags().String("elapsed", "", "Elapsed time")
	cmd.Flags().String("defects", "", "Defect IDs (comma-separated)")
	cmd.Flags().Int64("assigned-to", 0, "User ID for assignment")
	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")
	_ = cmd.MarkFlagRequired("case-id")
	_ = cmd.MarkFlagRequired("status-id")

	return cmd
}

// newAddBulkCmd creates the 'result add-bulk' command.
// Endpoint: POST /add_results/{run_id}
func newAddBulkCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-bulk [run-id]",
		Short: "Bulk add results",
		Long: `Adds multiple results in a single request.

The JSON file must contain an array of results:
[
  {
    "test_id": 12345,
    "status_id": 1,
    "comment": "Test passed"
  },
  {
    "case_id": 98765,
    "status_id": 5,
    "comment": "Test failed",
    "defects": "BUG-123"
  }
]

Both formats are supported: with test_id and with case_id.

Examples:
	# Dry-run mode
	gotr result add-bulk 12345 --results-file results.json --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newResultServiceFromInterface(cli)
			runID, err := svc.ParseID(ctx, args, 0)
			if err != nil {
				return fmt.Errorf("invalid run ID: %w", err)
			}

			resultsFile, _ := cmd.Flags().GetString("results-file")
			// add-bulk intentionally stays manual-only: input file is required for deterministic batch execution.
			fileData, err := os.ReadFile(resultsFile)
			if err != nil {
				return fmt.Errorf("file read error: %w", err)
			}

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("result add-bulk")
				dr.PrintOperation(
					fmt.Sprintf("Add Bulk Results for Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_results/%d", runID),
					string(fileData),
				)
				return nil
			}

			// Parse and submit results
			results, err := svc.AddBulkResults(ctx, runID, fileData)
			if err != nil {
				return err
			}

			output.PrintSuccess(cmd, "Results added successfully:")
			return output.OutputResultWithFlags(cmd, results)
		},
	}

	cmd.Flags().String("results-file", "", "JSON file with results (required)")
	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")
	_ = cmd.MarkFlagRequired("results-file")

	return cmd
}

// buildAddResultRequest builds the request payload from command flags.
func buildAddResultRequest(cmd *cobra.Command) (*data.AddResultRequest, error) {
	// Ensure status-id is provided (required parameter)
	if !cmd.Flags().Changed("status-id") {
		return nil, fmt.Errorf("--status-id is required (use: 1=Passed, 2=Blocked, 3=Untested, 4=Retest, 5=Failed)")
	}

	statusID, _ := cmd.Flags().GetInt64("status-id")
	comment, _ := cmd.Flags().GetString("comment")
	version, _ := cmd.Flags().GetString("version")
	elapsed, _ := cmd.Flags().GetString("elapsed")
	defects, _ := cmd.Flags().GetString("defects")
	assignedTo, _ := cmd.Flags().GetInt64("assigned-to")

	return &data.AddResultRequest{
		StatusID:   statusID,
		Comment:    comment,
		Version:    version,
		Elapsed:    elapsed,
		Defects:    defects,
		AssignedTo: assignedTo,
	}, nil
}

// Backward compatibility: exported vars for registration in result.go
var (
	addCmd     = newAddCmd(getClientSafe)
	addCaseCmd = newAddCaseCmd(getClientSafe)
	addBulkCmd = newAddBulkCmd(getClientSafe)
)
