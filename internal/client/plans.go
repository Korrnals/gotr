// internal/client/plans.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetPlan получает тест-план по ID
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#getplan
func (c *HTTPClient) GetPlan(planID int64) (*data.Plan, error) {
	endpoint := fmt.Sprintf("get_plan/%d", planID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting plan %d: %w", planID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for plan %d: %s", resp.Status, planID, string(body))
	}

	var plan data.Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, fmt.Errorf("error decoding plan: %w", err)
	}
	return &plan, nil
}

// GetPlans получает список тест-планов проекта
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#getplans
func (c *HTTPClient) GetPlans(projectID int64) (data.GetPlansResponse, error) {
	endpoint := fmt.Sprintf("get_plans/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting plans for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for project %d: %s", resp.Status, projectID, string(body))
	}

	var plans data.GetPlansResponse
	if err := json.NewDecoder(resp.Body).Decode(&plans); err != nil {
		return nil, fmt.Errorf("error decoding plans: %w", err)
	}
	return plans, nil
}

// AddPlan создает новый тест-план
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#addplan
func (c *HTTPClient) AddPlan(projectID int64, req *data.AddPlanRequest) (*data.Plan, error) {
	endpoint := fmt.Sprintf("add_plan/%d", projectID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating plan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var plan data.Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, fmt.Errorf("error decoding created plan: %w", err)
	}
	return &plan, nil
}

// UpdatePlan обновляет существующий тест-план
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#updateplan
func (c *HTTPClient) UpdatePlan(planID int64, req *data.UpdatePlanRequest) (*data.Plan, error) {
	endpoint := fmt.Sprintf("update_plan/%d", planID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating plan %d: %w", planID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for plan %d: %s", resp.Status, planID, string(body))
	}

	var plan data.Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, fmt.Errorf("error decoding updated plan: %w", err)
	}
	return &plan, nil
}

// ClosePlan закрывает тест-план
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#closeplan
func (c *HTTPClient) ClosePlan(planID int64) (*data.Plan, error) {
	endpoint := fmt.Sprintf("close_plan/%d", planID)

	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return nil, fmt.Errorf("error closing plan %d: %w", planID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for plan %d: %s", resp.Status, planID, string(body))
	}

	var plan data.Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, fmt.Errorf("error decoding closed plan: %w", err)
	}
	return &plan, nil
}

// DeletePlan удаляет тест-план
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#deleteplan
func (c *HTTPClient) DeletePlan(planID int64) error {
	endpoint := fmt.Sprintf("delete_plan/%d", planID)

	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting plan %d: %w", planID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s for plan %d: %s", resp.Status, planID, string(body))
	}
	return nil
}

// AddPlanEntry добавляет entry в существующий тест-план
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#addplanentry
func (c *HTTPClient) AddPlanEntry(planID int64, req *data.AddPlanEntryRequest) (*data.Plan, error) {
	endpoint := fmt.Sprintf("add_plan_entry/%d", planID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error adding plan entry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var plan data.Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, fmt.Errorf("error decoding plan after adding entry: %w", err)
	}
	return &plan, nil
}

// UpdatePlanEntry обновляет entry в тест-плане
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#updateplanentry
func (c *HTTPClient) UpdatePlanEntry(planID int64, entryID string, req *data.UpdatePlanEntryRequest) (*data.Plan, error) {
	endpoint := fmt.Sprintf("update_plan_entry/%d/%s", planID, entryID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating plan entry %s: %w", entryID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for entry %s: %s", resp.Status, entryID, string(body))
	}

	var plan data.Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, fmt.Errorf("error decoding plan after updating entry: %w", err)
	}
	return &plan, nil
}

// DeletePlanEntry удаляет entry из тест-плана
// https://support.testrail.com/hc/en-us/articles/7077996481044-Plans#deleteplanentry
func (c *HTTPClient) DeletePlanEntry(planID int64, entryID string) error {
	endpoint := fmt.Sprintf("delete_plan_entry/%d/%s", planID, entryID)

	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting plan entry %s: %w", entryID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s for entry %s: %s", resp.Status, entryID, string(body))
	}
	return nil
}
