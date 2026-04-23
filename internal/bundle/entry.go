package bundle

import (
	"io/fs"
	"time"
)

// Entry describes one archive entry queued for writing.
//
// Exactly one of SourcePath or Content must be non-zero:
//   - SourcePath != ""  → file read from the local filesystem
//   - Content   != nil  → in-memory bytes (useful for manifest/README/SHA256SUMS)
type Entry struct {
	// ArchivePath is the forward-slash path inside the archive. Must not be
	// empty and must not start with "/".
	ArchivePath string
	// RelHome is the original location in ~/.gotr form; stored into the
	// manifest. May be empty for virtual entries (manifest, README, SHA256SUMS).
	RelHome string

	SourcePath string
	Content    []byte

	Mode    fs.FileMode
	ModTime time.Time
}

// defaultMode is the mode used when Entry.Mode is zero.
const defaultMode fs.FileMode = 0o644

// defaultTime returns a stable deterministic timestamp used when
// Entry.ModTime is zero so bundles are reproducible.
func defaultTime() time.Time { return time.Unix(0, 0).UTC() }
