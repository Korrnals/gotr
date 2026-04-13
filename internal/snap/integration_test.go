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

// ---------------------------------------------------------------------------
// statefulCasesAPI — mock with in-memory state for integration tests.
// Simulates a minimal TestRail server: store cases, track mutations.
// ---------------------------------------------------------------------------

type statefulCasesAPI struct {
	cases   map[int64]*data.Case // in-memory "database"
	nextID  int64                // auto-increment for AddCase
	calls   []string             // audit log: "GetCase:42", "UpdateCase:42", etc.
	failOn  map[string]error     // inject errors: "UpdateCase:42" → error
}

func newStatefulAPI(initial ...*data.Case) *statefulCasesAPI {
	api := &statefulCasesAPI{
		cases:  make(map[int64]*data.Case),
		nextID: 1000,
		failOn: make(map[string]error),
	}
	for _, c := range initial {
		api.cases[c.ID] = c
		if c.ID >= api.nextID {
			api.nextID = c.ID + 1
		}
	}
	return api
}

func (a *statefulCasesAPI) log(method string, id int64) {
	a.calls = append(a.calls, fmt.Sprintf("%s:%d", method, id))
}

func (a *statefulCasesAPI) err(method string, id int64) error {
	key := fmt.Sprintf("%s:%d", method, id)
	if e, ok := a.failOn[key]; ok {
		return e
	}
	return nil
}

func (a *statefulCasesAPI) GetCase(_ context.Context, caseID int64) (*data.Case, error) {
	a.log("GetCase", caseID)
	if e := a.err("GetCase", caseID); e != nil {
		return nil, e
	}
	c, ok := a.cases[caseID]
	if !ok {
		return nil, fmt.Errorf("API returned 400: case %d not found", caseID)
	}
	// Return a copy to prevent aliasing.
	cp := *c
	return &cp, nil
}

func (a *statefulCasesAPI) UpdateCase(_ context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
	a.log("UpdateCase", caseID)
	if e := a.err("UpdateCase", caseID); e != nil {
		return nil, e
	}
	c, ok := a.cases[caseID]
	if !ok {
		return nil, fmt.Errorf("API returned 400: case %d not found", caseID)
	}
	// Apply update.
	if req.Title != nil {
		c.Title = *req.Title
	}
	if req.PriorityID != nil {
		c.PriorityID = *req.PriorityID
	}
	if req.TypeID != nil {
		c.TypeID = *req.TypeID
	}
	if req.Estimate != nil {
		c.Estimate = *req.Estimate
	}
	if req.Refs != nil {
		c.Refs = *req.Refs
	}
	if req.SectionID != nil {
		c.SectionID = *req.SectionID
	}
	if req.MilestoneID != nil {
		c.MilestoneID = *req.MilestoneID
	}
	if req.TemplateID != nil {
		c.TemplateID = *req.TemplateID
	}
	if req.CustomPreconds != nil {
		c.CustomPreconds = *req.CustomPreconds
	}
	if req.CustomSteps != nil {
		c.CustomSteps = *req.CustomSteps
	}
	if req.CustomExpected != nil {
		c.CustomExpected = *req.CustomExpected
	}
	if len(req.CustomStepsSeparated) > 0 {
		c.CustomStepsSeparated = req.CustomStepsSeparated
	}
	cp := *c
	return &cp, nil
}

func (a *statefulCasesAPI) AddCase(_ context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
	a.log("AddCase", sectionID)
	if e := a.err("AddCase", sectionID); e != nil {
		return nil, e
	}
	id := a.nextID
	a.nextID++
	c := &data.Case{
		ID:                   id,
		Title:                req.Title,
		SectionID:            sectionID,
		TypeID:               req.TypeID,
		PriorityID:           req.PriorityID,
		Estimate:             req.Estimate,
		Refs:                 req.Refs,
		MilestoneID:          req.MilestoneID,
		TemplateID:           req.TemplateID,
		CustomPreconds:       req.CustomPreconds,
		CustomSteps:          req.CustomSteps,
		CustomExpected:       req.CustomExpected,
		CustomStepsSeparated: req.CustomStepsSeparated,
	}
	a.cases[id] = c
	cp := *c
	return &cp, nil
}

func (a *statefulCasesAPI) DeleteCase(_ context.Context, caseID int64) error {
	a.log("DeleteCase", caseID)
	if e := a.err("DeleteCase", caseID); e != nil {
		return e
	}
	if _, ok := a.cases[caseID]; !ok {
		return fmt.Errorf("API returned 400: case %d not found", caseID)
	}
	delete(a.cases, caseID)
	return nil
}

// ---------------------------------------------------------------------------
// Helper: create store + manifest in t.TempDir
// ---------------------------------------------------------------------------

func setupIntegration(t *testing.T) (*Store, *Manifest) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest, err := LoadManifest(store)
	require.NoError(t, err)
	return store, manifest
}

// ---------------------------------------------------------------------------
// Smoke test 1: Update → Snap → Mutate → Rollback → Verify original state
// ---------------------------------------------------------------------------

func TestIntegration_UpdateRollback(t *testing.T) {
	store, manifest := setupIntegration(t)

	// Seed API with original case.
	original := &data.Case{
		ID:             42,
		Title:          "Login test",
		SectionID:      10,
		PriorityID:     3,
		TypeID:         1,
		Estimate:       "5m",
		CustomPreconds: "User exists",
		CustomSteps:    "1. Open login\n2. Enter creds\n3. Submit",
		CustomExpected: "Dashboard shown",
		Refs:           "REQ-100",
	}
	api := newStatefulAPI(original)

	ctx := context.Background()

	// 1. Take snapshot (simulates hook.Before).
	meta := BuildMeta(OpUpdate, "case", []int64{42}, Tier1, 1, 1, "", []string{"cases", "update", "42"})
	snap, err := TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 42)
	})
	require.NoError(t, err)
	require.NotNil(t, snap)
	snapID := snap.Meta.ID

	// 2. Mutate the case (simulates the actual update command).
	newTitle := "Login test (UPDATED)"
	newPriority := int64(1)
	_, err = api.UpdateCase(ctx, 42, &data.UpdateCaseRequest{
		Title:      &newTitle,
		PriorityID: &newPriority,
	})
	require.NoError(t, err)

	// Verify mutation applied.
	mutated, _ := api.GetCase(ctx, 42)
	assert.Equal(t, "Login test (UPDATED)", mutated.Title)
	assert.Equal(t, int64(1), mutated.PriorityID)

	// 3. Rollback.
	result, err := Rollback(ctx, api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "restored to pre-update state")

	// 4. Verify original state restored.
	restored, _ := api.GetCase(ctx, 42)
	assert.Equal(t, "Login test", restored.Title)
	assert.Equal(t, int64(3), restored.PriorityID)
	assert.Equal(t, int64(1), restored.TypeID)
	assert.Equal(t, "5m", restored.Estimate)
	assert.Equal(t, "User exists", restored.CustomPreconds)
	assert.Equal(t, "REQ-100", restored.Refs)

	// 5. Manifest shows rolled_back.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusRolledBack, entry.Status)
}

// ---------------------------------------------------------------------------
// Smoke test 2: Delete → Snap → Delete from API → Rollback → Re-created
// ---------------------------------------------------------------------------

func TestIntegration_DeleteRollback(t *testing.T) {
	store, manifest := setupIntegration(t)

	original := &data.Case{
		ID:             99,
		Title:          "Checkout flow",
		SectionID:      20,
		PriorityID:     2,
		TypeID:         3,
		CustomPreconds: "Cart not empty",
		CustomSteps:    "1. Go to checkout\n2. Pay",
		CustomExpected: "Order confirmed",
	}
	api := newStatefulAPI(original)
	ctx := context.Background()

	// 1. Take snapshot before delete.
	meta := BuildMeta(OpDelete, "case", []int64{99}, Tier2, 1, 1, "", []string{"cases", "delete", "99"})
	snap, err := TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 99)
	})
	require.NoError(t, err)
	snapID := snap.Meta.ID

	// 2. Delete the case.
	require.NoError(t, api.DeleteCase(ctx, 99))

	// Verify gone.
	_, err = api.GetCase(ctx, 99)
	assert.Error(t, err)

	// 3. Rollback (re-create).
	result, err := Rollback(ctx, api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotZero(t, result.NewEntityID)
	assert.Contains(t, result.Message, "re-created")

	// 4. The re-created case exists with original data (but new ID).
	recreated, err := api.GetCase(ctx, result.NewEntityID)
	require.NoError(t, err)
	assert.Equal(t, "Checkout flow", recreated.Title)
	assert.Equal(t, int64(20), recreated.SectionID)
	assert.Equal(t, int64(2), recreated.PriorityID)
	assert.Equal(t, int64(3), recreated.TypeID)
	assert.Equal(t, "Cart not empty", recreated.CustomPreconds)

	// 5. Old ID is still gone (Tier 2: new ID).
	_, err = api.GetCase(ctx, 99)
	assert.Error(t, err)

	// 6. Manifest shows rolled_back.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusRolledBack, entry.Status)
}

// ---------------------------------------------------------------------------
// Smoke test 3: Add → Snap + FinalizeAdd → Rollback → Deleted
// ---------------------------------------------------------------------------

func TestIntegration_AddRollback(t *testing.T) {
	store, manifest := setupIntegration(t)

	api := newStatefulAPI() // empty
	ctx := context.Background()

	// 1. Take snapshot for add (no fetchFn).
	meta := BuildMeta(OpAdd, "case", nil, Tier2, 1, 1, "", []string{"cases", "add"})
	snapObj, err := TakeSnapshot(ctx, store, manifest, meta, nil)
	require.NoError(t, err)
	snapID := snapObj.Meta.ID

	// 2. Add the case.
	created, err := api.AddCase(ctx, 50, &data.AddCaseRequest{
		Title:      "Brand new test",
		SectionID:  50,
		PriorityID: 2,
		TypeID:     1,
	})
	require.NoError(t, err)

	// 3. FinalizeAdd — record createdID in meta.
	require.NoError(t, snapObj.FinalizeAdd(created.ID))

	// Verify case exists.
	_, err = api.GetCase(ctx, created.ID)
	require.NoError(t, err)

	// 4. Rollback (delete the created case).
	result, err := Rollback(ctx, api, store, manifest, snapID)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "deleted")

	// 5. Case is gone.
	_, err = api.GetCase(ctx, created.ID)
	assert.Error(t, err)

	// 6. Manifest shows rolled_back.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusRolledBack, entry.Status)
}

// ---------------------------------------------------------------------------
// Smoke test 4: Double rollback is rejected
// ---------------------------------------------------------------------------

func TestIntegration_DoubleRollbackRejected(t *testing.T) {
	store, manifest := setupIntegration(t)

	original := &data.Case{ID: 10, Title: "Alpha", SectionID: 1, PriorityID: 1}
	api := newStatefulAPI(original)
	ctx := context.Background()

	meta := BuildMeta(OpUpdate, "case", []int64{10}, Tier1, 1, 1, "", []string{"cases", "update", "10"})
	snap, err := TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 10)
	})
	require.NoError(t, err)
	snapID := snap.Meta.ID

	// Mutate.
	newTitle := "Alpha v2"
	_, err = api.UpdateCase(ctx, 10, &data.UpdateCaseRequest{Title: &newTitle})
	require.NoError(t, err)

	// First rollback succeeds.
	_, err = Rollback(ctx, api, store, manifest, snapID)
	require.NoError(t, err)

	// Second rollback fails — status is "rolled_back", not "available".
	_, err = Rollback(ctx, api, store, manifest, snapID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rolled_back")
	assert.Contains(t, err.Error(), "available")
}

// ---------------------------------------------------------------------------
// Smoke test 5: Custom name → snapshot stored in custom/ category
// ---------------------------------------------------------------------------

func TestIntegration_CustomNameCategory(t *testing.T) {
	store, manifest := setupIntegration(t)

	original := &data.Case{ID: 7, Title: "Auth test", SectionID: 3, PriorityID: 2}
	api := newStatefulAPI(original)
	ctx := context.Background()

	meta := BuildMeta(OpUpdate, "case", []int64{7}, Tier1, 1, 1, "before-refactor", []string{"cases", "update", "7", "--snap-name=before-refactor"})
	snap, err := TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 7)
	})
	require.NoError(t, err)
	snapID := snap.Meta.ID

	// ID must start with "custom/"
	assert.Contains(t, snapID, "custom/")
	assert.Contains(t, snapID, "before-refactor")

	// Category in manifest is "custom".
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, CatCustom, entry.Category)

	// Data file exists on disk.
	assert.True(t, store.Exists(snapID))
}

// ---------------------------------------------------------------------------
// Smoke test 6: API error during rollback → status becomes rollback_partial
// ---------------------------------------------------------------------------

func TestIntegration_APIErrorPartialRollback(t *testing.T) {
	store, manifest := setupIntegration(t)

	original := &data.Case{ID: 55, Title: "Flaky test", SectionID: 5, PriorityID: 1}
	api := newStatefulAPI(original)
	ctx := context.Background()

	meta := BuildMeta(OpUpdate, "case", []int64{55}, Tier1, 1, 1, "", []string{"cases", "update", "55"})
	snap, err := TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 55)
	})
	require.NoError(t, err)
	snapID := snap.Meta.ID

	// Inject API failure.
	api.failOn["UpdateCase:55"] = fmt.Errorf("API returned 503: service unavailable")

	// Rollback fails.
	result, err := Rollback(ctx, api, store, manifest, snapID)
	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "503")

	// Status becomes rollback_partial.
	entry := manifest.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, StatusRollbackPartial, entry.Status)
}

// ---------------------------------------------------------------------------
// Smoke test 7: GC orphans — snapshot on disk without manifest entry
// ---------------------------------------------------------------------------

func TestIntegration_GCOrphans(t *testing.T) {
	store, manifest := setupIntegration(t)

	original := &data.Case{ID: 1, Title: "Tracked", SectionID: 1}
	api := newStatefulAPI(original)
	ctx := context.Background()

	// 1. Create a tracked snapshot.
	meta := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 1, 1, "", []string{"cases", "update", "1"})
	_, err := TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 1)
	})
	require.NoError(t, err)

	// 2. Create an orphan: write files directly without manifest entry.
	orphanID := "cases/orphan_test"
	orphanMeta := &Meta{
		ID:        orphanID,
		Category:  "cases",
		Operation: OpUpdate,
		Status:    StatusAvailable,
		Timestamp: time.Now().UTC(),
	}
	require.NoError(t, store.SaveMeta(orphanMeta))

	// Verify orphan exists.
	assert.True(t, store.Exists(orphanID))

	// 3. Run GC.
	cleaned, err := store.CleanOrphans(manifest.ManifestIDs())
	require.NoError(t, err)
	assert.Equal(t, 1, cleaned)

	// 4. Orphan is gone.
	assert.False(t, store.Exists(orphanID))

	// 5. Tracked snapshot still exists.
	ids, _ := store.List()
	assert.Len(t, ids, 1)
}

// ---------------------------------------------------------------------------
// Smoke test 8: Multiple snapshots → List + filter by category / status
// ---------------------------------------------------------------------------

func TestIntegration_ManifestFiltering(t *testing.T) {
	store, manifest := setupIntegration(t)

	api := newStatefulAPI(
		&data.Case{ID: 1, Title: "A", SectionID: 1},
		&data.Case{ID: 2, Title: "B", SectionID: 2},
	)
	ctx := context.Background()

	// Snapshot 1: update case 1.
	m1 := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 1, 1, "", []string{"cases", "update", "1"})
	_, err := TakeSnapshot(ctx, store, manifest, m1, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 1)
	})
	require.NoError(t, err)

	// Snapshot 2: delete case 2.
	m2 := BuildMeta(OpDelete, "case", []int64{2}, Tier2, 1, 1, "", []string{"cases", "delete", "2"})
	_, err = TakeSnapshot(ctx, store, manifest, m2, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 2)
	})
	require.NoError(t, err)

	// Snapshot 3: custom name.
	m3 := BuildMeta(OpUpdate, "case", []int64{1}, Tier1, 1, 1, "my-backup", []string{"cases", "update", "1", "--snap-name=my-backup"})
	_, err = TakeSnapshot(ctx, store, manifest, m3, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 1)
	})
	require.NoError(t, err)

	// Total: 3 entries.
	assert.Equal(t, 3, manifest.Len())

	// Filter by category "cases".
	cases := manifest.ListByCategory("cases")
	assert.Len(t, cases, 2)

	// Filter by category "custom".
	custom := manifest.ListByCategory(CatCustom)
	assert.Len(t, custom, 1)
	assert.Equal(t, "my-backup", custom[0].Name)

	// Filter by status.
	available := manifest.ListByStatus(StatusAvailable)
	assert.Len(t, available, 3)

	// All.
	all := manifest.All()
	assert.Len(t, all, 3)
}

// ---------------------------------------------------------------------------
// Smoke test 9: Snapshot data persists across store reload
// ---------------------------------------------------------------------------

func TestIntegration_PersistenceAcrossReload(t *testing.T) {
	dir := t.TempDir()
	store1, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest1, err := LoadManifest(store1)
	require.NoError(t, err)

	original := &data.Case{ID: 77, Title: "Persistent", SectionID: 8, PriorityID: 5}
	api := newStatefulAPI(original)
	ctx := context.Background()

	meta := BuildMeta(OpUpdate, "case", []int64{77}, Tier1, 1, 1, "", []string{"cases", "update", "77"})
	snap, err := TakeSnapshot(ctx, store1, manifest1, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 77)
	})
	require.NoError(t, err)
	snapID := snap.Meta.ID

	// Reload store + manifest from the same directory.
	store2, err := NewStoreAt(dir)
	require.NoError(t, err)
	manifest2, err := LoadManifest(store2)
	require.NoError(t, err)

	// Manifest entry survives reload.
	entry := manifest2.Find(snapID)
	require.NotNil(t, entry)
	assert.Equal(t, OpUpdate, entry.Operation)
	assert.Equal(t, StatusAvailable, entry.Status)

	// Data survives reload.
	var loaded data.Case
	require.NoError(t, store2.LoadData(snapID, "data.json", &loaded))
	assert.Equal(t, int64(77), loaded.ID)
	assert.Equal(t, "Persistent", loaded.Title)
	assert.Equal(t, int64(5), loaded.PriorityID)
}

// ---------------------------------------------------------------------------
// Smoke test 10: Full cycle — snapshot + delete + rollback (re-create) + update re-created
// ---------------------------------------------------------------------------

func TestIntegration_FullCycleDeleteAndReuse(t *testing.T) {
	store, manifest := setupIntegration(t)

	original := &data.Case{
		ID:             200,
		Title:          "Payment flow",
		SectionID:      30,
		PriorityID:     4,
		TypeID:         2,
		CustomPreconds: "User logged in, cart has items",
		CustomSteps:    "1. Click pay\n2. Enter card\n3. Confirm",
		CustomExpected: "Payment success",
	}
	api := newStatefulAPI(original)
	ctx := context.Background()

	// 1. Snapshot + delete.
	meta := BuildMeta(OpDelete, "case", []int64{200}, Tier2, 1, 1, "", []string{"cases", "delete", "200"})
	snap, err := TakeSnapshot(ctx, store, manifest, meta, func(ctx context.Context) (interface{}, error) {
		return api.GetCase(ctx, 200)
	})
	require.NoError(t, err)
	snapID := snap.Meta.ID

	require.NoError(t, api.DeleteCase(ctx, 200))

	// 2. Rollback → re-creates with new ID.
	result, err := Rollback(ctx, api, store, manifest, snapID)
	require.NoError(t, err)
	newID := result.NewEntityID
	require.NotZero(t, newID)

	// 3. The re-created entity can be mutated again.
	newTitle := "Payment flow v2"
	updated, err := api.UpdateCase(ctx, newID, &data.UpdateCaseRequest{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "Payment flow v2", updated.Title)

	// 4. Verify all API calls were in expected order.
	expected := []string{
		"GetCase:200",       // snapshot fetch
		"DeleteCase:200",    // user mutation
		"AddCase:30",        // rollback re-create (section 30)
		"UpdateCase:" + fmt.Sprintf("%d", newID), // subsequent update
	}
	assert.Equal(t, expected, api.calls)
}
