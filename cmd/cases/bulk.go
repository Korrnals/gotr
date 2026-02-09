package cases

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newBulkCmd creates 'cases bulk' parent command
func newBulkCmd(getClient GetClientFunc) *cobra.Command {
	bulkCmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk operations on test cases",
		Long:  `Bulk update, delete, copy, or move multiple test cases at once.`,
	}

	bulkCmd.AddCommand(newBulkUpdateCmd(getClient))
	bulkCmd.AddCommand(newBulkDeleteCmd(getClient))
	bulkCmd.AddCommand(newBulkCopyCmd(getClient))
	bulkCmd.AddCommand(newBulkMoveCmd(getClient))

	return bulkCmd
}

// newBulkUpdateCmd creates 'cases bulk update' command
func newBulkUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <case_ids...>",
		Short: "Bulk update cases",
		Long:  `Update multiple test cases at once with the same field values.`,
		Example: `  gotr cases bulk update 1,2,3 --suite-id=100 --priority-id=1
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
				return fmt.Errorf("--suite-id is required")
			}

			req := data.UpdateCasesRequest{CaseIDs: caseIDs}
			if v, _ := cmd.Flags().GetInt64("priority-id"); v > 0 {
				req.PriorityID = v
			}
			if v, _ := cmd.Flags().GetString("estimate"); v != "" {
				req.Estimate = v
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk update")
				dr.PrintSimple("Bulk Update Cases", fmt.Sprintf("Suite: %d, Cases: %v", suiteID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateCases(suiteID, &req)
			if err != nil {
				return fmt.Errorf("failed to update cases: %w", err)
			}

			fmt.Printf("✅ Updated %d cases\n", len(caseIDs))
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().StringP("output", "o", "", "Save response to file")
	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")
	cmd.Flags().Int64("priority-id", 0, "Priority ID to set")
	cmd.Flags().String("estimate", "", "Estimate to set (e.g., '1h 30m')")

	return cmd
}

// newBulkDeleteCmd creates 'cases bulk delete' command
func newBulkDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <case_ids...>",
		Short: "Bulk delete cases",
		Long:  `Delete multiple test cases at once.`,
		Example: `  gotr cases bulk delete 1,2,3 --suite-id=100`,
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
				return fmt.Errorf("--suite-id is required")
			}

			req := data.DeleteCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk delete")
				dr.PrintSimple("Bulk Delete Cases", fmt.Sprintf("Suite: %d, Cases: %v", suiteID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteCases(suiteID, &req); err != nil {
				return fmt.Errorf("failed to delete cases: %w", err)
			}

			fmt.Printf("✅ Deleted %d cases\n", len(caseIDs))
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")

	return cmd
}

// newBulkCopyCmd creates 'cases bulk copy' command
func newBulkCopyCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy <case_ids...>",
		Short: "Copy cases to section",
		Long:  `Copy multiple test cases to another section.`,
		Example: `  gotr cases bulk copy 1,2,3 --section-id=50`,
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
				return fmt.Errorf("--section-id is required")
			}

			req := data.CopyCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk copy")
				dr.PrintSimple("Copy Cases", fmt.Sprintf("Section: %d, Cases: %v", sectionID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.CopyCasesToSection(sectionID, &req); err != nil {
				return fmt.Errorf("failed to copy cases: %w", err)
			}

			fmt.Printf("✅ Copied %d cases to section %d\n", len(caseIDs), sectionID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().Int64("section-id", 0, "Target section ID (required)")

	return cmd
}

// newBulkMoveCmd creates 'cases bulk move' command
func newBulkMoveCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <case_ids...>",
		Short: "Move cases to section",
		Long:  `Move multiple test cases to another section.`,
		Example: `  gotr cases bulk move 1,2,3 --section-id=50`,
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
				return fmt.Errorf("--section-id is required")
			}

			req := data.MoveCasesRequest{CaseIDs: caseIDs}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases bulk move")
				dr.PrintSimple("Move Cases", fmt.Sprintf("Section: %d, Cases: %v", sectionID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.MoveCasesToSection(sectionID, &req); err != nil {
				return fmt.Errorf("failed to move cases: %w", err)
			}

			fmt.Printf("✅ Moved %d cases to section %d\n", len(caseIDs), sectionID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().Int64("section-id", 0, "Target section ID (required)")

	return cmd
}

// parseIDList parses comma-separated or space-separated IDs
func parseIDList(args []string) []int64 {
	var ids []int64
	for _, arg := range args {
		for _, part := range strings.Split(arg, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseInt(part, 10, 64)
			if err == nil && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// outputResult outputs result as JSON or to file
func outputResult(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if output != "" {
		return os.WriteFile(output, jsonBytes, 0644)
	}

	fmt.Println(string(jsonBytes))
	return nil
}
