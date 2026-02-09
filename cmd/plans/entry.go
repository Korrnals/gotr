package plans

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newEntryCmd creates 'plans entry' parent command
func newEntryCmd(getClient GetClientFunc) *cobra.Command {
	entryCmd := &cobra.Command{
		Use:   "entry",
		Short: "Manage plan entries",
		Long:  `Add, update, or delete entries within a test plan.`,
	}

	entryCmd.AddCommand(newEntryAddCmd(getClient))
	entryCmd.AddCommand(newEntryUpdateCmd(getClient))
	entryCmd.AddCommand(newEntryDeleteCmd(getClient))

	return entryCmd
}

// newEntryAddCmd creates 'plans entry add' command
func newEntryAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <plan_id>",
		Short: "Add entry to plan",
		Long:  `Add a new entry (test run) to an existing plan.`,
		Example: `  gotr plans entry add 100 --suite-id=50 --name="Entry 1"
  gotr plans entry add 100 --suite-id=50 --config-ids="1,2,3"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			if suiteID <= 0 {
				return fmt.Errorf("--suite-id is required")
			}

			req := data.AddPlanEntryRequest{
				SuiteID: suiteID,
			}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}
			if v, _ := cmd.Flags().GetString("config-ids"); v != "" {
				req.ConfigIDs = parseIntList(v)
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans entry add")
				dr.PrintSimple("Add Plan Entry", fmt.Sprintf("Plan ID: %d, Suite ID: %d", planID, suiteID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddPlanEntry(planID, &req)
			if err != nil {
				return fmt.Errorf("failed to add plan entry: %w", err)
			}

			fmt.Printf("✅ Entry added to plan %d\n", planID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().StringP("output", "o", "", "Save response to file")
	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")
	cmd.Flags().String("name", "", "Entry name")
	cmd.Flags().String("config-ids", "", "Comma-separated config IDs")

	return cmd
}

// newEntryUpdateCmd creates 'plans entry update' command
func newEntryUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <plan_id> <entry_id>",
		Short: "Update plan entry",
		Long:  `Update an existing plan entry.`,
		Example: `  gotr plans entry update 100 abc123 --name="Updated Entry"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			entryID := args[1]
			if entryID == "" {
				return fmt.Errorf("entry_id is required")
			}

			req := data.UpdatePlanEntryRequest{}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans entry update")
				dr.PrintSimple("Update Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdatePlanEntry(planID, entryID, &req)
			if err != nil {
				return fmt.Errorf("failed to update plan entry: %w", err)
			}

			fmt.Printf("✅ Entry %s updated in plan %d\n", entryID, planID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().StringP("output", "o", "", "Save response to file")
	cmd.Flags().String("name", "", "New entry name")

	return cmd
}

// newEntryDeleteCmd creates 'plans entry delete' command
func newEntryDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <plan_id> <entry_id>",
		Short: "Delete plan entry",
		Long:  `Delete an entry from a test plan.`,
		Example: `  gotr plans entry delete 100 abc123`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			entryID := args[1]
			if entryID == "" {
				return fmt.Errorf("entry_id is required")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans entry delete")
				dr.PrintSimple("Delete Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeletePlanEntry(planID, entryID); err != nil {
				return fmt.Errorf("failed to delete plan entry: %w", err)
			}

			fmt.Printf("✅ Entry %s deleted from plan %d\n", entryID, planID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")

	return cmd
}

// parseIntList parses comma-separated integers
func parseIntList(s string) []int64 {
	var ids []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
