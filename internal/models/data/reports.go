// internal/models/data/reports.go
package data

// ReportTemplate — шаблон отчёта
type ReportTemplate struct {
	ID          int64  `json:"id"`           // Уникальный ID шаблона
	Name        string `json:"name"`         // Название шаблона
	Description string `json:"description"`  // Описание
}

// Report — сгенерированный отчёт
type Report struct {
	ID        int64  `json:"id"`         // ID отчёта
	Name      string `json:"name"`       // Название
	URL       string `json:"url"`        // URL для скачивания
	Status    string `json:"status"`     // Статус (completed, pending, error)
	CreatedOn int64  `json:"created_on"` // Timestamp создания
}

// GetReportsResponse — ответ на get_reports
type GetReportsResponse []ReportTemplate

// RunReportResponse — ответ на run_report
type RunReportResponse struct {
	ReportID  int64  `json:"report_id"`
	URL       string `json:"url"`
	Status    string `json:"status"`
}
