package snap

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeSnapshot_WithFetch(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := BuildMeta(OpUpdate, "case", []int64{42}, Tier1, 30, 100, "", []string{"cases", "update", "42"})

	fetchData := map[string]interface{}{"id": float64(42), "title": "Original"}
	fetchFn := func(ctx context.Context) (interface{}, error) {
		return fetchData, nil
	}

	s, err := TakeSnapshot(context.Background(), store, manifest, meta, fetchFn)
	require.NoError(t, err)
	require.NotNil(t, s)

	assert.Equal(t, StatusAvailable, s.Meta.Status)
	assert.False(t, s.Meta.Timestamp.IsZero())
	assert.Greater(t, s.Meta.DataSizeBytes, int64(0))
	assert.Equal(t, 1, manifest.Len())
	assert.True(t, strings.HasPrefix(s.Meta.ID, "cases/"), "ID should have cases/ prefix")
	assert.Equal(t, Category("cases"), s.Meta.Category)

	// Verify data on disk.
	var loaded map[string]interface{}
	err = store.LoadData(s.Meta.ID, "data.json", &loaded)
	require.NoError(t, err)
	assert.Equal(t, "Original", loaded["title"])
}

func TestTakeSnapshot_AddNoFetch(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := BuildMeta(OpAdd, "case", nil, Tier2, 30, 100, "", []string{"add", "case"})

	s, err := TakeSnapshot(context.Background(), store, manifest, meta, nil)
	require.NoError(t, err)
	require.NotNil(t, s)

	assert.Equal(t, StatusAvailable, s.Meta.Status)
	assert.Equal(t, 1, manifest.Len())
	assert.True(t, strings.HasPrefix(s.Meta.ID, "cases/"), "ID should have cases/ prefix")
}

func TestTakeSnapshot_FinalizeAdd(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := BuildMeta(OpAdd, "section", nil, Tier2, 30, 0, "", []string{"add", "section"})

	s, err := TakeSnapshot(context.Background(), store, manifest, meta, nil)
	require.NoError(t, err)

	err = s.FinalizeAdd(999)
	require.NoError(t, err)

	assert.Equal(t, []int64{999}, s.Meta.EntityIDs)
	assert.Equal(t, int64(999), s.Meta.Entities[0].ID)

	// Verify persisted.
	loaded, err := store.LoadMeta(s.Meta.ID)
	require.NoError(t, err)
	assert.Equal(t, []int64{999}, loaded.EntityIDs)
}

func TestTakeSnapshot_FetchError(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 30, 0, "", nil)

	fetchFn := func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("API error 404")
	}

	s, err := TakeSnapshot(context.Background(), store, manifest, meta, fetchFn)
	assert.Error(t, err)
	assert.Nil(t, s)
	assert.Contains(t, err.Error(), "API error 404")
	assert.Equal(t, 0, manifest.Len())
}

func TestTakeSnapshot_CustomName(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	meta := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 30, 0, "before-migration", nil)

	fetchFn := func(ctx context.Context) (interface{}, error) {
		return map[string]string{"hello": "world"}, nil
	}

	s, err := TakeSnapshot(context.Background(), store, manifest, meta, fetchFn)
	require.NoError(t, err)
	assert.Equal(t, "custom/before-migration", s.Meta.ID)
	assert.Equal(t, "before-migration", s.Meta.Name)
	assert.Equal(t, CatCustom, s.Meta.Category)
}

func TestBuildMeta(t *testing.T) {
	meta := BuildMeta(OpDelete, "section", []int64{10, 20}, Tier2, 3, 100, "", []string{"delete", "section", "10,20"})
	assert.Equal(t, OpDelete, meta.Operation)
	assert.Equal(t, "section", meta.EntityType)
	assert.Equal(t, Tier2, meta.RollbackTier)
	assert.Equal(t, int64(3), meta.ProjectID)
	assert.Contains(t, meta.CLICommand, "gotr delete section")
	assert.True(t, strings.HasPrefix(meta.ID, "sections/"), "ID should have sections/ prefix")
	assert.Equal(t, Category("sections"), meta.Category)
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"before migration", "before-migration"},
		{"path/to/snap", "path_to_snap"},
		{"time:12:00", "time_12_00"},
		{"normal-name", "normal-name"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, sanitizeName(tt.input))
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 30, cfg.RetentionDays)
	assert.Equal(t, 100, cfg.MaxSnapshots)
	assert.Equal(t, "auto", cfg.AttachSaveBinary)
	assert.Equal(t, 10, cfg.AttachMaxFileMB)
	assert.True(t, cfg.AttachCompress)
	assert.True(t, cfg.AttachPromptAboveThresh)
}

func TestTimestampInMeta(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	before := time.Now().UTC().Add(-time.Second)
	meta := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 1, 0, "", nil)
	s, err := TakeSnapshot(context.Background(), store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return "data", nil
	})
	require.NoError(t, err)
	after := time.Now().UTC().Add(time.Second)

	assert.True(t, s.Meta.Timestamp.After(before))
	assert.True(t, s.Meta.Timestamp.Before(after))
}
