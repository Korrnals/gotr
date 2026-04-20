package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type hubEntry struct {
	label      string
	subcommand string
}

// workCompareHub presents an interactive menu for compare subcommands.
func workCompareHub(ctx context.Context, cmd *cobra.Command) error {
	entries := []hubEntry{
		{"All resources (full scan)", "all"},
		{"Cases", "cases"},
		{"Suites", "suites"},
		{"Sections", "sections"},
		{"Shared steps", "sharedsteps"},
		{"Runs", "runs"},
		{"Plans", "plans"},
		{"Milestones", "milestones"},
		{"Datasets", "datasets"},
		{"Groups", "groups"},
		{"Labels", "labels"},
		{"Templates", "templates"},
		{"Configurations", "configurations"},
	}
	return genericHub(ctx, cmd, "compare", "Compare — what to compare?", entries)
}

// workSyncHub presents an interactive menu for sync subcommands.
func workSyncHub(ctx context.Context, cmd *cobra.Command) error {
	entries := []hubEntry{
		{"Full migration (cases + shared steps) ★", "full"},
		{"Cases", "cases"},
		{"Shared steps", "shared-steps"},
		{"Suites", "suites"},
		{"Sections", "sections"},
	}
	return genericHub(ctx, cmd, "sync", "Sync — what to migrate?", entries)
}

// workSnapHub presents an interactive menu for snap subcommands.
func workSnapHub(ctx context.Context, cmd *cobra.Command) error {
	entries := []hubEntry{
		{"📋 List snapshots", "list"},
		{"🔍 View snapshot details", "info"},
		{"↻ Rollback a snapshot", "rollback"},
		{"📤 Export snapshot", "export"},
		{"🗑  Delete snapshot", "delete"},
		{"🧹 Garbage collection", "gc"},
	}
	return genericHub(ctx, cmd, "snap", "Snap — what to do?", entries)
}

// runSubcommandHub handles generic command groups (get, add, delete, update, export, config)
// by discovering subcommands from the cobra tree.
func runSubcommandHub(ctx context.Context, cmd *cobra.Command, groupName string) error {
	root := cmd.Root()
	groupCmd, err := interactive.FindSubcommand(root, groupName)
	if err != nil {
		return fmt.Errorf("command group '%s' not found: %w", groupName, err)
	}

	subs := groupCmd.Commands()
	if len(subs) == 0 {
		// Leaf command — run it directly.
		groupCmd.SetContext(ctx)
		return groupCmd.RunE(groupCmd, nil)
	}

	entries := make([]hubEntry, 0, len(subs))
	for _, s := range subs {
		if s.Hidden || !s.IsAvailableCommand() {
			continue
		}
		entries = append(entries, hubEntry{
			label:      s.Name() + " — " + s.Short,
			subcommand: s.Name(),
		})
	}

	return genericHub(ctx, cmd, groupName, groupCmd.Short+" — choose action:", entries)
}

// genericHub is the core loop: show a Browse menu, find and run the subcommand,
// then return to the menu. Back returns to the work hub.
func genericHub(ctx context.Context, cmd *cobra.Command, group, prompt string, entries []hubEntry) error {
	p := interactive.PrompterFromContext(ctx)
	root := cmd.Root()

	labels := make([]string, len(entries))
	for i, e := range entries {
		labels[i] = e.label
	}

	for {
		idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
			Prompt:    prompt,
			Items:     labels,
			AllowBack: true,
		})
		if err != nil {
			return err // ErrGoBack → back to work hub
		}

		subName := entries[idx].subcommand
		subCmd, findErr := interactive.FindSubcommand(root, group, subName)
		if findErr != nil {
			ui.Error(os.Stdout, findErr.Error())
			continue
		}

		subCmd.SetContext(ctx)
		if runErr := subCmd.RunE(subCmd, nil); runErr != nil {
			if interactive.IsGoBack(runErr) || interactive.IsExit(runErr) {
				continue
			}
			ui.Error(os.Stdout, fmt.Sprintf("%v", runErr))
		}
		// Reset changed flags to avoid stale state on repeated invocations.
		subCmd.Flags().Visit(func(f *pflag.Flag) {
			f.Changed = false
		})
		fmt.Println()
	}
}
