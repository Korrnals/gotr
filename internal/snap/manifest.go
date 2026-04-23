package snap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// ManifestFile is the name of the manifest index file.
const ManifestFile = "manifest.json"

// ManifestEntry is a lightweight index entry for a snapshot.
type ManifestEntry struct {
	ID              string    `json:"id"`
	Name            string    `json:"name,omitempty"`
	Label           string    `json:"label,omitempty"`
	ServerURL       string    `json:"server_url,omitempty"`
	Category        Category  `json:"category"`
	Operation       Operation `json:"operation"`
	EntityType      string    `json:"entity_type"`
	EntityIDs       []int64   `json:"entity_ids,omitempty"`
	ProjectID       int64     `json:"project_id,omitempty"`
	SourceProjectID int64     `json:"source_project_id,omitempty"`
	DstProjectID    int64     `json:"dst_project_id,omitempty"`
	DstSuiteID      int64     `json:"dst_suite_id,omitempty"`
	RollbackTier    Tier      `json:"rollback_tier"`
	Status          Status    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
	DataSize        int64     `json:"data_size_bytes"`
}

// Manifest is the top-level snapshot index stored in manifest.json.
type Manifest struct {
	mu      sync.Mutex
	path    string
	Entries []ManifestEntry `json:"entries"`
}

// LoadManifest reads or creates the manifest at the store root.
func LoadManifest(store *Store) (*Manifest, error) {
	p := filepath.Join(store.BaseDir(), ManifestFile)
	m := &Manifest{path: p}

	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			m.Entries = []ManifestEntry{}
			return m, nil
		}
		return nil, fmt.Errorf("snap: open manifest: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(m); err != nil {
		if errors.Is(err, io.EOF) {
			m.Entries = []ManifestEntry{}
			return m, nil
		}
		return nil, fmt.Errorf("snap: decode manifest: %w", err)
	}
	return m, nil
}

// Add inserts a new entry into the manifest and saves to disk.
func (m *Manifest) Add(meta *Meta) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := ManifestEntry{
		ID:              meta.ID,
		Name:            meta.Name,
		Label:           meta.Label,
		ServerURL:       meta.ServerURL,
		Category:        meta.Category,
		Operation:       meta.Operation,
		EntityType:      meta.EntityType,
		EntityIDs:       meta.EntityIDs,
		ProjectID:       meta.ProjectID,
		SourceProjectID: meta.SourceProjectID,
		DstProjectID:    meta.DstProjectID,
		DstSuiteID:      meta.DstSuiteID,
		RollbackTier:    meta.RollbackTier,
		Status:          meta.Status,
		Timestamp:       meta.Timestamp,
		DataSize:        meta.DataSizeBytes,
	}

	m.Entries = append(m.Entries, entry)
	return m.save()
}

// Remove deletes an entry by ID and saves to disk.
func (m *Manifest) Remove(snapID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filtered := m.Entries[:0]
	for _, e := range m.Entries {
		if e.ID != snapID {
			filtered = append(filtered, e)
		}
	}
	m.Entries = filtered
	return m.save()
}

// UpdateStatus changes the status of an entry and saves.
func (m *Manifest) UpdateStatus(snapID string, status Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.Entries {
		if m.Entries[i].ID == snapID {
			m.Entries[i].Status = status
			return m.save()
		}
	}
	return fmt.Errorf("snap: entry %q not found in manifest", snapID)
}

// UpdateLabel changes the label of an entry and saves.
func (m *Manifest) UpdateLabel(snapID, label string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.Entries {
		if m.Entries[i].ID == snapID {
			m.Entries[i].Label = label
			return m.save()
		}
	}
	return fmt.Errorf("snap: entry %q not found in manifest", snapID)
}

// UpdateDataSize refreshes the data_size_bytes field for an entry. Used when
// the data.json is written after the initial manifest entry (e.g. sync flows
// that save their payload only after the mutating phase succeeds).
func (m *Manifest) UpdateDataSize(snapID string, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.Entries {
		if m.Entries[i].ID == snapID {
			m.Entries[i].DataSize = size
			return m.save()
		}
	}
	return fmt.Errorf("snap: entry %q not found in manifest", snapID)
}

// Find returns the entry with the given ID or name, or nil.
func (m *Manifest) Find(idOrName string) *ManifestEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.Entries {
		if m.Entries[i].ID == idOrName || m.Entries[i].Name == idOrName {
			return &m.Entries[i]
		}
	}
	return nil
}

// Latest returns the most recent entry, or nil if empty.
func (m *Manifest) Latest() *ManifestEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.Entries) == 0 {
		return nil
	}

	sorted := make([]ManifestEntry, len(m.Entries))
	copy(sorted, m.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})
	return &sorted[0]
}

// ListByEntity returns entries filtered by entity type.
func (m *Manifest) ListByEntity(entityType string) []ManifestEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []ManifestEntry
	for _, e := range m.Entries {
		if e.EntityType == entityType {
			result = append(result, e)
		}
	}
	return result
}

// ListByStatus returns entries filtered by status.
func (m *Manifest) ListByStatus(status Status) []ManifestEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []ManifestEntry
	for _, e := range m.Entries {
		if e.Status == status {
			result = append(result, e)
		}
	}
	return result
}

// ListByCategory returns entries filtered by category.
func (m *Manifest) ListByCategory(cat Category) []ManifestEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []ManifestEntry
	for _, e := range m.Entries {
		if e.Category == cat {
			result = append(result, e)
		}
	}
	return result
}

// ManifestIDs returns all entry IDs as a set (for orphan detection).
func (m *Manifest) ManifestIDs() map[string]struct{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make(map[string]struct{}, len(m.Entries))
	for _, e := range m.Entries {
		ids[e.ID] = struct{}{}
	}
	return ids
}

// Len returns the number of entries.
func (m *Manifest) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Entries)
}

// All returns a copy of all entries (thread-safe).
func (m *Manifest) All() []ManifestEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]ManifestEntry, len(m.Entries))
	copy(result, m.Entries)
	return result
}

func (m *Manifest) save() error {
	return atomicWriteJSON(m.path, m)
}
