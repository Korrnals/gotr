package snap

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Hook — disabled path
// ---------------------------------------------------------------------------

func TestNewHook_Disabled(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	_ = cmd.Flags().Set(FlagSnapshot, "false")

	hook := NewHook(cmd)
	assert.False(t, hook.Enabled)
	assert.Nil(t, hook.Store)
	assert.Nil(t, hook.Manifest)
}

// ---------------------------------------------------------------------------
// Hook — enabled path (real temp store)
// ---------------------------------------------------------------------------

func TestNewHook_Enabled(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)

	hook := NewHook(cmd)
	require.True(t, hook.Enabled)
	assert.NotNil(t, hook.Store)
	assert.NotNil(t, hook.Manifest)
}

// ---------------------------------------------------------------------------
// Hook.Before / FinalizeAdd / FinalizeSyncData when disabled — no-ops
// ---------------------------------------------------------------------------

func TestHook_Before_Disabled(t *testing.T) {
	hook := &Hook{Enabled: false}
	hook.Before(context.Background(), Meta{}, nil)
	assert.Nil(t, hook.Snap)
}

func TestHook_FinalizeAdd_Disabled(t *testing.T) {
	hook := &Hook{Enabled: false}
	hook.FinalizeAdd(42) // should not panic
}

func TestHook_FinalizeAdd_NilSnap(t *testing.T) {
	hook := &Hook{Enabled: true, Snap: nil}
	hook.FinalizeAdd(42) // should not panic
}

func TestHook_FinalizeSyncData_Disabled(t *testing.T) {
	hook := &Hook{Enabled: false}
	hook.FinalizeSyncData(map[string]int{"a": 1}) // should not panic
}

func TestHook_FinalizeSyncData_NilSnap(t *testing.T) {
	hook := &Hook{Enabled: true, Snap: nil}
	hook.FinalizeSyncData(map[string]int{"a": 1}) // should not panic
}

// ---------------------------------------------------------------------------
// Hook.Before + FinalizeAdd — full happy path
// ---------------------------------------------------------------------------

func TestHook_Before_Happy(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)

	hook := NewHook(cmd)
	require.True(t, hook.Enabled)

	meta := BuildMeta(OpAdd, "case", nil, Tier1, 1, 1, "", nil, "https://x.com")
	hook.Before(t.Context(), meta, nil)
	assert.NotNil(t, hook.Snap)

	// FinalizeAdd should work without error.
	hook.FinalizeAdd(999)
}

// ---------------------------------------------------------------------------
// Hook.Before with label from flag
// ---------------------------------------------------------------------------

func TestHook_Before_WithLabel(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	_ = cmd.Flags().Set(FlagSnapLabel, "my-label")

	hook := NewHook(cmd)
	require.True(t, hook.Enabled)
	assert.Equal(t, "my-label", hook.label)

	meta := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 1, 1, "", nil, "https://x.com")
	hook.Before(t.Context(), meta, nil)
	require.NotNil(t, hook.Snap)
	assert.Equal(t, "my-label", hook.Snap.Meta.Label)
}

// ---------------------------------------------------------------------------
// Hook.FinalizeSyncData — happy path
// ---------------------------------------------------------------------------

func TestHook_FinalizeSyncData_Happy(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)

	hook := NewHook(cmd)
	require.True(t, hook.Enabled)

	meta := BuildMeta(OpSyncFull, "case", nil, Tier1, 1, 1, "", nil, "https://x.com")
	hook.Before(t.Context(), meta, nil)
	require.NotNil(t, hook.Snap)

	// Should save without panic.
	hook.FinalizeSyncData(map[string]int{"created": 5})
}

// ---------------------------------------------------------------------------
// HookMutation convenience
// ---------------------------------------------------------------------------

func TestHookMutation(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	_ = cmd.Flags().Set(FlagSnapLabel, "hook-mut-label")

	hook := HookMutation(t.Context(), Mutation{
		Cmd:        cmd,
		Op:         OpUpdate,
		EntityType: "case",
		EntityIDs:  []int64{10},
		Tier:       Tier1,
		ProjectID:  1,
		SuiteID:    1,
	})
	require.True(t, hook.Enabled)
	require.NotNil(t, hook.Snap)
	assert.Equal(t, "hook-mut-label", hook.Snap.Meta.Label)
}

func TestHookMutation_WithExplicitLabel(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)

	hook := HookMutation(t.Context(), Mutation{
		Cmd:        cmd,
		Op:         OpDelete,
		EntityType: "suite",
		EntityIDs:  []int64{5},
		Tier:       Tier2,
		ProjectID:  2,
		SuiteID:    5,
		Label:      "explicit-label",
	})
	require.True(t, hook.Enabled)
	require.NotNil(t, hook.Snap)
	assert.Equal(t, "explicit-label", hook.Snap.Meta.Label)
}

func TestHookMutation_Disabled(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	_ = cmd.Flags().Set(FlagSnapshot, "false")

	hook := HookMutation(t.Context(), Mutation{
		Cmd:        cmd,
		Op:         OpUpdate,
		EntityType: "case",
		EntityIDs:  []int64{1},
		Tier:       Tier1,
		ProjectID:  1,
		SuiteID:    1,
	})
	assert.False(t, hook.Enabled)
	assert.Nil(t, hook.Snap)
}

// ---------------------------------------------------------------------------
// SnapOrWarn
// ---------------------------------------------------------------------------

func TestSnapOrWarn_Happy(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := NewStore()
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 1, 1, "", nil, "https://x.com")
	snap := SnapOrWarn(t.Context(), store, manifest, meta, nil)
	assert.NotNil(t, snap)
}

func TestSnapOrWarn_FailedFetch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := NewStore()
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 1, 1, "", nil, "https://x.com")
	failFetch := func(_ context.Context) (interface{}, error) {
		return nil, assert.AnError
	}
	snap := SnapOrWarn(t.Context(), store, manifest, meta, failFetch)
	// Fetch failure is fatal for snapshot; SnapOrWarn logs warning and returns nil.
	assert.Nil(t, snap)
}
