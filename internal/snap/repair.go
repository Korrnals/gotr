package snap

import (
	"fmt"
	"sort"
)

// RepairAction describes a single change planned or applied by RepairManifest.
type RepairAction struct {
	// Op is one of "add" or "remove".
	Op string
	// SnapID is the snapshot identifier the action targets.
	SnapID string
	// Reason is a human-readable explanation.
	Reason string
}

// RepairResult summarizes the outcome of a manifest repair pass.
type RepairResult struct {
	// Added lists meta-on-disk snapshots that were missing from the manifest
	// and (unless DryRun) have been re-indexed.
	Added []RepairAction
	// Removed lists manifest entries whose meta.json no longer exists on disk
	// and (unless DryRun) have been pruned from the manifest.
	Removed []RepairAction
	// MetaErrors lists snapshot directories whose meta.json failed to load.
	// Such entries are reported but not auto-pruned.
	MetaErrors []RepairAction
	// DryRun, if true, indicates no changes were written to the manifest.
	DryRun bool
}

// HasChanges reports whether the repair pass found any drift between the
// on-disk snapshot directories and the manifest entries.
func (r *RepairResult) HasChanges() bool {
	return len(r.Added) > 0 || len(r.Removed) > 0
}

// RepairManifest reconciles the manifest with the snapshot directories on
// disk. Drift sources it handles:
//
//   - Snapshot directories present on disk with a valid meta.json but missing
//     from manifest.entries → re-indexed via Manifest.Add (status preserved
//     from meta).
//   - Manifest entries whose snapshot directory is gone → removed from the
//     manifest.
//
// Snapshot directories whose meta.json cannot be read are reported in
// MetaErrors but left untouched in both the manifest and on disk; the
// operator is expected to investigate them manually.
//
// When dryRun is true the manifest file is not modified; the returned
// RepairResult still describes the actions that would have been taken.
//
//nolint:gocyclo // Repair walks store directory, reconciles entries and applies orphan/missing fixes inline.
func RepairManifest(store *Store, manifest *Manifest, dryRun bool) (*RepairResult, error) {
	result := &RepairResult{DryRun: dryRun}

	diskIDs, err := store.List()
	if err != nil {
		return nil, fmt.Errorf("snap repair: list store: %w", err)
	}
	diskSet := make(map[string]struct{}, len(diskIDs))
	for _, id := range diskIDs {
		diskSet[id] = struct{}{}
	}

	manifestIDs := manifest.ManifestIDs()

	// 1. Find directories on disk missing from the manifest.
	missing := make([]string, 0)
	for _, id := range diskIDs {
		if _, ok := manifestIDs[id]; ok {
			continue
		}
		missing = append(missing, id)
	}
	sort.Strings(missing)

	toAdd := make([]*Meta, 0, len(missing))
	for _, id := range missing {
		meta, err := store.LoadMeta(id)
		if err != nil {
			result.MetaErrors = append(result.MetaErrors, RepairAction{
				Op:     "skip",
				SnapID: id,
				Reason: fmt.Sprintf("load meta failed: %v", err),
			})
			continue
		}
		// Trust the on-disk ID: meta.ID is allowed to drift from the dir
		// path (e.g. after a manual rename), but the manifest must index by
		// the directory ID since that is what every other code path looks up.
		meta.ID = id
		result.Added = append(result.Added, RepairAction{
			Op:     "add",
			SnapID: id,
			Reason: fmt.Sprintf("meta.json present, missing from manifest (status=%s)", meta.Status),
		})
		toAdd = append(toAdd, meta)
	}
	if !dryRun && len(toAdd) > 0 {
		if err := manifest.AddMany(toAdd); err != nil {
			return result, fmt.Errorf("snap repair: re-index batch: %w", err)
		}
	}

	// 2. Find manifest entries whose snapshot directory is gone.
	orphans := make([]string, 0)
	for id := range manifestIDs {
		if _, ok := diskSet[id]; ok {
			continue
		}
		orphans = append(orphans, id)
	}
	sort.Strings(orphans)

	for _, id := range orphans {
		result.Removed = append(result.Removed, RepairAction{
			Op:     "remove",
			SnapID: id,
			Reason: "manifest entry has no matching directory on disk",
		})
	}
	if !dryRun && len(orphans) > 0 {
		if err := manifest.RemoveMany(orphans); err != nil {
			return result, fmt.Errorf("snap repair: prune orphans: %w", err)
		}
	}

	return result, nil
}
