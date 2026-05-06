// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"strconv"
	"strings"
)

// NameResolver is the minimal contract the report needs to enrich
// records with human-readable names. The cleanup package keeps this
// interface narrow and local so the heavy implementation (lazy cache,
// HTTP client, mocks) can live elsewhere (internal/resolver).
//
// Implementations MUST return "" (not an error) when a name cannot be
// resolved, including for malformed/zero IDs. EnrichNames will then
// silently leave ParentName/ProjectName untouched.
type NameResolver interface {
	Project(ctx context.Context, id int64) string
	ByKind(ctx context.Context, kind string, id int64) string
}

// EnrichNames populates Report.Projects[*].ProjectName and
// Report.Projects[*].Items[*].ParentName by querying r through the
// supplied NameResolver. Existing non-empty names are preserved (the
// scanner-side ProjectName is treated as authoritative). All errors
// are swallowed by the resolver per its contract; this function never
// returns one. Calling EnrichNames with rep==nil or nr==nil is a
// no-op.
func EnrichNames(ctx context.Context, rep *Report, nr NameResolver) {
	if rep == nil || nr == nil {
		return
	}
	for i := range rep.Projects {
		pg := &rep.Projects[i]
		if pg.ProjectName == "" && pg.ProjectID > 0 {
			pg.ProjectName = nr.Project(ctx, pg.ProjectID)
		}
		for j := range pg.Items {
			it := &pg.Items[j]
			if it.ParentName != "" || it.ParentKind == "" || it.ParentID == "" {
				continue
			}
			id, ok := parseParentID(it.ParentID)
			if !ok {
				continue
			}
			it.ParentName = nr.ByKind(ctx, it.ParentKind, id)
		}
	}
	// Mirror project names into the entity breakdown so the matrix
	// table also benefits from the lookup.
	for i := range rep.EntityBreakdown {
		row := &rep.EntityBreakdown[i]
		if row.ProjectName != "" || row.ProjectID <= 0 {
			continue
		}
		row.ProjectName = nr.Project(ctx, row.ProjectID)
	}
}

// parseParentID extracts the int64 ID from a Record.ParentID string.
// Cloud-format EntityIDs (e.g. "P12.S3.case_42") are not numeric and
// cannot be resolved through this path — caller falls back to "".
func parseParentID(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	if id <= 0 {
		return 0, false
	}
	return id, true
}
