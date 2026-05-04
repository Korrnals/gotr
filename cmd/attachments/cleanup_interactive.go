// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// promptCleanupOptions runs the interactive survey for `gotr attachments
// cleanup` and overlays the resulting choices onto the flag-derived
// opts. It is a no-op when the prompter is missing or non-interactive.
//
// The survey only asks about a value when the corresponding CLI flag was
// not explicitly provided, so users can mix flags and prompts freely.
//nolint:gocyclo // Linear survey: each block guards one CLI flag with the same shape, splitting per-question would obscure the prompt order.
func promptCleanupOptions(ctx context.Context, cmd *cobra.Command, opts *cleanupOptions) error {
	if !interactive.HasPrompterInContext(ctx) || interactive.IsNonInteractive(ctx) {
		return nil
	}
	p := interactive.PrompterFromContext(ctx)
	if p == nil {
		return nil
	}

	// 1. Project scope.
	if !flagExplicit(cmd, "project") && !flagExplicit(cmd, "all-projects") {
		_, choice, err := p.Select("Project scope", []string{"All visible projects", "Specific project IDs"})
		if err != nil {
			return fmt.Errorf("project scope: %w", err)
		}
		if choice == "All visible projects" {
			opts.AllProjects = true
		} else {
			raw, err := p.Input("Project IDs (comma-separated)", "")
			if err != nil {
				return fmt.Errorf("project ids: %w", err)
			}
			ids, err := parseInt64List(raw)
			if err != nil {
				return fmt.Errorf("project ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("at least one project ID is required")
			}
			opts.ProjectIDs = ids
		}
	}

	// 2. Entity types — multi-select via comma-separated input with
	//    inline help. Empty answer keeps the flag default.
	if !flagExplicit(cmd, "entity-type") {
		raw, err := p.Input(
			"Parent kinds to clean (comma-separated: case,run,plan,plan_entry,result,test)",
			strings.Join(opts.EntityTypes, ","),
		)
		if err != nil {
			return fmt.Errorf("entity types: %w", err)
		}
		types, err := parseEntityTypeList(raw)
		if err != nil {
			return fmt.Errorf("entity types: %w", err)
		}
		opts.EntityTypes = types
	}

	// 3. Age cutoff.
	if !flagExplicit(cmd, "older-than") {
		raw, err := p.Input("Older than (e.g. 7d, 3M, 1y)", "90d")
		if err != nil {
			return fmt.Errorf("older-than: %w", err)
		}
		d, err := parseHumanDuration(raw)
		if err != nil {
			return fmt.Errorf("older-than %q: %w", raw, err)
		}
		opts.OlderThan = d
	}

	// 4. Concurrency.
	if !flagExplicit(cmd, "concurrency") {
		raw, err := p.Input("Worker concurrency", strconv.Itoa(opts.Concurrency))
		if err != nil {
			return fmt.Errorf("concurrency: %w", err)
		}
		if raw != "" {
			n, perr := strconv.Atoi(raw)
			if perr != nil || n < 1 {
				return fmt.Errorf("concurrency %q: must be a positive integer", raw)
			}
			opts.Concurrency = n
		}
	}

	// 5. Snapshot toggle + retention.
	if !flagExplicit(cmd, "no-snapshot") {
		take, err := p.Confirm("Take a snapshot before deleting (recommended)?", true)
		if err != nil {
			return fmt.Errorf("snapshot toggle: %w", err)
		}
		opts.SkipSnapshot = !take
	}
	if !opts.SkipSnapshot && !flagExplicit(cmd, "snapshot-retention") {
		raw, err := p.Input("Snapshot retention (e.g. 7d, 14d, 30d)", "7d")
		if err != nil {
			return fmt.Errorf("snapshot-retention: %w", err)
		}
		if raw != "" {
			d, perr := parseHumanDuration(raw)
			if perr != nil {
				return fmt.Errorf("snapshot-retention %q: %w", raw, perr)
			}
			opts.SnapshotRetention = d
		}
	}

	// 6. Dry-run prompt — only offered when the user did not pass it
	//    on the CLI; defaults to NO so the survey-driven flow performs
	//    the real cleanup once confirmed.
	if !flagExplicit(cmd, "dry-run") {
		dry, err := p.Confirm("Dry-run only (no snapshot, no deletes)?", false)
		if err != nil {
			return fmt.Errorf("dry-run: %w", err)
		}
		opts.DryRun = dry
	}
	return nil
}

// confirmCleanupExecution is the final pre-flight gate. Returns false
// when the user declines, true when --force is set or the user
// confirms.
func confirmCleanupExecution(ctx context.Context, opts *cleanupOptions) (bool, error) {
	if opts.Force || opts.DryRun {
		return true, nil
	}
	if !interactive.HasPrompterInContext(ctx) || interactive.IsNonInteractive(ctx) {
		// Outside of a TTY we proceed without an extra confirmation —
		// callers should pass --force explicitly in scripts.
		return true, nil
	}
	p := interactive.PrompterFromContext(ctx)
	if p == nil {
		return true, nil
	}
	ok, err := p.Confirm("Proceed with deletion?", false)
	if err != nil {
		return false, fmt.Errorf("confirm: %w", err)
	}
	return ok, nil
}

// flagExplicit reports whether the user passed the named flag on the
// command line (vs. the cobra default).
func flagExplicit(cmd *cobra.Command, name string) bool {
	f := cmd.Flags().Lookup(name)
	return f != nil && f.Changed
}

// parseInt64List parses a comma- (or whitespace-) separated list of
// signed 64-bit integers, ignoring empty fragments.
func parseInt64List(raw string) ([]int64, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	splitter := func(r rune) bool { return r == ',' || r == ' ' || r == '\t' }
	parts := strings.FieldsFunc(raw, splitter)
	out := make([]int64, 0, len(parts))
	seen := map[int64]struct{}{}
	for _, p := range parts {
		n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", p, err)
		}
		if _, dup := seen[n]; dup {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func parseEntityTypeList(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return []string{"result"}, nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		v := strings.ToLower(strings.TrimSpace(p))
		if v == "" {
			continue
		}
		if !isValidCleanupEntity(v) {
			return nil, fmt.Errorf("%q invalid (allowed: %s)", v, strings.Join(validCleanupEntityTypes, ", "))
		}
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one entity type is required")
	}
	return out, nil
}
