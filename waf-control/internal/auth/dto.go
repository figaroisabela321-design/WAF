package auth

import "time"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresIn int          `json:"expires_in"`
	User      UserResponse `json:"user"`
}

type CreateUserRequest struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Status    string   `json:"status"`
	RoleCodes []string `json:"role_codes"`
}

type UpdateUserRequest struct {
	Password  *string  `json:"password"`
	Status    *string  `json:"status"`
	RoleCodes []string `json:"role_codes"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type PermissionResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateRoleRequest struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	PermissionCodes []string `json:"permission_codes"`
}
