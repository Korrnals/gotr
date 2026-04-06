// internal/client/sections.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/concurrency"
	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
)

// GetSections fetches sections for a suite in a project (with pagination).
// suite_id is required for multi-suite projects.
func (c *HTTPClient) GetSections(ctx context.Context, projectID, suiteID int64) (data.GetSectionsResponse, error) {
	endpoint := fmt.Sprintf("get_sections/%d", projectID)
	var baseQuery map[string]string
	if suiteID != 0 {
		baseQuery = map[string]string{"suite_id": fmt.Sprintf("%d", suiteID)}
	}
	sections, err := fetchAllPages[data.Section](ctx, c, endpoint, baseQuery, "sections")
	if err != nil {
		return nil, fmt.Errorf("request error GetSections for project %d, suite %d: %w", projectID, suiteID, err)
	}
	return data.GetSectionsResponse(sections), nil
}

// GetSectionsParallelCtx gets sections from multiple suites with shared runtime controls.
// If suiteIDs is empty, it falls back to unfiltered sections request (suite_id=0).
func (c *HTTPClient) GetSectionsParallelCtx(
	ctx context.Context,
	projectID int64,
	suiteIDs []int64,
	config *concurrency.ControllerConfig,
) (data.GetSectionsResponse, error) {
	if config == nil {
		config = concurrency.DefaultControllerConfig()
	} else {
		config.Validate()
	}

	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	if len(suiteIDs) == 0 {
		return c.GetSections(ctx, projectID, 0)
	}

	var limiter *concurrent.AdaptiveRateLimiter
	if config.RequestsPerMinute > 0 {
		limiter = concurrent.NewAdaptiveRateLimiter(config.RequestsPerMinute)
	}

	maxRetries := config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	opts := []concurrency.FetchOption{
		concurrency.WithContinueOnError(),
		concurrency.WithMaxConcurrency(config.MaxConcurrentSuites),
	}
	if config.Reporter != nil {
		opts = append(opts, concurrency.WithReporter(config.Reporter))
	}

	sections, err := concurrency.FetchParallelBySuite(ctx, suiteIDs,
		func(suiteID int64) ([]data.Section, error) {
			if limiter != nil {
				if waitErr := limiter.WaitCtx(ctx); waitErr != nil {
					return nil, waitErr
				}
			}

			var lastErr error
			for attempt := 0; attempt <= maxRetries; attempt++ {
				sections, fetchErr := c.GetSections(ctx, projectID, suiteID)
				if fetchErr == nil {
					return sections, nil
				}

				if ctx.Err() != nil {
					return nil, ctx.Err()
				}

				lastErr = fetchErr
				if attempt == maxRetries {
					break
				}

				delay := time.Duration(100*(1<<attempt)) * time.Millisecond
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
			}

			return nil, lastErr
		},
		opts...,
	)
	if err != nil && len(sections) == 0 {
		return nil, err
	}

	return data.GetSectionsResponse(sections), err
}

// GetSection fetches a single section by ID.
func (c *HTTPClient) GetSection(ctx context.Context, sectionID int64) (*data.Section, error) {
	endpoint := fmt.Sprintf("get_section/%d", sectionID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("decode error section %d: %w", sectionID, err)
	}

	return &section, nil
}

// AddSection creates a new section in a project's suite.
func (c *HTTPClient) AddSection(ctx context.Context, projectID int64, req *data.AddSectionRequest) (*data.Section, error) {
	bodyBytes, _ := json.Marshal(req)
	endpoint := fmt.Sprintf("add_section/%d", projectID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error AddSection for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("decode error created section: %w", err)
	}

	return &section, nil
}

// UpdateSection updates a section (name, description, parent_id for moving).
func (c *HTTPClient) UpdateSection(ctx context.Context, sectionID int64, req *data.UpdateSectionRequest) (*data.Section, error) {
	bodyBytes, _ := json.Marshal(req)
	endpoint := fmt.Sprintf("update_section/%d", sectionID)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	var section data.Section
	if err := json.NewDecoder(resp.Body).Decode(&section); err != nil {
		return nil, fmt.Errorf("decode error updated section %d: %w", sectionID, err)
	}

	return &section, nil
}

// DeleteSection deletes a section (irreversible, deletes cases/results).
func (c *HTTPClient) DeleteSection(ctx context.Context, sectionID int64) error {
	endpoint := fmt.Sprintf("delete_section/%d", sectionID)
	resp, err := c.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("request error DeleteSection %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	return nil
}
