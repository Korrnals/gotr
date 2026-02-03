// models/data/shared.go
package data

// Pagination — общая структура для пагинированных ответов
type Pagination struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
	Size   int64 `json:"size"`
	Links  struct {
		Next string `json:"next,omitempty"`
		Prev string `json:"prev,omitempty"`
	} `json:"_links,omitempty"`
}

// Step — шаг кейса (custom_steps_separated)
type Step struct {
	Content        string `json:"content,omitempty"`
	AdditionalInfo string `json:"additional_info,omitempty"`
	Expected       string `json:"expected,omitempty"`
	Refs           string `json:"refs,omitempty"`
	SharedStepID   int64  `json:"shared_step_id,omitempty"`
}

// Label — метка кейса (используется в Case.Labels)
type Label struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
