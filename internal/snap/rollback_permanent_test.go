package snap

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// permanentTrackingAPI tracks which Delete variant is invoked.
type permanentTrackingAPI struct {
	*mockCasesAPI

	deleteCalled          bool
	deleteCasePermCalled  bool
	deleteSectionCalled   bool
	deleteSectPermCalled  bool
	deleteSuiteCalled     bool
	deleteSuitePermCalled bool
}

func newPermanentTrackingAPI() *permanentTrackingAPI {
	return &permanentTrackingAPI{mockCasesAPI: &mockCasesAPI{}}
}

func (p *permanentTrackingAPI) DeleteCase(ctx context.Context, caseID int64) error {
	p.deleteCalled = true
	return nil
}
func (p *permanentTrackingAPI) DeleteCasePermanent(ctx context.Context, caseID int64) error {
	p.deleteCasePermCalled = true
	return nil
}
func (p *permanentTrackingAPI) DeleteSection(ctx context.Context, sectionID int64) error {
	p.deleteSectionCalled = true
	return nil
}
func (p *permanentTrackingAPI) DeleteSectionPermanent(ctx context.Context, sectionID int64) error {
	p.deleteSectPermCalled = true
	return nil
}
func (p *permanentTrackingAPI) DeleteSuite(ctx context.Context, suiteID int64) error {
	p.deleteSuiteCalled = true
	return nil
}
func (p *permanentTrackingAPI) DeleteSuitePermanent(ctx context.Context, suiteID int64) error {
	p.deleteSuitePermCalled = true
	return nil
}

func setupCaseAddSnap(t *testing.T) (*Store, *Manifest, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260422T000000_add_bulk_0"
	meta := &Meta{
		ID:           snapID,
		Category:     Category("cases"),
		Operation:    OpAdd,
		EntityType:   "case",
		EntityIDs:    []int64{123},
		Status:       StatusAvailable,
		DataFile:     "data.json",
		RollbackTier: Tier2,
		Timestamp:    time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))
	return store, manifest, snapID
}

// TestRollback_CaseAdd_SoftByDefault ensures rollback of an add uses soft-delete
// when RollbackOpts.Permanent is not set (default path — the Stage 0.5 fix).
func TestRollback_CaseAdd_SoftByDefault(t *testing.T) {
	store, manifest, snapID := setupCaseAddSnap(t)
	api := newPermanentTrackingAPI()

	_, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, api.deleteCalled, "expected DeleteCase (soft) to be called")
	assert.False(t, api.deleteCasePermCalled, "DeleteCasePermanent must not be called by default")
}

// TestRollback_CaseAdd_Permanent ensures --permanent flag routes through the
// DeleteCasePermanent path.
func TestRollback_CaseAdd_Permanent(t *testing.T) {
	store, manifest, snapID := setupCaseAddSnap(t)
	api := newPermanentTrackingAPI()

	_, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{Permanent: true})
	require.NoError(t, err)
	assert.False(t, api.deleteCalled, "DeleteCase (soft) must not be called when Permanent=true")
	assert.True(t, api.deleteCasePermCalled, "expected DeleteCasePermanent to be called")
}
