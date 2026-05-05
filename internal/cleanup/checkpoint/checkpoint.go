// Package checkpoint provides a persistent, resumable state store for
// long-running attachment cleanups.
//
// A Checkpoint captures the per-project status of a `gotr attachments
// cleanup` run together with the filter fingerprint that produced it.
// The store writes each checkpoint atomically (tmp file + rename +
// fsync) under ~/.gotr/cache/cleanup-attachments/<run-id>/ so a SIGINT
// or process crash never leaves a half-written file behind, and a
// subsequent `--resume <run-id>` can pick up where the previous run
// left off.
package checkpoint

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Korrnals/gotr/internal/paths"
)

// Schema version embedded in every persisted Checkpoint. Bump only on
// breaking layout changes.
const Schema = 1

// File names within a run-id directory.
const (
	FileCheckpoint  = "checkpoint.json"
	FilePartialPlan = "partial-plan.json"
	subdir          = "cleanup-attachments"
)

// Project status values.
const (
	StatePending      = "pending"
	StateDone         = "done"
	StateFailed       = "failed"
	StateTimeout      = "timeout"
	StateRetryPending = "retry_pending"
)

// ErrCheckpointNotFound signals that the requested run-id has no
// persisted checkpoint on disk.
var ErrCheckpointNotFound = errors.New("checkpoint not found")

// ErrCheckpointMalformed wraps a JSON / IO failure encountered while
// loading an existing checkpoint.
var ErrCheckpointMalformed = errors.New("checkpoint malformed")

// FilterSnapshot captures the parts of the cleanup filter that must
// match between an original run and its --resume invocation.
type FilterSnapshot struct {
	OlderThanRaw          string   `json:"older_than_raw,omitempty"`
	EntityTypes           []string `json:"entity_types,omitempty"`
	Limit                 int      `json:"limit,omitempty"`
	Concurrency           int      `json:"concurrency,omitempty"`
	ScanStrategy          string   `json:"scan_strategy,omitempty"`
	ScanTimeoutPerProject string   `json:"scan_timeout_per_project,omitempty"`
}

// ProjectStatus tracks the lifecycle of a single project inside a run.
type ProjectStatus struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	State     string    `json:"state"`
	Reason    string    `json:"reason,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Found     int       `json:"found,omitempty"`
	Eligible  int       `json:"eligible,omitempty"`
	Bytes     int64     `json:"bytes,omitempty"`
}

// Checkpoint is the on-disk shape of a resumable cleanup run.
type Checkpoint struct {
	Schema      int             `json:"schema"`
	RunID       string          `json:"run_id"`
	StartedAt   time.Time       `json:"started_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CLIArgs     []string        `json:"cli_args,omitempty"`
	Filter      FilterSnapshot  `json:"filter"`
	AllProjects bool            `json:"all_projects"`
	ProjectIDs  []int64         `json:"project_ids,omitempty"`
	ChunkSize   int             `json:"chunk_size"`
	Projects    []ProjectStatus `json:"projects"`
}

// CheckpointSummary is the shape returned by Store.List.
type CheckpointSummary struct {
	RunID     string    `json:"run_id"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Total     int       `json:"total"`
	Done      int       `json:"done"`
	Failed    int       `json:"failed"`
	Pending   int       `json:"pending"`
	Timeout   int       `json:"timeout"`
}

// NewRunID returns a deterministic, sortable run identifier of the
// shape "20260505T123456-<6hex>". The timestamp is in UTC; the suffix
// is 3 random bytes (6 hex chars).
func NewRunID(now time.Time) string {
	var buf [3]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// rand.Read on Linux/macOS practically never fails. As an
		// absolute last resort fall back to time-based entropy so the
		// caller still gets a unique-ish ID.
		t := now.UnixNano()
		buf[0] = byte(t)
		buf[1] = byte(t >> 8)
		buf[2] = byte(t >> 16)
	}
	return now.UTC().Format("20060102T150405") + "-" + hex.EncodeToString(buf[:])
}

// Store is the disk-backed persistence layer for cleanup checkpoints.
// All writes are atomic.
type Store struct {
	root string
}

// NewStore returns a Store rooted at ~/.gotr/cache/cleanup-attachments/.
// The directory is created on demand; callers do not need to MkdirAll
// themselves.
func NewStore() (*Store, error) {
	cache, err := paths.CacheDirPath()
	if err != nil {
		return nil, fmt.Errorf("checkpoint: resolve cache dir: %w", err)
	}
	root := filepath.Join(cache, subdir)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("checkpoint: create root: %w", err)
	}
	return &Store{root: root}, nil
}

// NewStoreAt is the test-friendly constructor: it points the store at
// an explicit directory and skips the ~/.gotr resolution.
func NewStoreAt(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("checkpoint: create root: %w", err)
	}
	return &Store{root: root}, nil
}

// Root returns the directory all run-id subdirectories live under.
func (s *Store) Root() string { return s.root }

// runDir returns the per-run directory and ensures it exists.
func (s *Store) runDir(runID string) (string, error) {
	if runID == "" {
		return "", errors.New("checkpoint: empty run id")
	}
	dir := filepath.Join(s.root, runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("checkpoint: create run dir: %w", err)
	}
	return dir, nil
}

// Save persists the checkpoint atomically. UpdatedAt is set to now
// when the caller has not supplied a value.
func (s *Store) Save(runID string, cp *Checkpoint) error {
	if cp == nil {
		return errors.New("checkpoint: nil checkpoint")
	}
	if cp.Schema == 0 {
		cp.Schema = Schema
	}
	if cp.UpdatedAt.IsZero() {
		cp.UpdatedAt = time.Now().UTC()
	}
	dir, err := s.runDir(runID)
	if err != nil {
		return err
	}
	payload, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	return atomicWrite(filepath.Join(dir, FileCheckpoint), payload)
}

// Load reads and decodes the checkpoint for runID. It returns
// ErrCheckpointNotFound when the file is absent and
// ErrCheckpointMalformed (wrapping the cause) on JSON / IO failures.
func (s *Store) Load(runID string) (*Checkpoint, error) {
	if runID == "" {
		return nil, errors.New("checkpoint: empty run id")
	}
	path := filepath.Join(s.root, runID, FileCheckpoint)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrCheckpointNotFound, runID)
		}
		return nil, fmt.Errorf("%w: read: %w", ErrCheckpointMalformed, err)
	}
	var cp Checkpoint
	if err := json.Unmarshal(b, &cp); err != nil {
		return nil, fmt.Errorf("%w: decode: %w", ErrCheckpointMalformed, err)
	}
	return &cp, nil
}

// SavePartialPlan persists an arbitrary JSON-encodable value
// (typically *cleanup.Plan) atomically alongside the checkpoint. The
// package keeps the shape opaque to avoid an import cycle with
// internal/cleanup.
func (s *Store) SavePartialPlan(runID string, plan any) error {
	dir, err := s.runDir(runID)
	if err != nil {
		return err
	}
	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint: marshal partial plan: %w", err)
	}
	return atomicWrite(filepath.Join(dir, FilePartialPlan), payload)
}

// LoadPartialPlan reads partial-plan.json into dst (must be a pointer).
// Returns ErrCheckpointNotFound when no partial plan has been written
// yet for runID.
func (s *Store) LoadPartialPlan(runID string, dst any) error {
	path := filepath.Join(s.root, runID, FilePartialPlan)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s/%s", ErrCheckpointNotFound, runID, FilePartialPlan)
		}
		return fmt.Errorf("%w: read partial plan: %w", ErrCheckpointMalformed, err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return fmt.Errorf("%w: decode partial plan: %w", ErrCheckpointMalformed, err)
	}
	return nil
}

// Delete removes the run-id directory and every artifact in it.
func (s *Store) Delete(runID string) error {
	if runID == "" {
		return errors.New("checkpoint: empty run id")
	}
	dir := filepath.Join(s.root, runID)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("checkpoint: delete: %w", err)
	}
	return nil
}

// List returns a summary for every run-id directory under the store
// root, sorted by StartedAt descending (most recent first).
// Directories that are not valid checkpoints are skipped silently.
func (s *Store) List() ([]CheckpointSummary, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("checkpoint: list: %w", err)
	}
	out := make([]CheckpointSummary, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cp, err := s.Load(e.Name())
		if err != nil {
			continue
		}
		out = append(out, summarize(cp))
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	return out, nil
}

func summarize(cp *Checkpoint) CheckpointSummary {
	s := CheckpointSummary{
		RunID:     cp.RunID,
		StartedAt: cp.StartedAt,
		UpdatedAt: cp.UpdatedAt,
		Total:     len(cp.Projects),
	}
	for _, p := range cp.Projects {
		switch p.State {
		case StateDone:
			s.Done++
		case StateFailed:
			s.Failed++
		case StateTimeout:
			s.Timeout++
		case StatePending, StateRetryPending, "":
			s.Pending++
		}
	}
	return s
}

// atomicWrite persists data to path via a sibling tempfile +
// fsync + rename, guaranteeing that an interrupted write never
// truncates the destination.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("checkpoint: tempfile: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("checkpoint: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("checkpoint: fsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("checkpoint: close: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return fmt.Errorf("checkpoint: rename: %w", err)
	}
	return nil
}
