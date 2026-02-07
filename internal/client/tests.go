// internal/client/tests.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetTest получает информацию о тесте по ID
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests#gettest
func (c *HTTPClient) GetTest(testID int64) (*data.Test, error) {
	endpoint := fmt.Sprintf("get_test/%d", testID)
	
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetTest для теста %d: %w", testID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении теста %d: %s",
			resp.Status, testID, string(body))
	}

	var test data.Test
	if err := json.NewDecoder(resp.Body).Decode(&test); err != nil {
		return nil, fmt.Errorf("ошибка декодирования теста: %w", err)
	}

	return &test, nil
}

// GetTests получает список тестов для тест-рана
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests#gettests
// Поддерживает фильтры: status_id, assignedto_id
func (c *HTTPClient) GetTests(runID int64, filters map[string]string) ([]data.Test, error) {
	endpoint := fmt.Sprintf("get_tests/%d", runID)
	
	// Преобразуем фильтры в query параметры
	var queryParams map[string]string
	if len(filters) > 0 {
		queryParams = make(map[string]string)
		for key, value := range filters {
			queryParams[key] = value
		}
	}
	
	resp, err := c.Get(endpoint, queryParams)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetTests для рана %d: %w", runID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении тестов для рана %d: %s",
			resp.Status, runID, string(body))
	}

	var tests []data.Test
	if err := json.NewDecoder(resp.Body).Decode(&tests); err != nil {
		return nil, fmt.Errorf("ошибка декодирования списка тестов: %w", err)
	}

	return tests, nil
}

// UpdateTest обновляет тест (статус, назначение)
// https://support.testrail.com/hc/en-us/articles/7077723946004-Tests#updatetest
func (c *HTTPClient) UpdateTest(testID int64, req *data.UpdateTestRequest) (*data.Test, error) {
	if req == nil {
		return nil, fmt.Errorf("тело запроса обязательно")
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateTestRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_test/%d", testID)
	
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateTest для теста %d: %w", testID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при обновлении теста %d: %s",
			resp.Status, testID, string(body))
	}

	var test data.Test
	if err := json.NewDecoder(resp.Body).Decode(&test); err != nil {
		return nil, fmt.Errorf("ошибка декодирования теста: %w", err)
	}

	return &test, nil
}

// Helper методы для удобства

// GetTestsByStatus получает тесты с определенным статусом
func (c *HTTPClient) GetTestsByStatus(runID int64, statusID int64) ([]data.Test, error) {
	filters := map[string]string{
		"status_id": strconv.FormatInt(statusID, 10),
	}
	return c.GetTests(runID, filters)
}

// GetTestsAssignedTo получает тесты, назначенные на пользователя
func (c *HTTPClient) GetTestsAssignedTo(runID int64, userID int64) ([]data.Test, error) {
	filters := map[string]string{
		"assignedto_id": strconv.FormatInt(userID, 10),
	}
	return c.GetTests(runID, filters)
}
