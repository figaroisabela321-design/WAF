package auth

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository persists users and auth relations.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, u *User) error
	List(ctx context.Context, offset, limit int) ([]User, int, error)
	SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	GetUserRoleCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	CountAdmins(ctx context.Context) (int, error)
}

// RoleRepository persists roles and permissions.
type RoleRepository interface {
	ListRoles(ctx context.Context) ([]Role, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetRoleByCode(ctx context.Context, code string) (*Role, error)
	CreateRole(ctx context.Context, role *Role, permissionCodes []string) error
	GetRoleIDsByCodes(ctx context.Context, codes []string) ([]uuid.UUID, error)
}
