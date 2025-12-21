package data

// Структура для 'get_shared_step'
type GetSharedStepResponse struct {
	ID                   int    `json:"id"`
	Title                string `json:"title"`
	ProjectID            int    `json:"project_id"`
	CreatedBy            int    `json:"created_by"`
	CreatedOn            int    `json:"created_on"`
	UpdatedBy            int    `json:"updated_by"`
	UpdatedOn            int    `json:"updated_on"`
	CustomStepsSeparated []struct {
		Content        string `json:"content"`
		AdditionalInfo any    `json:"additional_info"`
		Expected       string `json:"expected"`
		Refs           string `json:"refs"`
	} `json:"custom_steps_separated"`
	CaseIds []int `json:"case_ids"`
}

// Структура для 'get_shared_step_history'
type GetSharedStepHistoryResponse struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Size   int `json:"size"`
	Links  struct {
		Next any `json:"next"`
		Prev any `json:"prev"`
	} `json:"_links"`
	StepHistory []struct {
		ID                   string `json:"id"`
		Timestamp            int    `json:"timestamp"`
		UserID               string `json:"user_id"`
		CustomStepsSeparated []struct {
			Content        string `json:"content"`
			AdditionalInfo string `json:"additional_info"`
			Expected       string `json:"expected"`
			Refs           string `json:"refs"`
		} `json:"custom_steps_separated"`
		Title string `json:"title"`
	} `json:"step_history"`
}

// Структура для 'get_shared_steps'
type GetSharedStepsResponse struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Size   int `json:"size"`
	Links  struct {
		Next string `json:"next"`
		Prev any    `json:"prev"`
	} `json:"_links"`
	SharedSteps []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"shared_steps"`
}

// Структура для 'add_shared_step'
type AddSharedStepRequest struct {
	Title                string `json:"title"`
	CustomStepsSeparated []struct {
		Content        string `json:"content"`
		AdditionalInfo string `json:"additional_info,omitempty"`
		Expected       string `json:"expected,omitempty"`
		Refs           string `json:"refs,omitempty"`
	} `json:"custom_steps_separated"`
}

// Структура для 'update_shared_step'
type UpdateSharedStepResponse struct {
	ID                   int    `json:"id"`
	Title                string `json:"title"`
	ProjectID            int    `json:"project_id"`
	CreatedBy            int    `json:"created_by"`
	CreatedOn            int    `json:"created_on"`
	UpdatedBy            int    `json:"updated_by"`
	UpdatedOn            int    `json:"updated_on"`
	CustomStepsSeparated []struct {
		Content        string `json:"content"`
		AdditionalInfo any    `json:"additional_info"`
		Expected       string `json:"expected"`
		Refs           string `json:"refs"`
	} `json:"custom_steps_separated"`
	CaseIds []int `json:"case_ids"`
}

// Структура для 'delete_shared_step'
type DeleteSharedStepRequest struct {
	KeepInCases int `json:"keep_in_cases"`
}