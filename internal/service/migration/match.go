// internal/service/migration/match.go
package migration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/Korrnals/gotr/internal/models/data"
)

// MatchKey is a composite comparable key used by Bucket to match source
// entities against target entities with multiset semantics.
//
// Scope carries either the destination section_id (for cases) or the
// destination parent section_id (for sections), or 0 when the kind does
// not need scope (suites, shared steps).
// Name holds the case title / section name / suite name / shared-step title.
// Hash is used only for shared steps (sha256 of steps payload) to avoid
// collapsing different shared steps that happen to share a title.
type MatchKey struct {
	Kind  string
	Scope int64
	Name  string
	Hash  string
}

// Bucket holds a multiset of target IDs grouped by MatchKey. Each call to
// ConsumeOne pops one ID from the bucket for the given key, so N source items
// with the same key match at most N target items — fixing the historical bug
// where a flat map[string]int64 collapsed duplicates and dropped source rows.
type Bucket struct {
	items map[MatchKey][]int64
}

// NewBucket returns an empty Bucket.
func NewBucket() *Bucket {
	return &Bucket{items: make(map[MatchKey][]int64)}
}

// Add appends id to the bucket for key.
func (b *Bucket) Add(key MatchKey, id int64) {
	b.items[key] = append(b.items[key], id)
}

// ConsumeOne pops the first id for key (FIFO). Returns (0,false) if no
// target id remains for that key.
func (b *Bucket) ConsumeOne(key MatchKey) (int64, bool) {
	ids, ok := b.items[key]
	if !ok || len(ids) == 0 {
		return 0, false
	}
	id := ids[0]
	b.items[key] = ids[1:]
	return id, true
}

// Len returns the number of distinct keys in the bucket.
func (b *Bucket) Len() int { return len(b.items) }

// TotalItems returns the total number of target ids across all keys.
func (b *Bucket) TotalItems() int {
	n := 0
	for _, ids := range b.items {
		n += len(ids)
	}
	return n
}

// caseMatchKey builds the MatchKey for a case given its destination
// section_id and the value of the compare field (defaults to title).
func caseMatchKey(dstSectionID int64, compareValue string) MatchKey {
	return MatchKey{Kind: "case", Scope: dstSectionID, Name: compareValue}
}

// sectionMatchKey builds the MatchKey for a section given its destination
// parent_id and the section name.
func sectionMatchKey(dstParentID int64, name string) MatchKey {
	return MatchKey{Kind: "section", Scope: dstParentID, Name: name}
}

// suiteMatchKey builds the MatchKey for a suite by name.
func suiteMatchKey(name string) MatchKey {
	return MatchKey{Kind: "suite", Name: name}
}

// sharedStepMatchKey builds the MatchKey for a shared step using both the
// compare value (usually title) and a stable hash of the step payload to
// keep shared steps with the same title but different bodies distinguishable.
func sharedStepMatchKey(compareValue string, steps []data.Step) MatchKey {
	return MatchKey{Kind: "shared_step", Name: compareValue, Hash: stepsHash(steps)}
}

// stepsHash returns a stable sha256 hex of the step slice in JSON form.
// Empty slice yields the sha256 of "[]" so empty steps still compare equal
// to other empty steps.
func stepsHash(steps []data.Step) string {
	if len(steps) == 0 {
		// Keep hashing consistent — still hash the JSON representation.
		steps = []data.Step{}
	}
	buf, err := json.Marshal(steps)
	if err != nil {
		// Fallback: deterministic but coarse marker so encoding failure does
		// not accidentally collapse unrelated shared steps.
		return "unhashable"
	}
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:])
}

// fieldValue extracts a struct field by name using reflection. Returns empty
// string when the field is missing, matching the previous behavior used by
// Filter* functions.
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

// resolveDstSectionIDForFilter resolves a source section_id to a destination
// section_id at Filter time. Sections migration (filter + import) runs before
// cases migration, so m.mapping is already populated for both pre-existing
// and freshly created destination sections.
//
// When the section cannot be resolved we return -srcSectionID to keep the
// scope unique and negative so it can never collide with a legitimate
// positive destination section_id — source cases from unmapped sections are
// therefore treated as new and never incorrectly matched against target.
func (m *Migration) resolveDstSectionIDForFilter(srcSectionID int64) int64 {
	if srcSectionID == 0 {
		return m.dstSuite
	}
	if mapped, ok := m.mapping.GetTargetBySource(srcSectionID); ok {
		return mapped
	}
	return -srcSectionID
}
