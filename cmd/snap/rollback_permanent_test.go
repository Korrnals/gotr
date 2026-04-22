package snap

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI_SnapRollback_Add_PermanentFlag verifies that --permanent routes the
// rollback delete through DeleteCasePermanent instead of the soft-delete default.
func TestCLI_SnapRollback_Add_PermanentFlag(t *testing.T) {
	redirectHome(t)

	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	meta := snaplib.BuildMeta(snaplib.OpAdd, "case", nil, snaplib.Tier2, 1, 1, "", []string{"cases", "add"}, "")
	snap, err := snaplib.TakeSnapshot(context.Background(), store, manifest, meta, nil)
	require.NoError(t, err)
	require.NoError(t, snap.FinalizeAdd(500))
	snapID := snap.Meta.ID

	var softCalled, permCalled bool
	mock := &client.MockClient{
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			softCalled = true
			return nil
		},
		DeleteCasePermanentFunc: func(ctx context.Context, caseID int64) error {
			permCalled = true
			return nil
		},
	}

	cmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd.SetContext(ctx)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs([]string{snapID, "--permanent"})
	err = cmd.Execute()
	require.NoError(t, err)

	w.Close()
	var captured bytes.Buffer
	_, _ = captured.ReadFrom(r)
	os.Stdout = old

	assert.False(t, softCalled, "DeleteCase (soft) must not be called when --permanent is set")
	assert.True(t, permCalled, "DeleteCasePermanent must be called when --permanent is set")
}

// TestCLI_SnapRollback_Add_DefaultIsSoft verifies that without --permanent the
// rollback delete uses the soft-delete path (DeleteCase), matching the Stage 0.5
// fix that makes TestRail's UI-compatible trash behaviour the default.
func TestCLI_SnapRollback_Add_DefaultIsSoft(t *testing.T) {
	redirectHome(t)

	store, err := snaplib.NewStore()
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	meta := snaplib.BuildMeta(snaplib.OpAdd, "case", nil, snaplib.Tier2, 1, 1, "", []string{"cases", "add"}, "")
	snap, err := snaplib.TakeSnapshot(context.Background(), store, manifest, meta, nil)
	require.NoError(t, err)
	require.NoError(t, snap.FinalizeAdd(501))
	snapID := snap.Meta.ID

	var softCalled, permCalled bool
	mock := &client.MockClient{
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			softCalled = true
			assert.Equal(t, int64(501), caseID)
			return nil
		},
		DeleteCasePermanentFunc: func(ctx context.Context, caseID int64) error {
			permCalled = true
			return nil
		},
	}

	cmd := newRollbackCmd(getClientForTests)
	ctx := nonInteractiveCtx(mock)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{snapID})
	err = cmd.Execute()
	require.NoError(t, err)

	assert.True(t, softCalled, "DeleteCase (soft) must be called by default")
	assert.False(t, permCalled, "DeleteCasePermanent must not be called by default")

	// Sanity: keep the data type referenced so imports stay stable in the future.
	_ = data.Case{}
}
