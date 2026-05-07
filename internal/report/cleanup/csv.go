// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"
)

// csvHeader is the stable column order. Keep in sync with golden tests.
var csvHeader = []string{
	"project_id",
	"project_name",
	"attachment_id",
	"name",
	"size_bytes",
	"parent_kind",
	"parent_id",
	"parent_name",
	"created_unix",
	"created_utc",
	"deleted",
	"dry_run",
	"snapshot_id",
}

// RenderCSV produces a flat per-attachment table. One row per attachment;
// every row carries the run-level dry-run/snapshot context for self-contained
// downstream filtering.
func RenderCSV(r *Report) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(csvHeader); err != nil {
		return nil, fmt.Errorf("cleanup-report: write csv header: %w", err)
	}

	deleted := !r.DryRun
	for _, p := range r.Projects {
		for _, it := range p.Items {
			createdUTC := ""
			if it.CreatedUnix > 0 {
				createdUTC = time.Unix(it.CreatedUnix, 0).UTC().Format(time.RFC3339)
			}
			rec := []string{
				strconv.FormatInt(p.ProjectID, 10),
				p.ProjectName,
				strconv.FormatInt(it.AttachmentID, 10),
				it.Name,
				strconv.FormatInt(it.Size, 10),
				it.ParentKind,
				it.ParentID,
				it.ParentName,
				strconv.FormatInt(it.CreatedUnix, 10),
				createdUTC,
				strconv.FormatBool(deleted),
				strconv.FormatBool(r.DryRun),
				r.SnapshotID,
			}
			if err := w.Write(rec); err != nil {
				return nil, fmt.Errorf("cleanup-report: write csv row: %w", err)
			}
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("cleanup-report: flush csv: %w", err)
	}
	return buf.Bytes(), nil
}
