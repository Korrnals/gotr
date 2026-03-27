package migration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
)

func countFilesByPrefix(t *testing.T, dir, prefix string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error: %v", dir, err)
	}
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			count++
		}
	}
	return count
}

func TestMigrationExportFunctions(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})

	t.Run("empty inputs", func(t *testing.T) {
		dir := t.TempDir()

		if err := m.ExportSharedSteps(nil, false, dir); err != nil {
			t.Fatalf("ExportSharedSteps() error = %v", err)
		}
		if err := m.ExportSuites(nil, false, dir); err != nil {
			t.Fatalf("ExportSuites() error = %v", err)
		}
		if err := m.ExportCases(nil, false, dir); err != nil {
			t.Fatalf("ExportCases() error = %v", err)
		}
		if err := m.ExportSections(nil, false, dir); err != nil {
			t.Fatalf("ExportSections() error = %v", err)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir() error = %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected no files for empty exports, got %d", len(entries))
		}
	})

	t.Run("non-empty exports write files", func(t *testing.T) {
		dir := t.TempDir()

		if err := m.ExportSharedSteps(data.GetSharedStepsResponse{{ID: 1, Title: "S"}}, true, dir); err != nil {
			t.Fatalf("ExportSharedSteps() error = %v", err)
		}
		if err := m.ExportSuites(data.GetSuitesResponse{{ID: 2, Name: "Suite"}}, false, dir); err != nil {
			t.Fatalf("ExportSuites() error = %v", err)
		}
		if err := m.ExportCases(data.GetCasesResponse{{ID: 3, Title: "Case"}}, true, dir); err != nil {
			t.Fatalf("ExportCases() error = %v", err)
		}
		if err := m.ExportSections(data.GetSectionsResponse{{ID: 4, Name: "Sec"}}, false, dir); err != nil {
			t.Fatalf("ExportSections() error = %v", err)
		}

		if got := countFilesByPrefix(t, dir, "shared_steps_filtered_"); got != 1 {
			t.Fatalf("expected 1 shared_steps export file, got %d", got)
		}
		if got := countFilesByPrefix(t, dir, "suites_all_"); got != 1 {
			t.Fatalf("expected 1 suites export file, got %d", got)
		}
		if got := countFilesByPrefix(t, dir, "cases_filtered_"); got != 1 {
			t.Fatalf("expected 1 cases export file, got %d", got)
		}
		if got := countFilesByPrefix(t, dir, "sections_all_"); got != 1 {
			t.Fatalf("expected 1 sections export file, got %d", got)
		}
	})

	t.Run("export mapping", func(t *testing.T) {
		dir := t.TempDir()
		m.mapping = NewSharedStepMapping(1, 2)
		m.mapping.AddPair(10, 20, "existing")

		if err := m.ExportMapping(dir); err != nil {
			t.Fatalf("ExportMapping() error = %v", err)
		}
		if got := countFilesByPrefix(t, dir, "mapping_"); got != 1 {
			t.Fatalf("expected 1 mapping export file, got %d", got)
		}
	})
}

func TestMigrationLogHelpers(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})

	// Verify helper methods are callable and don't panic.
	m.LogInfo("info", "k1", "v1")
	m.LogWarn("warn", "k2", 42)
	m.LogError("err", os.ErrNotExist, "k3", true)
}

func TestMigrationLoadMappingFromFile(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})

	t.Run("full format", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "mapping_full.json")
		sm := NewSharedStepMapping(1, 2)
		sm.AddPair(11, 111, "existing")

		payload, err := json.Marshal(sm)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		if err := os.WriteFile(file, payload, 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		m.mapping = NewSharedStepMapping(1, 2)
		if err := m.LoadMappingFromFile(file); err != nil {
			t.Fatalf("LoadMappingFromFile() error = %v", err)
		}
		if got, ok := m.mapping.GetTargetBySource(11); !ok || got != 111 {
			t.Fatalf("expected mapping 11 -> 111, got %d, ok=%v", got, ok)
		}
	})

	t.Run("simple map format", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "mapping_simple.json")
		if err := os.WriteFile(file, []byte(`{"12":212,"13":213}`), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		m.mapping = NewSharedStepMapping(1, 2)
		if err := m.LoadMappingFromFile(file); err != nil {
			t.Fatalf("LoadMappingFromFile() error = %v", err)
		}
		if got, ok := m.mapping.GetTargetBySource(12); !ok || got != 212 {
			t.Fatalf("expected mapping 12 -> 212, got %d, ok=%v", got, ok)
		}
	})

	t.Run("wrapper pairs format", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "mapping_wrapper.json")
		wrapper := []byte(`{"pairs":[{"source_id":21,"target_id":221}]}`)
		if err := os.WriteFile(file, wrapper, 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		m.mapping = NewSharedStepMapping(1, 2)
		if err := m.LoadMappingFromFile(file); err != nil {
			t.Fatalf("LoadMappingFromFile() error = %v", err)
		}
		if got, ok := m.mapping.GetTargetBySource(21); !ok || got != 221 {
			t.Fatalf("expected mapping 21 -> 221, got %d, ok=%v", got, ok)
		}
	})

	t.Run("invalid file", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "invalid.json")
		if err := os.WriteFile(file, []byte(`{invalid`), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		m.mapping = NewSharedStepMapping(1, 2)
		if err := m.LoadMappingFromFile(file); err == nil {
			t.Fatalf("expected error for invalid mapping format")
		}
	})
}
