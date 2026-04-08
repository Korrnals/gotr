package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/ui"

	"github.com/spf13/cobra"
)

var casesCmd = &cobra.Command{
	Use:   "cases",
	Short: "Synchronize test cases between suites",
	Long: `Complete procedure for transferring test cases from one suite to another.

Features:
• Automatic interactive selection of projects and suites (if flags are not specified)
• Support for shared_step_id replacement via mapping file
• Interactive confirmation before import
• Dry-run mode (without creating objects)
• Saving JSON result log

If project/suite IDs are not specified, you will be prompted to select them from a list.

Examples:
	# Fully interactive mode (select all parameters)
	gotr sync cases

	# Partially interactive (only source project specified)
	gotr sync cases --src-project 30

	# Fully via flags
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859

	# With mapping file and dry-run
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping-file mapping.json --dry-run
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cli := getClientInterface(cmd)
		ctx := cmd.Context()

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")
		outputFile, _ := cmd.Flags().GetString("output")
		mappingFile, _ := cmd.Flags().GetString("mapping-file")

		p := interactive.PrompterFromContext(ctx)
		var err error

		// Interactive source project selection
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project (copy from):")
			if err != nil {
				return err
			}
		}

		// Interactive source suite selection
		if srcSuite == 0 {
			srcSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, srcProject, "Select SOURCE suite:")
			if err != nil {
				return err
			}
		}

		// Interactive destination project selection
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project (copy to):")
			if err != nil {
				return err
			}
		}

		// Interactive destination suite selection
		if dstSuite == 0 {
			dstSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, dstProject, "Select DESTINATION suite:")
			if err != nil {
				return err
			}
		}

		// Log directory
		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return err
		}
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFile := filepath.Join(logDir, fmt.Sprintf("sync_cases_%s.json", timestamp))
		// Use the additional output file if specified
		if outputFile != "" {
			logFile = outputFile
		}

		// Create migration object
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		op := newSyncOperation("Sync cases", quiet)
		defer op.Finish()

		// If a mapping file is specified, load it into m.mapping
		if mappingFile != "" {
			op.Phase("Loading mapping")
			_, err := runSyncStatus(ctx, "Loading mapping...", quiet, func(context.Context) (struct{}, error) {
				return struct{}{}, m.LoadMappingFromFile(mappingFile)
			})
			if err != nil {
				return fmt.Errorf("failed to load mapping: %w", err)
			}
			ui.Infof(os.Stdout, "Mapping loaded: %d entries", len(m.Mapping()))
		} else {
			ui.Warning(os.Stdout, "mapping not loaded — shared_step_id will NOT be replaced")
		}

		op.Phase("Loading cases")
		loaded, err := runSyncStatus(ctx, "Loading cases...", quiet, func(ctx context.Context) (struct {
			Source data.GetCasesResponse
			Target data.GetCasesResponse
		}, error) {
			sourceCases, targetCases, err := m.FetchCasesData(ctx)
			if err != nil {
				return struct {
					Source data.GetCasesResponse
					Target data.GetCasesResponse
				}{}, err
			}
			return struct {
				Source data.GetCasesResponse
				Target data.GetCasesResponse
			}{Source: sourceCases, Target: targetCases}, nil
		})
		if err != nil {
			return err
		}
		sourceCases := loaded.Source
		targetCases := loaded.Target

		filtered, err := m.FilterCases(sourceCases, targetCases)
		if err != nil {
			return err
		}

		// Count matches
		var matches data.GetCasesResponse
		filteredIDs := make(map[int64]struct{})
		for _, f := range filtered {
			filteredIDs[f.ID] = struct{}{}
		}
		for _, s := range sourceCases {
			if _, ok := filteredIDs[s.ID]; !ok {
				matches = append(matches, s)
			}
		}

		if !quiet {
			if !quiet {
				fmt.Printf("\nAnalysis result:\n")
				fmt.Printf("  Matches: %d\n", len(matches))
				fmt.Printf("  New: %d\n", len(filtered))
			}
		}

		if dryRun {
			ui.Info(os.Stdout, "Dry-run: import NOT performed (safe).")
			saveLog(logFile, matches, filtered, nil, m.Mapping(), quiet)
			return nil
		}

		op.Phase("Awaiting confirmation")
		ui.Infof(os.Stdout, "Confirm import of %d new cases...", len(filtered))
		ok, err := p.Confirm("Continue?", false)
		if err != nil {
			return err
		}
		if !ok {
			ui.Canceled(os.Stdout)
			saveLog(logFile, matches, filtered, nil, m.Mapping(), quiet)
			return nil
		}

		op.Phase("Importing cases")
		imported, err := runSyncStatus(ctx, fmt.Sprintf("Importing %d cases...", len(filtered)), quiet, func(ctx context.Context) (struct {
			IDs    []int64
			Errors []string
		}, error) {
			createdIDs, importErrors, err := m.ImportCasesReport(ctx, filtered, false)
			if err != nil {
				return struct {
					IDs    []int64
					Errors []string
				}{}, err
			}
			return struct {
				IDs    []int64
				Errors []string
			}{IDs: createdIDs, Errors: importErrors}, nil
		})
		if err != nil {
			return err
		}
		createdIDs := imported.IDs
		importErrors := imported.Errors

		ui.Successf(os.Stdout, "Import complete: %d new cases", len(createdIDs))

		if len(importErrors) > 0 {
			ui.Error(os.Stdout, "Errors:")
			for _, e := range importErrors {
				fmt.Printf("  - %s\n", e)
			}
		}

		// Save log and mapping
		saveLog(logFile, matches, filtered, importErrors, m.Mapping(), quiet)

		return nil
	},
}

func saveLog(file string, matches, filtered data.GetCasesResponse, errors []string, mapping map[int64]int64, quiet bool) {
	result := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"matches":   len(matches),
		"filtered":  len(filtered),
		"errors":    errors,
		"mapping":   mapping,
	}
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	if err := os.WriteFile(file, jsonData, 0o644); err != nil {
		ui.Warningf(os.Stderr, "Failed to save log %s: %v", file, err)
		return
	}
	if !quiet {
		ui.Infof(os.Stdout, "Log saved: %s", file)
	}
}
