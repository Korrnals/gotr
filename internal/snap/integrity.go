// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package snap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// IntegritySchemaVersion is the on-disk schema for integrity.json.
const IntegritySchemaVersion = 2

// IntegrityIndex is the top-level Merkle integrity manifest written
// next to attachments.json/references.json. The file enumerates
// every artifact in the snapshot directory together with its SHA-256
// and a Merkle root computed over the sorted "path|sha256\n" lines.
type IntegrityIndex struct {
	SchemaVersion int              `json:"schema_version"`
	GeneratedAt   time.Time        `json:"generated_at"`
	SnapID        string           `json:"snap_id"`
	Files         []IntegrityEntry `json:"files"`
	Root          string           `json:"merkle_root"`
}

// IntegrityEntry is a single file enumerated under IntegrityIndex.Files.
type IntegrityEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// integrityIgnore lists basenames inside the snap dir that are not
// considered part of the integrity-protected artifact set. integrity.json
// itself is excluded so its hash isn't self-referential.
var integrityIgnore = map[string]bool{
	"integrity.json": true,
}

// BuildIntegrityIndex walks <snap>/ and computes a SHA-256 for every
// regular file, then derives a Merkle root over the sorted set. The
// returned index is the on-disk shape of integrity.json. Files are
// reported as paths relative to <snap>/ with forward slashes for
// cross-platform stability.
func BuildIntegrityIndex(store *Store, snapID string) (*IntegrityIndex, error) {
	root := store.SnapDir(snapID)
	idx := &IntegrityIndex{
		SchemaVersion: IntegritySchemaVersion,
		GeneratedAt:   time.Now().UTC(),
		SnapID:        snapID,
	}
	err := filepath.Walk(root, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if integrityIgnore[info.Name()] {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		sum, err := sha256OfFile(p)
		if err != nil {
			return fmt.Errorf("hash %s: %w", rel, err)
		}
		idx.Files = append(idx.Files, IntegrityEntry{Path: rel, SHA256: sum, Size: info.Size()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(idx.Files, func(i, j int) bool { return idx.Files[i].Path < idx.Files[j].Path })
	idx.Root = computeMerkleRoot(idx.Files)
	return idx, nil
}

// WriteIntegrityIndex builds the index and persists it as
// <snap>/integrity.json (atomic temp+rename via the snap store).
func WriteIntegrityIndex(store *Store, snapID string) (*IntegrityIndex, error) {
	idx, err := BuildIntegrityIndex(store, snapID)
	if err != nil {
		return nil, err
	}
	if _, err := store.SaveData(snapID, "integrity.json", idx); err != nil {
		return nil, fmt.Errorf("save integrity.json: %w", err)
	}
	return idx, nil
}

// VerifyIntegrityIndex reads <snap>/integrity.json and re-hashes every
// listed file. It returns the first mismatching entry as an error, or
// nil if every file is intact and the Merkle root matches.
func VerifyIntegrityIndex(store *Store, snapID string) error {
	var on IntegrityIndex
	if err := store.LoadData(snapID, "integrity.json", &on); err != nil {
		return err
	}
	root := store.SnapDir(snapID)
	for _, e := range on.Files {
		full := filepath.Join(root, filepath.FromSlash(e.Path))
		got, err := sha256OfFile(full)
		if err != nil {
			return fmt.Errorf("verify %s: %w", e.Path, err)
		}
		if got != e.SHA256 {
			return fmt.Errorf("integrity mismatch for %s: got %s, want %s", e.Path, got, e.SHA256)
		}
	}
	if want := computeMerkleRoot(on.Files); want != on.Root {
		return fmt.Errorf("integrity merkle_root mismatch: got %s, want %s", want, on.Root)
	}
	return nil
}

// sha256OfFile returns the lowercase hex SHA-256 of the file at path.
func sha256OfFile(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // path computed from store-rooted walk.
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// computeMerkleRoot folds the sorted file list into a single SHA-256
// over the "path|sha256\n" lines. It is not a tree-Merkle (tree depth
// gives no benefit at this scale) but the name reflects the contract:
// any change to any file changes the root.
func computeMerkleRoot(entries []IntegrityEntry) string {
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s|%s\n", e.Path, e.SHA256)
	}
	return hex.EncodeToString(h.Sum(nil))
}
