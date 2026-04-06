package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// LoadMappingFromFile loads a mapping from a file and populates m.mapping.
// Supports two formats: full SharedStepMapping (pairs) and simple key-value JSON.
func (m *Migration) LoadMappingFromFile(file string) error {
	// Try the full mapping format first (with pairs)
	sm, err := LoadSharedStepMapping(file)
	if err == nil && sm != nil && len(sm.Pairs) > 0 {
		m.mapping = sm
		return nil
	}

	// Otherwise try the simple key-value format
	simple, err := loadSimpleMapping(file)
	if err == nil && len(simple) > 0 {
		for s, t := range simple {
			m.mapping.AddPair(s, t, "existing")
		}
		return nil
	}

	return fmt.Errorf("failed to load mapping from file %s", file)
}

func loadSimpleMapping(file string) (map[int64]int64, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var m1 map[string]int64
	if err := json.Unmarshal(data, &m1); err == nil && len(m1) > 0 {
		res := make(map[int64]int64, len(m1))
		for k, v := range m1 {
			id, parseErr := strconv.ParseInt(k, 10, 64)
			if parseErr != nil {
				continue
			}
			res[id] = v
		}
		return res, nil
	}

	var wrapper struct {
		Pairs []struct {
			SourceID int64 `json:"source_id"`
			TargetID int64 `json:"target_id"`
		} `json:"pairs"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Pairs) > 0 {
		res := make(map[int64]int64, len(wrapper.Pairs))
		for _, p := range wrapper.Pairs {
			res[p.SourceID] = p.TargetID
		}
		return res, nil
	}

	return map[int64]int64{}, nil
}
