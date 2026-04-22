package snap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCasesAPI is a test double for RollbackAPI.
type mockCasesAPI struct {
	getCaseFunc    func(ctx context.Context, caseID int64) (*data.Case, error)
	updateCaseFunc func(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	addCaseFunc    func(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	deleteCaseFunc func(ctx context.Context, caseID int64) error

	getSectionFunc     func(ctx context.Context, sectionID int64) (*data.Section, error)
	addSectionFunc     func(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error)
	deleteSectionFunc  func(ctx context.Context, sectionID int64) error
	deleteSharedFunc   func(ctx context.Context, stepID int64, keep int) error
	deleteSuiteFunc    func(ctx context.Context, suiteID int64) error

	// Suite add
	addSuiteFunc func(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error)

	// Run
	addRunFunc    func(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error)
	deleteRunFunc func(ctx context.Context, runID int64) error

	// Milestone
	addMilestoneFunc    func(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error)
	deleteMilestoneFunc func(ctx context.Context, milestoneID int64) error

	// Plan
	addPlanFunc    func(ctx context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error)
	deletePlanFunc func(ctx context.Context, planID int64) error

	// Project
	addProjectFunc    func(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error)
	deleteProjectFunc func(ctx context.Context, projectID int64) error

	// Configuration
	addConfigGroupFunc    func(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error)
	addConfigFunc         func(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error)
	deleteConfigGroupFunc func(ctx context.Context, groupID int64) error
	deleteConfigFunc      func(ctx context.Context, configID int64) error

	// Group
	addGroupFunc    func(ctx context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error)
	deleteGroupFunc func(ctx context.Context, groupID int64) error

	// Dataset
	addDatasetFunc    func(ctx context.Context, projectID int64, name string) (*data.Dataset, error)
	deleteDatasetFunc func(ctx context.Context, datasetID int64) error

	// Variable
	addVariableFunc    func(ctx context.Context, datasetID int64, name string) (*data.Variable, error)
	deleteVariableFunc func(ctx context.Context, variableID int64) error
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
func (m *mockCasesAPI) DeleteCasePermanent(ctx context.Context, caseID int64) error {
	return m.DeleteCase(ctx, caseID)
}
func (m *mockCasesAPI) GetSection(ctx context.Context, sectionID int64) (*data.Section, error) {
	if m.getSectionFunc != nil {
		return m.getSectionFunc(ctx, sectionID)
	}
	return nil, fmt.Errorf("GetSection not mocked")
}
func (m *mockCasesAPI) AddSection(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	if m.addSectionFunc != nil {
		return m.addSectionFunc(ctx, projectID, req)
	}
	return nil, fmt.Errorf("AddSection not mocked")
}
func (m *mockCasesAPI) DeleteSection(ctx context.Context, sectionID int64) error {
	if m.deleteSectionFunc != nil {
		return m.deleteSectionFunc(ctx, sectionID)
	}
	return nil
}
func (m *mockCasesAPI) DeleteSectionPermanent(ctx context.Context, sectionID int64) error {
	return m.DeleteSection(ctx, sectionID)
}
func (m *mockCasesAPI) DeleteSharedStep(ctx context.Context, stepID int64, keep int) error {
	if m.deleteSharedFunc != nil {
		return m.deleteSharedFunc(ctx, stepID, keep)
	}
	return nil
}
func (m *mockCasesAPI) DeleteSuite(ctx context.Context, suiteID int64) error {
	if m.deleteSuiteFunc != nil {
		return m.deleteSuiteFunc(ctx, suiteID)
	}
	return nil
}
func (m *mockCasesAPI) DeleteSuitePermanent(ctx context.Context, suiteID int64) error {
	return m.DeleteSuite(ctx, suiteID)
}
func (m *mockCasesAPI) AddSuite(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	if m.addSuiteFunc != nil {
		return m.addSuiteFunc(ctx, projectID, req)
	}
	return &data.Suite{ID: 1}, nil
}
func (m *mockCasesAPI) AddRun(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	if m.addRunFunc != nil {
		return m.addRunFunc(ctx, projectID, req)
	}
	return &data.Run{ID: 1}, nil
}
func (m *mockCasesAPI) DeleteRun(ctx context.Context, runID int64) error {
	if m.deleteRunFunc != nil {
		return m.deleteRunFunc(ctx, runID)
	}
	return nil
}
func (m *mockCasesAPI) AddMilestone(ctx context.Context, projectID int64, req *data.AddMilestoneRequest) (*data.Milestone, error) {
	if m.addMilestoneFunc != nil {
		return m.addMilestoneFunc(ctx, projectID, req)
	}
	return &data.Milestone{ID: 1}, nil
}
func (m *mockCasesAPI) DeleteMilestone(ctx context.Context, milestoneID int64) error {
	if m.deleteMilestoneFunc != nil {
		return m.deleteMilestoneFunc(ctx, milestoneID)
	}
	return nil
}
func (m *mockCasesAPI) AddPlan(ctx context.Context, projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
	if m.addPlanFunc != nil {
		return m.addPlanFunc(ctx, projectID, req)
	}
	return &data.Plan{ID: 1}, nil
}
func (m *mockCasesAPI) DeletePlan(ctx context.Context, planID int64) error {
	if m.deletePlanFunc != nil {
		return m.deletePlanFunc(ctx, planID)
	}
	return nil
}
func (m *mockCasesAPI) AddProject(ctx context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
	if m.addProjectFunc != nil {
		return m.addProjectFunc(ctx, req)
	}
	return &data.GetProjectResponse{ID: 1}, nil
}
func (m *mockCasesAPI) DeleteProject(ctx context.Context, projectID int64) error {
	if m.deleteProjectFunc != nil {
		return m.deleteProjectFunc(ctx, projectID)
	}
	return nil
}
func (m *mockCasesAPI) AddConfigGroup(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
	if m.addConfigGroupFunc != nil {
		return m.addConfigGroupFunc(ctx, projectID, req)
	}
	return &data.ConfigGroup{ID: 1}, nil
}
func (m *mockCasesAPI) AddConfig(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
	if m.addConfigFunc != nil {
		return m.addConfigFunc(ctx, groupID, req)
	}
	return &data.Config{ID: 1}, nil
}
func (m *mockCasesAPI) DeleteConfigGroup(ctx context.Context, groupID int64) error {
	if m.deleteConfigGroupFunc != nil {
		return m.deleteConfigGroupFunc(ctx, groupID)
	}
	return nil
}
func (m *mockCasesAPI) DeleteConfig(ctx context.Context, configID int64) error {
	if m.deleteConfigFunc != nil {
		return m.deleteConfigFunc(ctx, configID)
	}
	return nil
}
func (m *mockCasesAPI) AddGroup(ctx context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error) {
	if m.addGroupFunc != nil {
		return m.addGroupFunc(ctx, projectID, name, userIDs)
	}
	return &data.Group{ID: 1}, nil
}
func (m *mockCasesAPI) DeleteGroup(ctx context.Context, groupID int64) error {
	if m.deleteGroupFunc != nil {
		return m.deleteGroupFunc(ctx, groupID)
	}
	return nil
}
func (m *mockCasesAPI) AddDataset(ctx context.Context, projectID int64, name string) (*data.Dataset, error) {
	if m.addDatasetFunc != nil {
		return m.addDatasetFunc(ctx, projectID, name)
	}
	return &data.Dataset{ID: 1}, nil
}
func (m *mockCasesAPI) DeleteDataset(ctx context.Context, datasetID int64) error {
	if m.deleteDatasetFunc != nil {
		return m.deleteDatasetFunc(ctx, datasetID)
	}
	return nil
}
func (m *mockCasesAPI) AddVariable(ctx context.Context, datasetID int64, name string) (*data.Variable, error) {
	if m.addVariableFunc != nil {
		return m.addVariableFunc(ctx, datasetID, name)
	}
	return &data.Variable{ID: 1}, nil
}
func (m *mockCasesAPI) DeleteVariable(ctx context.Context, variableID int64) error {
	if m.deleteVariableFunc != nil {
		return m.deleteVariableFunc(ctx, variableID)
	}
	return nil
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

	snapID := "unknown/20260413T120000_update_1"
	meta := &Meta{
		ID:         snapID,
		Category:   Category("unknown"),
		Operation:  OpUpdate,
		EntityType: "unknown_entity",
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

// ---------------------------------------------------------------------------
// Phase 2 tests: dry-run, entity-ids filter, resume, per-entity log, export
// ---------------------------------------------------------------------------

func TestRollback_DryRun_CaseUpdate(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 42, Title: "Original", SectionID: 100, PriorityID: 3}
	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	// Mock: GetCase returns modified values; UpdateCase should NOT be called.
	updateCalled := false
	api := &mockCasesAPI{
		getCaseFunc: func(_ context.Context, caseID int64) (*data.Case, error) {
			return &data.Case{ID: caseID, Title: "Modified", SectionID: 100, PriorityID: 5}, nil
		},
		updateCaseFunc: func(_ context.Context, _ int64, _ *data.UpdateCaseRequest) (*data.Case, error) {
			updateCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.DryRun)
	assert.False(t, updateCalled, "UpdateCase should not be called in dry-run")

	// Preview should contain changed fields.
	require.NotEmpty(t, result.Preview)
	fields := make(map[string]bool)
	for _, d := range result.Preview {
		fields[d.Field] = true
		assert.Equal(t, int64(42), d.EntityID)
	}
	assert.True(t, fields["title"], "title diff expected")
	assert.True(t, fields["priority_id"], "priority_id diff expected")

	// Status should remain available (not rolled_back).
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusAvailable, entry.Status)
}

func TestRollback_DryRun_CaseDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 99, Title: "Deleted Case", SectionID: 200}
	snapID := "cases/20260413T120000_delete_99"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpDelete,
		EntityType: "case", EntityIDs: []int64{99}, Status: StatusAvailable,
		DataFile: "data.json", RollbackTier: Tier2, Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{} // No API calls expected in dry-run

	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	require.NotEmpty(t, result.Preview)
	assert.Equal(t, "action", result.Preview[0].Field)
	assert.Equal(t, "DELETED", result.Preview[0].Current)
}

func TestRollback_DryRun_CaseAdd(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260413T120000_add_bulk_0"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpAdd,
		EntityType: "case", EntityIDs: []int64{500}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{} // No API calls expected in dry-run

	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	require.NotEmpty(t, result.Preview)
	assert.Equal(t, "action", result.Preview[0].Field)
	assert.Equal(t, "DELETE", result.Preview[0].Saved)
}

func TestRollback_EntityIDsFilter_Skipped(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 42, Title: "Original", SectionID: 100}
	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	updateCalled := false
	api := &mockCasesAPI{
		updateCaseFunc: func(_ context.Context, _ int64, _ *data.UpdateCaseRequest) (*data.Case, error) {
			updateCalled = true
			return nil, nil
		},
	}

	// Filter to entity 999 — should skip entity 42.
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{EntityIDs: []int64{999}})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.False(t, updateCalled, "UpdateCase should not be called for filtered entity")
	assert.Contains(t, result.Message, "skipped")
}

func TestRollback_EntityIDsFilter_Allowed(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 42, Title: "Original", SectionID: 100}
	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	updateCalled := false
	api := &mockCasesAPI{
		updateCaseFunc: func(_ context.Context, caseID int64, _ *data.UpdateCaseRequest) (*data.Case, error) {
			updateCalled = true
			return &data.Case{ID: caseID}, nil
		},
	}

	// Include entity 42 in filter — should proceed.
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{EntityIDs: []int64{42}})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, updateCalled, "UpdateCase should be called for allowed entity")
}

func TestRollback_Resume_SkipRestored(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 42, Title: "Original", SectionID: 100}
	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusRollbackPartial,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
		// Pre-set log entry as already restored.
		RollbackLog: []RollbackLogEntry{{Type: "case", ID: 42, Status: RBRestored}},
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	updateCalled := false
	api := &mockCasesAPI{
		updateCaseFunc: func(_ context.Context, _ int64, _ *data.UpdateCaseRequest) (*data.Case, error) {
			updateCalled = true
			return nil, nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.False(t, updateCalled, "Should skip already-restored entity")
	assert.Contains(t, result.Message, "resume skip")
}

func TestRollback_Resume_RetryPartial(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 42, Title: "Original", SectionID: 100}
	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusRollbackPartial,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
		// Pre-set log entry as failed (should be retried).
		RollbackLog: []RollbackLogEntry{{Type: "case", ID: 42, Status: RBFailed, Error: "API error 500"}},
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	updateCalled := false
	api := &mockCasesAPI{
		updateCaseFunc: func(_ context.Context, caseID int64, _ *data.UpdateCaseRequest) (*data.Case, error) {
			updateCalled = true
			return &data.Case{ID: caseID}, nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, updateCalled, "Should retry failed entity")
}

func TestRollback_PerEntityLog(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 42, Title: "Original", SectionID: 100}
	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		updateCaseFunc: func(_ context.Context, caseID int64, _ *data.UpdateCaseRequest) (*data.Case, error) {
			return &data.Case{ID: caseID}, nil
		},
	}

	_, err = Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)

	// Verify rollback log was written to meta.
	updatedMeta, err := store.LoadMeta(snapID)
	require.NoError(t, err)
	require.Len(t, updatedMeta.RollbackLog, 1)
	assert.Equal(t, "case", updatedMeta.RollbackLog[0].Type)
	assert.Equal(t, int64(42), updatedMeta.RollbackLog[0].ID)
	assert.Equal(t, RBRestored, updatedMeta.RollbackLog[0].Status)
}

func TestBuildCaseDiff(t *testing.T) {
	current := &data.Case{ID: 1, Title: "Changed", PriorityID: 5, SectionID: 100}
	saved := &data.Case{ID: 1, Title: "Original", PriorityID: 3, SectionID: 100}

	diffs := buildCaseDiff(1, current, saved)
	fields := make(map[string]bool)
	for _, d := range diffs {
		fields[d.Field] = true
	}
	assert.True(t, fields["title"])
	assert.True(t, fields["priority_id"])
	assert.False(t, fields["section_id"], "identical fields should not appear")
}

func TestEntityAllowed(t *testing.T) {
	assert.True(t, entityAllowed(42, nil), "nil filter allows all")
	assert.True(t, entityAllowed(42, []int64{}), "empty filter allows all")
	assert.True(t, entityAllowed(42, []int64{10, 42, 99}), "ID in list")
	assert.False(t, entityAllowed(42, []int64{10, 99}), "ID not in list")
}

func TestStore_Export(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)

	snapID := "cases/20260413T120000_update_42"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpUpdate,
		EntityType: "case", EntityIDs: []int64{42}, Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	savedCase := data.Case{ID: 42, Title: "Original", SectionID: 100}
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)

	outPath := dir + "/export_test.json"
	err = store.Export(snapID, outPath)
	require.NoError(t, err)

	// Verify the exported file contains meta and data.
	var envelope struct {
		Meta *Meta     `json:"meta"`
		Data data.Case `json:"data"`
	}
	f, err := os.Open(outPath)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, json.NewDecoder(f).Decode(&envelope))

	assert.Equal(t, snapID, envelope.Meta.ID)
	assert.Equal(t, "Original", envelope.Data.Title)
	assert.Equal(t, int64(42), envelope.Data.ID)
}

// ---------------------------------------------------------------------------
// Phase 3 tests: cascade section rollback
// ---------------------------------------------------------------------------

func TestRollback_SectionCascade_Delete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	cascade := CascadeData{
		Section: data.Section{ID: 10, Name: "Auth", SuiteID: 1, Description: "Auth tests"},
		Cases: []data.Case{
			{ID: 100, Title: "Login", SectionID: 10, PriorityID: 3},
			{ID: 101, Title: "Logout", SectionID: 10, PriorityID: 2},
		},
	}
	snapID := "sections/20260413T120000_delete_10"
	meta := &Meta{
		ID: snapID, Category: Category("sections"), Operation: OpDelete,
		EntityType: "section", EntityIDs: []int64{10, 100, 101},
		Status: StatusAvailable, DataFile: "data.json",
		RollbackTier: Tier2, ProjectID: 1,
		Entities: []Entity{
			{Type: "section", ID: 10},
			{Type: "case", ID: 100, ParentID: 10},
			{Type: "case", ID: 101, ParentID: 10},
		},
		Timestamp: time.Now().UTC(),
	}

	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", cascade)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	var createdSectionReq *data.AddSectionRequest
	var createdCases []*data.AddCaseRequest
	api := &mockCasesAPI{
		addSectionFunc: func(_ context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
			assert.Equal(t, int64(1), projectID)
			createdSectionReq = req
			return &data.Section{ID: 50, Name: req.Name}, nil
		},
		addCaseFunc: func(_ context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
			assert.Equal(t, int64(50), sectionID) // should use new section ID
			createdCases = append(createdCases, req)
			return &data.Case{ID: 200 + int64(len(createdCases)), Title: req.Title}, nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(50), result.NewEntityID)

	require.NotNil(t, createdSectionReq)
	assert.Equal(t, "Auth", createdSectionReq.Name)
	assert.Len(t, createdCases, 2)
	assert.Equal(t, "Login", createdCases[0].Title)
	assert.Equal(t, "Logout", createdCases[1].Title)
}

func TestRollback_SectionCascade_DryRun(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	cascade := CascadeData{
		Section: data.Section{ID: 10, Name: "Auth"},
		Cases:   []data.Case{{ID: 100, Title: "Login"}, {ID: 101, Title: "Logout"}},
	}
	snapID := "sections/20260413T120000_delete_10"
	meta := &Meta{
		ID: snapID, Category: Category("sections"), Operation: OpDelete,
		EntityType: "section", EntityIDs: []int64{10}, Status: StatusAvailable,
		DataFile: "data.json", ProjectID: 1, Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", cascade)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{} // No API calls expected
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	assert.Len(t, result.Preview, 3) // 1 section + 2 cases
	assert.Contains(t, result.Preview[0].Saved, "RE-CREATE SECTION")
}

func TestRollback_SectionCascade_EntityFilter(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	cascade := CascadeData{
		Section: data.Section{ID: 10, Name: "Auth", SuiteID: 1},
		Cases: []data.Case{
			{ID: 100, Title: "Login", SectionID: 10},
			{ID: 101, Title: "Logout", SectionID: 10},
		},
	}
	snapID := "sections/20260413T120000_delete_10"
	meta := &Meta{
		ID: snapID, Category: Category("sections"), Operation: OpDelete,
		EntityType: "section", EntityIDs: []int64{10, 100, 101},
		Status: StatusAvailable, DataFile: "data.json", ProjectID: 1,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", cascade)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	createdCaseCount := 0
	api := &mockCasesAPI{
		addSectionFunc: func(_ context.Context, _ int64, req *data.AddSectionRequest) (*data.Section, error) {
			return &data.Section{ID: 50, Name: req.Name}, nil
		},
		addCaseFunc: func(_ context.Context, _ int64, req *data.AddCaseRequest) (*data.Case, error) {
			createdCaseCount++
			return &data.Case{ID: 200, Title: req.Title}, nil
		},
	}

	// Only restore case 100.
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{EntityIDs: []int64{100}})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 1, createdCaseCount, "should only restore filtered case")
}

// ---------------------------------------------------------------------------
// Sync rollback tests
// ---------------------------------------------------------------------------

func TestRollback_Sync_DeleteCreatedEntities(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	syncData := SyncData{
		SrcProject: 1, DstProject: 2,
		SrcSuite: 10, DstSuite: 20,
		Created: []SyncCreatedEntity{
			{Type: "suite", SourceID: 10, TargetID: 200},
			{Type: "section", SourceID: 11, TargetID: 201},
			{Type: "shared_step", SourceID: 12, TargetID: 202},
			{Type: "case", SourceID: 13, TargetID: 203},
			{Type: "case", SourceID: 14, TargetID: 204},
		},
	}

	snapID := "sync/20260413T120000_sync_full_1"
	meta := &Meta{
		ID: snapID, Category: CatSync, Operation: OpSyncFull,
		EntityType: "sync", Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))
	_, err = store.SaveData(snapID, "data.json", syncData)
	require.NoError(t, err)

	var deletedIDs []int64
	api := &mockCasesAPI{
		deleteCaseFunc: func(_ context.Context, id int64) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		},
		deleteSectionFunc: func(_ context.Context, id int64) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		},
		deleteSharedFunc: func(_ context.Context, id int64, _ int) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		},
		deleteSuiteFunc: func(_ context.Context, id int64) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "5/5")
	// Verify deletion order: cases first, then sections, then shared_steps, then suites.
	assert.Equal(t, []int64{203, 204, 201, 202, 200}, deletedIDs)
}

func TestRollback_Sync_DryRun(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	syncData := SyncData{
		Created: []SyncCreatedEntity{
			{Type: "case", SourceID: 1, TargetID: 100},
			{Type: "section", SourceID: 2, TargetID: 200},
		},
	}

	snapID := "sync/20260413T120000_sync_cases_1"
	meta := &Meta{
		ID: snapID, Category: CatSync, Operation: OpSyncCases,
		EntityType: "sync", Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))
	_, err = store.SaveData(snapID, "data.json", syncData)
	require.NoError(t, err)

	api := &mockCasesAPI{}
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.DryRun)
	assert.Len(t, result.Preview, 2)
}

func TestRollback_Sync_EntityFilter(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	syncData := SyncData{
		Created: []SyncCreatedEntity{
			{Type: "case", SourceID: 1, TargetID: 100},
			{Type: "case", SourceID: 2, TargetID: 200},
			{Type: "case", SourceID: 3, TargetID: 300},
		},
	}

	snapID := "sync/20260413T120000_sync_cases_1"
	meta := &Meta{
		ID: snapID, Category: CatSync, Operation: OpSyncCases,
		EntityType: "sync", Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))
	_, err = store.SaveData(snapID, "data.json", syncData)
	require.NoError(t, err)

	var deletedIDs []int64
	api := &mockCasesAPI{
		deleteCaseFunc: func(_ context.Context, id int64) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{EntityIDs: []int64{100, 300}})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, []int64{100, 300}, deletedIDs, "should only delete filtered entities")
}

// ---------------------------------------------------------------------------
// Phase 7: isGoneError
// ---------------------------------------------------------------------------

func TestIsGoneError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{"nil error", nil, false},
		{"400 error", fmt.Errorf("API returned 400: Bad Request"), true},
		{"404 error", fmt.Errorf("API returned 404: Not Found"), true},
		{"not found lower", fmt.Errorf("entity not found"), true},
		{"no longer exists", fmt.Errorf("The case no longer exists"), true},
		{"500 error", fmt.Errorf("API returned 500: Internal Server Error"), false},
		{"timeout", fmt.Errorf("connection timeout"), false},
		{"generic", fmt.Errorf("something went wrong"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, isGoneError(tc.err))
		})
	}
}

// ---------------------------------------------------------------------------
// Phase 7: rollbackCaseAdd graceful skip on 404
// ---------------------------------------------------------------------------

func TestRollbackCaseAdd_GracefulSkip_404(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260414T120000_add_bulk_0"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpAdd,
		EntityType: "case", EntityIDs: []int64{500}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	// DeleteCase returns 404 → should skip gracefully.
	api := &mockCasesAPI{
		deleteCaseFunc: func(_ context.Context, caseID int64) error {
			return fmt.Errorf("API returned 404: Not Found")
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "already deleted")

	// Log entry should be marked as restored (not failed).
	updatedMeta, err := store.LoadMeta(snapID)
	require.NoError(t, err)
	require.Len(t, updatedMeta.RollbackLog, 1)
	assert.Equal(t, RBRestored, updatedMeta.RollbackLog[0].Status)
}

func TestRollbackCaseAdd_GracefulSkip_400(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "cases/20260414T120000_add_bulk_1"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpAdd,
		EntityType: "case", EntityIDs: []int64{600}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		deleteCaseFunc: func(_ context.Context, caseID int64) error {
			return fmt.Errorf("API returned 400: Bad Request")
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "already deleted")
}

// ---------------------------------------------------------------------------
// Phase 7: rollbackCaseDelete graceful skip on 404 (section gone)
// ---------------------------------------------------------------------------

func TestRollbackCaseDelete_GracefulSkip_SectionGone(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	savedCase := data.Case{ID: 99, Title: "Deleted Case", SectionID: 200, TypeID: 2}
	snapID := "cases/20260414T120000_delete_99"
	meta := &Meta{
		ID: snapID, Category: Category("cases"), Operation: OpDelete,
		EntityType: "case", EntityIDs: []int64{99}, Status: StatusAvailable,
		DataFile: "data.json", RollbackTier: Tier2, Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", savedCase)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	// AddCase returns 404 (section deleted) → should skip gracefully.
	api := &mockCasesAPI{
		addCaseFunc: func(_ context.Context, sectionID int64, _ *data.AddCaseRequest) (*data.Case, error) {
			return nil, fmt.Errorf("API returned 404: section not found")
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "section")
	assert.Contains(t, result.Message, "no longer exists")

	// Log entry should be marked as failed (not restored, since the entity couldn't be re-created).
	updatedMeta, err := store.LoadMeta(snapID)
	require.NoError(t, err)
	require.Len(t, updatedMeta.RollbackLog, 1)
	assert.Equal(t, RBFailed, updatedMeta.RollbackLog[0].Status)
	assert.Contains(t, updatedMeta.RollbackLog[0].Error, "section")
}

// ---------------------------------------------------------------------------
// New entity type rollback tests
// ---------------------------------------------------------------------------

func TestRollback_SectionAdd(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "sections/20260413T120000_add_20"
	meta := &Meta{
		ID: snapID, Category: Category("sections"), Operation: OpAdd,
		EntityType: "section", EntityIDs: []int64{20}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteSectionFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(20), deletedID)
	assert.Contains(t, result.Message, "deleted (undo add)")
}

func TestRollback_SectionAdd_DryRun(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "sections/20260413T120000_add_30"
	meta := &Meta{
		ID: snapID, Category: Category("sections"), Operation: OpAdd,
		EntityType: "section", EntityIDs: []int64{30}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	require.NotEmpty(t, result.Preview)
	assert.Equal(t, "DELETE", result.Preview[0].Saved)
}

func TestRollback_SectionAdd_GoneSkip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "sections/20260413T120000_add_40"
	meta := &Meta{
		ID: snapID, Category: Category("sections"), Operation: OpAdd,
		EntityType: "section", EntityIDs: []int64{40}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		deleteSectionFunc: func(_ context.Context, _ int64) error {
			return fmt.Errorf("API returned 404: Not Found")
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "already deleted")
}

func TestRollback_ProjectAdd(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "projects/20260413T120000_add_5"
	meta := &Meta{
		ID: snapID, Category: Category("projects"), Operation: OpAdd,
		EntityType: "project", EntityIDs: []int64{5}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteProjectFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(5), deletedID)
	assert.Contains(t, result.Message, "deleted (undo add)")
}

func TestRollback_ProjectDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	projectData := ProjectData{
		Project: data.Project{ID: 7, Name: "Test Project", SuiteMode: 1},
	}
	snapID := "projects/20260413T120000_delete_7"
	meta := &Meta{
		ID: snapID, Category: Category("projects"), Operation: OpDelete,
		EntityType: "project", EntityIDs: []int64{7}, Status: StatusAvailable,
		DataFile: "data.json", RollbackTier: Tier2, Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", projectData)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	var capturedReq *data.AddProjectRequest
	api := &mockCasesAPI{
		addProjectFunc: func(_ context.Context, req *data.AddProjectRequest) (*data.GetProjectResponse, error) {
			capturedReq = req
			return &data.GetProjectResponse{ID: 50}, nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(50), result.NewEntityID)
	require.NotNil(t, capturedReq)
	assert.Equal(t, "Test Project", capturedReq.Name)
	assert.Equal(t, 1, capturedReq.SuiteMode)
	assert.Contains(t, result.Message, "re-created as ID 50")
}

func TestRollback_ProjectDelete_DryRun(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	projectData := ProjectData{
		Project: data.Project{ID: 7, Name: "Test Project", SuiteMode: 1},
	}
	snapID := "projects/20260413T120000_delete_7"
	meta := &Meta{
		ID: snapID, Category: Category("projects"), Operation: OpDelete,
		EntityType: "project", EntityIDs: []int64{7}, Status: StatusAvailable,
		DataFile: "data.json", Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	_, err = store.SaveData(snapID, "data.json", projectData)
	require.NoError(t, err)
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	require.NotEmpty(t, result.Preview)
	assert.Equal(t, "RE-CREATE", result.Preview[0].Saved)
	assert.Contains(t, result.Message, "Dry-run")
}

func TestRollback_SimpleEntity_Milestone(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "milestones/20260413T120000_add_10"
	meta := &Meta{
		ID: snapID, Category: Category("milestones"), Operation: OpAdd,
		EntityType: "milestone", EntityIDs: []int64{10}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteMilestoneFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(10), deletedID)
	assert.Contains(t, result.Message, "milestone 10 deleted (undo add)")
}

func TestRollback_SimpleEntity_Run(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "runs/20260413T120000_add_15"
	meta := &Meta{
		ID: snapID, Category: Category("runs"), Operation: OpAdd,
		EntityType: "run", EntityIDs: []int64{15}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteRunFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(15), deletedID)
	assert.Contains(t, result.Message, "run 15 deleted (undo add)")
}

func TestRollback_SimpleEntity_Plan(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "plans/20260413T120000_add_3"
	meta := &Meta{
		ID: snapID, Category: Category("plans"), Operation: OpAdd,
		EntityType: "plan", EntityIDs: []int64{3}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deletePlanFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(3), deletedID)
	assert.Contains(t, result.Message, "plan 3 deleted (undo add)")
}

func TestRollback_SimpleEntity_Group(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "groups/20260413T120000_add_8"
	meta := &Meta{
		ID: snapID, Category: Category("groups"), Operation: OpAdd,
		EntityType: "group", EntityIDs: []int64{8}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteGroupFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(8), deletedID)
	assert.Contains(t, result.Message, "group 8 deleted (undo add)")
}

func TestRollback_SimpleEntity_Variable(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "variables/20260413T120000_add_12"
	meta := &Meta{
		ID: snapID, Category: Category("variables"), Operation: OpAdd,
		EntityType: "variable", EntityIDs: []int64{12}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteVariableFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(12), deletedID)
	assert.Contains(t, result.Message, "variable 12 deleted (undo add)")
}

func TestRollback_SimpleEntity_Dataset(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "datasets/20260413T120000_add_4"
	meta := &Meta{
		ID: snapID, Category: Category("datasets"), Operation: OpAdd,
		EntityType: "dataset", EntityIDs: []int64{4}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedID int64
	api := &mockCasesAPI{
		deleteDatasetFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(4), deletedID)
	assert.Contains(t, result.Message, "dataset 4 deleted (undo add)")
}

func TestRollback_SimpleEntity_Configuration(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "configurations/20260413T120000_add_6"
	meta := &Meta{
		ID: snapID, Category: Category("configurations"), Operation: OpAdd,
		EntityType: "configuration", EntityIDs: []int64{6}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedConfigGroupID int64
	api := &mockCasesAPI{
		deleteConfigGroupFunc: func(_ context.Context, id int64) error {
			deletedConfigGroupID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(6), deletedConfigGroupID)
	assert.Contains(t, result.Message, "configuration 6 deleted (undo add)")
}

func TestRollback_SimpleEntity_Configuration_FallbackToConfig(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "configurations/20260413T120000_add_9"
	meta := &Meta{
		ID: snapID, Category: Category("configurations"), Operation: OpAdd,
		EntityType: "configuration", EntityIDs: []int64{9}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	var deletedConfigID int64
	api := &mockCasesAPI{
		deleteConfigGroupFunc: func(_ context.Context, _ int64) error {
			return fmt.Errorf("API returned 404: Not Found") // not a config group
		},
		deleteConfigFunc: func(_ context.Context, id int64) error {
			deletedConfigID = id
			return nil
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, int64(9), deletedConfigID)
	assert.Contains(t, result.Message, "configuration 9 deleted (undo add)")
}

func TestRollback_SimpleEntity_DryRun(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "milestones/20260413T120000_add_10"
	meta := &Meta{
		ID: snapID, Category: Category("milestones"), Operation: OpAdd,
		EntityType: "milestone", EntityIDs: []int64{10}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	result, err := Rollback(context.Background(), api, store, manifest, snapID, RollbackOpts{DryRun: true})
	require.NoError(t, err)
	assert.True(t, result.DryRun)
	require.NotEmpty(t, result.Preview)
	assert.Equal(t, "DELETE", result.Preview[0].Saved)
}

func TestRollback_SimpleEntity_UnsupportedOp(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "milestones/20260413T120000_update_10"
	meta := &Meta{
		ID: snapID, Category: Category("milestones"), Operation: OpUpdate,
		EntityType: "milestone", EntityIDs: []int64{10}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{}
	_, err = Rollback(context.Background(), api, store, manifest, snapID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported operation for milestone rollback")
}

func TestRollback_SimpleEntity_GoneSkip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)

	snapID := "runs/20260413T120000_add_25"
	meta := &Meta{
		ID: snapID, Category: Category("runs"), Operation: OpAdd,
		EntityType: "run", EntityIDs: []int64{25}, Status: StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(meta))
	require.NoError(t, manifest.Add(meta))

	api := &mockCasesAPI{
		deleteRunFunc: func(_ context.Context, _ int64) error {
			return fmt.Errorf("API returned 404: Not Found")
		},
	}

	result, err := Rollback(context.Background(), api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "already deleted")
}
