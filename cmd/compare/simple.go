package compare

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newSimpleCompareCmd creates a generic compare subcommand that loads resources
// from two projects in parallel and compares them.
//
// This factory replaces 9 nearly identical command files (groups, labels,
// milestones, plans, runs, sharedsteps, templates, configurations, datasets).
func newSimpleCompareCmd(resource, use, short, long string, fetchFn FetchFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClientSafe(cmd)
			ctx := cmd.Context()
			quiet, _ := cmd.Flags().GetBool("quiet")
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			pid1, pid2, format, savePath, err := parseCommonFlags(cmd, cli)
			if err != nil {
				return err
			}

			project1Name, project2Name, err := GetProjectNames(ctx, cli, pid1, pid2)
			if err != nil {
				return err
			}

			startTime := time.Now()

			result, err := compareSimpleInternal(ctx, cli, pid1, pid2, resource, fetchFn, quiet)
			if err != nil {
				return fmt.Errorf("comparison error %s: %w", resource, err)
			}

			elapsed := time.Since(startTime)

			// In interactive mode with table format and no save — skip
			// dumping the detail table. The user can view it via
			// "📋 View detailed results" in the post-action menu.
			isInteractiveTable := format == "table" && savePath == "" && !quiet &&
				interactive.HasPrompterInContext(ctx) && !interactive.IsNonInteractive(ctx)

			if !isInteractiveTable {
				if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
					return err
				}
			}

			if !quiet {
				PrintCompareStats(resource, pid1, pid2,
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed, result.Status)
			}

			// Post-action menu (interactive only, non-save).
			if savePath == "" && !quiet {
				comparePostAction(ctx, cmd, *result, project1Name, project2Name)
			}

			return nil
		},
	}

	addCommonFlags(cmd)
	return cmd
}

// compareSimpleInternal loads resources from both projects in parallel
// using FetchParallel[T] and compares them.
func compareSimpleInternal(
	ctx context.Context,
	cli client.ClientInterface,
	pid1, pid2 int64,
	resource string,
	fetchFn FetchFunc,
	quiet ...bool,
) (*CompareResult, error) {
	isQuiet := len(quiet) > 0 && quiet[0]

	return ui.RunWithStatus(ctx, ui.StatusConfig{
		Title:  fmt.Sprintf("Loading %s from projects %d and %d...", resource, pid1, pid2),
		Writer: os.Stderr,
		Quiet:  isQuiet,
	}, func(ctx context.Context) (*CompareResult, error) {
		results, err := concurrency.FetchParallel(ctx, []int64{pid1, pid2},
			func(pid int64) ([]ItemInfo, error) {
				return fetchFn(ctx, cli, pid)
			},
		)
		if err != nil {
			return nil, err
		}

		items1 := results[pid1]
		items2 := results[pid2]

		return compareItemInfos(resource, pid1, pid2, items1, items2), nil
	})
}
