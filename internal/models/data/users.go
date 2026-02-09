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
