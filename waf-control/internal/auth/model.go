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

// Claims carried in JWT / request context.
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
