package snap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCasesAPI is a test double for CasesAPI.
type mockCasesAPI struct {
	getCaseFunc    func(ctx context.Context, caseID int64) (*data.Case, error)
	updateCaseFunc func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	addCaseFunc    func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	deleteCaseFunc func(ctx context.Context, caseID int64) error
}

func (m *mockCasesAPI) GetCase(ctx context.Context, caseID int64) (*data.Case, error) {
	if m.getCaseFunc != nil {
		return m.getCaseFunc(ctx, caseID)
	}
	return nil, fmt.Errorf("GetCase not mocked")
}
func (m *mockCasesAPI) UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
	if m.updateCaseFunc != nil {
		return m.updateCaseFunc(ctx, caseID, req)
	}
	return nil, fmt.Errorf("UpdateCase not mocked")
}
func (m *mockCasesAPI) AddCase(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
	if m.addCaseFunc != nil {
		return m.addCaseFunc(ctx, sectionID, req)
	}
	return nil, fmt.Errorf("AddCase not mocked")
}
func (m *mockCasesAPI) DeleteCase(ctx context.Context, caseID int64) error {
	if m.deleteCaseFunc != nil {
		return m.deleteCaseFunc(ctx, caseID)
	}
	return fmt.Errorf("DeleteCase not mocked")
}

func TestRollback_CaseUpdate(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	// Simulate a snapshot: save original case data before an update.
	savedCase := data.Case{
		ID:         42,
		Title:      "Original Title",
		SectionID:  100,
		PriorityID: 3,
		TypeID:     1,
	}
	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID:           snapID,
		Category:     Category("cases"),
		Operation:    OpUpdate,
		EntityType:   "case",
		EntityIDs:    []int64{42},
		Status:       StatusAvailable,
		DataFile:     "data.json",
		RollbackTier: Tier1,
		Timestamp:    time.Now().UTC(),
	}

	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	// Mock API: UpdateCase should receive original values.
	var capturedCaseID int64
	var capturedReq *data.UpdateCaseRequest
	api := &mockCasesAPI{
		updateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			capturedCaseID = caseID
			capturedReq = req
			return &data.Case{ID: caseID, Title: "Original Title"}, nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(42), capturedCaseID)
	assert.Equal(t, "Original Title", *capturedReq.Title)
	assert.Equal(t, int64(3), *capturedReq.PriorityID)
	assert.Contains(t, result.Message, "restored to pre-update state")

	// Status should be rolled_back.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusRolledBack, entry.Status)
}

func TestRollback_CaseDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{
		ID:        99,
		Title:     "Deleted Case",
		SectionID: 200,
		TypeID:    2,
	}
	snapID := "cases/20260413T120000_delete_99"
	meta := &Meta{
		ID:           snapID,
		Category:     Category("cases"),
		Operation:    OpDelete,
		EntityType:   "case",
		EntityIDs:    []int64{99},
		Status:       StatusAvailable,
		DataFile:     "data.json",
		RollbackTier: Tier2,
		Timestamp:    time.Now().UTC(),
	}

	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	var capturedSectionID int64
	var capturedReq *data.AddCaseRequest
	api := &mockCasesAPI{
		addCaseFunc: func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			capturedSectionID = sectionID
			capturedReq = req
			return &data.Case{ID: 150, Title: req.Title}, nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(200), capturedSectionID)
	assert.Equal(t, "Deleted Case", capturedReq.Title)
	assert.Equal(t, int64(150), result.NewEntityID)
	assert.Contains(t, result.Message, "re-created as ID 150")
}

func TestRollback_CaseAdd(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260413T120000_add_bulk_0"
	meta := &Meta{
		ID:           snapID,
		Category:     Category("cases"),
		Operation:    OpAdd,
		EntityType:   "case",
		EntityIDs:    []int64{500}, // set by FinalizeAdd
		Status:       StatusAvailable,
		DataFile:     "data.json",
		RollbackTier: Tier2,
		Timestamp:    time.Now().UTC(),
	}

	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteCaseFunc: func(ctx context.Context, caseID int64) error {
			deletedID = caseID
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(500), deletedID)
	assert.Contains(t, result.Message, "deleted (undo add)")
}

func TestRollback_AlreadyRolledBack(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260413T120000_update_1"
	meta := &Meta{
		ID:        snapID,
		Category:  Category("cases"),
		Operation: OpUpdate,
		Status:    StatusRolledBack, // already rolled back
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	_, err = Rollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status is \"rolled_back\"")
}

func TestRollback_APIError(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 10, Title: "T", SectionID: 1}
	snapID := "cases/20260413T120000_update_10"
	meta := &Meta{
		ID:        snapID,
		Category:  Category("cases"),
		Operation: OpUpdate,
		EntityType: "case",
		EntityIDs: []int64{10},
		Status:    StatusAvailable,
		DataFile:  "data.json",
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		updateCaseFunc: func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
			return nil, fmt.Errorf("API error 500")
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 500")
	assert.False(t, result.Success)

	// Status should be rollback_partial.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusRollbackPartial, entry.Status)
}

func TestRollback_UnsupportedEntityType(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "sections/20260413T120000_update_1"
	meta := &Meta{
		ID:         snapID,
		Category:   Category("sections"),
		Operation:  OpUpdate,
		EntityType: "section",
		Status:     StatusAvailable,
		Timestamp:  time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	_, err = Rollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported entity type")
}

func TestCaseToUpdateRequest(t *testing.T) {
	c := &data.Case{
		Title:                "Test",
		TypeID:               2,
		PriorityID:           3,
		Estimate:             "5m",
		CustomPreconds:       "Login",
		CustomSteps:          "Step 1",
		CustomExpected:       "Pass",
		CustomStepsSeparated: []data.Step{{Content: "A", Expected: "B"}},
		Refs:                 "REF-1",
		MilestoneID:          10,
		TemplateID:           1,
		SectionID:            50,
	}

	req := caseToUpdateRequest(c)
	assert.Equal(t, "Test", *req.Title)
	assert.Equal(t, int64(2), *req.TypeID)
	assert.Equal(t, int64(3), *req.PriorityID)
	assert.Equal(t, "5m", *req.Estimate)
	assert.Equal(t, "Login", *req.CustomPreconds)
	assert.Equal(t, "Step 1", *req.CustomSteps)
	assert.Equal(t, "Pass", *req.CustomExpected)
	assert.Len(t, req.CustomStepsSeparated, 1)
	assert.Equal(t, "REF-1", *req.Refs)
	assert.Equal(t, int64(10), *req.MilestoneID)
	assert.Equal(t, int64(1), *req.TemplateID)
	assert.Equal(t, int64(50), *req.SectionID)
}

func TestCaseToAddRequest(t *testing.T) {
	c := &data.Case{
		Title:      "Recreate",
		SectionID:  100,
		TypeID:     2,
		PriorityID: 1,
		Refs:       "R-1",
		TemplateID: 3,
	}

	req := caseToAddRequest(c)
	assert.Equal(t, "Recreate", req.Title)
	assert.Equal(t, int64(100), req.SectionID)
	assert.Equal(t, int64(2), req.TypeID)
	assert.Equal(t, int64(1), req.PriorityID)
	assert.Equal(t, "R-1", req.Refs)
	assert.Equal(t, int64(3), req.TemplateID)
}
