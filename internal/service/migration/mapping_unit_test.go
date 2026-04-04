package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSharedStepMapping_Save_MarshalError(t *testing.T) {
	sm := NewSharedStepMapping(1, 2)
	sm.AddPair(1, 101, "created")

	// time.Time with year > 9999 causes json.Marshal to fail deterministically.
	sm.CreatedAt = time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)

	err := sm.Save(t.TempDir())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "year outside of range")
}

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

func TestSharedStepMapping_Save_TableDriven(t *testing.T) {
	testCases := []struct {
		name        string
		dirFactory  func(t *testing.T) string
		prepare     func(sm *SharedStepMapping)
		wantErr     bool
		checkResult func(t *testing.T, dir string)
	}{
		{
			name: "default path when dir is empty",
			dirFactory: func(t *testing.T) string {
				wd, err := os.Getwd()
				if err != nil {
					t.Fatalf("Getwd() error: %v", err)
				}
				base := t.TempDir()
				if err := os.Chdir(base); err != nil {
					t.Fatalf("Chdir(%s) error: %v", base, err)
				}
				t.Cleanup(func() {
					_ = os.Chdir(wd)
				})
				return ""
			},
			prepare: func(sm *SharedStepMapping) {
				sm.AddPair(1, 101, "created")
			},
			wantErr: false,
			checkResult: func(t *testing.T, _ string) {
				entries, err := os.ReadDir(".testrail")
				if err != nil {
					t.Fatalf("ReadDir(.testrail) error: %v", err)
				}
				assert.Len(t, entries, 1)
				assert.Contains(t, entries[0].Name(), "mapping_")
			},
		},
		{
			name: "mkdir error when path collides with file",
			dirFactory: func(t *testing.T) string {
				tmp := t.TempDir()
				filePath := filepath.Join(tmp, "not_a_dir")
				if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error: %v", filePath, err)
				}
				return filepath.Join(filePath, "child")
			},
			prepare: func(sm *SharedStepMapping) {
				sm.AddPair(2, 202, "existing")
			},
			wantErr: true,
		},
		{
			name: "write error on read-only directory",
			dirFactory: func(t *testing.T) string {
				roDir := filepath.Join(t.TempDir(), "readonly")
				if err := os.MkdirAll(roDir, 0o755); err != nil {
					t.Fatalf("MkdirAll(%s) error: %v", roDir, err)
				}
				if err := os.Chmod(roDir, 0o555); err != nil {
					t.Fatalf("Chmod(%s) error: %v", roDir, err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(roDir, 0o755)
				})
				return roDir
			},
			prepare: func(sm *SharedStepMapping) {
				sm.AddPair(3, 303, "created")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := tc.dirFactory(t)
			sm := NewSharedStepMapping(1, 2)
			tc.prepare(sm)

			err := sm.Save(dir)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tc.checkResult != nil {
				tc.checkResult(t, dir)
			}
		})
	}
}

func TestSharedStepMapping_Save_EmptyDataCases(t *testing.T) {
	testCases := []struct {
		name string
		sm   *SharedStepMapping
		dir  string
	}{
		{
			name: "nil pairs with zero count",
			sm: &SharedStepMapping{
				SrcProjectID: 1,
				DstProjectID: 2,
				Count:        0,
				Pairs:        nil,
				index:        map[int64]int64{},
			},
			dir: t.TempDir(),
		},
		{
			name: "empty pairs with zero count",
			sm: &SharedStepMapping{
				SrcProjectID: 1,
				DstProjectID: 2,
				Count:        0,
				Pairs:        []MappingPair{},
				index:        map[int64]int64{},
			},
			dir: t.TempDir(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sm.Save(tc.dir)
			assert.NoError(t, err)

			entries, readErr := os.ReadDir(tc.dir)
			assert.NoError(t, readErr)
			assert.Len(t, entries, 0, fmt.Sprintf("directory should stay empty for case %q", tc.name))
		})
	}
}
