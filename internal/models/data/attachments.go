// internal/models/data/attachments.go
package data

// Attachment представляет файл вложения в TestRail
type Attachment struct {
	ID          int64  `json:"id"`           // Уникальный ID вложения
	Name        string `json:"name"`         // Имя файла
	Filename    string `json:"filename"`     // Имя файла (альтернативное поле)
	Size        int64  `json:"size"`         // Размер файла в байтах
	ContentType string `json:"content_type"` // MIME-тип содержимого
	CreatedOn   int64  `json:"created_on"`   // Timestamp создания
	ProjectID   int64  `json:"project_id"`   // ID проекта
	CaseID      int64  `json:"case_id"`      // ID кейса (если привязано к кейсу)
	RunID       int64  `json:"run_id"`       // ID рана (если привязано к рану)
	PlanID      int64  `json:"plan_id"`      // ID плана (если привязано к плану)
	ResultID    int64  `json:"result_id"`    // ID результата (если привязано к результату)
	UserID      int64  `json:"user_id"`      // ID пользователя, загрузившего файл
}

// AttachmentResponse ответ от API при добавлении вложения
type AttachmentResponse struct {
	AttachmentID int64  `json:"attachment_id"` // ID созданного вложения
	URL          string `json:"url"`           // URL для доступа к вложению
	Name         string `json:"name"`          // Имя файла
	Size         int64  `json:"size"`          // Размер файла
}

// GetAttachmentsResponse ответ на получение списка вложений
type GetAttachmentsResponse []Attachment
