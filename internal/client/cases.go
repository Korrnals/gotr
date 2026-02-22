package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetCases получает **все** кейсы проекта (с пагинацией).
// suiteID и sectionID — опционально (0 = не использовать).
// Возвращает полный список кейсов.
func (c *HTTPClient) GetCases(projectID int64, suiteID int64, sectionID int64) (data.GetCasesResponse, error) {
	return c.GetCasesWithProgress(projectID, suiteID, sectionID, nil)
}

// GetCasesWithProgress получает **все** кейсы проекта с отслеживанием прогресса.
// monitor вызывается после каждой страницы (каждые 250 кейсов).
func (c *HTTPClient) GetCasesWithProgress(projectID int64, suiteID int64, sectionID int64, monitor ProgressMonitor) (data.GetCasesResponse, error) {
	var all data.GetCasesResponse
	offset := int64(0)
	limit := int64(250)

	for {
		endpoint := fmt.Sprintf("get_cases/%d", projectID)
		query := map[string]string{
			"offset": fmt.Sprintf("%d", offset),
			"limit":  fmt.Sprintf("%d", limit),
		}
		if suiteID != 0 {
			query["suite_id"] = fmt.Sprintf("%d", suiteID)
		}
		if sectionID != 0 {
			query["section_id"] = fmt.Sprintf("%d", sectionID)
		}

		resp, err := c.Get(endpoint, query)
		if err != nil {
			return nil, fmt.Errorf("ошибка запроса GetCases для проекта %d: %w", projectID, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API вернул %s при получении кейсов проекта %d: %s", resp.Status, projectID, string(body))
		}

		var page data.GetCasesResponse
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			return nil, fmt.Errorf("ошибка декодирования страницы кейсов (offset=%d): %w", offset, err)
		}

		all = append(all, page...)

		// Update progress after each page
		if monitor != nil {
			monitor.Increment()
		}

		// Break if we got fewer items than limit, or if page is empty (safety check)
		if len(page) == 0 || len(page) < int(limit) {
			break
		}

		offset += limit
	}

	return all, nil
}

// GetCase получает информацию о конкретном кейсе по ID.
// Возвращает указатель на Case.
func (c *HTTPClient) GetCase(caseID int64) (*data.Case, error) {
	endpoint := fmt.Sprintf("get_case/%d", caseID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении кейса %d: %s",
			resp.Status, caseID, string(body))
	}

	var kase data.Case
	if err := json.NewDecoder(resp.Body).Decode(&kase); err != nil {
		return nil, fmt.Errorf("ошибка декодирования кейса %d: %w", caseID, err)
	}

	return &kase, nil
}

// GetHistoryForCase получает историю изменений кейса.
func (c *HTTPClient) GetHistoryForCase(caseID int64) (*data.GetHistoryForCaseResponse, error) {
	endpoint := fmt.Sprintf("get_history_for_case/%d", caseID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetHistoryForCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении истории кейса %d: %s",
			resp.Status, caseID, string(body))
	}

	var result data.GetHistoryForCaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования истории кейса %d: %w", caseID, err)
	}

	return &result, nil
}

// AddCase создаёт новый кейс в секции.
// Требует sectionID и Title в запросе.
func (c *HTTPClient) AddCase(sectionID int64, req *data.AddCaseRequest) (*data.Case, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddCaseRequest: %w", err)
	}

	endpoint := fmt.Sprintf("add_case/%d", sectionID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddCase в секции %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при создании кейса в секции %d: %s",
			resp.Status, sectionID, string(body))
	}

	var result data.Case
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования созданного кейса: %w", err)
	}

	return &result, nil
}

// UpdateCase обновляет существующий кейс.
// Поддерживает частичные обновления.
func (c *HTTPClient) UpdateCase(caseID int64, req *data.UpdateCaseRequest) (*data.Case, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateCaseRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_case/%d", caseID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при обновлении кейса %d: %s",
			resp.Status, caseID, string(body))
	}

	var result data.Case
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования обновлённого кейса %d: %w", caseID, err)
	}

	return &result, nil
}

// UpdateCases — bulk-обновление кейсов в suite.
func (c *HTTPClient) UpdateCases(suiteID int64, req *data.UpdateCasesRequest) (*data.GetCasesResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга UpdateCasesRequest: %w", err)
	}

	endpoint := fmt.Sprintf("update_cases/%d", suiteID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса UpdateCases в suite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при bulk-обновлении в suite %d: %s",
			resp.Status, suiteID, string(body))
	}

	var result data.GetCasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа bulk-обновления: %w", err)
	}

	return &result, nil
}

// DeleteCase удаляет кейс по ID.
// Удаление необратимо.
func (c *HTTPClient) DeleteCase(caseID int64) error {
	endpoint := fmt.Sprintf("delete_case/%d", caseID)
	resp, err := c.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteCase %d: %w", caseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка удаления кейса %d: %s, тело: %s", caseID, resp.Status, string(body))
	}

	return nil
}

// DeleteCases — bulk-удаление кейсов в suite.
func (c *HTTPClient) DeleteCases(suiteID int64, req *data.DeleteCasesRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга DeleteCasesRequest: %w", err)
	}

	endpoint := fmt.Sprintf("delete_cases/%d", suiteID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса DeleteCases в suite %d: %w", suiteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка bulk-удаления кейсов в suite %d: %s, тело: %s", suiteID, resp.Status, string(body))
	}

	return nil
}

// GetCaseTypes получает список всех типов кейсов.
func (c *HTTPClient) GetCaseTypes() (data.GetCaseTypesResponse, error) {
	endpoint := "get_case_types"
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetCaseTypes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении типов кейсов: %s", resp.Status, string(body))
	}

	var result data.GetCaseTypesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования типов кейсов: %w", err)
	}

	return result, nil
}

// GetCaseFields получает список всех полей кейсов.
func (c *HTTPClient) GetCaseFields() (data.GetCaseFieldsResponse, error) {
	endpoint := "get_case_fields"
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса GetCaseFields: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при получении полей кейсов: %s", resp.Status, string(body))
	}

	var result data.GetCaseFieldsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования полей кейсов: %w", err)
	}

	return result, nil
}

// AddCaseField создаёт новое поле кейса.
func (c *HTTPClient) AddCaseField(req *data.AddCaseFieldRequest) (*data.AddCaseFieldResponse, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга AddCaseFieldRequest: %w", err)
	}

	endpoint := "add_case_field"
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса AddCaseField: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул %s при создании поля кейса: %s", resp.Status, string(body))
	}

	var result data.AddCaseFieldResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования созданного поля кейса: %w", err)
	}

	return &result, nil
}

// DiffCasesData — сравнивает кейсы двух проектов по указанному полю.
// Возвращает DiffCasesResponse с разницей.
// Использует параллельную загрузку для ускорения.
func (c *HTTPClient) DiffCasesData(pid1, pid2 int64, field string) (*data.DiffCasesResponse, error) {
	// Параллельная загрузка кейсов из обоих проектов
	type result struct {
		cases data.GetCasesResponse
		err   error
		pid   int64
	}
	
	resultChan := make(chan result, 2)
	
	go func() {
		cases, err := c.GetCases(pid1, 0, 0)
		resultChan <- result{cases, err, pid1}
	}()
	
	go func() {
		cases, err := c.GetCases(pid2, 0, 0)
		resultChan <- result{cases, err, pid2}
	}()
	
	var cases1, cases2 data.GetCasesResponse
	for i := 0; i < 2; i++ {
		res := <-resultChan
		if res.err != nil {
			return nil, fmt.Errorf("ошибка получения кейсов из проекта %d: %w", res.pid, res.err)
		}
		if res.pid == pid1 {
			cases1 = res.cases
		} else {
			cases2 = res.cases
		}
	}

	firstCases := make(map[int64]data.Case)
	for _, c := range cases1 {
		firstCases[c.ID] = c
	}

	secondCases := make(map[int64]data.Case)
	for _, c := range cases2 {
		secondCases[c.ID] = c
	}

	diffResult := &data.DiffCasesResponse{}

	// Только в первом
	for id, c := range firstCases {
		if _, ok := secondCases[id]; !ok {
			diffResult.OnlyInFirst = append(diffResult.OnlyInFirst, c)
		}
	}

	// Только во втором
	for id, c := range secondCases {
		if _, ok := firstCases[id]; !ok {
			diffResult.OnlyInSecond = append(diffResult.OnlyInSecond, c)
		}
	}

	// Отличаются по полю
	for id, c1 := range firstCases {
		if c2, ok := secondCases[id]; ok {
			if !casesEqualByField(c1, c2, field) {
				diffResult.DiffByField = append(diffResult.DiffByField, struct {
					CaseID int64     `json:"case_id"`
					First  data.Case `json:"first"`
					Second data.Case `json:"second"`
				}{id, c1, c2})
			}
		}
	}

	return diffResult, nil
}

// casesEqualByField — сравнивает два кейса по указанному полю
func casesEqualByField(c1, c2 data.Case, field string) bool {
	switch field {
	case "title":
		return c1.Title == c2.Title
	case "priority_id":
		return c1.PriorityID == c2.PriorityID
	case "custom_preconds":
		return c1.CustomPreconds == c2.CustomPreconds
	case "id":
		return c1.ID == c2.ID
	case "suite_id":
		return c1.SuiteID == c2.SuiteID
	case "created_by":
		return c1.CreatedBy == c2.CreatedBy
	case "section_id":
		return c1.SectionID == c2.SectionID
	default:
		return false
	}
}

// CopyCasesToSection копирует кейсы в указанную секцию
// POST index.php?/api/v2/copy_cases_to_section/:section_id
func (c *HTTPClient) CopyCasesToSection(sectionID int64, req *data.CopyCasesRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга CopyCasesRequest: %w", err)
	}

	endpoint := fmt.Sprintf("copy_cases_to_section/%d", sectionID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса CopyCasesToSection для секции %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка копирования кейсов в секцию %d: %s, тело: %s", sectionID, resp.Status, string(body))
	}

	return nil
}

// MoveCasesToSection перемещает кейсы в указанную секцию
// POST index.php?/api/v2/move_cases_to_section/:section_id
func (c *HTTPClient) MoveCasesToSection(sectionID int64, req *data.MoveCasesRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга MoveCasesRequest: %w", err)
	}

	endpoint := fmt.Sprintf("move_cases_to_section/%d", sectionID)
	resp, err := c.Post(endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return fmt.Errorf("ошибка запроса MoveCasesToSection для секции %d: %w", sectionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка перемещения кейсов в секцию %d: %s, тело: %s", sectionID, resp.Status, string(body))
	}

	return nil
}
