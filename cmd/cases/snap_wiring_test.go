package cases

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Wiring test: cases update creates a snapshot
// ---------------------------------------------------------------------------

func TestUpdateCmd_CreatesSnapshot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	mock := &client.MockClient{
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			return &data.Case{ID: caseID, Title: "Original", SectionID: 1, PriorityID: 3}, nil
		},
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return &data.Case{ID: caseID, Title: *req.Title}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42", "--title=Updated"})
	cmd.SetOut(&devNull{})
	cmd.SetErr(&devNull{})

	err := cmd.Execute()
	require.NoError(t, err)

	// Snapshot must exist in $HOME/.gotr/snaps/.
	snapsDir := filepath.Join(home, ".gotr", "snaps")
	store, err := snaplib.NewStoreAt(snapsDir)
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	assert.Equal(t, 1, manifest.Len())
	entries := manifest.All()
	assert.Equal(t, snaplib.OpUpdate, entries[0].Operation)
	assert.Equal(t, "case", entries[0].EntityType)
	assert.Equal(t, snaplib.StatusAvailable, entries[0].Status)
}

// ---------------------------------------------------------------------------
// Wiring test: cases delete creates a snapshot
// ---------------------------------------------------------------------------

func TestDeleteCmd_CreatesSnapshot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	mock := &client.MockClient{
		GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
			return &data.Case{ID: caseID, Title: "ToDelete", SectionID: 5, PriorityID: 2}, nil
		},
		DeleteCaseFunc: func(ctx context.Context, caseID int64) error {
			return nil
		},
	}

	cmd := newDeleteCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"99"})
	cmd.SetOut(&devNull{})
	cmd.SetErr(&devNull{})

	err := cmd.Execute()
	require.NoError(t, err)

	snapsDir := filepath.Join(home, ".gotr", "snaps")
	store, err := snaplib.NewStoreAt(snapsDir)
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	assert.Equal(t, 1, manifest.Len())
	entries := manifest.All()
	assert.Equal(t, snaplib.OpDelete, entries[0].Operation)
	assert.Equal(t, "case", entries[0].EntityType)
}

// ---------------------------------------------------------------------------
// Wiring test: cases add creates a snapshot with FinalizeAdd
// ---------------------------------------------------------------------------

func TestAddCmd_CreatesSnapshot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	mock := &client.MockClient{
		AddCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			return &data.Case{ID: 777, Title: req.Title, SectionID: sectionID}, nil
		},
	}

	cmd := newAddCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"10", "--title=New test"})
	cmd.SetOut(&devNull{})
	cmd.SetErr(&devNull{})

	err := cmd.Execute()
	require.NoError(t, err)

	snapsDir := filepath.Join(home, ".gotr", "snaps")
	store, err := snaplib.NewStoreAt(snapsDir)
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)

	assert.Equal(t, 1, manifest.Len())
	entries := manifest.All()
	assert.Equal(t, snaplib.OpAdd, entries[0].Operation)
	assert.Equal(t, "case", entries[0].EntityType)

	// FinalizeAdd should have saved the created ID in meta.
	meta, err := store.LoadMeta(entries[0].ID)
	require.NoError(t, err)
	assert.Contains(t, meta.EntityIDs, int64(777))
}

// ---------------------------------------------------------------------------
// Wiring test: --snapshot=false disables snapshotting
// ---------------------------------------------------------------------------

func TestUpdateCmd_SnapshotDisabled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	mock := &client.MockClient{
		UpdateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return &data.Case{ID: caseID, Title: "Updated"}, nil
		},
	}

	cmd := newUpdateCmd(getClientForTests)
	cmd.SetContext(setupTestCmd(t, mock).Context())
	cmd.SetArgs([]string{"42", "--title=Updated", "--snapshot=false"})
	cmd.SetOut(&devNull{})
	cmd.SetErr(&devNull{})

	err := cmd.Execute()
	require.NoError(t, err)

	// No snapshot should exist.
	snapsDir := filepath.Join(home, ".gotr", "snaps")
	if _, err := os.Stat(snapsDir); os.IsNotExist(err) {
		return // dir doesn't exist — correct
	}
	store, err := snaplib.NewStoreAt(snapsDir)
	require.NoError(t, err)
	manifest, err := snaplib.LoadManifest(store)
	require.NoError(t, err)
	assert.Equal(t, 0, manifest.Len())
}

// devNull discards all writes (for suppressing command output in tests).
type devNull struct{}

func (d *devNull) Write(p []byte) (int, error) { return len(p), nil }
