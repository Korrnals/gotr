// internal/models/data/users.go
package data

// User — пользователь TestRail
type User struct {
	ID        int64  `json:"id"`         // Уникальный ID пользователя
	Name      string `json:"name"`       // Имя пользователя
	Email     string `json:"email"`      // Email пользователя
	IsActive  bool   `json:"is_active"`  // Активен ли пользователь
	RoleID    int64  `json:"role_id"`    // ID роли пользователя
	Role      string `json:"role"`       // Название роли
	IsAdmin   bool   `json:"is_admin"`   // Является ли администратором
	MfaAuth   int    `json:"mfa_auth"`   // MFA статус
	MfaSecret string `json:"mfa_secret,omitempty"` // MFA секрет (обычно не возвращается)
}

// GetUsersResponse — ответ на get_users
type GetUsersResponse []User

// GetUserResponse — ответ на get_user
type GetUserResponse User

// AddUserRequest — запрос для add_user
type AddUserRequest struct {
	Name     string `json:"name"`              // Имя пользователя (обязательно)
	Email    string `json:"email"`             // Email пользователя (обязательно)
	RoleID   int64  `json:"role_id,omitempty"` // ID роли пользователя
	IsAdmin  int    `json:"is_admin,omitempty"` // 1 = администратор, 0 = нет
	Password string `json:"password,omitempty"` // Пароль пользователя
}

// UpdateUserRequest — запрос для update_user
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`      // Имя пользователя
	Email    string `json:"email,omitempty"`     // Email пользователя
	RoleID   int64  `json:"role_id,omitempty"`   // ID роли пользователя
	IsAdmin  int    `json:"is_admin,omitempty"`  // 1 = администратор, 0 = нет
	IsActive int    `json:"is_active,omitempty"` // 1 = активен, 0 = неактивен
}
