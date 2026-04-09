package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetSuites fetches all test suites for a project.
func (c *HTTPClient) GetSuites(ctx context.Context, projectID int64) (data.GetSuitesResponse, error) {
	endpoint := fmt.Sprintf("get_suites/%d", projectID)
	suites, err := fetchAllPages[data.Suite](ctx, c, endpoint, nil, "suites")
	if err != nil {
		return nil, fmt.Errorf("request error GetSuites for project %d: %w", projectID, err)
	}
	return data.GetSuitesResponse(suites), nil
}

// GetSuite fetches a single test suite by ID.
func (c *HTTPClient) GetSuite(ctx context.Context, suiteID int64) (*data.Suite, error) {
	endpoint := fmt.Sprintf("get_suite/%d", suiteID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetSuite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	var suite data.Suite
	if err := json.NewDecoder(resp.Body).Decode(&suite); err != nil {
		return nil, fmt.Errorf("decode error suite %d: %w", suiteID, err)
	}

	return &suite, nil
}

// AddSuite creates a new test suite in a project.
func (c *HTTPClient) AddSuite(ctx context.Context, projectID int64, req *data.AddSuiteRequest) (*data.Suite, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	endpoint := fmt.Sprintf("add_suite/%d", projectID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddSuite for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var suite data.Suite
	if err := json.NewDecoder(resp.Body).Decode(&suite); err != nil {
		return nil, fmt.Errorf("decode error created suite: %w", err)
	}

	return &suite, nil
}

// UpdateSuite updates an existing test suite.
func (c *HTTPClient) UpdateSuite(ctx context.Context, suiteID int64, req *data.UpdateSuiteRequest) (*data.Suite, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	endpoint := fmt.Sprintf("update_suite/%d", suiteID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateSuite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	var suite data.Suite
	if err := json.NewDecoder(resp.Body).Decode(&suite); err != nil {
		return nil, fmt.Errorf("decode error updated suite %d: %w", suiteID, err)
	}

	return &suite, nil
}

// DeleteSuite deletes a test suite by ID.
// This is irreversible — all cases in the suite will be deleted.
func (c *HTTPClient) DeleteSuite(ctx context.Context, suiteID int64) error {
	endpoint := fmt.Sprintf("delete_suite/%d", suiteID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteSuite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	return nil
}
