package snap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/paths"
)

// Store manages snapshot files on disk with category-based directory layout.
type Store struct {
	baseDir string
}

// NewStore creates a Store rooted at ~/.gotr/snaps/.
func NewStore() (*Store, error) {
	dir, err := paths.SnapsDirPath()
	if err != nil {
		return nil, fmt.Errorf("snap store: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("snap store: cannot create directory: %w", err)
	}
	return &Store{baseDir: dir}, nil
}

// NewStoreAt creates a Store at a custom directory (for testing).
func NewStoreAt(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("snap store: cannot create directory: %w", err)
	}
	return &Store{baseDir: dir}, nil
}

// BaseDir returns the store root directory.
func (s *Store) BaseDir() string {
	return s.baseDir
}

// ResolveCategoryDir determines the storage subdirectory for a snapshot.
//
//	--snap-name set   → "custom"
//	sync operation    → "sync"
//	everything else   → entity_type + "s" (cases, sections, runs, ...)
func ResolveCategoryDir(meta *Meta) Category {
	if meta.Name != "" {
		return CatCustom
	}
	if meta.IsSyncOp() {
		return CatSync
	}
	return Category(meta.EntityType + "s")
}

// GenerateID creates a snapshot ID including the category prefix.
// Returns "category/timestamp_op_entityID" (e.g. "cases/20260413T143000_update_1234").
func GenerateID(meta *Meta) string {
	cat := ResolveCategoryDir(meta)
	meta.Category = cat

	if meta.Name != "" {
		return string(CatCustom) + "/" + sanitizeName(meta.Name)
	}

	ts := time.Now().UTC().Format("20060102T150405")
	var dirname string

	if meta.IsSyncOp() {
		opShort := strings.TrimPrefix(string(meta.Operation), "sync_")
		// For sync ops, naming reflects src→dst projects so that operators can
		// see at a glance which target a rollback would affect.
		// Prefer DstProjectID when present; fall back to legacy p<src>_p<proj>.
		switch {
		case meta.DstProjectID != 0:
			dirname = fmt.Sprintf("%s_%s_p%d_to_p%d", ts, opShort, meta.ProjectID, meta.DstProjectID)
		default:
			dirname = fmt.Sprintf("%s_%s_p%d_p%d", ts, opShort, meta.SourceProjectID, meta.ProjectID)
		}
	} else if len(meta.EntityIDs) == 1 {
		dirname = fmt.Sprintf("%s_%s_%d", ts, meta.Operation, meta.EntityIDs[0])
	} else {
		dirname = fmt.Sprintf("%s_%s_bulk_%d", ts, meta.Operation, len(meta.EntityIDs))
	}

	return string(cat) + "/" + dirname
}

// snapDir returns the full filesystem path for a snapshot ID (which includes category/).
func (s *Store) snapDir(snapID string) string {
	return filepath.Join(s.baseDir, filepath.FromSlash(snapID))
}

// SaveMeta writes meta.json into the snapshot directory using atomic write.
func (s *Store) SaveMeta(meta *Meta) error {
	dir := s.snapDir(meta.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("snap: create dir %s: %w", meta.ID, err)
	}
	return atomicWriteJSON(filepath.Join(dir, "meta.json"), meta)
}

// SaveData writes entity data to a named JSON file in the snapshot directory using atomic write.
func (s *Store) SaveData(snapID, filename string, data interface{}) (int64, error) {
	dir := s.snapDir(snapID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, fmt.Errorf("snap: create dir %s: %w", snapID, err)
	}

	p := filepath.Join(dir, filename)
	if err := atomicWriteJSON(p, data); err != nil {
		return 0, err
	}

	info, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// LoadMeta reads meta.json from a snapshot directory.
func (s *Store) LoadMeta(snapID string) (*Meta, error) {
	p := filepath.Join(s.snapDir(snapID), "meta.json")
	return readJSON[Meta](p, snapID)
}

// LoadData reads a named JSON file from a snapshot directory into dst.
func (s *Store) LoadData(snapID, filename string, dst interface{}) error {
	p := filepath.Join(s.snapDir(snapID), filename)
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("snap: open %s/%s: %w", snapID, filename, err)
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(dst)
}

// Delete removes a snapshot directory.
func (s *Store) Delete(snapID string) error {
	dir := s.snapDir(snapID)
	return os.RemoveAll(dir)
}

// List returns all snapshot IDs (category/dirname) by walking category directories.
func (s *Store) List() ([]string, error) {
	catEntries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("snap: list: %w", err)
	}

	var ids []string
	for _, cat := range catEntries {
		if !cat.IsDir() {
			continue
		}
		catPath := filepath.Join(s.baseDir, cat.Name())
		snapEntries, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}
		for _, snap := range snapEntries {
			if snap.IsDir() {
				ids = append(ids, cat.Name()+"/"+snap.Name())
			}
		}
	}
	return ids, nil
}

// Exists checks whether a snapshot directory exists.
func (s *Store) Exists(snapID string) bool {
	dir := s.snapDir(snapID)
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// SnapDir returns the full filesystem path for a snapshot ID.
func (s *Store) SnapDir(snapID string) string {
	return s.snapDir(snapID)
}

// Export copies meta.json and data files from a snapshot into a single JSON file.
// The exported file contains {"meta": {...}, "data": {...}} for portability.
func (s *Store) Export(snapID, outPath string) error {
	meta, err := s.LoadMeta(snapID)
	if err != nil {
		return fmt.Errorf("snap export: %w", err)
	}

	var data json.RawMessage
	if meta.DataFile != "" {
		p := filepath.Join(s.snapDir(snapID), meta.DataFile)
		raw, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("snap export: read data %s: %w", meta.DataFile, err)
		}
		data = raw
	}

	envelope := struct {
		Meta *Meta           `json:"meta"`
		Data json.RawMessage `json:"data,omitempty"`
	}{Meta: meta, Data: data}

	dir := filepath.Dir(outPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("snap export: create output dir: %w", err)
		}
	}

	return atomicWriteJSON(outPath, envelope)
}

// CollectOrphans returns snapshot IDs on disk that are not present in the manifest.
func (s *Store) CollectOrphans(manifestIDs map[string]struct{}) ([]string, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	var orphans []string
	for _, id := range all {
		if _, ok := manifestIDs[id]; !ok {
			orphans = append(orphans, id)
		}
	}
	return orphans, nil
}

// CleanOrphans removes snapshot directories not tracked in manifest.
func (s *Store) CleanOrphans(manifestIDs map[string]struct{}) (int, error) {
	orphans, err := s.CollectOrphans(manifestIDs)
	if err != nil {
		return 0, err
	}
	for _, id := range orphans {
		if err := s.Delete(id); err != nil {
			return 0, fmt.Errorf("snap: clean orphan %s: %w", id, err)
		}
	}
	return len(orphans), nil
}

// --- atomic write helpers ---

// atomicWriteJSON encodes data as indented JSON, writes to a .tmp file, then renames.
func atomicWriteJSON(path string, data interface{}) error {
	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("snap: create tmp %s: %w", tmp, err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("snap: encode %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("snap: close tmp %s: %w", tmp, err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("snap: rename %s: %w", path, err)
	}
	return nil
}

// readJSON is a generic helper to read and decode a JSON file.
func readJSON[T any](path, snapID string) (*T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("snap: open meta %s: %w", snapID, err)
	}
	defer f.Close()

	var v T
	if err := json.NewDecoder(f).Decode(&v); err != nil {
		return nil, fmt.Errorf("snap: decode meta %s: %w", snapID, err)
	}
	return &v, nil
}
