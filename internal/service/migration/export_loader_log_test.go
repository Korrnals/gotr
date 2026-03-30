package migration

import (
	"context"
	"encoding/json"
	"errors"
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

func TestLoadSimpleMapping_TableDriven(t *testing.T) {
	t.Run("read file error", func(t *testing.T) {
		_, err := loadSimpleMapping(filepath.Join(t.TempDir(), "missing.json"))
		if err == nil {
			t.Fatalf("expected read error for missing file")
		}
	})

	t.Run("simple map ignores invalid keys", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "simple.json")
		if err := os.WriteFile(file, []byte(`{"11":111,"bad":999}`), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := loadSimpleMapping(file)
		if err != nil {
			t.Fatalf("loadSimpleMapping() error = %v", err)
		}
		if len(got) != 1 || got[11] != 111 {
			t.Fatalf("unexpected mapping result: %#v", got)
		}
	})

	t.Run("simple map with only invalid keys returns empty map", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "invalid_keys.json")
		if err := os.WriteFile(file, []byte(`{"bad":123}`), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := loadSimpleMapping(file)
		if err != nil {
			t.Fatalf("loadSimpleMapping() error = %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty mapping, got %#v", got)
		}
	})

	t.Run("wrapper pairs format", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "wrapper.json")
		if err := os.WriteFile(file, []byte(`{"pairs":[{"source_id":21,"target_id":221},{"source_id":21,"target_id":222}]}`), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := loadSimpleMapping(file)
		if err != nil {
			t.Fatalf("loadSimpleMapping() error = %v", err)
		}
		if len(got) != 1 || got[21] != 222 {
			t.Fatalf("unexpected wrapper mapping: %#v", got)
		}
	})

	t.Run("invalid json returns empty mapping without error", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "invalid.json")
		if err := os.WriteFile(file, []byte(`{invalid`), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := loadSimpleMapping(file)
		if err != nil {
			t.Fatalf("loadSimpleMapping() unexpected error = %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty mapping, got %#v", got)
		}
	})
}

func TestMigrationFetchData_TableDriven(t *testing.T) {
	t.Run("FetchSharedStepsData source error", func(t *testing.T) {
		calls := 0
		mock := &MockClient{
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				calls++
				if p == 1 {
					return nil, errors.New("source_shared_steps_error")
				}
				return data.GetSharedStepsResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSharedStepsData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "source_shared_steps_error") {
			t.Fatalf("expected source error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on source error, got src=%#v dst=%#v", src, dst)
		}
		if calls != 1 {
			t.Fatalf("expected single source call, got %d", calls)
		}
	})

	t.Run("FetchSharedStepsData target error", func(t *testing.T) {
		mock := &MockClient{
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				if p == 1 {
					return data.GetSharedStepsResponse{{ID: 10, Title: "src"}}, nil
				}
				return nil, errors.New("target_shared_steps_error")
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSharedStepsData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "target_shared_steps_error") {
			t.Fatalf("expected target error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on target error, got src=%#v dst=%#v", src, dst)
		}
	})

	t.Run("FetchSharedStepsData success with empty responses", func(t *testing.T) {
		mock := &MockClient{
			GetSharedStepsFunc: func(ctx context.Context, p int64) (data.GetSharedStepsResponse, error) {
				return data.GetSharedStepsResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSharedStepsData(context.Background())
		if err != nil {
			t.Fatalf("FetchSharedStepsData() error = %v", err)
		}
		if len(src) != 0 || len(dst) != 0 {
			t.Fatalf("expected empty responses, got src=%d dst=%d", len(src), len(dst))
		}
	})

	t.Run("FetchCasesData source error", func(t *testing.T) {
		calls := 0
		mock := &MockClient{
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				calls++
				if p == 1 {
					return nil, errors.New("source_cases_error")
				}
				return data.GetCasesResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchCasesData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "source_cases_error") {
			t.Fatalf("expected source error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on source error, got src=%#v dst=%#v", src, dst)
		}
		if calls != 1 {
			t.Fatalf("expected single source call, got %d", calls)
		}
	})

	t.Run("FetchCasesData target error", func(t *testing.T) {
		mock := &MockClient{
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				if p == 1 {
					return data.GetCasesResponse{{ID: 101, Title: "src"}}, nil
				}
				return nil, errors.New("target_cases_error")
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchCasesData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "target_cases_error") {
			t.Fatalf("expected target error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on target error, got src=%#v dst=%#v", src, dst)
		}
	})

	t.Run("FetchCasesData success with empty responses", func(t *testing.T) {
		mock := &MockClient{
			GetCasesFunc: func(ctx context.Context, p, s, sec int64) (data.GetCasesResponse, error) {
				return data.GetCasesResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchCasesData(context.Background())
		if err != nil {
			t.Fatalf("FetchCasesData() error = %v", err)
		}
		if len(src) != 0 || len(dst) != 0 {
			t.Fatalf("expected empty responses, got src=%d dst=%d", len(src), len(dst))
		}
	})

	t.Run("FetchSuitesData source error", func(t *testing.T) {
		calls := 0
		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				calls++
				if p == 1 {
					return nil, errors.New("source_suites_error")
				}
				return data.GetSuitesResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSuitesData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "source_suites_error") {
			t.Fatalf("expected source error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on source error, got src=%#v dst=%#v", src, dst)
		}
		if calls != 1 {
			t.Fatalf("expected single source call, got %d", calls)
		}
	})

	t.Run("FetchSuitesData target error", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				if p == 1 {
					return data.GetSuitesResponse{{ID: 7, Name: "src"}}, nil
				}
				return nil, errors.New("target_suites_error")
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSuitesData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "target_suites_error") {
			t.Fatalf("expected target error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on target error, got src=%#v dst=%#v", src, dst)
		}
	})

	t.Run("FetchSuitesData success with empty responses", func(t *testing.T) {
		mock := &MockClient{
			GetSuitesFunc: func(ctx context.Context, p int64) (data.GetSuitesResponse, error) {
				return data.GetSuitesResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSuitesData(context.Background())
		if err != nil {
			t.Fatalf("FetchSuitesData() error = %v", err)
		}
		if len(src) != 0 || len(dst) != 0 {
			t.Fatalf("expected empty responses, got src=%d dst=%d", len(src), len(dst))
		}
	})

	t.Run("FetchSectionsData source error", func(t *testing.T) {
		calls := 0
		mock := &MockClient{
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				calls++
				if p == 1 {
					return nil, errors.New("source_sections_error")
				}
				return data.GetSectionsResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSectionsData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "source_sections_error") {
			t.Fatalf("expected source error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on source error, got src=%#v dst=%#v", src, dst)
		}
		if calls != 1 {
			t.Fatalf("expected single source call, got %d", calls)
		}
	})

	t.Run("FetchSectionsData target error", func(t *testing.T) {
		mock := &MockClient{
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				if p == 1 {
					return data.GetSectionsResponse{{ID: 8, Name: "src"}}, nil
				}
				return nil, errors.New("target_sections_error")
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSectionsData(context.Background())
		if err == nil || !strings.Contains(err.Error(), "target_sections_error") {
			t.Fatalf("expected target error, got err=%v", err)
		}
		if src != nil || dst != nil {
			t.Fatalf("expected nil responses on target error, got src=%#v dst=%#v", src, dst)
		}
	})

	t.Run("FetchSectionsData success with empty responses", func(t *testing.T) {
		mock := &MockClient{
			GetSectionsFunc: func(ctx context.Context, p, s int64) (data.GetSectionsResponse, error) {
				return data.GetSectionsResponse{}, nil
			},
		}

		m := setupTestMigration(t, mock)
		src, dst, err := m.FetchSectionsData(context.Background())
		if err != nil {
			t.Fatalf("FetchSectionsData() error = %v", err)
		}
		if len(src) != 0 || len(dst) != 0 {
			t.Fatalf("expected empty responses, got src=%d dst=%d", len(src), len(dst))
		}
	})
}

func TestMigrationExportSuitesAndSections_Hotspots(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})

	t.Run("default dir and filtered branch", func(t *testing.T) {
		wd := t.TempDir()
		t.Chdir(wd)

		err := m.ExportSuites(data.GetSuitesResponse{{ID: 1, Name: "Suite"}}, true, "")
		if err != nil {
			t.Fatalf("ExportSuites() error = %v", err)
		}

		err = m.ExportSections(data.GetSectionsResponse{{ID: 2, Name: "Section"}}, true, "")
		if err != nil {
			t.Fatalf("ExportSections() error = %v", err)
		}

		exportDir := filepath.Join(wd, ".testrail")
		if got := countFilesByPrefix(t, exportDir, "suites_filtered_"); got != 1 {
			t.Fatalf("expected 1 suites filtered export, got %d", got)
		}
		if got := countFilesByPrefix(t, exportDir, "sections_filtered_"); got != 1 {
			t.Fatalf("expected 1 sections filtered export, got %d", got)
		}
	})

	t.Run("mkdir failure when path points to file", func(t *testing.T) {
		dir := t.TempDir()
		notDir := filepath.Join(dir, "not_a_dir")
		if err := os.WriteFile(notDir, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		err := m.ExportSuites(data.GetSuitesResponse{{ID: 3, Name: "Suite"}}, false, notDir)
		if err == nil {
			t.Fatalf("expected ExportSuites() error for file path as dir")
		}

		err = m.ExportSections(data.GetSectionsResponse{{ID: 4, Name: "Section"}}, false, notDir)
		if err == nil {
			t.Fatalf("expected ExportSections() error for file path as dir")
		}
	})
}

func TestMigrationExportSharedStepsAndCases_Hotspots(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})

	t.Run("default dir and all branch", func(t *testing.T) {
		wd := t.TempDir()
		t.Chdir(wd)

		err := m.ExportSharedSteps(data.GetSharedStepsResponse{{ID: 11, Title: "Step"}}, false, "")
		if err != nil {
			t.Fatalf("ExportSharedSteps() error = %v", err)
		}

		err = m.ExportCases(data.GetCasesResponse{{ID: 22, Title: "Case"}}, false, "")
		if err != nil {
			t.Fatalf("ExportCases() error = %v", err)
		}

		exportDir := filepath.Join(wd, ".testrail")
		if got := countFilesByPrefix(t, exportDir, "shared_steps_all_"); got != 1 {
			t.Fatalf("expected 1 shared_steps all export, got %d", got)
		}
		if got := countFilesByPrefix(t, exportDir, "cases_all_"); got != 1 {
			t.Fatalf("expected 1 cases all export, got %d", got)
		}
	})

	t.Run("mkdir failure when path points to file", func(t *testing.T) {
		dir := t.TempDir()
		notDir := filepath.Join(dir, "not_a_dir")
		if err := os.WriteFile(notDir, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		err := m.ExportSharedSteps(data.GetSharedStepsResponse{{ID: 1, Title: "Step"}}, false, notDir)
		if err == nil {
			t.Fatalf("expected ExportSharedSteps() error for file path as dir")
		}

		err = m.ExportCases(data.GetCasesResponse{{ID: 2, Title: "Case"}}, false, notDir)
		if err == nil {
			t.Fatalf("expected ExportCases() error for file path as dir")
		}
	})

	t.Run("marshal error on invalid raw json", func(t *testing.T) {
		err := m.ExportSharedSteps(data.GetSharedStepsResponse{{
			ID:           3,
			Title:        "Broken",
			CustomFields: []byte("{invalid"),
		}}, true, t.TempDir())
		if err == nil {
			t.Fatalf("expected ExportSharedSteps() marshal error")
		}

		err = m.ExportCases(data.GetCasesResponse{{
			ID:           4,
			Title:        "Broken",
			CustomFields: []byte("{invalid"),
		}}, true, t.TempDir())
		if err == nil {
			t.Fatalf("expected ExportCases() marshal error")
		}
	})

	t.Run("write failure on read-only directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "readonly")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatalf("Chmod() error = %v", err)
		}
		defer func() {
			_ = os.Chmod(dir, 0o700)
		}()

		err := m.ExportSharedSteps(data.GetSharedStepsResponse{{ID: 5, Title: "Step"}}, false, dir)
		if err == nil {
			t.Fatalf("expected ExportSharedSteps() write error")
		}

		err = m.ExportCases(data.GetCasesResponse{{ID: 6, Title: "Case"}}, false, dir)
		if err == nil {
			t.Fatalf("expected ExportCases() write error")
		}
	})
}

func TestMigrationExportMapping_ErrorPath(t *testing.T) {
	m := setupTestMigration(t, &MockClient{})
	m.mapping = NewSharedStepMapping(1, 2)
	m.mapping.AddPair(10, 20, "existing")

	dir := t.TempDir()
	notDir := filepath.Join(dir, "not_a_dir")
	if err := os.WriteFile(notDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := m.ExportMapping(notDir); err == nil {
		t.Fatalf("expected ExportMapping() error for file path as dir")
	}
}
