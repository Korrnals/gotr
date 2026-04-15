package snap

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/viper"
)

// CurrentServerURL returns the configured base_url from viper.
// Falls back to "https://localhost" if not configured.
func CurrentServerURL() string {
	if u := viper.GetString("base_url"); u != "" {
		return u
	}
	return "https://localhost"
}

// Snapshot is the handle returned after a successful pre-mutation save.
type Snapshot struct {
	store    *Store
	manifest *Manifest
	Meta     *Meta
}

// TakeSnapshot performs a pre-mutation GET via fetchFn and saves the result.
// For add operations, pass nil fetchFn — use FinalizeAdd after mutation.
func TakeSnapshot(
	ctx context.Context,
	store *Store,
	manifest *Manifest,
	meta Meta,
	fetchFn func(ctx context.Context) (interface{}, error),
) (*Snapshot, error) {
	meta.Timestamp = time.Now().UTC()
	meta.Status = StatusAvailable
	meta.DataFile = "data.json"

	s := &Snapshot{
		store:    store,
		manifest: manifest,
		Meta:     &meta,
	}

	// For add operations, no pre-fetch needed.
	if fetchFn == nil {
		if err := store.SaveMeta(&meta); err != nil {
			return nil, fmt.Errorf("snap: save meta: %w", err)
		}
		if err := manifest.Add(&meta); err != nil {
			return nil, fmt.Errorf("snap: add to manifest: %w", err)
		}
		return s, nil
	}

	// Fetch current state before mutation.
	data, err := fetchFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("snap: fetch pre-mutation data: %w", err)
	}

	// Save data.
	size, err := store.SaveData(meta.ID, meta.DataFile, data)
	if err != nil {
		return nil, fmt.Errorf("snap: save data: %w", err)
	}
	meta.DataSizeBytes = size

	// Save meta.
	if err := store.SaveMeta(&meta); err != nil {
		return nil, fmt.Errorf("snap: save meta: %w", err)
	}

	// Update manifest.
	if err := manifest.Add(&meta); err != nil {
		return nil, fmt.Errorf("snap: add to manifest: %w", err)
	}

	return s, nil
}

// FinalizeAdd saves the created entity ID after an add operation for rollback (delete).
func (s *Snapshot) FinalizeAdd(createdID int64) error {
	s.Meta.EntityIDs = []int64{createdID}
	s.Meta.Entities = []Entity{
		{Type: s.Meta.EntityType, ID: createdID},
	}

	// Re-save meta with the new ID.
	return s.store.SaveMeta(s.Meta)
}

// MockServerURL is the canonical URL used in tests.
const MockServerURL = "https://mock.testrail.local"

// BuildMeta is a helper to construct Meta for common operations.
// If serverURL is empty, falls back to CurrentServerURL().
func BuildMeta(
	op Operation,
	entityType string,
	entityIDs []int64,
	tier Tier,
	projectID int64,
	suiteID int64,
	customName string,
	cliArgs []string,
	serverURL string,
) Meta {
	if serverURL == "" {
		serverURL = CurrentServerURL()
	}
	meta := Meta{
		Name:         customName,
		ServerURL:    serverURL,
		Operation:    op,
		EntityType:   entityType,
		EntityIDs:    entityIDs,
		ProjectID:    projectID,
		SuiteID:      suiteID,
		RollbackTier: tier,
		CLICommand:   "gotr " + strings.Join(cliArgs, " "),
		DataFile:     "data.json",
	}
	meta.ID = GenerateID(&meta)
	return meta
}

// SnapOrWarn attempts snapshot; on failure logs warning and returns nil.
func SnapOrWarn(
	ctx context.Context,
	store *Store,
	manifest *Manifest,
	meta Meta,
	fetchFn func(ctx context.Context) (interface{}, error),
) *Snapshot {
	s, err := TakeSnapshot(ctx, store, manifest, meta, fetchFn)
	if err != nil {
		ui.Warning(os.Stderr, fmt.Sprintf("Snapshot failed (non-fatal): %v", err))
		return nil
	}
	return s
}

func sanitizeName(name string) string {
	r := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		" ", "-",
		":", "_",
	)
	return r.Replace(name)
}
