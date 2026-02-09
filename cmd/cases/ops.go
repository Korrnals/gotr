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

// newUpdateCmd creates 'cases update' command for bulk update
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <case_ids...>",
		Short: "Bulk update cases",
		Long:  `Update multiple test cases at once with the same field values.`,
		Example: `  gotr cases update 1,2,3 --suite-id=100 --priority-id=1
  gotr cases update 1 2 3 --suite-id=100 --estimate="1h 30m"`,
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

			// Build update request
			req := data.UpdateCasesRequest{
				CaseIDs: caseIDs,
			}

			if v, _ := cmd.Flags().GetInt64("priority-id"); v > 0 {
				req.PriorityID = v
			}
			if v, _ := cmd.Flags().GetString("estimate"); v != "" {
				req.Estimate = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases update")
				dr.PrintSimple("Bulk Update Cases", fmt.Sprintf("Suite ID: %d, Cases: %v, Priority ID: %d, Estimate: %s",
					suiteID, caseIDs, req.PriorityID, req.Estimate))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateCases(suiteID, &req)
			if err != nil {
				return fmt.Errorf("failed to update cases: %w", err)
			}

			fmt.Printf("✅ Updated %d cases in suite %d\n", len(caseIDs), suiteID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")
	cmd.Flags().Int64("priority-id", 0, "Priority ID to set")
	cmd.Flags().String("estimate", "", "Estimate to set (e.g., '1h 30m')")

	cmd.MarkFlagRequired("suite-id")

	return cmd
}

// newDeleteCmd creates 'cases delete' command for bulk delete
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <case_ids...>",
		Short: "Bulk delete cases",
		Long:  `Delete multiple test cases at once.`,
		Example: `  gotr cases delete 1,2,3 --suite-id=100
  gotr cases delete 10 20 30 --suite-id=100`,
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

			req := data.DeleteCasesRequest{
				CaseIDs: caseIDs,
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases delete")
				dr.PrintSimple("Bulk Delete Cases", fmt.Sprintf("Suite ID: %d, Cases: %v", suiteID, caseIDs))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeleteCases(suiteID, &req); err != nil {
				return fmt.Errorf("failed to delete cases: %w", err)
			}

			fmt.Printf("✅ Deleted %d cases from suite %d\n", len(caseIDs), suiteID)
			return nil
		},
	}

	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")
	cmd.MarkFlagRequired("suite-id")

	return cmd
}

// newCopyCmd creates 'cases copy' command
func newCopyCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy <case_ids...>",
		Short: "Copy cases to section",
		Long:  `Copy test cases to another section.`,
		Example: `  gotr cases copy 1,2,3 --section-id=50
  gotr cases copy 10 20 --section-id=100`,
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

			req := data.CopyCasesRequest{
				CaseIDs: caseIDs,
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases copy")
				dr.PrintSimple("Copy Cases", fmt.Sprintf("Section ID: %d, Cases: %v", sectionID, caseIDs))
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

	cmd.Flags().Int64("section-id", 0, "Target section ID (required)")
	cmd.MarkFlagRequired("section-id")

	return cmd
}

// newMoveCmd creates 'cases move' command
func newMoveCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <case_ids...>",
		Short: "Move cases to section",
		Long:  `Move test cases to another section.`,
		Example: `  gotr cases move 1,2,3 --section-id=50
  gotr cases move 10 20 --section-id=100`,
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

			req := data.MoveCasesRequest{
				CaseIDs: caseIDs,
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases move")
				dr.PrintSimple("Move Cases", fmt.Sprintf("Section ID: %d, Cases: %v", sectionID, caseIDs))
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

	cmd.Flags().Int64("section-id", 0, "Target section ID (required)")
	cmd.MarkFlagRequired("section-id")

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
