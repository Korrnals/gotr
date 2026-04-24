// Package state provides a lightweight JSON-backed store under ~/.gotr for
// persistent CLI state (one-time hint flags, feature migration breadcrumbs).
//
// The store is best-effort: read/write failures are surfaced as errors but
// should typically be treated as non-fatal by callers so that transient FS
// issues never block the user's primary command.
package state

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/paths"
)

// FileName is the name of the state file under ~/.gotr.
const FileName = "state.json"

// State captures persistent CLI flags. Fields are additive; unknown fields in
// existing files are preserved across round-trips via json.RawMessage.
type State struct {
	// FlatLayoutWarned is set to true after the first-time hint about the
	// legacy flat ~/.gotr/reports/ layout has been shown to the user.
	FlatLayoutWarned bool `json:"flat_layout_warned,omitempty"`
}

// Path returns the absolute path to ~/.gotr/state.json.
func Path() (string, error) {
	base, err := paths.BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, FileName), nil
}

// Load reads the state file; a missing file yields a zero-valued State and
// no error so first-run callers can write without a separate existence check.
func Load() (*State, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p) //nolint:gosec // path resolved via paths.BaseDir
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &State{}, nil
		}
		return nil, err
	}
	s := &State{}
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Save writes the state file atomically (write+rename). Creates ~/.gotr
// if necessary.
func Save(s *State) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
