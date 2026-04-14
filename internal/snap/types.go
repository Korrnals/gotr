// Package snap provides pre-mutation snapshot and rollback capabilities.
package snap

import "time"

// Operation describes the type of mutation being performed.
type Operation string

const (
	OpAdd    Operation = "add"
	OpUpdate Operation = "update"
	OpDelete Operation = "delete"
	OpMove   Operation = "move"
	OpCopy   Operation = "copy"
	OpClose  Operation = "close"
	OpBulk   Operation = "bulk"

	// Sync operations.
	OpSyncFull        Operation = "sync_full"
	OpSyncCases       Operation = "sync_cases"
	OpSyncSections    Operation = "sync_sections"
	OpSyncSharedSteps Operation = "sync_shared_steps"
	OpSyncSuites      Operation = "sync_suites"
)

// Category determines the storage subdirectory for a snapshot.
type Category string

const (
	CatCustom Category = "custom"
	CatSync   Category = "sync"
	// Entity-based categories are constructed dynamically: "cases", "sections", etc.
)

// Tier describes how well an operation can be rolled back.
type Tier int

const (
	// Tier1 — full rollback (update/move).
	Tier1 Tier = 1
	// Tier2 — partial rollback, new ID (delete/add).
	Tier2 Tier = 2
	// Tier3 — info only, no rollback possible (close/add results).
	Tier3 Tier = 3
)

// Status describes the current state of a snapshot.
type Status string

const (
	StatusAvailable       Status = "available"
	StatusRolledBack      Status = "rolled_back"
	StatusRollbackPartial Status = "rollback_partial"
	StatusExpired         Status = "expired"
)

// RollbackStatus describes the state of a single entity rollback.
type RollbackStatus string

const (
	RBPending  RollbackStatus = "pending"
	RBRestored RollbackStatus = "restored"
	RBFailed   RollbackStatus = "failed"
)

// Entity represents a single snapshotted entity within a snapshot.
type Entity struct {
	Type     string `json:"type"`
	ID       int64  `json:"id"`
	ParentID int64  `json:"parent_id,omitempty"`
}

// RollbackLogEntry tracks the rollback state of a single entity (Phase 2).
type RollbackLogEntry struct {
	Type   string         `json:"type"`
	ID     int64          `json:"id"`
	NewID  int64          `json:"new_id,omitempty"` // ID of re-created entity (for delete rollbacks)
	Status RollbackStatus `json:"status"`
	Error  string         `json:"error,omitempty"`
}

// Meta holds metadata for one snapshot.
type Meta struct {
	ID              string    `json:"id"`
	Name            string    `json:"name,omitempty"`
	ServerURL       string    `json:"server_url,omitempty"`
	Category        Category  `json:"category"`
	Operation       Operation `json:"operation"`
	EntityType      string    `json:"entity_type"`
	EntityIDs       []int64   `json:"entity_ids"`
	ProjectID       int64     `json:"project_id,omitempty"`
	SuiteID         int64     `json:"suite_id,omitempty"`
	SourceProjectID int64     `json:"source_project_id,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
	RollbackTier    Tier      `json:"rollback_tier"`
	Status          Status    `json:"status"`
	CLICommand      string    `json:"cli_command"`
	DataFile        string    `json:"data_file"`
	DataSizeBytes   int64     `json:"data_size_bytes"`
	Entities        []Entity  `json:"entities,omitempty"`

	RollbackLog []RollbackLogEntry `json:"rollback_log,omitempty"`
}

// IsSyncOp returns true if the operation is a sync variant.
func (m *Meta) IsSyncOp() bool {
	switch m.Operation {
	case OpSyncFull, OpSyncCases, OpSyncSections, OpSyncSharedSteps, OpSyncSuites:
		return true
	}
	return false
}

// SnapConfig holds snap-related configuration values read from Viper.
type SnapConfig struct {
	Enabled       bool
	RetentionDays int
	MaxSnapshots  int

	AttachSaveBinary        string // "auto" | "always" | "never"
	AttachMaxFileMB         int
	AttachCompress          bool
	AttachPromptAboveThresh bool
}

// DefaultConfig returns the default snap configuration values.
func DefaultConfig() SnapConfig {
	return SnapConfig{
		Enabled:                 true,
		RetentionDays:           30,
		MaxSnapshots:            100,
		AttachSaveBinary:        "auto",
		AttachMaxFileMB:         10,
		AttachCompress:          true,
		AttachPromptAboveThresh: true,
	}
}
