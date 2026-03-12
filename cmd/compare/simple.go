package compare

import (
	"context"
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/concurrency"
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
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			pid1, pid2, format, savePath, err := parseCommonFlags(cmd)
			if err != nil {
				return err
			}

			project1Name, project2Name, err := GetProjectNames(ctx, cli, pid1, pid2)
			if err != nil {
				return err
			}

			startTime := time.Now()

			result, err := compareSimpleInternal(cli, pid1, pid2, resource, fetchFn)
			if err != nil {
				return fmt.Errorf("comparison error %s: %w", resource, err)
			}

			elapsed := time.Since(startTime)

			if err := PrintCompareResult(cmd, *result, project1Name, project2Name, format, savePath); err != nil {
				return err
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				PrintCompareStats(resource, pid1, pid2,
					len(result.OnlyInFirst), len(result.OnlyInSecond), len(result.Common), elapsed)
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
	cli client.ClientInterface,
	pid1, pid2 int64,
	resource string,
	fetchFn FetchFunc,
) (*CompareResult, error) {
	ctx := context.Background()

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
}
