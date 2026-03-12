package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ResponseData — для универсальных запросов (когда структура неизвестна).
type ResponseData struct {
	Status     string              `json:"status"`
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       interface{}         `json:"body"`               // распарсенный JSON
	RawBody    []byte              `json:"raw_body,omitempty"` // сырые байты (опционально в JSON)
	Timestamp  time.Time           `json:"timestamp"`
	Duration   time.Duration       `json:"duration"`
}

// PrintResponse — красивый вывод универсального (с типом interface{}) response
// Чтение ЛЮБОГО response
func (c *HTTPClient) ReadResponse(ctx context.Context, resp *http.Response, duration time.Duration, outputFormat string) (ResponseData, error) {
	// Считываем поток response (сырые данные) в переменную
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseData{}, err
	}
	// Преобразуем сырые данные в формат JSON, для дальнейшей записи в структуру 'ResponseData'
	var bodyData interface{}
	if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
		bodyData = string(bodyBytes)
	}
	// Записываем (маппим) полученные данные в структуру
	data := ResponseData{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       bodyData,
		RawBody:    bodyBytes,
		Timestamp:  time.Now(),
		Duration:   duration,
	}
	// Возвращаем целиком заполненную  данными структуру 'ResponseData'
	return data, nil
}

// ReadJSONResponse — универсальный метод для чтения response в любую структуру (не в interface{})
func (c *HTTPClient) ReadJSONResponse(ctx context.Context, resp *http.Response, target any) error {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s, body: %s", resp.Status, string(body))
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

// PrintResponseFromData — вывод из уже готовой структуры (без чтения resp.Body), с нетипизированным телом response
func (c *HTTPClient) PrintResponseFromData(ctx context.Context, data ResponseData, outputFormat string) {
	switch outputFormat {
	case "json":
		pretty, _ := json.MarshalIndent(data.Body, "", "  ")
		fmt.Println(string(pretty))
	case "json-full":
		pretty, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(pretty))
	default: // table
		printTable(data)
	}
}

// SaveResponseToFile — сохранение не типизированного response
func (c *HTTPClient) SaveResponseToFile(ctx context.Context, data ResponseData, filename string, outputFormat string) error {
	var toSave []byte
	switch outputFormat {
	case "json":
		toSave, _ = json.MarshalIndent(data.Body, "", "  ")
	case "json-full":
		toSave, _ = json.MarshalIndent(data, "", "  ")
	default: // table
		// Для table сохраняем как json-full (или можно сделать отдельный формат)
		toSave, _ = json.MarshalIndent(data, "", "  ")
	}

	if err := os.WriteFile(filename, toSave, 0644); err != nil {
		return err
	}
	fmt.Printf("Response saved to %s (format: %s)\n", filename, outputFormat)

	return nil
}

// Вспомогательные приватные функции //
// 'printTable' - формирует таблицу response
func printTable(data ResponseData) {
	fmt.Printf("Status: %s (%d)\n", data.Status, data.StatusCode)
	fmt.Printf("Duration: %v\n", data.Duration)
	fmt.Printf("Timestamp: %s\n", data.Timestamp.Format(time.RFC3339))
	fmt.Printf("\nHeaders:\n")
	for k, v := range data.Headers {
		for _, val := range v {
			fmt.Printf("  %s: %s\n", k, val)
		}
	}
	fmt.Printf("\nBody:\n")
	jsonBody, _ := json.MarshalIndent(data.Body, "", "  ")
	fmt.Println(string(jsonBody))
}
