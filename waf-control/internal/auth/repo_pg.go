package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGUserRepository struct{ pool *pgxpool.Pool }
type PGRoleRepository struct{ pool *pgxpool.Pool }

func NewPGUserRepository(pool *pgxpool.Pool) *PGUserRepository {
	return &PGUserRepository{pool: pool}
}
func NewPGRoleRepository(pool *pgxpool.Pool) *PGRoleRepository {
	return &PGRoleRepository{pool: pool}
}

func (r *PGUserRepository) Create(ctx context.Context, u *User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		u.ID, u.Username, u.PasswordHash, u.Status, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *PGUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, status, created_at, updated_at FROM users WHERE id=$1`, id)
	return scanUser(row)
}

func (r *PGUserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, status, created_at, updated_at FROM users WHERE username=$1`, username)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PGUserRepository) Update(ctx context.Context, u *User) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE users SET password_hash=$2, status=$3, updated_at=$4 WHERE id=$1`,
		u.ID, u.PasswordHash, u.Status, u.UpdatedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PGUserRepository) List(ctx context.Context, offset, limit int) ([]User, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, password_hash, status, created_at, updated_at
		FROM users ORDER BY created_at DESC OFFSET $1 LIMIT $2`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, u)
	}
	return list, total, rows.Err()
}

func (r *PGUserRepository) SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1`, userID); err != nil {
		return err
	}
	for _, rid := range roleIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1,$2)`, userID, rid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PGUserRepository) GetUserRoleCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ro.code FROM roles ro
		JOIN user_roles ur ON ur.role_id = ro.id
		WHERE ur.user_id = $1 ORDER BY ro.code`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *PGUserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT p.code FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1 ORDER BY p.code`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *PGUserRepository) CountAdmins(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE r.code = 'admin'`).Scan(&n)
	return n, err
}

func (r *PGRoleRepository) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, created_at FROM roles ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, role)
	}
	return list, rows.Err()
}

func (r *PGRoleRepository) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name FROM permissions ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Code, &p.Name); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *PGRoleRepository) GetRoleByCode(ctx context.Context, code string) (*Role, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, code, name, created_at FROM roles WHERE code=$1`, code)
	var role Role
	err := row.Scan(&role.ID, &role.Code, &role.Name, &role.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *PGRoleRepository) CreateRole(ctx context.Context, role *Role, permissionCodes []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO roles (id, code, name, created_at) VALUES ($1,$2,$3,$4)`,
		role.ID, role.Code, role.Name, role.CreatedAt); err != nil {
		return err
	}
	for _, code := range permissionCodes {
		var pid uuid.UUID
		err := tx.QueryRow(ctx, `SELECT id FROM permissions WHERE code=$1`, code).Scan(&pid)
		if err != nil {
			return fmt.Errorf("permission %s: %w", code, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1,$2)`, role.ID, pid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PGRoleRepository) GetRoleIDsByCodes(ctx context.Context, codes []string) ([]uuid.UUID, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `SELECT id FROM roles WHERE code = ANY($1)`, codes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
