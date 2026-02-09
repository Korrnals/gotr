package plans

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newAddCmd creates 'plans add' command
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <project_id>",
		Short: "Create a new test plan",
		Long:  `Create a new test plan in the specified project.`,
		Example: `  gotr plans add 1 --name="Sprint 1 Plan"
  gotr plans add 1 --name="Regression" --description="Full regression suite"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			req := data.AddPlanRequest{
				Name: name,
			}

			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetInt64("milestone-id"); v > 0 {
				req.MilestoneID = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans add")
				dr.PrintSimple("Create Plan", fmt.Sprintf("Project ID: %d, Name: %s", projectID, req.Name))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.AddPlan(projectID, &req)
			if err != nil {
				return fmt.Errorf("failed to create plan: %w", err)
			}

			fmt.Printf("✅ Plan created (ID: %d)\n", resp.ID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().StringP("output", "o", "", "Save response to file")
	cmd.Flags().String("name", "", "Plan name (required)")
	cmd.Flags().String("description", "", "Plan description")
	cmd.Flags().Int64("milestone-id", 0, "Milestone ID")

	return cmd
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
