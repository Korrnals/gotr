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

// MappingPair — пара source → target ID с метаданными
type MappingPair struct {
	SourceID  int64     `json:"source_id"`
	TargetID  int64     `json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // "created" или "existing"
}

// SharedStepMapping — полная структура mapping с контекстом проектов
type SharedStepMapping struct {
	SrcProjectID int64         `json:"src_project_id"`
	DstProjectID int64         `json:"dst_project_id"`
	CreatedAt    time.Time     `json:"created_at"`
	Count        int           `json:"count"`
	Pairs        []MappingPair `json:"pairs"`

	index map[int64]int64 // быстрый поиск (не экспортируется)
}

// NewSharedStepMapping — конструктор
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

// AddPair — добавляет пару
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

// GetTargetBySource — поиск target по source
func (sm *SharedStepMapping) GetTargetBySource(sourceID int64) (int64, bool) {
	targetID, ok := sm.index[sourceID]
	return targetID, ok
}

// SortPairs — сортировка для экспорта
func (sm *SharedStepMapping) SortPairs() {
	sort.Slice(sm.Pairs, func(i, j int) bool {
		return sm.Pairs[i].SourceID < sm.Pairs[j].SourceID
	})
}

// Save — сохраняет в JSON
func (sm *SharedStepMapping) Save(dir string) error {
	if sm.Count == 0 {
		return nil
	}

	if dir == "" {
		dir = ".testrail"
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	sm.SortPairs()

	file := filepath.Join(dir, fmt.Sprintf("mapping_%s.json", time.Now().Format("2006-01-02_15-04-05")))
	data := sm // marshal всей структуры

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, jsonData, 0644)
}

// LoadSharedStepMapping — загружает маппу общих шагов из файла
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
