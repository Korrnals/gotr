package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ResponseData — структура для красивого вывода
type ResponseData struct {
	Status     string              `json:"status"`
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       interface{}         `json:"body"`       		   // распарсенный JSON
    RawBody    []byte              `json:"raw_body,omitempty"` // сырые байты (опционально в JSON)
	Timestamp  time.Time           `json:"timestamp"`
	Duration   time.Duration       `json:"duration"`
}

// PrintResponse — красивый вывод ответа
func (c *HTTPClient) ReadResponse(resp *http.Response, duration time.Duration, outputFormat string) (ResponseData, error) {
	// Считываем поток ответа (сырые данные) в переменную
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

// PrintResponseFromData — вывод из уже готовой структуры (без чтения resp.Body)
func (c *HTTPClient) PrintResponseFromData(data ResponseData, outputFormat string) {
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

// SaveResponseToFile — сохранение ответа
func (c *HTTPClient) SaveResponseToFile(data ResponseData, filename string, outputFormat string) error {
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
    fmt.Printf("Ответ сохранён в файл %s (формат: %s)\n", filename, outputFormat)

	return nil
}