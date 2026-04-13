package snap

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// CasesAPI defines the API methods needed for case rollback operations.
// client.ClientInterface satisfies this interface.
type CasesAPI interface {
	GetCase(ctx context.Context, caseID int64) (*data.Case, error)
	UpdateCase(ctx context.Context, caseID int64, req *data.UpdateCaseRequest) (*data.Case, error)
	AddCase(ctx context.Context, sectionID int64, req *data.AddCaseRequest) (*data.Case, error)
	DeleteCase(ctx context.Context, caseID int64) error
}

// RollbackResult holds the outcome of a rollback operation.
type RollbackResult struct {
	SnapID      string
	Operation   Operation
	EntityType  string
	Success     bool
	NewEntityID int64  // non-zero if a new entity was created (e.g. delete rollback)
	Message     string
}

// Rollback reverses a mutation using the saved snapshot data.
func Rollback(ctx context.Context, api CasesAPI, store *Store, manifest *Manifest, snapID string) (*RollbackResult, error) {
	meta, err := store.LoadMeta(snapID)
	if err != nil {
		return nil, fmt.Errorf("load snapshot: %w", err)
	}

	if meta.Status != StatusAvailable {
		return nil, fmt.Errorf("snapshot %q status is %q, expected %q", snapID, meta.Status, StatusAvailable)
	}

	result := &RollbackResult{
		SnapID:     snapID,
		Operation:  meta.Operation,
		EntityType: meta.EntityType,
	}

	switch meta.EntityType {
	case "case":
		err = rollbackCase(ctx, api, store, meta, result)
	default:
		return nil, fmt.Errorf("unsupported entity type for rollback: %q", meta.EntityType)
	}

	if err != nil {
		meta.Status = StatusRollbackPartial
		_ = store.SaveMeta(meta)
		_ = manifest.UpdateStatus(snapID, StatusRollbackPartial)
		return result, err
	}

	result.Success = true
	meta.Status = StatusRolledBack
	_ = store.SaveMeta(meta)
	_ = manifest.UpdateStatus(snapID, StatusRolledBack)
	return result, nil
}

func rollbackCase(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult) error {
	switch meta.Operation {
	case OpUpdate:
		return rollbackCaseUpdate(ctx, api, store, meta, result)
	case OpDelete:
		return rollbackCaseDelete(ctx, api, store, meta, result)
	case OpAdd:
		return rollbackCaseAdd(ctx, api, meta, result)
	default:
		return fmt.Errorf("unsupported operation for case rollback: %q", meta.Operation)
	}
}

// rollbackCaseUpdate restores the pre-update state of a case.
func rollbackCaseUpdate(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult) error {
	var saved data.Case
	if err := store.LoadData(meta.ID, meta.DataFile, &saved); err != nil {
		return fmt.Errorf("load saved case data: %w", err)
	}

	if len(meta.EntityIDs) == 0 {
		return fmt.Errorf("no entity ID in snapshot meta")
	}
	caseID := meta.EntityIDs[0]

	req := caseToUpdateRequest(&saved)
	_, err := api.UpdateCase(ctx, caseID, req)
	if err != nil {
		return fmt.Errorf("API update_case %d: %w", caseID, err)
	}

	result.Message = fmt.Sprintf("Case %d restored to pre-update state", caseID)
	return nil
}

// rollbackCaseDelete re-creates a deleted case from snapshot data.
func rollbackCaseDelete(ctx context.Context, api CasesAPI, store *Store, meta *Meta, result *RollbackResult) error {
	var saved data.Case
	if err := store.LoadData(meta.ID, meta.DataFile, &saved); err != nil {
		return fmt.Errorf("load saved case data: %w", err)
	}

	req := caseToAddRequest(&saved)
	created, err := api.AddCase(ctx, saved.SectionID, req)
	if err != nil {
		return fmt.Errorf("API add_case (re-create): %w", err)
	}

	result.NewEntityID = created.ID
	result.Message = fmt.Sprintf("Case re-created as ID %d (original: %d, Tier 2: new ID)", created.ID, saved.ID)
	return nil
}

// rollbackCaseAdd deletes a case that was created by an add operation.
func rollbackCaseAdd(ctx context.Context, api CasesAPI, meta *Meta, result *RollbackResult) error {
	if len(meta.EntityIDs) == 0 {
		return fmt.Errorf("no created entity ID in snapshot meta (FinalizeAdd may not have been called)")
	}
	caseID := meta.EntityIDs[0]

	if err := api.DeleteCase(ctx, caseID); err != nil {
		return fmt.Errorf("API delete_case %d: %w", caseID, err)
	}

	result.Message = fmt.Sprintf("Case %d deleted (undo add)", caseID)
	return nil
}

// caseToUpdateRequest converts a saved Case to an UpdateCaseRequest.
func caseToUpdateRequest(c *data.Case) *data.UpdateCaseRequest {
	req := &data.UpdateCaseRequest{}
	if c.Title != "" {
		req.Title = &c.Title
	}
	if c.TypeID != 0 {
		req.TypeID = &c.TypeID
	}
	if c.PriorityID != 0 {
		req.PriorityID = &c.PriorityID
	}
	if c.Estimate != "" {
		req.Estimate = &c.Estimate
	}
	if c.CustomPreconds != "" {
		req.CustomPreconds = &c.CustomPreconds
	}
	if c.CustomSteps != "" {
		req.CustomSteps = &c.CustomSteps
	}
	if c.CustomExpected != "" {
		req.CustomExpected = &c.CustomExpected
	}
	if len(c.CustomStepsSeparated) > 0 {
		req.CustomStepsSeparated = c.CustomStepsSeparated
	}
	if c.Refs != "" {
		req.Refs = &c.Refs
	}
	if c.MilestoneID != 0 {
		req.MilestoneID = &c.MilestoneID
	}
	if c.TemplateID != 0 {
		req.TemplateID = &c.TemplateID
	}
	if c.SectionID != 0 {
		req.SectionID = &c.SectionID
	}
	return req
}

// caseToAddRequest converts a saved Case to an AddCaseRequest for re-creation.
func caseToAddRequest(c *data.Case) *data.AddCaseRequest {
	return &data.AddCaseRequest{
		Title:                c.Title,
		SectionID:            c.SectionID,
		TypeID:               c.TypeID,
		PriorityID:           c.PriorityID,
		Estimate:             c.Estimate,
		TemplateID:           c.TemplateID,
		Refs:                 c.Refs,
		MilestoneID:          c.MilestoneID,
		CustomPreconds:       c.CustomPreconds,
		CustomSteps:          c.CustomSteps,
		CustomExpected:       c.CustomExpected,
		CustomStepsSeparated: c.CustomStepsSeparated,
	}
}
