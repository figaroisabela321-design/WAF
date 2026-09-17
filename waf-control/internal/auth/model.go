package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	ID        uuid.UUID
	Code      string
	Name      string
	CreatedAt time.Time
}

type Permission struct {
	ID   uuid.UUID
	Code string
	Name string
}

// Claims carried after JWT verification + live RBAC load.
// JWT itself only embeds identity (user_id, username, iat, exp).
// Roles/Permissions are populated from the database on each request
// and are the source of truth for authorization.
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}
