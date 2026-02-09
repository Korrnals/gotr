package cases

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates 'cases update' command
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <case_id>",
		Short: "Update a test case",
		Long:  `Update an existing test case.`,
		Example: `  gotr cases update 12345 --title="Updated Title" --priority-id=2
  gotr cases update 12345 --json-file=update.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || caseID <= 0 {
				return fmt.Errorf("invalid case_id: %s", args[0])
			}

			jsonFile, _ := cmd.Flags().GetString("json-file")
			var req data.UpdateCaseRequest

			if jsonFile != "" {
				jsonData, err := os.ReadFile(jsonFile)
				if err != nil {
					return fmt.Errorf("error reading JSON file: %w", err)
				}
				if err := json.Unmarshal(jsonData, &req); err != nil {
					return fmt.Errorf("error parsing JSON: %w", err)
				}
			} else {
				title, _ := cmd.Flags().GetString("title")
				if title != "" {
					req.Title = &title
				}
				if v, _ := cmd.Flags().GetInt64("type-id"); v > 0 {
					req.TypeID = &v
				}
				if v, _ := cmd.Flags().GetInt64("priority-id"); v > 0 {
					req.PriorityID = &v
				}
				if v, _ := cmd.Flags().GetString("refs"); v != "" {
					req.Refs = &v
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("cases update")
				dr.PrintSimple("Update Case", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateCase(caseID, &req)
			if err != nil {
				return fmt.Errorf("failed to update case: %w", err)
			}

			fmt.Printf("✅ Case %d updated\n", caseID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().StringP("output", "o", "", "Save response to file")
	cmd.Flags().String("json-file", "", "Path to JSON file with update data")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().Int64("type-id", 0, "New type ID")
	cmd.Flags().Int64("priority-id", 0, "New priority ID")
	cmd.Flags().String("refs", "", "New references")

	return cmd
}
