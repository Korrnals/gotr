// internal/client/extended.go
// Extended APIs: Groups, Roles, ResultFields, Variables, Datasets, BDDs, Labels

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// ==================== Groups API ====================

// GetGroups получает список групп проекта
func (c *HTTPClient) GetGroups(projectID int64) (data.GetGroupsResponse, error) {
	endpoint := fmt.Sprintf("get_groups/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting groups for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var groups data.GetGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("error decoding groups: %w", err)
	}
	return groups, nil
}

// GetGroup получает группу по ID
func (c *HTTPClient) GetGroup(groupID int64) (*data.Group, error) {
	endpoint := fmt.Sprintf("get_group/%d", groupID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting group %d: %w", groupID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var group data.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding group: %w", err)
	}
	return &group, nil
}

// AddGroup создает новую группу
func (c *HTTPClient) AddGroup(projectID int64, name string, userIDs []int64) (*data.Group, error) {
	endpoint := fmt.Sprintf("add_group/%d", projectID)
	req := map[string]interface{}{
		"name":     name,
		"user_ids": userIDs,
	}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var group data.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding group: %w", err)
	}
	return &group, nil
}

// UpdateGroup обновляет группу
func (c *HTTPClient) UpdateGroup(groupID int64, name string, userIDs []int64) (*data.Group, error) {
	endpoint := fmt.Sprintf("update_group/%d", groupID)
	req := map[string]interface{}{
		"name":     name,
		"user_ids": userIDs,
	}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var group data.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding group: %w", err)
	}
	return &group, nil
}

// DeleteGroup удаляет группу
func (c *HTTPClient) DeleteGroup(groupID int64) error {
	endpoint := fmt.Sprintf("delete_group/%d", groupID)
	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}
	return nil
}

// ==================== Roles API ====================

// GetRoles получает список ролей
func (c *HTTPClient) GetRoles() (data.GetRolesResponse, error) {
	resp, err := c.Get("get_roles", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting roles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var roles data.GetRolesResponse
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, fmt.Errorf("error decoding roles: %w", err)
	}
	return roles, nil
}

// GetRole получает роль по ID
func (c *HTTPClient) GetRole(roleID int64) (*data.Role, error) {
	endpoint := fmt.Sprintf("get_role/%d", roleID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting role %d: %w", roleID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var role data.Role
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return nil, fmt.Errorf("error decoding role: %w", err)
	}
	return &role, nil
}

// ==================== ResultFields API ====================

// GetResultFields получает список полей результата
func (c *HTTPClient) GetResultFields() (data.GetResultFieldsResponse, error) {
	resp, err := c.Get("get_result_fields", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting result fields: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var fields data.GetResultFieldsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fields); err != nil {
		return nil, fmt.Errorf("error decoding result fields: %w", err)
	}
	return fields, nil
}

// ==================== Datasets API ====================

// GetDatasets получает список наборов данных проекта
func (c *HTTPClient) GetDatasets(projectID int64) (data.GetDatasetsResponse, error) {
	endpoint := fmt.Sprintf("get_datasets/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting datasets for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var datasets data.GetDatasetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&datasets); err != nil {
		return nil, fmt.Errorf("error decoding datasets: %w", err)
	}
	return datasets, nil
}

// GetDataset получает набор данных по ID
func (c *HTTPClient) GetDataset(datasetID int64) (*data.Dataset, error) {
	endpoint := fmt.Sprintf("get_dataset/%d", datasetID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting dataset %d: %w", datasetID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var dataset data.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&dataset); err != nil {
		return nil, fmt.Errorf("error decoding dataset: %w", err)
	}
	return &dataset, nil
}

// AddDataset создает новый набор данных
func (c *HTTPClient) AddDataset(projectID int64, name string) (*data.Dataset, error) {
	endpoint := fmt.Sprintf("add_dataset/%d", projectID)
	req := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating dataset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var dataset data.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&dataset); err != nil {
		return nil, fmt.Errorf("error decoding dataset: %w", err)
	}
	return &dataset, nil
}

// UpdateDataset обновляет набор данных
func (c *HTTPClient) UpdateDataset(datasetID int64, name string) (*data.Dataset, error) {
	endpoint := fmt.Sprintf("update_dataset/%d", datasetID)
	req := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating dataset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var dataset data.Dataset
	if err := json.NewDecoder(resp.Body).Decode(&dataset); err != nil {
		return nil, fmt.Errorf("error decoding dataset: %w", err)
	}
	return &dataset, nil
}

// DeleteDataset удаляет набор данных
func (c *HTTPClient) DeleteDataset(datasetID int64) error {
	endpoint := fmt.Sprintf("delete_dataset/%d", datasetID)
	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting dataset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}
	return nil
}

// ==================== Variables API ====================

// GetVariables получает список переменных набора данных
func (c *HTTPClient) GetVariables(datasetID int64) (data.GetVariablesResponse, error) {
	endpoint := fmt.Sprintf("get_variables/%d", datasetID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting variables for dataset %d: %w", datasetID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var variables data.GetVariablesResponse
	if err := json.NewDecoder(resp.Body).Decode(&variables); err != nil {
		return nil, fmt.Errorf("error decoding variables: %w", err)
	}
	return variables, nil
}

// AddVariable добавляет переменную в набор данных
func (c *HTTPClient) AddVariable(datasetID int64, name string) (*data.Variable, error) {
	endpoint := fmt.Sprintf("add_variable/%d", datasetID)
	req := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating variable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var variable data.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("error decoding variable: %w", err)
	}
	return &variable, nil
}

// UpdateVariable обновляет переменную
func (c *HTTPClient) UpdateVariable(variableID int64, name string) (*data.Variable, error) {
	endpoint := fmt.Sprintf("update_variable/%d", variableID)
	req := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating variable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var variable data.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("error decoding variable: %w", err)
	}
	return &variable, nil
}

// DeleteVariable удаляет переменную
func (c *HTTPClient) DeleteVariable(variableID int64) error {
	endpoint := fmt.Sprintf("delete_variable/%d", variableID)
	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting variable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}
	return nil
}

// ==================== BDDs API ====================

// GetBDD получает BDD сценарий для кейса
func (c *HTTPClient) GetBDD(caseID int64) (*data.BDD, error) {
	endpoint := fmt.Sprintf("get_bdd/%d", caseID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting BDD for case %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var bdd data.BDD
	if err := json.NewDecoder(resp.Body).Decode(&bdd); err != nil {
		return nil, fmt.Errorf("error decoding BDD: %w", err)
	}
	return &bdd, nil
}

// AddBDD добавляет BDD сценарий к кейсу
func (c *HTTPClient) AddBDD(caseID int64, content string) (*data.BDD, error) {
	endpoint := fmt.Sprintf("add_bdd/%d", caseID)
	req := map[string]string{"content": content}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating BDD: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var bdd data.BDD
	if err := json.NewDecoder(resp.Body).Decode(&bdd); err != nil {
		return nil, fmt.Errorf("error decoding BDD: %w", err)
	}
	return &bdd, nil
}

// ==================== Labels API ====================

// UpdateTestLabels обновляет labels для теста
func (c *HTTPClient) UpdateTestLabels(testID int64, labels []string) error {
	endpoint := fmt.Sprintf("update_test_labels/%d", testID)
	req := data.UpdateLabelsRequest{Labels: labels}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return fmt.Errorf("error updating test labels: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}
	return nil
}

// UpdateTestsLabels обновляет labels для нескольких тестов
func (c *HTTPClient) UpdateTestsLabels(runID int64, testIDs []int64, labels []string) error {
	endpoint := fmt.Sprintf("update_tests_labels/%d", runID)
	req := data.UpdateTestsLabelsRequest{
		TestIDs: testIDs,
		Labels:  labels,
	}
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return fmt.Errorf("error updating tests labels: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}
	return nil
}

// GetLabels получает список меток проекта
func (c *HTTPClient) GetLabels(projectID int64) (data.GetLabelsResponse, error) {
	endpoint := fmt.Sprintf("get_labels/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting labels for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var labels data.GetLabelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&labels); err != nil {
		return nil, fmt.Errorf("error decoding labels: %w", err)
	}
	return labels, nil
}

// GetLabel получает метку по ID
func (c *HTTPClient) GetLabel(labelID int64) (*data.Label, error) {
	endpoint := fmt.Sprintf("get_label/%d", labelID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting label %d: %w", labelID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var label data.Label
	if err := json.NewDecoder(resp.Body).Decode(&label); err != nil {
		return nil, fmt.Errorf("error decoding label: %w", err)
	}
	return &label, nil
}

// UpdateLabel обновляет метку
func (c *HTTPClient) UpdateLabel(labelID int64, req data.UpdateLabelRequest) (*data.Label, error) {
	endpoint := fmt.Sprintf("update_label/%d", labelID)
	jsonBody, _ := json.Marshal(req)

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating label: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var label data.Label
	if err := json.NewDecoder(resp.Body).Decode(&label); err != nil {
		return nil, fmt.Errorf("error decoding label: %w", err)
	}
	return &label, nil
}
