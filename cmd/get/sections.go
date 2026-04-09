package get

import (
	"context"
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newSectionsCmd creates the parent command for section operations.
func newSectionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sections",
		Short: "Section management",
		Long: `Section management (sections) — organizational units
within test suites for grouping test cases.

Sections allow structuring tests into a hierarchy for convenient navigation
and organization of test documentation.

Available operations:
  • get    — get section information by ID
  • list   — list project/suite sections`,
	}
}

// newSectionGetCmd creates the command for retrieving a single section.
func newSectionGetCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "section [section_id]",
		Short: "Get section information",
		Long:  `Get detailed section information by its ID.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			var sectionID int64
			var err error
			if len(args) > 0 {
				sectionID, err = flags.ValidateRequiredID(args, 0, "section_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("section_id required: gotr get section [section_id]")
				}

				projectID, err := interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}

				sections, err := cli.GetSections(ctx, projectID, 0)
				if err != nil {
					return fmt.Errorf("failed to get sections list: %w", err)
				}
				if len(sections) == 0 {
					return fmt.Errorf("no sections found in project %d", projectID)
				}

				sectionID, err = selectSectionID(ctx, sections)
				if err != nil {
					return err
				}
			}

			section, err := runGetStatus(command, "Loading section...", func(ctx context.Context) (any, error) {
				return cli.GetSection(ctx, sectionID)
			})
			if err != nil {
				return err
			}

			return handleOutput(command, section, start)
		},
	}
}

func selectSectionID(ctx context.Context, sections data.GetSectionsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(sections))
	for i, section := range sections {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, section.ID, section.Name))
	}

	idx, _, err := p.Select("Select section:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select section: %w", err)
	}

	return sections[idx].ID, nil
}

// newSectionsListCmd creates the command for listing project sections.
func newSectionsListCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Get project sections list",
		Long: `Get a list of all sections for the specified project.

To filter by a specific suite, use the --suite-id flag.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			ctx := command.Context()
			projectIDStr := ""
			if len(args) > 0 {
				projectIDStr = args[0]
			}
			if pid, _ := command.Flags().GetString("project-id"); pid != "" {
				projectIDStr = pid
			}

			var projectID int64
			var err error
			if projectIDStr == "" {
				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			} else {
				projectID, err = flags.ParseID(projectIDStr)
				if err != nil {
					return fmt.Errorf("invalid project_id: %w", err)
				}
				if projectID <= 0 {
					return fmt.Errorf("invalid project_id: %s", projectIDStr)
				}
			}

			suiteID, _ := command.Flags().GetInt64("suite-id")

			sections, err := runGetStatus(command, "Loading sections...", func(ctx context.Context) (any, error) {
				return cli.GetSections(ctx, projectID, suiteID)
			})
			if err != nil {
				return err
			}

			return handleOutput(command, sections, start)
		},
	}

	// Filter by suite_id
	cmd.Flags().Int64P("suite-id", "s", 0, "Suite ID for filtering")
	cmd.Flags().String("project-id", "", "Project ID (alternative to positional argument)")

	return cmd
}

// sectionsCmd is the exported parent command.
var sectionsCmd = newSectionsCmd()

// sectionGetCmd is the exported command registered with the root.
var sectionGetCmd = newSectionGetCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// sectionsListCmd is the exported command registered with the root.
var sectionsListCmd = newSectionsListCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

func init() {
	// Register subcommands under the parent
	sectionsCmd.AddCommand(sectionGetCmd)
	sectionsCmd.AddCommand(sectionsListCmd)
}
