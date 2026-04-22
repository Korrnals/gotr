// internal/migration/filter.go
package migration

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Korrnals/gotr/internal/models/data"
)

// FilterStats holds summary statistics from the last filter operation.
type FilterStats struct {
	Source     int // total items in source
	Target     int // total items in target
	Duplicates int // items already existing in target (skipped)
	New        int // items ready for import
}

// LastFilterStats returns the statistics from the last Filter* call.
func (m *Migration) LastFilterStats() FilterStats {
	return m.lastFilterStats
}

// FilterSharedSteps filters shared steps by usage in suite and duplicates in target.
// Candidates are shared steps whose IDs appear in usedStepIDs (collected from case steps).
// Duplicates are added to the mapping with status "existing".
// New (non-duplicate) steps are returned for import.
func (m *Migration) FilterSharedSteps(source, target data.GetSharedStepsResponse, usedStepIDs map[int64]struct{}) (filtered data.GetSharedStepsResponse, err error) {
	m.logger.Info("Starting shared steps filtering by usage in suite")

	var candidates data.GetSharedStepsResponse
	for _, step := range source {
		if _, ok := usedStepIDs[step.ID]; ok {
			candidates = append(candidates, step)
		}
	}
	m.logger.Infow("Found candidates for transfer (used in suite)", "count", len(candidates))

	m.logger.Info("Checking for duplicates in target project")
	targetMap := make(map[string]int64)
	for _, t := range target {
		val := fieldValue(t, m.compareField)
		if val != "" {
			targetMap[val] = t.ID
		}
	}

	for _, step := range candidates {
		val := fieldValue(step, m.compareField)
		if existingID, ok := targetMap[val]; ok {
			m.mapping.AddPair(step.ID, existingID, "existing")
			m.logger.Debugw("Duplicate found — added to mapping", "title", step.Title, "old_id", step.ID, "existing_id", existingID)
		} else {
			filtered = append(filtered, step)
		}
	}

	duplicates := len(candidates) - len(filtered)
	m.lastFilterStats = FilterStats{Source: len(source), Target: len(target), Duplicates: duplicates, New: len(filtered)}
	m.logger.Infow("Ready to import new shared steps", "count", len(filtered))
	return filtered, nil
}

// FilterSuites filters suites by duplicate detection (by name).
// Duplicates are added to the mapping with status "existing".
// New (non-duplicate) suites are returned for import.
func (m *Migration) FilterSuites(source, target data.GetSuitesResponse) (filtered data.GetSuitesResponse, err error) {
	m.logger.Info("Starting suites filtering by duplicates (by name)")

	targetMap := make(map[string]int64)
	for _, t := range target {
		if t.Name != "" {
			targetMap[t.Name] = t.ID
		}
	}

	for _, s := range source {
		if existingID, ok := targetMap[s.Name]; ok {
			m.mapping.AddPair(s.ID, existingID, "existing")
			m.logger.Debugw("Duplicate suite found — added to mapping", "name", s.Name, "old_id", s.ID, "existing_id", existingID)
		} else {
			filtered = append(filtered, s)
		}
	}

	duplicates := len(source) - len(filtered)
	m.lastFilterStats = FilterStats{Source: len(source), Target: len(target), Duplicates: duplicates, New: len(filtered)}
	m.logger.Infow("Ready to import new suites", "count", len(filtered))
	return filtered, nil
}

// FilterCases filters cases by duplicate detection (using compareField).
func (m *Migration) FilterCases(source, target data.GetCasesResponse) (filtered data.GetCasesResponse, err error) {
	m.logger.Info("Starting cases filtering by duplicates")

	targetMap := make(map[string]int64)
	for _, t := range target {
		val := fieldValue(t, m.compareField)
		if val != "" {
			targetMap[val] = t.ID
		}
	}

	for _, c := range source {
		val := fieldValue(c, m.compareField)
		if _, exists := targetMap[val]; !exists {
			filtered = append(filtered, c)
		} else {
			m.logger.Debugw("Duplicate case found — skipped", "title", c.Title)
		}
	}

	duplicates := len(source) - len(filtered)
	m.lastFilterStats = FilterStats{Source: len(source), Target: len(target), Duplicates: duplicates, New: len(filtered)}
	m.logger.Infow("Ready to import new cases", "count", len(filtered))
	return filtered, nil
}

func fieldValue(obj interface{}, field string) string {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() {
		return ""
	}

	f := v.FieldByName(field)
	if f.IsValid() {
		return fmt.Sprintf("%v", f.Interface())
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if strings.EqualFold(t.Field(i).Name, field) {
			f = v.Field(i)
			if f.IsValid() {
				return fmt.Sprintf("%v", f.Interface())
			}
		}
	}

	return ""
}

// FilterSections filters sections by duplicate detection in the target suite (by name).
func (m *Migration) FilterSections(source, target data.GetSectionsResponse) (filtered data.GetSectionsResponse, err error) {
	m.logger.Info("Starting sections filtering by duplicates (by name in suite)")

	targetMap := make(map[string]int64)
	for _, t := range target {
		if t.Name != "" {
			targetMap[t.Name] = t.ID
		}
	}

	for _, s := range source {
		if existingID, ok := targetMap[s.Name]; ok {
			m.mapping.AddPair(s.ID, existingID, "existing")
			m.logger.Debugw("Duplicate section found — mapping added", "name", s.Name, "old_id", s.ID, "existing_id", existingID)
		} else {
			filtered = append(filtered, s)
		}
	}

	duplicates := len(source) - len(filtered)
	m.lastFilterStats = FilterStats{Source: len(source), Target: len(target), Duplicates: duplicates, New: len(filtered)}
	m.logger.Infow("Ready to import new sections", "count", len(filtered))
	return filtered, nil
}
