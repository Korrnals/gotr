package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/debug"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// exportCmd exports data from TestRail.
var exportCmd = &cobra.Command{
	Use:   "export <resource> <endpoint> [id]",
	Short: "Export data from TestRail to a JSON file",
	Long: `Exports data from TestRail to a JSON file.

Output file name:
    • Via --output (-o) flag: gotr export cases get_cases 30 -o my_cases.json
    • Without flag: saved to .testrail directory as <resource>_[id]_<timestamp>.json

Examples:
    gotr export projects get_projects
    gotr export cases get_cases 1 --suite-id 5 -o cases_suite5.json`,

	Args: cobra.MaximumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		resource, endpoint, mainID, err := resolveExportInputs(cmd, args)
		if err != nil {
			return err
		}

		client := GetClient(cmd)

		// Build full endpoint path and query parameters
		fullEndpoint, queryParams, err := buildRequestParams(endpoint, mainID, cmd)
		if err != nil {
			return err
		}

		debug.DebugPrint("{exportCmd} - Final endpoint: %s", fullEndpoint)
		debug.DebugPrint("{exportCmd} - Query params: %v", queryParams)

		// Request
		start := time.Now()
		resp, err := client.Get(ctx, fullEndpoint, queryParams)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		data, err := client.ReadResponse(ctx, resp, time.Since(start), "json")
		if err != nil {
			return fmt.Errorf("response reading error: %w", err)
		}

		// Flags
		quiet, _ := cmd.Flags().GetBool("quiet")
		saveFlag, _ := cmd.Flags().GetBool("save")

		if saveFlag {
			// Save via output.Output to ~/.gotr/exports/export/
			filepath, err := output.Output(cmd, data, "export", "json")
			if err != nil {
				return fmt.Errorf("save error: %w", err)
			}
			if !quiet && filepath != "" {
				ui.Infof(os.Stdout, "Data exported to %s", filepath)
			}
		} else {
			// Save to .testrail/ (legacy behavior)
			exportDir := ".testrail"
			if err := os.MkdirAll(exportDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", exportDir, err)
			}
			filename := fmt.Sprintf("%s/%s_%s.json", exportDir, resource, time.Now().Format("20060102_150405"))
			if mainID != "" {
				filename = fmt.Sprintf("%s/%s_%s_%s.json", exportDir, resource, mainID, time.Now().Format("20060102_150405"))
			}
			if err := client.SaveResponseToFile(ctx, data, filename, "json"); err != nil {
				return fmt.Errorf("file export error %s: %w", filename, err)
			}
			if !quiet {
				ui.Infof(os.Stdout, "Data exported to %s", filename)
			}
		}

		return nil
	},
}

func resolveExportInputs(cmd *cobra.Command, args []string) (string, string, string, error) {
	ctx := cmd.Context()
	p := interactive.PrompterFromContext(ctx)

	resource := ""
	if len(args) > 0 {
		resource = strings.ToLower(args[0])
	} else {
		if !interactive.HasPrompterInContext(ctx) {
			return "", "", "", fmt.Errorf("resource required: gotr export <resource> <endpoint> [id]")
		}
		idx, _, err := p.Select("Select export resource:", ValidResources)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to select resource: %w", err)
		}
		resource = strings.ToLower(ValidResources[idx])
	}

	endpoint := ""
	if len(args) > 1 {
		endpoint = args[1]
	} else {
		if !interactive.HasPrompterInContext(ctx) {
			return "", "", "", fmt.Errorf("endpoint required: gotr export <resource> <endpoint> [id]")
		}
		endpointOptions, endpointsErr := getResourceEndpoints(resource, "list")
		if endpointsErr != nil {
			return "", "", "", fmt.Errorf("failed to get endpoints for %s: %w", resource, endpointsErr)
		}
		endpointOptions = filterEmpty(endpointOptions)
		if len(endpointOptions) == 0 {
			return "", "", "", fmt.Errorf("no export endpoints found for resource: %s", resource)
		}
		idx, _, err := p.Select("Select export endpoint:", endpointOptions)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to select endpoint: %w", err)
		}
		endpoint = endpointOptions[idx]
	}

	mainID := ""
	if pid, _ := cmd.Flags().GetString("project-id"); pid != "" {
		mainID = pid
	} else if len(args) > 2 {
		mainID = args[2]
	}

	if mainID == "" && strings.Contains(endpoint, "{") {
		if !interactive.HasPrompterInContext(ctx) {
			return "", "", "", fmt.Errorf("id required for endpoint: %s", endpoint)
		}
		input, err := p.Input("Enter main ID:", "")
		if err != nil {
			return "", "", "", fmt.Errorf("failed to read id: %w", err)
		}
		mainID = strings.TrimSpace(input)
	}

	return resource, endpoint, mainID, nil
}

func filterEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}
