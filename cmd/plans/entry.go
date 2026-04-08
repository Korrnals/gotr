package plans

import (
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

// newEntryCmd creates the parent 'plans entry' command.
// Groups subcommands for managing plan entries.
func newEntryCmd(getClient GetClientFunc) *cobra.Command {
	entryCmd := &cobra.Command{
		Use:   "entry",
		Short: "Manage plan entries",
		Long: `Manage test plan entries — test runs within a plan.

Subcommands:
  • add    — add an entry (test run) to a plan
  • update — update an existing entry
  • delete — delete an entry from a plan`,
	}

	entryCmd.AddCommand(newEntryAddCmd(getClient))
	entryCmd.AddCommand(newEntryUpdateCmd(getClient))
	entryCmd.AddCommand(newEntryDeleteCmd(getClient))

	return entryCmd
}

// newEntryAddCmd creates the 'plans entry add' command.
// Endpoint: POST /add_plan_entry/{plan_id}
func newEntryAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [plan_id]",
		Short: "Add an entry to a plan",
		Long:  `Adds a new entry (test run) to an existing plan.`,
		Example: `  # Add a run with a name
  gotr plans entry add 100 --suite-id=50 --name="Run 1"

  # Add with configurations
  gotr plans entry add 100 --suite-id=50 --config-ids="1,2,3"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planID int64
			if len(args) > 0 {
				var err error
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans entry add [plan_id]")
				}
				var err error
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
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
				dr := output.NewDryRunPrinter("plans entry add")
				dr.PrintSimple("Add Plan Entry", fmt.Sprintf("Plan ID: %d, Suite ID: %d", planID, suiteID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.AddPlanEntry(ctx, planID, &req)
			if err != nil {
				return fmt.Errorf("failed to add plan entry: %w", err)
			}

			ui.Successf(os.Stdout, "Entry added to plan %d", planID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without adding")
	output.AddFlag(cmd)
	cmd.Flags().Int64("suite-id", 0, "Suite ID (required)")
	cmd.Flags().String("name", "", "Entry name")
	cmd.Flags().String("config-ids", "", "Comma-separated configuration IDs")

	return cmd
}

// newEntryUpdateCmd creates the 'plans entry update' command.
// Endpoint: POST /update_plan_entry/{plan_id}/{entry_id}
func newEntryUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [plan_id] [entry_id]",
		Short: "Update a plan entry",
		Long:  `Updates an existing entry in a test plan.`,
		Example: `  # Change entry name
  gotr plans entry update 100 abc123 --name="Updated entry"`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planID int64
			var err error
			if len(args) > 0 {
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans entry update [plan_id] [entry_id]")
				}
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			var entryID string
			if len(args) > 1 {
				entryID = args[1]
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("entry_id is required in non-interactive mode: gotr plans entry update [plan_id] [entry_id]")
				}
				entryID, err = resolvePlanEntryIDInteractive(cmd.Context(), getClient(cmd), planID)
				if err != nil {
					return err
				}
			}
			if entryID == "" {
				return fmt.Errorf("entry_id is required")
			}

			req := data.UpdatePlanEntryRequest{}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans entry update")
				dr.PrintSimple("Update Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.UpdatePlanEntry(ctx, planID, entryID, &req)
			if err != nil {
				return fmt.Errorf("failed to update plan entry: %w", err)
			}

			ui.Successf(os.Stdout, "Entry %s updated in plan %d", entryID, planID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "New entry name")

	return cmd
}

// newEntryDeleteCmd creates the 'plans entry delete' command.
// Endpoint: POST /delete_plan_entry/{plan_id}/{entry_id}
func newEntryDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [plan_id] [entry_id]",
		Short: "Delete a plan entry",
		Long:  `Deletes an entry from a test plan.`,
		Example: `  # Delete an entry from a plan
  gotr plans entry delete 100 abc123

  # Preview before deleting
  gotr plans entry delete 100 abc123 --dry-run`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planID int64
			var err error
			if len(args) > 0 {
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans entry delete [plan_id] [entry_id]")
				}
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			var entryID string
			if len(args) > 1 {
				entryID = args[1]
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("entry_id is required in non-interactive mode: gotr plans entry delete [plan_id] [entry_id]")
				}
				entryID, err = resolvePlanEntryIDInteractive(cmd.Context(), getClient(cmd), planID)
				if err != nil {
					return err
				}
			}
			if entryID == "" {
				return fmt.Errorf("entry_id is required")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans entry delete")
				dr.PrintSimple("Delete Plan Entry", fmt.Sprintf("Plan ID: %d, Entry ID: %s", planID, entryID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeletePlanEntry(ctx, planID, entryID); err != nil {
				return fmt.Errorf("failed to delete plan entry: %w", err)
			}

			ui.Successf(os.Stdout, "Entry %s deleted from plan %d", entryID, planID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be deleted")

	return cmd
}

// parseIntList splits a comma-separated string into a list of int64 IDs.
func parseIntList(s string) []int64 {
	var ids []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := flags.ParseID(part)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
