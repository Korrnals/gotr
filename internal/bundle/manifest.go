package bundle

import (
	"encoding/json"
	"fmt"
	"time"
)

// SchemaVersion is the current bundle schema version. Bumped on any
// incompatible manifest change.
const SchemaVersion = 1

// Bundle layout filenames (forward-slash relative paths inside archives).
const (
	ManifestName  = "manifest.json"
	ChecksumsName = "SHA256SUMS"
	ReadmeName    = "README.txt"
)

// Kind identifies the type of payload carried by a bundle.
type Kind string

// Supported bundle kinds.
const (
	KindSnap    Kind = "snap"
	KindReports Kind = "reports"
)

// File describes one archived payload file (manifest/checksums/readme
// themselves are NOT listed here — only the user payload).
type File struct {
	// Path is the file path inside the archive (forward-slash, no leading
	// slash). For snap bundles this is `snaps/<id>/...`; for reports it is
	// `reports/<basename>`.
	Path string `json:"path"`
	// RelHome is the original location in ~/.gotr form, e.g.
	// "~/.gotr/snaps/sync/20260413T143000_full_p48_to_p49/meta.json".
	RelHome string `json:"rel_home"`
	// SHA256 is the hex-encoded SHA-256 of the file's bytes.
	SHA256 string `json:"sha256"`
	// Size is the uncompressed byte size.
	Size int64 `json:"size"`
}

// Manifest is the top-level bundle metadata stored at manifest.json.
type Manifest struct {
	SchemaVersion int       `json:"schema_version"`
	Kind          Kind      `json:"kind"`
	GotrVersion   string    `json:"gotr_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	Files         []File    `json:"files"`

	// Snap-specific (when Kind == KindSnap with a single snapshot).
	SnapID   string   `json:"snap_id,omitempty"`
	Redacted []string `json:"redacted_fields,omitempty"`

	// Reports-specific. Left as zero value when irrelevant.
	ReportCount int `json:"report_count,omitempty"`
}

// MarshalJSON returns stable indented JSON so manifests diff nicely.
func (m *Manifest) MarshalJSON() ([]byte, error) {
	type alias Manifest
	return json.MarshalIndent((*alias)(m), "", "  ")
}

// ValidateSchema returns an error when the manifest's schema version is not
// supported by the current build. The error message is actionable.
func (m *Manifest) ValidateSchema() error {
	if m.SchemaVersion == 0 {
		return fmt.Errorf("bundle manifest: missing schema_version (corrupt or pre-v3.3 archive)")
	}
	if m.SchemaVersion > SchemaVersion {
		return fmt.Errorf("bundle manifest: schema version %d is newer than supported %d — upgrade gotr", m.SchemaVersion, SchemaVersion)
	}
	return nil
}
