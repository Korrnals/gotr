// internal/models/data/users.go
package data

// User represents a TestRail user.
type User struct {
	ID        int64  `json:"id"`                   // Unique user ID
	Name      string `json:"name"`                 // User name
	Email     string `json:"email"`                // User email
	IsActive  bool   `json:"is_active"`            // Whether the user is active
	RoleID    int64  `json:"role_id"`              // User role ID
	Role      string `json:"role"`                 // Role name
	IsAdmin   bool   `json:"is_admin"`             // Whether the user is an administrator
	MfaAuth   int    `json:"mfa_auth"`             // MFA status
	MfaSecret string `json:"mfa_secret,omitempty"` // MFA secret (usually not returned)
}

// GetUsersResponse is the response for get_users.
type GetUsersResponse []User

// GetUserResponse is the response for get_user.
type GetUserResponse User

// AddUserRequest is the request for add_user.
type AddUserRequest struct {
	Name     string `json:"name"`               // User name (required)
	Email    string `json:"email"`              // User email (required)
	RoleID   int64  `json:"role_id,omitempty"`  // User role ID
	IsAdmin  int    `json:"is_admin,omitempty"` // 1 = administrator, 0 = no
	Password string `json:"password,omitempty"` // User password
}

// UpdateUserRequest is the request for update_user.
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`      // User name
	Email    string `json:"email,omitempty"`     // User email
	RoleID   int64  `json:"role_id,omitempty"`   // User role ID
	IsAdmin  int    `json:"is_admin,omitempty"`  // 1 = administrator, 0 = no
	IsActive int    `json:"is_active,omitempty"` // 1 = active, 0 = inactive
}
