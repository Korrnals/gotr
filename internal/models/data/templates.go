// internal/models/data/templates.go
package data

// Template — шаблон тест-кейса
type Template struct {
	ID          int64  `json:"id"`          // Уникальный ID шаблона
	Name        string `json:"name"`        // Название шаблона
	IsDefault   bool   `json:"is_default"`  // Шаблон по умолчанию
}

// GetTemplatesResponse — ответ на get_templates
type GetTemplatesResponse []Template
