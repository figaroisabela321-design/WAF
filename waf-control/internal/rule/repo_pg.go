package rule

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepository struct{ pool *pgxpool.Pool }

func NewPGRepository(pool *pgxpool.Pool) *PGRepository { return &PGRepository{pool: pool} }

func (r *PGRepository) Create(ctx context.Context, rule *Rule) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rules (id,rule_id,name,category,source,severity,enabled,scope,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		rule.ID, rule.RuleID, rule.Name, rule.Category, rule.Source, rule.Severity,
		rule.Enabled, rule.Scope, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert rule: %w", err)
	}
	return nil
}

func (r *PGRepository) GetByID(ctx context.Context, id uuid.UUID) (*Rule, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id,rule_id,name,category,source,severity,enabled,scope,created_at,updated_at
		FROM rules WHERE id=$1`, id)
	var rule Rule
	err := row.Scan(&rule.ID, &rule.RuleID, &rule.Name, &rule.Category, &rule.Source,
		&rule.Severity, &rule.Enabled, &rule.Scope, &rule.CreatedAt, &rule.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *PGRepository) Update(ctx context.Context, rule *Rule) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE rules SET rule_id=$2,name=$3,category=$4,source=$5,severity=$6,
			enabled=$7,scope=$8,updated_at=$9 WHERE id=$1`,
		rule.ID, rule.RuleID, rule.Name, rule.Category, rule.Source, rule.Severity,
		rule.Enabled, rule.Scope, rule.UpdatedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PGRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM rules WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PGRepository) List(ctx context.Context, offset, limit int) ([]Rule, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM rules`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id,rule_id,name,category,source,severity,enabled,scope,created_at,updated_at
		FROM rules ORDER BY created_at DESC OFFSET $1 LIMIT $2`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Rule
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(&rule.ID, &rule.RuleID, &rule.Name, &rule.Category, &rule.Source,
			&rule.Severity, &rule.Enabled, &rule.Scope, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, rule)
	}
	return list, total, rows.Err()
}
