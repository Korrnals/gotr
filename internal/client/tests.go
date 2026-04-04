// internal/client/tests.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetTest получает информацию о тесте по ID
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests#gettest
func (c *HTTPClient) GetTest(ctx context.Context, testID int64) (*data.Test, error) {
	endpoint := fmt.Sprintf("get_test/%d", testID)

	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request error GetTest for test %d: %w", testID, err)
	}
	defer resp.Body.Close()

	var test data.Test
	if err := json.NewDecoder(resp.Body).Decode(&test); err != nil {
		return nil, fmt.Errorf("decode error test: %w", err)
	}

	return &test, nil
}

// GetTests получает список тестов для тест-run (поддерживает пагинацию)
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests#gettests
// Поддерживает фильтры: status_id, assignedto_id
func (c *HTTPClient) GetTests(ctx context.Context, runID int64, filters map[string]string) ([]data.Test, error) {
	endpoint := fmt.Sprintf("get_tests/%d", runID)
	tests, err := fetchAllPages[data.Test](ctx, c, endpoint, filters, "tests")
	if err != nil {
		return nil, fmt.Errorf("request error GetTests for run %d: %w", runID, err)
	}
	return tests, nil
}

// UpdateTest обновляет тест (статус, назначение)
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests#updatetest
func (c *HTTPClient) UpdateTest(ctx context.Context, testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}

	bodyBytes, _ := json.Marshal(req)
	endpoint := fmt.Sprintf("update_test/%d", testID)

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("request error UpdateTest for test %d: %w", testID, err)
	}
	defer resp.Body.Close()

	var test data.Test
	if err := json.NewDecoder(resp.Body).Decode(&test); err != nil {
		return nil, fmt.Errorf("decode error test: %w", err)
	}

	return &test, nil
}

// Helper методы для удобства

// GetTestsByStatus получает тесты с определенным статусом
func (c *HTTPClient) GetTestsByStatus(ctx context.Context, runID int64, statusID int64) ([]data.Test, error) {
	filters := map[string]string{
		"status_id": strconv.FormatInt(statusID, 10),
	}
	return c.GetTests(ctx, runID, filters)
}

// GetTestsAssignedTo получает тесты, назначенные на пользователя
func (c *HTTPClient) GetTestsAssignedTo(ctx context.Context, runID int64, userID int64) ([]data.Test, error) {
	filters := map[string]string{
		"assignedto_id": strconv.FormatInt(userID, 10),
	}
	return c.GetTests(ctx, runID, filters)
}
