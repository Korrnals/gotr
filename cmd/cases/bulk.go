package cases

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func runBulkStatus[T any](cmd *cobra.Command, total int, fn func(context.Context) (T, error)) (T, error) {
	return ui.RunWithStatus(cmd.Context(), ui.StatusConfig{
		Title:  fmt.Sprintf("Processing %d cases...", total),
		Writer: os.Stderr,
	}, fn)
}

// newBulkCmd creates the parent 'cases bulk' command.
func newBulkCmd(getClient GetClientFunc) *cobra.Command {
	bulkCmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk operations on test cases",
		Long: `Bulk operations: update, delete, copy or move multiple test cases.

Subcommands:
  • update — bulk field update
  • delete — bulk deletion
  • copy   — copy to another section
  • move   — move to another section`,
	}

	bulkCmd.AddCommand(newBulkUpdateCmd(getClient))
	bulkCmd.AddCommand(newBulkDeleteCmd(getClient))
	bulkCmd.AddCommand(newBulkCopyCmd(getClient))
	bulkCmd.AddCommand(newBulkMoveCmd(getClient))

	return bulkCmd
}

// newBulkUpdateCmd creates the 'cases bulk update' command.
// Endpoint: POST /update_cases/{suite_id}
func newBulkUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <case_ids...>",
		Short: "Bulk update test cases",
		Long:  `Updates multiple test cases at once with the same field values.`,
		Example: `  # Update priority of several cases
  gotr cases bulk update 1,2,3 --suite-id=100 --priority-id=1

  # Update time estimate
  gotr cases bulk update 1 2 3 --suite-id=100 --estimate="1h 30m"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("--suite-id is required")
				}

				selectedSuiteID, err := resolveSuiteIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
				suiteID = selectedSuiteID
			}

			req := data.UpdateCasesRequest{CaseIDs: caseIDs}
			if v, _ := cmd.Flags().GetInt64("priority-id"); v > 0 {
				req.PriorityID = v
			}
			if v, _ := cmd.Flags().GetString("estimate"); v != "" {
				req.Estimate = v
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases bulk update")
				dr.PrintSimple("Bulk Update Cases", fmt.Sprintf("Suite: %d, Cases: %v", suiteID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			resp, err := runBulkStatus(cmd, len(caseIDs), func(ctx context.Context) (*data.GetCasesResponse, error) {
				return cli.UpdateCases(ctx, suiteID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to update cases: %w", err)
			}

			ui.Successf(os.Stdout, "Updated %d cases", len(caseIDs))
			return output.OutputResult(cmd, resp, "cases")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	output.AddFlag(cmd)
	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")
	cmd.Flags().Int64("priority-id", 0, "Priority ID to set")
	cmd.Flags().String("estimate", "", "Time estimate (e.g. '1h 30m')")

	return cmd
}

// newBulkDeleteCmd creates the 'cases bulk delete' command.
// Endpoint: POST /delete_cases/{suite_id}
func newBulkDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <case_ids...>",
		Short: "Bulk delete test cases",
		Long:  `Deletes multiple test cases at once.`,
		Example: `  # Delete several cases
  gotr cases bulk delete 1,2,3 --suite-id=100

  # Preview before deleting
  gotr cases bulk delete 1,2,3 --suite-id=100 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("--suite-id is required")
				}

				selectedSuiteID, err := resolveSuiteIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
				suiteID = selectedSuiteID
			}

			req := data.DeleteCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases bulk delete")
				dr.PrintSimple("Bulk Delete Cases", fmt.Sprintf("Suite: %d, Cases: %v", suiteID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			_, err := runBulkStatus(cmd, len(caseIDs), func(ctx context.Context) (struct{}, error) {
				return struct{}{}, cli.DeleteCases(ctx, suiteID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to delete cases: %w", err)
			}

			ui.Successf(os.Stdout, "Deleted %d cases", len(caseIDs))
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what will be deleted")
	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")

	return cmd
}

// newBulkCopyCmd creates the 'cases bulk copy' command.
// Endpoint: POST /copy_cases_to_section/{section_id}
func newBulkCopyCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy <case_ids...>",
		Short: "Copy cases to a section",
		Long:  `Copies multiple test cases to another section.`,
		Example: `  # Copy cases to another section
  gotr cases bulk copy 1,2,3 --section-id=50

  # Preview before copying
  gotr cases bulk copy 1,2,3 --section-id=50 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			sectionID, _ := cmd.Flags().GetInt64("section-id")
			if sectionID <= 0 {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("--section-id is required")
				}

				selectedSectionID, err := resolveSectionIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
				sectionID = selectedSectionID
			}

			req := data.CopyCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases bulk copy")
				dr.PrintSimple("Copy Cases", fmt.Sprintf("Section: %d, Cases: %v", sectionID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			_, err := runBulkStatus(cmd, len(caseIDs), func(ctx context.Context) (struct{}, error) {
				return struct{}{}, cli.CopyCasesToSection(ctx, sectionID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to copy cases: %w", err)
			}

			ui.Successf(os.Stdout, "Copied %d cases to section %d", len(caseIDs), sectionID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	cmd.Flags().Int64("section-id", 0, "Target section ID (required)")

	return cmd
}

// newBulkMoveCmd creates the 'cases bulk move' command.
// Endpoint: POST /move_cases_to_section/{section_id}
func newBulkMoveCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <case_ids...>",
		Short: "Move cases to a section",
		Long:  `Moves multiple test cases to another section.`,
		Example: `  # Move cases to another section
  gotr cases bulk move 1,2,3 --section-id=50

  # Preview before moving
  gotr cases bulk move 1,2,3 --section-id=50 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("case IDs required")
			}

			caseIDs := parseIDList(args)
			if len(caseIDs) == 0 {
				return fmt.Errorf("no valid case IDs provided")
			}

			sectionID, _ := cmd.Flags().GetInt64("section-id")
			if sectionID <= 0 {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("--section-id is required")
				}

				selectedSectionID, err := resolveSectionIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
				sectionID = selectedSectionID
			}

			req := data.MoveCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases bulk move")
				dr.PrintSimple("Move Cases", fmt.Sprintf("Section: %d, Cases: %v", sectionID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			_, err := runBulkStatus(cmd, len(caseIDs), func(ctx context.Context) (struct{}, error) {
				return struct{}{}, cli.MoveCasesToSection(ctx, sectionID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to move cases: %w", err)
			}

			ui.Successf(os.Stdout, "Moved %d cases to section %d", len(caseIDs), sectionID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	cmd.Flags().Int64("section-id", 0, "Target section ID (required)")

	return cmd
}

// parseIDList parses IDs separated by commas or spaces.
func parseIDList(args []string) []int64 {
	var ids []int64
	for _, arg := range args {
		for _, part := range strings.Split(arg, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := flags.ParseID(part)
			if err == nil && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}
