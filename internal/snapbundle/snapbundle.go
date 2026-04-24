// Package snapbundle implements portable tar.gz export/import of gotr
// snapshots. It builds on internal/bundle for archive mechanics and on
// internal/snap for on-disk snapshot layout.
package snapbundle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/bundle"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/snap"
)

// ExportOptions controls snapshot export behavior.
type ExportOptions struct {
	// GotrVersion is recorded into manifest.json for reader context.
	GotrVersion string
	// Redact turns on redaction of known sensitive fields in meta.json
	// (assignee_email, assignee, and top-level "email" keys). The list of
	// redacted field paths is reported back in the returned Result.
	Redact bool
}

// Result summarizes an export or import operation.
type Result struct {
	ArchivePath string
	SnapID      string
	Files       []bundle.File
	Redacted    []string
}

// ExportOne writes the given snapshot into destPath (a .tar.gz file). The
// snapshot directory at ~/.gotr/snaps/<id>/ is archived in full.
func ExportOne(store *snap.Store, snapID, destPath string, opts ExportOptions) (*Result, error) {
	if !store.Exists(snapID) {
		return nil, fmt.Errorf("snapbundle: snapshot %q does not exist", snapID)
	}
	entries, files, redacted, err := collectEntries(store, snapID, opts)
	if err != nil {
		return nil, err
	}

	manifest := &bundle.Manifest{
		SchemaVersion: bundle.SchemaVersion,
		Kind:          bundle.KindSnap,
		GotrVersion:   opts.GotrVersion,
		GeneratedAt:   time.Now().UTC(),
		SnapID:        snapID,
		Files:         files,
		Redacted:      redacted,
	}
	if err := appendMetaEntries(&entries, manifest); err != nil {
		return nil, err
	}

	if err := bundle.WriteTarGz(destPath, entries); err != nil {
		return nil, fmt.Errorf("snapbundle: write archive: %w", err)
	}
	return &Result{ArchivePath: destPath, SnapID: snapID, Files: files, Redacted: redacted}, nil
}

// collectEntries walks the snapshot directory and builds archive Entries +
// manifest File descriptors. It also applies redaction when requested.
func collectEntries(store *snap.Store, snapID string, opts ExportOptions) ([]bundle.Entry, []bundle.File, []string, error) {
	snapDir := store.SnapDir(snapID)
	home, _ := os.UserHomeDir()

	var entries []bundle.Entry
	var files []bundle.File
	var redacted []string

	err := filepath.WalkDir(snapDir, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(snapDir, p)
		if err != nil {
			return err
		}
		archivePath := "snaps/" + filepath.ToSlash(filepath.Join(snapID, rel))

		content, redFields, err := readAndMaybeRedact(p, rel, opts.Redact)
		if err != nil {
			return err
		}
		redacted = append(redacted, redFields...)

		sum := bundle.SHA256Bytes(content)
		entries = append(entries, bundle.Entry{
			ArchivePath: archivePath,
			RelHome:     relHome(home, p),
			Content:     content,
		})
		files = append(files, bundle.File{
			Path:    archivePath,
			RelHome: relHome(home, p),
			SHA256:  sum,
			Size:    int64(len(content)),
		})
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("snapbundle: walk %s: %w", snapDir, err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	sort.Slice(entries, func(i, j int) bool { return entries[i].ArchivePath < entries[j].ArchivePath })
	sort.Strings(redacted)
	redacted = dedupe(redacted)

	return entries, files, redacted, nil
}

// appendMetaEntries adds manifest.json, SHA256SUMS, and README.txt to entries.
func appendMetaEntries(entries *[]bundle.Entry, manifest *bundle.Manifest) error {
	manifestJSON, err := manifest.MarshalJSON()
	if err != nil {
		return fmt.Errorf("snapbundle: marshal manifest: %w", err)
	}
	sums := bundle.FormatSHA256Sums(manifest.Files)
	readme := readmeSnap(manifest)

	*entries = append(*entries,
		bundle.Entry{ArchivePath: bundle.ManifestName, Content: manifestJSON},
		bundle.Entry{ArchivePath: bundle.ChecksumsName, Content: []byte(sums)},
		bundle.Entry{ArchivePath: bundle.ReadmeName, Content: []byte(readme)},
	)
	return nil
}

func readAndMaybeRedact(path, rel string, redact bool) (content []byte, redacted []string, err error) {
	raw, err := os.ReadFile(path) //nolint:gosec // path enumerated from snap store
	if err != nil {
		return nil, nil, fmt.Errorf("snapbundle: read %s: %w", path, err)
	}
	if !redact {
		return raw, nil, nil
	}
	// Only attempt redaction on JSON files. meta.json is the main target.
	if filepath.Ext(rel) != ".json" {
		return raw, nil, nil
	}
	var obj any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return raw, nil, nil
	}
	fields := redactSensitive(obj, "")
	if len(fields) == 0 {
		return raw, nil, nil
	}
	out, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return raw, nil, nil
	}
	return out, fields, nil
}

// sensitiveKeys lists JSON object keys that redactSensitive strips. Values
// are replaced by "[redacted]" and the fully-qualified dotted path is
// returned so operators can audit what left the machine.
var sensitiveKeys = map[string]struct{}{
	"assignee":       {},
	"assignee_email": {},
	"assignee_name":  {},
	"email":          {},
}

func redactSensitive(node any, prefix string) []string {
	var out []string
	switch v := node.(type) {
	case map[string]any:
		for k, child := range v {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			if _, hit := sensitiveKeys[k]; hit && child != nil {
				v[k] = "[redacted]"
				out = append(out, p)
				continue
			}
			out = append(out, redactSensitive(child, p)...)
		}
	case []any:
		for i, child := range v {
			out = append(out, redactSensitive(child, fmt.Sprintf("%s[%d]", prefix, i))...)
		}
	}
	return out
}

func dedupe(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func relHome(home, p string) string {
	if home == "" {
		return p
	}
	if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + filepath.ToSlash(rel)
	}
	return p
}

func readmeSnap(m *bundle.Manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "gotr snapshot bundle\n")
	fmt.Fprintf(&b, "generated_at: %s\n", m.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "gotr_version: %s\n", m.GotrVersion)
	fmt.Fprintf(&b, "schema_version: %d\n", m.SchemaVersion)
	fmt.Fprintf(&b, "snap_id: %s\n", m.SnapID)
	fmt.Fprintf(&b, "files: %d\n", len(m.Files))
	if len(m.Redacted) > 0 {
		fmt.Fprintf(&b, "redacted_fields: %s\n", strings.Join(m.Redacted, ", "))
	}
	fmt.Fprintf(&b, "\nExtract with: tar -xzf <file>.tar.gz\n")
	fmt.Fprintf(&b, "Import with: gotr import snap <file>.tar.gz\n")
	return b.String()
}

// DefaultExportPath returns the conventional export destination for a
// snapshot, under ~/.gotr/exports/snaps/.
func DefaultExportPath(snapID, label string) (string, error) {
	dir, err := paths.EnsureExportsSnapsDirPath()
	if err != nil {
		return "", err
	}
	ts := time.Now().UTC().Format("20060102")
	slug := sanitizeFilename(snapID)
	if label != "" {
		slug += "_" + sanitizeFilename(label)
	}
	return filepath.Join(dir, fmt.Sprintf("snap_%s_%s.tar.gz", slug, ts)), nil
}

func sanitizeFilename(s string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return r.Replace(s)
}

// ImportOptions controls how a bundle is imported.
type ImportOptions struct {
	Overwrite bool
	RenameID  string // when set, import under this snap_id instead of the one in manifest
	DryRun    bool
}

// Import extracts a snap bundle from srcPath into the local snap store.
// Returns the effective snapshot ID and metadata about what was done.
func Import(store *snap.Store, srcPath string, opts ImportOptions) (*Result, error) {
	manifest, sums, err := peekManifest(srcPath)
	if err != nil {
		return nil, err
	}
	if manifest.Kind != bundle.KindSnap {
		return nil, fmt.Errorf("snapbundle: archive kind %q is not a snapshot bundle", manifest.Kind)
	}
	if err := verifyChecksums(manifest, sums); err != nil {
		return nil, err
	}
	targetID := effectiveSnapID(manifest.SnapID, opts.RenameID)
	if targetID == "" {
		return nil, errors.New("snapbundle: manifest is missing snap_id")
	}

	if store.Exists(targetID) {
		switch {
		case opts.DryRun:
			// fine — just report
		case opts.Overwrite:
			if err := backupExistingSnap(store, targetID); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("snapbundle: snapshot %q already exists; use --overwrite or --rename-id", targetID)
		}
	}

	result := &Result{ArchivePath: srcPath, SnapID: targetID, Files: manifest.Files, Redacted: manifest.Redacted}
	if opts.DryRun {
		return result, nil
	}

	if err := extractAndRelocate(srcPath, store, manifest.SnapID, targetID); err != nil {
		return nil, err
	}
	return result, nil
}

func peekManifest(srcPath string) (*bundle.Manifest, string, error) {
	tmp, err := os.MkdirTemp("", "gotr-snap-peek-*")
	if err != nil {
		return nil, "", fmt.Errorf("snapbundle: tmpdir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	if _, err := bundle.ReadTarGz(srcPath, tmp); err != nil {
		return nil, "", fmt.Errorf("snapbundle: extract for peek: %w", err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(tmp, bundle.ManifestName))
	if err != nil {
		return nil, "", fmt.Errorf("snapbundle: read manifest: %w", err)
	}
	var m bundle.Manifest
	if err := json.Unmarshal(manifestBytes, &m); err != nil {
		return nil, "", fmt.Errorf("snapbundle: parse manifest: %w", err)
	}
	if err := m.ValidateSchema(); err != nil {
		return nil, "", err
	}
	sumsBytes, _ := os.ReadFile(filepath.Join(tmp, bundle.ChecksumsName))
	return &m, string(sumsBytes), nil
}

func verifyChecksums(m *bundle.Manifest, sumsText string) error {
	if sumsText == "" {
		// No external SHA256SUMS file; trust manifest.
		return nil
	}
	parsed, err := bundle.ParseSHA256Sums(sumsText)
	if err != nil {
		return fmt.Errorf("snapbundle: bad SHA256SUMS: %w", err)
	}
	for _, f := range m.Files {
		want, ok := parsed[f.Path]
		if !ok {
			return fmt.Errorf("snapbundle: SHA256SUMS missing entry for %s", f.Path)
		}
		if want != f.SHA256 {
			return fmt.Errorf("snapbundle: checksum mismatch for %s (manifest=%s sums=%s)", f.Path, f.SHA256, want)
		}
	}
	return nil
}

func effectiveSnapID(manifestID, rename string) string {
	if rename != "" {
		return rename
	}
	return manifestID
}

// backupExistingSnap moves ~/.gotr/snaps/<id> to ~/.gotr/snaps/.trash/<id>_<ts>/
// so the user can recover a clobbered snapshot.
func backupExistingSnap(store *snap.Store, snapID string) error {
	trash := filepath.Join(store.BaseDir(), ".trash")
	if err := os.MkdirAll(trash, 0o755); err != nil {
		return fmt.Errorf("snapbundle: mkdir trash: %w", err)
	}
	src := store.SnapDir(snapID)
	ts := time.Now().UTC().Format("20060102T150405")
	dst := filepath.Join(trash, sanitizeFilename(snapID)+"_"+ts)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("snapbundle: mkdir backup parent: %w", err)
	}
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("snapbundle: backup existing snap: %w", err)
	}
	return nil
}

// extractAndRelocate unpacks the archive to a temp dir and moves the
// snapshot subtree into the store, optionally under a renamed id.
func extractAndRelocate(srcPath string, store *snap.Store, archiveID, targetID string) error {
	tmp, err := os.MkdirTemp("", "gotr-snap-import-*")
	if err != nil {
		return fmt.Errorf("snapbundle: tmpdir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	if _, err := bundle.ReadTarGz(srcPath, tmp); err != nil {
		return fmt.Errorf("snapbundle: extract: %w", err)
	}
	srcSnap := filepath.Join(tmp, "snaps", filepath.FromSlash(archiveID))
	if _, err := os.Stat(srcSnap); err != nil {
		return fmt.Errorf("snapbundle: extracted archive missing snaps/%s: %w", archiveID, err)
	}
	dst := store.SnapDir(targetID)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("snapbundle: mkdir target parent: %w", err)
	}
	// If target still exists (e.g. --rename-id target clashes unexpectedly), fail.
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("snapbundle: target %s already exists", dst)
	}
	if err := os.Rename(srcSnap, dst); err != nil {
		// os.Rename may fail across filesystems; fall back to copy+remove.
		if err := copyTree(srcSnap, dst); err != nil {
			return fmt.Errorf("snapbundle: relocate %s -> %s: %w", srcSnap, dst, err)
		}
	}
	if targetID != archiveID {
		if err := rewriteMetaID(dst, targetID); err != nil {
			return fmt.Errorf("snapbundle: rewrite meta.json id: %w", err)
		}
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(src, p)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, rerr := os.ReadFile(p) //nolint:gosec // source is our own tmp dir
		if rerr != nil {
			return rerr
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644) //nolint:gosec // restoring user-owned data
	})
}

func rewriteMetaID(dir, newID string) error {
	p := filepath.Join(dir, "meta.json")
	raw, err := os.ReadFile(p) //nolint:gosec // dir is inside snap store
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	m["id"] = newID
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, out, 0o644) //nolint:gosec // meta.json under trusted ~/.gotr/snaps path
}
