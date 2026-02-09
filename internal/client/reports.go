// internal/client/reports.go
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetReports получает список шаблонов отчётов для проекта
// https://support.testrail.com/hc/en-us/articles/7077721635988-Reports#getreports
func (c *HTTPClient) GetReports(projectID int64) (data.GetReportsResponse, error) {
	endpoint := fmt.Sprintf("get_reports/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting reports for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for project %d: %s", resp.Status, projectID, string(body))
	}

	var reports data.GetReportsResponse
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return nil, fmt.Errorf("error decoding reports: %w", err)
	}
	return reports, nil
}

// RunReport запускает генерацию отчёта по шаблону
// https://support.testrail.com/hc/en-us/articles/7077721635988-Reports#runreport
func (c *HTTPClient) RunReport(templateID int64) (*data.RunReportResponse, error) {
	endpoint := fmt.Sprintf("run_report/%d", templateID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error running report %d: %w", templateID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for template %d: %s", resp.Status, templateID, string(body))
	}

	var reportResp data.RunReportResponse
	if err := json.NewDecoder(resp.Body).Decode(&reportResp); err != nil {
		return nil, fmt.Errorf("error decoding report response: %w", err)
	}
	return &reportResp, nil
}

// RunCrossProjectReport запускает кросс-проектный отчёт
// https://support.testrail.com/hc/en-us/articles/7077721635988-Reports#runcrossprojectreport
func (c *HTTPClient) RunCrossProjectReport(templateID int64) (*data.RunReportResponse, error) {
	endpoint := fmt.Sprintf("run_cross_project_report/%d", templateID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error running cross-project report %d: %w", templateID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for template %d: %s", resp.Status, templateID, string(body))
	}

	var reportResp data.RunReportResponse
	if err := json.NewDecoder(resp.Body).Decode(&reportResp); err != nil {
		return nil, fmt.Errorf("error decoding report response: %w", err)
	}
	return &reportResp, nil
}
