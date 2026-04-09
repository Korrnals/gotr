// internal/client/extended.go
// Extended APIs: Groups, Roles, ResultFields, Variables, Datasets, BDDs, Labels

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// ==================== Groups API ====================

// GetGroups fetches the group list for a project.
func (c *HTTPClient) GetGroups(ctx context.Context, projectID int64) (data.GetGroupsResponse, error) {
	endpoint := fmt.Sprintf("get_groups/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting groups for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var groups data.GetGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("error decoding groups: %w", err)
	}
	return groups, nil
}

// GetGroup fetches a group by ID.
func (c *HTTPClient) GetGroup(ctx context.Context, groupID int64) (*data.Group, error) {
	endpoint := fmt.Sprintf("get_group/%d", groupID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting group %d: %w", groupID, err)
	}
	defer resp.Body.Close()

	var group data.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding group: %w", err)
	}
	return &group, nil
}

// AddGroup creates a new group.
func (c *HTTPClient) AddGroup(ctx context.Context, projectID int64, name string, userIDs []int64) (*data.Group, error) {
	endpoint := fmt.Sprintf("add_group/%d", projectID)
	req := map[string]interface{}{
		"name":     name,
		"user_ids": userIDs,
	}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating group: %w", err)
	}
	defer resp.Body.Close()

	var group data.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding group: %w", err)
	}
	return &group, nil
}

// UpdateGroup updates a group.
func (c *HTTPClient) UpdateGroup(ctx context.Context, groupID int64, name string, userIDs []int64) (*data.Group, error) {
	endpoint := fmt.Sprintf("update_group/%d", groupID)
	req := map[string]interface{}{
		"name":     name,
		"user_ids": userIDs,
	}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating group: %w", err)
	}
	defer resp.Body.Close()

	var group data.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding group: %w", err)
	}
	return &group, nil
}

// DeleteGroup deletes a group.
func (c *HTTPClient) DeleteGroup(ctx context.Context, groupID int64) error {
	endpoint := fmt.Sprintf("delete_group/%d", groupID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting group: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// ==================== Roles API ====================

// GetRoles fetches the list of roles.
func (c *HTTPClient) GetRoles(ctx context.Context) (data.GetRolesResponse, error) {
	resp, err := c.Get(ctx, "get_roles", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting roles: %w", err)
	}
	defer resp.Body.Close()

	var roles data.GetRolesResponse
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, fmt.Errorf("error decoding roles: %w", err)
	}
	return roles, nil
}

// GetRole fetches a role by ID.
func (c *HTTPClient) GetRole(ctx context.Context, roleID int64) (*data.Role, error) {
	endpoint := fmt.Sprintf("get_role/%d", roleID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting role %d: %w", roleID, err)
	}
	defer resp.Body.Close()

	var role data.Role
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return nil, fmt.Errorf("error decoding role: %w", err)
	}
	return &role, nil
}

// ==================== ResultFields API ====================

// GetResultFields fetches the list of result fields.
func (c *HTTPClient) GetResultFields(ctx context.Context) (data.GetResultFieldsResponse, error) {
	resp, err := c.Get(ctx, "get_result_fields", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting result fields: %w", err)
	}
	defer resp.Body.Close()

	var fields data.GetResultFieldsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fields); err != nil {
		return nil, fmt.Errorf("error decoding result fields: %w", err)
	}
	return fields, nil
}

// ==================== Datasets API ====================

// GetDatasets fetches the dataset list for a project.
func (c *HTTPClient) GetDatasets(ctx context.Context, projectID int64) (data.GetDatasetsResponse, error) {
	endpoint := fmt.Sprintf("get_datasets/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting datasets for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var datasets data.GetDatasetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&datasets); err != nil {
		return nil, fmt.Errorf("error decoding datasets: %w", err)
	}
	return datasets, nil
}

// GetDataset fetches a dataset by ID.
func (c *HTTPClient) GetDataset(ctx context.Context, datasetID int64) (*data.Dataset, error) {
	endpoint := fmt.Sprintf("get_dataset/%d", datasetID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting dataset %d: %w", datasetID, err)
	}
	defer resp.Body.Close()

	var dataset data.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&dataset); err != nil {
		return nil, fmt.Errorf("error decoding dataset: %w", err)
	}
	return &dataset, nil
}

// AddDataset creates a new dataset.
func (c *HTTPClient) AddDataset(ctx context.Context, projectID int64, name string) (*data.Dataset, error) {
	endpoint := fmt.Sprintf("add_dataset/%d", projectID)
	req := map[string]string{"name": name}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating dataset: %w", err)
	}
	defer resp.Body.Close()

	var dataset data.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&dataset); err != nil {
		return nil, fmt.Errorf("error decoding dataset: %w", err)
	}
	return &dataset, nil
}

// UpdateDataset updates a dataset.
func (c *HTTPClient) UpdateDataset(ctx context.Context, datasetID int64, name string) (*data.Dataset, error) {
	endpoint := fmt.Sprintf("update_dataset/%d", datasetID)
	req := map[string]string{"name": name}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating dataset: %w", err)
	}
	defer resp.Body.Close()

	var dataset data.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&dataset); err != nil {
		return nil, fmt.Errorf("error decoding dataset: %w", err)
	}
	return &dataset, nil
}

// DeleteDataset deletes a dataset.
func (c *HTTPClient) DeleteDataset(ctx context.Context, datasetID int64) error {
	endpoint := fmt.Sprintf("delete_dataset/%d", datasetID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting dataset: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// ==================== Variables API ====================

// GetVariables fetches the variable list for a dataset.
func (c *HTTPClient) GetVariables(ctx context.Context, datasetID int64) (data.GetVariablesResponse, error) {
	endpoint := fmt.Sprintf("get_variables/%d", datasetID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting variables for dataset %d: %w", datasetID, err)
	}
	defer resp.Body.Close()

	var variables data.GetVariablesResponse
	if err := json.NewDecoder(resp.Body).Decode(&variables); err != nil {
		return nil, fmt.Errorf("error decoding variables: %w", err)
	}
	return variables, nil
}

// AddVariable adds a variable to a dataset.
func (c *HTTPClient) AddVariable(ctx context.Context, datasetID int64, name string) (*data.Variable, error) {
	endpoint := fmt.Sprintf("add_variable/%d", datasetID)
	req := map[string]string{"name": name}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating variable: %w", err)
	}
	defer resp.Body.Close()

	var variable data.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("error decoding variable: %w", err)
	}
	return &variable, nil
}

// UpdateVariable updates a variable.
func (c *HTTPClient) UpdateVariable(ctx context.Context, variableID int64, name string) (*data.Variable, error) {
	endpoint := fmt.Sprintf("update_variable/%d", variableID)
	req := map[string]string{"name": name}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating variable: %w", err)
	}
	defer resp.Body.Close()

	var variable data.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("error decoding variable: %w", err)
	}
	return &variable, nil
}

// DeleteVariable deletes a variable.
func (c *HTTPClient) DeleteVariable(ctx context.Context, variableID int64) error {
	endpoint := fmt.Sprintf("delete_variable/%d", variableID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting variable: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// ==================== BDDs API ====================

// GetBDD fetches the BDD scenario for a case.
func (c *HTTPClient) GetBDD(ctx context.Context, caseID int64) (*data.BDD, error) {
	endpoint := fmt.Sprintf("get_bdd/%d", caseID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting BDD for case %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	var bdd data.BDD
	if err := json.NewDecoder(resp.Body).Decode(&bdd); err != nil {
		return nil, fmt.Errorf("error decoding BDD: %w", err)
	}
	return &bdd, nil
}

// AddBDD adds a BDD scenario to a case.
func (c *HTTPClient) AddBDD(ctx context.Context, caseID int64, content string) (*data.BDD, error) {
	endpoint := fmt.Sprintf("add_bdd/%d", caseID)
	req := map[string]string{"content": content}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating BDD: %w", err)
	}
	defer resp.Body.Close()

	var bdd data.BDD
	if err := json.NewDecoder(resp.Body).Decode(&bdd); err != nil {
		return nil, fmt.Errorf("error decoding BDD: %w", err)
	}
	return &bdd, nil
}

// ==================== Labels API ====================

// UpdateTestLabels updates labels for a test.
func (c *HTTPClient) UpdateTestLabels(ctx context.Context, testID int64, labels []string) error {
	endpoint := fmt.Sprintf("update_test_labels/%d", testID)
	req := data.UpdateLabelsRequest{Labels: labels}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return fmt.Errorf("error updating test labels: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// UpdateTestsLabels updates labels for multiple tests.
func (c *HTTPClient) UpdateTestsLabels(ctx context.Context, runID int64, testIDs []int64, labels []string) error {
	endpoint := fmt.Sprintf("update_tests_labels/%d", runID)
	req := data.UpdateTestsLabelsRequest{
		TestIDs: testIDs,
		Labels:  labels,
	}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return fmt.Errorf("error updating tests labels: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetLabels fetches the label list for a project.
func (c *HTTPClient) GetLabels(ctx context.Context, projectID int64) (data.GetLabelsResponse, error) {
	endpoint := fmt.Sprintf("get_labels/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting labels for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var labels data.GetLabelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&labels); err != nil {
		return nil, fmt.Errorf("error decoding labels: %w", err)
	}
	return labels, nil
}

// GetLabel fetches a label by ID.
func (c *HTTPClient) GetLabel(ctx context.Context, labelID int64) (*data.Label, error) {
	endpoint := fmt.Sprintf("get_label/%d", labelID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting label %d: %w", labelID, err)
	}
	defer resp.Body.Close()

	var label data.Label
	if err := json.NewDecoder(resp.Body).Decode(&label); err != nil {
		return nil, fmt.Errorf("error decoding label: %w", err)
	}
	return &label, nil
}

// UpdateLabel updates a label.
func (c *HTTPClient) UpdateLabel(ctx context.Context, labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
	endpoint := fmt.Sprintf("update_label/%d", labelID)
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating label: %w", err)
	}
	defer resp.Body.Close()

	var label data.Label
	if err := json.NewDecoder(resp.Body).Decode(&label); err != nil {
		return nil, fmt.Errorf("error decoding label: %w", err)
	}
	return &label, nil
}
