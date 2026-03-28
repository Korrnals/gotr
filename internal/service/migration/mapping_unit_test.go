package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSharedStepMapping_BasicOperations(t *testing.T) {
	sm := NewSharedStepMapping(1, 2)
	assert.Equal(t, 0, sm.Count)

	sm.AddPair(30, 300, "created")
	sm.AddPair(10, 100, "existing")
	sm.AddPair(20, 200, "created")
	sm.AddPair(10, 999, "duplicate") // must be ignored

	assert.Equal(t, 3, sm.Count)
	target, ok := sm.GetTargetBySource(10)
	assert.True(t, ok)
	assert.Equal(t, int64(100), target)

	sm.SortPairs()
	assert.Equal(t, int64(10), sm.Pairs[0].SourceID)
	assert.Equal(t, int64(20), sm.Pairs[1].SourceID)
	assert.Equal(t, int64(30), sm.Pairs[2].SourceID)
}

func TestSharedStepMapping_SaveAndLoad(t *testing.T) {
	t.Run("empty mapping save is no-op", func(t *testing.T) {
		dir := t.TempDir()
		sm := NewSharedStepMapping(1, 2)

		err := sm.Save(dir)
		assert.NoError(t, err)

		entries, err := os.ReadDir(dir)
		assert.NoError(t, err)
		assert.Len(t, entries, 0)
	})

	t.Run("save and load", func(t *testing.T) {
		dir := t.TempDir()
		sm := NewSharedStepMapping(1, 2)
		sm.AddPair(11, 111, "created")
		sm.AddPair(12, 112, "existing")

		err := sm.Save(dir)
		assert.NoError(t, err)

		entries, err := os.ReadDir(dir)
		assert.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Contains(t, entries[0].Name(), "mapping_")

		loaded, err := LoadSharedStepMapping(filepath.Join(dir, entries[0].Name()))
		assert.NoError(t, err)
		assert.Equal(t, 2, loaded.Count)

		target, ok := loaded.GetTargetBySource(12)
		assert.True(t, ok)
		assert.Equal(t, int64(112), target)
	})

	t.Run("load missing file error", func(t *testing.T) {
		_, err := LoadSharedStepMapping(filepath.Join(t.TempDir(), "missing.json"))
		assert.Error(t, err)
	})
}
