# Feature Design: Migration Reports & Snapshot Retention Management

**Status:** Design Phase  
**Priority:** High  
**Complexity:** Medium (3-4 days)  
**Date:** 2026-04-18  
**Owner:** Architecture  

---

## Overview

Three interconnected features to improve migration auditability and snapshot lifecycle management:

1. **Migration Reports** — Save detailed, compact migration summaries to `~/.gotr/reports/`
2. **Auto-labeling** — Automatic meaningful labels for user-created snapshots
3. **Snapshot Retention** — Config-driven TTL, protection policy, and GC commands

---

## Feature 1: Migration Reports

### Design

**Location:** `~/.gotr/reports/`

**Naming:** `migration-{timestamp}-{snapshot-id}.md`  
**Example:** `migration-20260418T143015Z-snap-a7f42c9e.md`

**File Structure:**

```markdown
# Migration Report: `snapshot-id`
**Date:** 2026-04-18 14:30:15 UTC | **Duration:** 12.5s | **Status:** ✅ Success

## Configuration
| Parameter | Value |
|-----------|-------|
| Source Project | 125 |
| Target Project | 130 |
| Migration Type | sync full |
| User | alice |

## Summary
| Resource Type | Source Count | Created | Updated | Skipped | Failed |
|---------------|--------------|---------|---------|---------|--------|
| Cases | 342 | 285 | 0 | 57 | 0 |
| Shared Steps | 89 | 76 | 0 | 13 | 0 |
| Attachments | 1,205 | 1,102 | 0 | 103 | 0 |
| **TOTAL** | **1,636** | **1,463** | **0** | **173** | **0** |

## Details

**Cases:** 285 created (IDs mapped in snapshot)
**Shared Steps:** 76 created (source IDs not preserved per TestRail API)
**Skipped Cases (57):** Custom field mismatch, configuration conflict
**Skipped Shared Steps (13):** Duplicate detection

## Rollback
- **Snapshot ID:** `snap-a7f42c9e`
- **Enabled:** Yes
- **Command:** `gotr snap rollback snap-a7f42c9e`
- **Entity Deletion Order:** Cases → Sections → Shared Steps → Suites

## Performance
- **Total Time:** 12.5s
- **Average Rate:** ~130 entities/sec
- **Peak Memory:** 245 MB

## Related Files
- Snapshot: `~/.gotr/snapshots/snap-a7f42c9e.json`
- Config Used: `~/.gotr/config.yaml` (hash: abc123)
```

### Implementation

**New types:**

```go
// internal/report/types.go
type MigrationReport struct {
    ID              string
    SnapshotID      string
    Timestamp       time.Time
    Duration        time.Duration
    Status          string // "success", "partial", "failed"
    
    SourceProject   int64
    TargetProject   int64
    MigrationType   string // "sync_full", "sync_cases", etc
    User            string
    
    Summary         map[string]*ResourceStats // "cases", "shared_steps"
    Skipped         map[string][]SkipReason
    Rollback        RollbackInfo
    Performance     PerfMetrics
}

type ResourceStats struct {
    SourceCount int64
    Created     int64
    Updated     int64
    Skipped     int64
    Failed      int64
}

type SkipReason struct {
    ID       int64
    Reason   string
    Detail   string
}

type RollbackInfo struct {
    SnapshotID  string
    Enabled     bool
    DeleteOrder []string
}

type PerfMetrics struct {
    TotalTime   time.Duration
    EntitiesPerSec float64
    PeakMemory  int64
}
```

**New service method:**

```go
// internal/service/report/migration.go
func (s *ReportService) SaveMigrationReport(
    ctx context.Context,
    report *types.MigrationReport,
) (string, error) {
    // Generate markdown
    md := generateMigrationMarkdown(report)
    
    // Save to ~/.gotr/reports/migration-{ts}-{id}.md
    path := filepath.Join(reportsDir, fmt.Sprintf(
        "migration-%s-%s.md",
        report.Timestamp.Format("20060102T150405Z"),
        report.SnapshotID,
    ))
    
    if err := os.WriteFile(path, []byte(md), 0644); err != nil {
        return "", err
    }
    
    // Update index
    return updateMigrationIndex(reportsDir, report)
}
```

**Integration points:**

- `cmd/sync/sync_full.go` — Collect stats, call SaveMigrationReport
- `cmd/sync/sync_cases.go` — Collect stats per command
- `internal/snap/hook.go` — Attach report info to snapshot metadata

---

## Feature 2: Auto-labeling for Snapshots

### Design

**Label Format:**

```
{mode}_{command}_{YYYYMMDDHHMMSS}

Examples:
- interactive_sync_full_20260418143015
- interactive_sync_cases_20260418095500
- manual_snapshot_20260418120000
- auto_compare_20260418080000  (system-generated)
```

**Logic:**

1. **Default label** generated automatically
2. User prompted in interactive mode: accept default, customize, or pin
3. Non-interactive: use default unless `--label` or `--pin` flag provided

### Implementation

**New config fields:**

```yaml
snapshot:
  auto_label:
    enabled: true
    format: "{mode}_{command}_{timestamp}"  # Template
    
  labels:
    # Reserved prefixes for system usage
    reserved_prefixes:
      - "auto_"
      - "system_"
```

**New function:**

```go
// internal/snap/labeling.go
func GenerateDefaultLabel(command, mode string) string {
    timestamp := time.Now().Format("20060102150405")
    return fmt.Sprintf("%s_%s_%s", mode, command, timestamp)
}

func PromptForLabel(ctx context.Context, defaultLabel string) (string, error) {
    // Interactive prompt
    fmt.Printf("Snapshot label [%s]: ", defaultLabel)
    input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
    
    if strings.TrimSpace(input) == "" {
        return defaultLabel, nil
    }
    return strings.TrimSpace(input), nil
}
```

**Integration:**

- All sync commands pass `command` and `mode` ("interactive"/"batch"/"auto") to snap hook
- Hook generates default label
- If interactive + no `--label` flag: prompt user

---

## Feature 3: Snapshot Retention & GC

### Design

**Config Structure:**

```yaml
snapshot:
  retention:
    enabled: true
    default_ttl_days: 30          # Auto-delete snapshots older than 30d
    
    protected_prefixes:
      - "pinned_"                 # pinned_* never auto-deleted
      - "archived_"               # archived_* never auto-deleted
      - "important_"
    
    frozen_snapshots: []          # List of snapshot IDs to never delete
    
  gc:
    enabled: true
    run_on_startup: false
    run_before_sync: false        # Can be enabled by user recommendation
```

**New Commands:**

```bash
# GC Preview and Execution
gotr snap gc                      # Show what would be deleted
gotr snap gc --confirm            # Actually delete
gotr snap gc --dry-run            # Preview (alias for default)

# List with Filters
gotr snap list                    # All snapshots
gotr snap list --older-than 30d
gotr snap list --label-pattern "pinned_*"
gotr snap list --protected

# Manual Operations
gotr snap delete <id> [<id2>...]  # Delete specific snapshots (confirm required)
gotr snap pin <id>                # Add pinned_ prefix → freeze
gotr snap unpin <id>              # Remove pinned_ prefix → unfreeze
gotr snap freeze <id>             # Mark as frozen in config (permanent)
```

**New Flags (all sync commands):**

```bash
gotr sync full \
  --label custom_name             # Custom label
  --pin                           # Auto-label with pinned_ prefix
  --no-snapshot                   # Skip snapshot entirely
  --auto-gc                       # Run GC preview before operation
```

### Implementation

**Type Definitions:**

```go
// internal/snap/retention.go
type RetentionPolicy struct {
    Enabled           bool
    DefaultTTLDays    int
    ProtectedPrefixes []string
    FrozenSnapshots   []string
}

type GCResult struct {
    Analyzed    int
    ToDelete    []GCCandidate
    Protected   []GCProtected
    DeletedCount int
    FreedBytes  int64
}

type GCCandidate struct {
    SnapshotID string
    Label      string
    Age        time.Duration
    Size       int64
    Reason     string
}

type GCProtected struct {
    SnapshotID string
    Label      string
    Reason     string // "pinned", "frozen", "protected_prefix", etc
}
```

**GC Logic:**

```go
// internal/snap/gc.go
func (s *SnapService) AnalyzeForGC(
    ctx context.Context,
    policy RetentionPolicy,
) (*GCResult, error) {
    snapshots, _ := s.ListSnapshots(ctx)
    
    result := &GCResult{}
    cutoff := time.Now().AddDate(0, 0, -policy.DefaultTTLDays)
    
    for _, snap := range snapshots {
        // Check if protected
        if isProtected(snap.Label, policy) {
            result.Protected = append(result.Protected, ...)
            continue
        }
        
        // Check age
        if snap.Created.Before(cutoff) {
            result.ToDelete = append(result.ToDelete, ...)
        }
    }
    
    return result, nil
}

func (s *SnapService) ExecuteGC(
    ctx context.Context,
    result *GCResult,
) error {
    for _, candidate := range result.ToDelete {
        s.Delete(ctx, candidate.SnapshotID)
    }
    return nil
}
```

**Interactive Flow (after sync success):**

```go
// cmd/sync/post_action_menu.go
func postMigrationMenu(ctx context.Context, snapID string) {
    fmt.Println("\n✅ Migration completed successfully")
    fmt.Printf("Snapshot: %s\n\n", snapID)
    
    fmt.Println("Options:")
    fmt.Println("  [↻] Rollback this migration")
    fmt.Println("  [💾] Save with custom label")
    fmt.Println("  [📌] Pin for archival")
    fmt.Println("  [🗑] Delete snapshot")
    fmt.Println("  [Q] Quit")
    
    choice := prompt("Choose: ")
    
    switch choice {
    case "↻":
        // Rollback flow
    case "💾":
        label := prompt("Label: ")
        snap.Pin(ctx, snapID, label)
    case "📌":
        snap.Pin(ctx, snapID, "pinned_"+uuid.Short())
    // ...
    }
}
```

---

## Implementation Plan

### Phase 1: Migration Reports (Day 1)
- [ ] Define report types and schema
- [ ] Implement SaveMigrationReport service
- [ ] Create markdown template generator
- [ ] Update sync_full.go to collect and save report
- [ ] Add tests for report generation
- [ ] Checkpoint: `report-service-complete`

### Phase 2: Auto-labeling (Day 1.5)
- [ ] Add labeling service with DefaultLabel generation
- [ ] Add interactive prompt logic
- [ ] Update all sync commands to use auto-labels
- [ ] Update snap hook to attach labels
- [ ] Add `--label`, `--pin`, `--no-snapshot` flags
- [ ] Add tests
- [ ] Checkpoint: `auto-labeling-complete`

### Phase 3: Config & Retention Policy (Day 2)
- [ ] Add `snapshot.retention` to config schema
- [ ] Implement policy reading from config
- [ ] Add validation and defaults
- [ ] Checkpoint: `retention-config-complete`

### Phase 4: GC Commands & Logic (Day 2-3)
- [ ] Implement AnalyzeForGC service
- [ ] Implement ExecuteGC service
- [ ] Add `gotr snap gc` command (preview)
- [ ] Add `gotr snap gc --confirm` (execute)
- [ ] Add `gotr snap delete <id>`
- [ ] Add `gotr snap pin/unpin` commands
- [ ] Add output formatting
- [ ] Add tests
- [ ] Checkpoint: `gc-commands-complete`

### Phase 5: Integration & Finalization (Day 3-4)
- [ ] Add `--auto-gc` flag to sync commands
- [ ] Integrate post-migration interactive menu
- [ ] Update documentation
- [ ] Full regression testing
- [ ] Performance testing
- [ ] Final commit: `feat(snap): migration reports + retention management`

---

## Testing Strategy

### Unit Tests
- Report generation and formatting
- Label generation and validation
- GC candidate detection
- Policy parsing and validation
- Protection rules

### Integration Tests
- sync_full → report save → verify content
- Interactive label prompts
- GC execution with protected snapshots
- Config-based retention

### E2E Tests
- Full migration flow with report
- GC cleanup of old snapshots
- Pin/unpin operations
- Rollback of pinned snapshots

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Report size explosion | Implement rotating retention for reports (keep last 100) |
| GC deletes important snapshot | Require explicit confirmation, show what will be deleted |
| Config parsing errors | Validate during startup, fallback to safe defaults |
| Label conflicts | Auto-generate unique labels if duplicates detected |

---

## Acceptance Criteria

- [x] Design document created and reviewed
- [ ] Reports save to correct location with correct format
- [ ] Auto-labels generated and applied to all snapshots
- [ ] GC correctly identifies candidates and respects protection
- [ ] All tests pass (unit + integration + e2e)
- [ ] Documentation updated
- [ ] Config backward-compatible

---

## Decision Log

**2026-04-18:**
- ✅ Decided on label format: `{mode}_{command}_{YYYYMMDDHHMMSS}`
- ✅ Decided on default TTL: 30 days
- ✅ Decided on protection mechanism: prefix-based + explicit frozen list
- ✅ Decided on GC: manual + optional pre-sync recommendation (not automatic background)
- ✅ Decided on report location: `~/.gotr/reports/`
