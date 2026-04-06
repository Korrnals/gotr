package cmd

import (
	"fmt"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// listCmd is the main subcommand: gotr list <resource>
var listCmd = &cobra.Command{
	Use:   "list <resource>",
	Short: "List available TestRail API endpoints for a resource",
	Long: `Lists available TestRail API v2 endpoints for a given resource.

Examples:
	gotr list projects          # endpoints for projects
	gotr list cases             # endpoints for cases
	gotr list all               # all endpoints
	gotr list cases --json      # as JSON
	gotr list cases --short     # short output (Method URI)`,

	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resource := ""
		if len(args) > 0 {
			resource = strings.ToLower(args[0])
		} else {
			if !interactive.HasPrompterInContext(cmd.Context()) {
				return fmt.Errorf("resource required: gotr list <resource>")
			}
			p := interactive.PrompterFromContext(cmd.Context())
			idx, _, err := p.Select("Select resource:", ValidResources)
			if err != nil {
				return fmt.Errorf("failed to select resource: %w", err)
			}
			resource = strings.ToLower(ValidResources[idx])
		}

		// Read flags declared below in init()
		jsonOutput, _ := cmd.Flags().GetBool("json")
		shortOutput, _ := cmd.Flags().GetBool("short")

		// JSON output
		if jsonOutput {
			getResourceEndpoints(resource, "json")
			return nil
		}
		// Short output (Method + URI)
		if shortOutput {
			getResourceEndpoints(resource, "short")
			return nil
		}
		// Full, formatted output
		getResourceEndpoints(resource, "")
		return nil
	},
}
