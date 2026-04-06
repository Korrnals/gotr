package cases

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newAddCmd creates the 'cases add' command.
// Endpoint: POST /add_case/{section_id}
func newAddCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [section_id]",
		Short: "Create a new test case",
		Long:  `Creates a new test case in the specified section.`,
		Example: `  # Create a test case with parameters
  gotr cases add 100 --title="Auth test" --template-id=1

  # Create from a JSON file
  gotr cases add 100 --json-file=case.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var sectionID int64
			var err error
			if len(args) > 0 {
				sectionID, err = flags.ValidateRequiredID(args, 0, "section_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("section_id required: gotr cases add [section_id]")
				}

				sectionID, err = resolveSectionIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Check JSON file
			jsonFile, _ := cmd.Flags().GetString("json-file")
			var req data.AddCaseRequest

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
				if title == "" {
					return fmt.Errorf("--title is required (or use --json-file)")
				}
				req.Title = title
				req.TemplateID, _ = cmd.Flags().GetInt64("template-id")
				req.TypeID, _ = cmd.Flags().GetInt64("type-id")
				req.PriorityID, _ = cmd.Flags().GetInt64("priority-id")
				req.Refs, _ = cmd.Flags().GetString("refs")
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases add")
				dr.PrintSimple("Create Case", fmt.Sprintf("Section ID: %d, Title: %s", sectionID, req.Title))
				return nil
			}

			resp, err := cli.AddCase(ctx, sectionID, &req)
			if err != nil {
				return fmt.Errorf("failed to create case: %w", err)
			}

			ui.Successf(os.Stdout, "Case created (ID: %d)", resp.ID)
			return output.OutputResult(cmd, resp, "cases")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview the action without creating anything")
	output.AddFlag(cmd)
	cmd.Flags().String("json-file", "", "Path to a JSON file with case data")
	cmd.Flags().String("title", "", "Test case title")
	cmd.Flags().Int64("template-id", 0, "Template ID")
	cmd.Flags().Int64("type-id", 0, "Test type ID")
	cmd.Flags().Int64("priority-id", 0, "Priority ID")
	cmd.Flags().String("refs", "", "References")

	return cmd
}
