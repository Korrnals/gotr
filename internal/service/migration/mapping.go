// internal/migration/mapping.go
package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// MappingPair represents a source → target ID pair with metadata.
type MappingPair struct {
	SourceID  int64     `json:"source_id"`
	TargetID  int64     `json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // "created" or "existing"
}

// SharedStepMapping holds the full mapping structure with project context.
type SharedStepMapping struct {
	SrcProjectID int64         `json:"src_project_id"`
	DstProjectID int64         `json:"dst_project_id"`
	CreatedAt    time.Time     `json:"created_at"`
	Count        int           `json:"count"`
	Pairs        []MappingPair `json:"pairs"`

	index map[int64]int64 // fast lookup (not exported)
}

// NewSharedStepMapping creates a new SharedStepMapping for the given project pair.
func NewSharedStepMapping(srcProjectID, dstProjectID int64) *SharedStepMapping {
	return &SharedStepMapping{
		SrcProjectID: srcProjectID,
		DstProjectID: dstProjectID,
		CreatedAt:    time.Now(),
		Count:        0,
		Pairs:        make([]MappingPair, 0),
		index:        make(map[int64]int64),
	}
}

// AddPair adds a source→target ID pair to the mapping (skips duplicates).
func (sm *SharedStepMapping) AddPair(sourceID, targetID int64, status string) {
	if _, exists := sm.index[sourceID]; exists {
		return
	}

	sm.index[sourceID] = targetID
	sm.Pairs = append(sm.Pairs, MappingPair{
		SourceID:  sourceID,
		TargetID:  targetID,
		CreatedAt: time.Now(),
		Status:    status,
	})
	sm.Count++
}

// GetTargetBySource looks up a target ID by source ID.
func (sm *SharedStepMapping) GetTargetBySource(sourceID int64) (int64, bool) {
	targetID, ok := sm.index[sourceID]
	return targetID, ok
}

// SortPairs sorts pairs by source ID for consistent export output.
func (sm *SharedStepMapping) SortPairs() {
	sort.Slice(sm.Pairs, func(i, j int) bool {
		return sm.Pairs[i].SourceID < sm.Pairs[j].SourceID
	})
}

// Save writes the mapping to a JSON file in the given directory.
func (sm *SharedStepMapping) Save(dir string) error {
	if sm.Count == 0 {
		return nil
	}

	if dir == "" {
		dir = ".testrail"
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	sm.SortPairs()

	file := filepath.Join(dir, fmt.Sprintf("mapping_%s.json", time.Now().Format("2006-01-02_15-04-05")))
	data := sm // marshal the entire struct

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, jsonData, 0o644)
}

// LoadSharedStepMapping loads a shared step mapping from a JSON file.
func LoadSharedStepMapping(file string) (*SharedStepMapping, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var sm SharedStepMapping
	if err := json.Unmarshal(data, &sm); err != nil {
		return nil, err
	}

	sm.index = make(map[int64]int64)
	for _, p := range sm.Pairs {
		sm.index[p.SourceID] = p.TargetID
	}

	return &sm, nil
}
