// internal/models/data/extended.go
// Extended APIs: Groups, Roles, ResultFields, Variables, Datasets, BDDs, Labels

package data

// Group — группа пользователей
type Group struct {
	ID        int64  `json:"id"`         // Уникальный ID группы
	Name      string `json:"name"`       // Название группы
	UserIDs   []int64 `json:"user_ids"`  // ID пользователей в группе
	ProjectID int64  `json:"project_id"` // ID проекта
}

// GetGroupsResponse — ответ на get_groups
type GetGroupsResponse []Group

// Role — роль пользователя
type Role struct {
	ID   int64  `json:"id"`   // Уникальный ID роли
	Name string `json:"name"` // Название роли
}

// GetRolesResponse — ответ на get_roles
type GetRolesResponse []Role

// ResultField — поле результата
type ResultField struct {
	ID       int64  `json:"id"`        // Уникальный ID поля
	Name     string `json:"name"`      // Название поля
	SystemName string `json:"system_name"` // Системное название
	TypeID   int    `json:"type_id"`   // Тип поля
	IsActive bool   `json:"is_active"` // Активно ли поле
}

// GetResultFieldsResponse — ответ на get_result_fields
type GetResultFieldsResponse []ResultField

// Dataset — набор данных для параметризованных тестов
type Dataset struct {
	ID        int64  `json:"id"`         // Уникальный ID набора
	Name      string `json:"name"`       // Название набора
	ProjectID int64  `json:"project_id"` // ID проекта
}

// GetDatasetsResponse — ответ на get_datasets
type GetDatasetsResponse []Dataset

// Variable — переменная в наборе данных
type Variable struct {
	ID        int64  `json:"id"`         // Уникальный ID переменной
	Name      string `json:"name"`       // Название переменной
	DatasetID int64  `json:"dataset_id"` // ID набора данных
}

// GetVariablesResponse — ответ на get_variables
type GetVariablesResponse []Variable

// BDD — BDD сценарий для кейса
type BDD struct {
	ID      int64  `json:"id"`       // Уникальный ID
	CaseID  int64  `json:"case_id"`  // ID кейса
	Content string `json:"content"`  // Содержимое BDD сценария
}

// UpdateLabelsRequest — запрос на обновление labels теста
type UpdateLabelsRequest struct {
	Labels []string `json:"labels"` // Список labels
}

// UpdateTestsLabelsRequest — запрос на обновление labels для нескольких тестов
type UpdateTestsLabelsRequest struct {
	TestIDs []int64  `json:"test_ids"` // ID тестов
	Labels  []string `json:"labels"`   // Список labels
}

// UpdateLabelRequest — запрос на обновление метки
type UpdateLabelRequest struct {
	ProjectID int64  `json:"project_id"` // ID проекта
	Title     string `json:"title"`      // Название метки (max 20 символов)
}

// GetLabelsResponse — ответ на get_labels
type GetLabelsResponse []Label
