// internal/models/data/configs.go
package data

// ConfigGroup — группа конфигураций (например, "OS", "Browser")
type ConfigGroup struct {
	ID        int64    `json:"id"`         // Уникальный ID группы
	Name      string   `json:"name"`       // Название группы
	ProjectID int64    `json:"project_id"` // ID проекта
	Configs   []Config `json:"configs"`    // Список конфигураций в группе
}

// Config — конкретная конфигурация (например, "Windows", "Chrome")
type Config struct {
	ID            int64  `json:"id"`             // Уникальный ID конфигурации
	Name          string `json:"name"`           // Название конфигурации
	GroupID       int64  `json:"group_id"`       // ID группы
	ProjectID     int64  `json:"project_id"`     // ID проекта (опционально)
}

// GetConfigsResponse — ответ на get_configs (список групп с конфигурациями)
type GetConfigsResponse []ConfigGroup

// AddConfigGroupRequest — запрос на создание группы конфигураций
type AddConfigGroupRequest struct {
	Name string `json:"name"` // Название группы (обязательное)
}

// AddConfigRequest — запрос на создание конфигурации
type AddConfigRequest struct {
	Name string `json:"name"` // Название конфигурации (обязательное)
}

// UpdateConfigGroupRequest — запрос на обновление группы
type UpdateConfigGroupRequest struct {
	Name string `json:"name,omitempty"` // Новое название группы
}

// UpdateConfigRequest — запрос на обновление конфигурации
type UpdateConfigRequest struct {
	Name string `json:"name,omitempty"` // Новое название конфигурации
}
