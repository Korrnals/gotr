package snap

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// Hook is a convenience wrapper for pre-mutation snapshot operations.
// Commands use it to take snapshots with minimal boilerplate.
type Hook struct {
	Store    *Store
	Manifest *Manifest
	Snap     *Snapshot
	Enabled  bool
}

// NewHook creates a snapshot hook based on the command's flags and config.
// Returns a disabled hook on any initialization error (non-fatal).
func NewHook(cmd *cobra.Command) *Hook {
	if !ResolveDecision(cmd) {
		return &Hook{Enabled: false}
	}

	store, err := NewStore()
	if err != nil {
		ui.Warning(os.Stderr, fmt.Sprintf("snap: %v", err))
		return &Hook{Enabled: false}
	}

	manifest, err := LoadManifest(store)
	if err != nil {
		ui.Warning(os.Stderr, fmt.Sprintf("snap: %v", err))
		return &Hook{Enabled: false}
	}

	return &Hook{Store: store, Manifest: manifest, Enabled: true}
}

// Before takes a pre-mutation snapshot.
// For update/delete operations, fetchFn should return the current entity state.
// For add operations, pass nil fetchFn.
func (h *Hook) Before(ctx context.Context, meta Meta, fetchFn func(ctx context.Context) (interface{}, error)) {
	if !h.Enabled {
		return
	}
	h.Snap = SnapOrWarn(ctx, h.Store, h.Manifest, meta, fetchFn)
}

// FinalizeAdd records the created entity ID after an add mutation.
func (h *Hook) FinalizeAdd(createdID int64) {
	if !h.Enabled || h.Snap == nil {
		return
	}
	if err := h.Snap.FinalizeAdd(createdID); err != nil {
		ui.Warning(os.Stderr, fmt.Sprintf("snap: finalize: %v", err))
	}
}

// FinalizeSyncData writes sync result data (created entities mapping) to the snapshot.
func (h *Hook) FinalizeSyncData(data interface{}) {
	if !h.Enabled || h.Snap == nil {
		return
	}
	meta := h.Snap.Meta
	if _, err := h.Store.SaveData(meta.ID, meta.DataFile, data); err != nil {
		ui.Warning(os.Stderr, fmt.Sprintf("snap: save sync data: %v", err))
	}
}
