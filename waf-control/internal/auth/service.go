package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gov-waf/waf-control/internal/httpx"
)

const MinPasswordLen = 12

type Service struct {
	users     UserRepository
	roles     RoleRepository
	hasher    PasswordHasher
	tokens    TokenService
	expireHrs int
}

func NewService(users UserRepository, roles RoleRepository, hasher PasswordHasher, tokens TokenService, expireHrs int) *Service {
	return &Service{users: users, roles: roles, hasher: hasher, tokens: tokens, expireHrs: expireHrs}
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return nil, httpx.Validation("username and password are required")
	}
	u, err := s.users.GetByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		return nil, httpx.Internal("login failed", err)
	}
	if u == nil || u.Status != "enabled" {
		return nil, httpx.Unauthorized("invalid username or password")
	}
	if err := s.hasher.Compare(u.PasswordHash, req.Password); err != nil {
		return nil, httpx.Unauthorized("invalid username or password")
	}
	roleCodes, err := s.users.GetUserRoleCodes(ctx, u.ID)
	if err != nil {
		return nil, httpx.Internal("login failed", err)
	}
	// Issue identity-only JWT (no embedded permissions/roles as source of truth).
	claims := &Claims{UserID: u.ID.String(), Username: u.Username}
	token, expiresIn, err := s.tokens.Issue(claims, s.expireHrs)
	if err != nil {
		return nil, httpx.Internal("failed to issue token", err)
	}
	return &LoginResponse{
		Token: token, ExpiresIn: expiresIn,
		User: UserResponse{
			ID: u.ID.String(), Username: u.Username, Status: u.Status, Roles: roleCodes,
			CreatedAt: u.CreatedAt.UTC(), UpdatedAt: u.UpdatedAt.UTC(),
		},
	}, nil
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	if strings.TrimSpace(req.Username) == "" {
		return nil, httpx.Validation("username is required")
	}
	if len(req.Password) < MinPasswordLen {
		return nil, httpx.Validation("password must be at least 12 characters")
	}
	status := strings.ToLower(req.Status)
	if status == "" {
		status = "enabled"
	}
	if status != "enabled" && status != "disabled" {
		return nil, httpx.Validation("status must be enabled or disabled")
	}
	existing, err := s.users.GetByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		return nil, httpx.Internal("failed to check user", err)
	}
	if existing != nil {
		return nil, httpx.Conflict("username already exists")
	}
	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, httpx.Internal("failed to hash password", err)
	}
	now := time.Now().UTC()
	u := &User{
		ID: uuid.New(), Username: strings.TrimSpace(req.Username), PasswordHash: hash,
		Status: status, CreatedAt: now, UpdatedAt: now,
	}

	var roleIDs []uuid.UUID
	if len(req.RoleCodes) > 0 {
		ids, err := s.roles.GetRoleIDsByCodes(ctx, req.RoleCodes)
		if err != nil {
			return nil, httpx.Internal("failed to resolve roles", err)
		}
		if len(ids) != len(req.RoleCodes) {
			return nil, httpx.Validation("one or more role_codes are invalid")
		}
		roleIDs = ids
	}

	// Repository owns the transaction: success COMMIT; failure ROLLBACK (no leftover user).
	if err := s.users.CreateWithRoles(ctx, u, roleIDs); err != nil {
		return nil, httpx.Internal("failed to create user", err)
	}
	return s.toUserResponse(ctx, u)
}

func (s *Service) GetUser(ctx context.Context, idStr string) (*UserResponse, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid user id")
	}
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get user", err)
	}
	if u == nil {
		return nil, httpx.NotFound("user not found")
	}
	return s.toUserResponse(ctx, u)
}

func (s *Service) UpdateUser(ctx context.Context, idStr string, req *UpdateUserRequest) (*UserResponse, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid user id")
	}
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get user", err)
	}
	if u == nil {
		return nil, httpx.NotFound("user not found")
	}
	if req.Password != nil {
		if len(*req.Password) < MinPasswordLen {
			return nil, httpx.Validation("password must be at least 12 characters")
		}
		hash, err := s.hasher.Hash(*req.Password)
		if err != nil {
			return nil, httpx.Internal("failed to hash password", err)
		}
		u.PasswordHash = hash
	}
	if req.Status != nil {
		st := strings.ToLower(*req.Status)
		if st != "enabled" && st != "disabled" {
			return nil, httpx.Validation("status must be enabled or disabled")
		}
		u.Status = st
	}
	u.UpdatedAt = time.Now().UTC()

	var roleIDsPtr *[]uuid.UUID
	if req.RoleCodes != nil {
		ids, err := s.roles.GetRoleIDsByCodes(ctx, req.RoleCodes)
		if err != nil {
			return nil, httpx.Internal("failed to resolve roles", err)
		}
		if len(ids) != len(req.RoleCodes) {
			return nil, httpx.Validation("one or more role_codes are invalid")
		}
		roleIDsPtr = &ids
	}

	// Repository owns the transaction; failure leaves original data unchanged.
	if err := s.users.UpdateWithRoles(ctx, u, roleIDsPtr); err != nil {
		return nil, httpx.Internal("failed to update user", err)
	}
	return s.toUserResponse(ctx, u)
}

func (s *Service) ListUsers(ctx context.Context, offset, limit int) ([]UserResponse, int, error) {
	list, total, err := s.users.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, httpx.Internal("failed to list users", err)
	}
	out := make([]UserResponse, 0, len(list))
	for i := range list {
		ur, err := s.toUserResponse(ctx, &list[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *ur)
	}
	return out, total, nil
}

func (s *Service) ListRoles(ctx context.Context) ([]RoleResponse, error) {
	list, err := s.roles.ListRoles(ctx)
	if err != nil {
		return nil, httpx.Internal("failed to list roles", err)
	}
	out := make([]RoleResponse, 0, len(list))
	for _, r := range list {
		out = append(out, RoleResponse{ID: r.ID.String(), Code: r.Code, Name: r.Name})
	}
	return out, nil
}

func (s *Service) ListPermissions(ctx context.Context) ([]PermissionResponse, error) {
	list, err := s.roles.ListPermissions(ctx)
	if err != nil {
		return nil, httpx.Internal("failed to list permissions", err)
	}
	out := make([]PermissionResponse, 0, len(list))
	for _, p := range list {
		out = append(out, PermissionResponse{ID: p.ID.String(), Code: p.Code, Name: p.Name})
	}
	return out, nil
}

func (s *Service) CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error) {
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, httpx.Validation("code and name are required")
	}
	existing, err := s.roles.GetRoleByCode(ctx, strings.TrimSpace(req.Code))
	if err != nil {
		return nil, httpx.Internal("failed to check role", err)
	}
	if existing != nil {
		return nil, httpx.Conflict("role code already exists")
	}
	role := &Role{ID: uuid.New(), Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), CreatedAt: time.Now().UTC()}
	if err := s.roles.CreateRole(ctx, role, req.PermissionCodes); err != nil {
		return nil, httpx.Internal("failed to create role", err)
	}
	return &RoleResponse{ID: role.ID.String(), Code: role.Code, Name: role.Name}, nil
}

// SeedAdmin creates admin user from env if none exists.
// Password length is enforced by production config.Validate; development may use the documented short default.
func (s *Service) SeedAdmin(ctx context.Context, username, password string) error {
	n, err := s.users.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	existing, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		adminRole, err := s.roles.GetRoleByCode(ctx, "admin")
		if err != nil || adminRole == nil {
			return err
		}
		return s.users.SetUserRoles(ctx, existing.ID, []uuid.UUID{adminRole.ID})
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	u := &User{ID: uuid.New(), Username: username, PasswordHash: hash, Status: "enabled", CreatedAt: now, UpdatedAt: now}
	adminRole, err := s.roles.GetRoleByCode(ctx, "admin")
	if err != nil {
		return err
	}
	var roleIDs []uuid.UUID
	if adminRole != nil {
		roleIDs = []uuid.UUID{adminRole.ID}
	}
	return s.users.CreateWithRoles(ctx, u, roleIDs)
}

func (s *Service) toUserResponse(ctx context.Context, u *User) (*UserResponse, error) {
	roles, err := s.users.GetUserRoleCodes(ctx, u.ID)
	if err != nil {
		return nil, httpx.Internal("failed to get roles", err)
	}
	if roles == nil {
		roles = []string{}
	}
	return &UserResponse{
		ID: u.ID.String(), Username: u.Username, Status: u.Status, Roles: roles,
		CreatedAt: u.CreatedAt.UTC(), UpdatedAt: u.UpdatedAt.UTC(),
	}, nil
}
