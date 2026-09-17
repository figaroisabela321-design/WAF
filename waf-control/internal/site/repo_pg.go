package site

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) Create(ctx context.Context, s *Site) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sites (id, name, domain, upstream_host, upstream_port, protocol,
			protection_mode, policy_id, node_group_id, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		s.ID, s.Name, s.Domain, s.UpstreamHost, s.UpstreamPort, s.Protocol,
		s.ProtectionMode, s.PolicyID, s.NodeGroupID, s.Status, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert site: %w", err)
	}
	return nil
}

func (r *PGRepository) GetByID(ctx context.Context, id uuid.UUID) (*Site, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, domain, upstream_host, upstream_port, protocol,
			protection_mode, policy_id, node_group_id, status, created_at, updated_at
		FROM sites WHERE id = $1`, id)
	var s Site
	err := row.Scan(&s.ID, &s.Name, &s.Domain, &s.UpstreamHost, &s.UpstreamPort, &s.Protocol,
		&s.ProtectionMode, &s.PolicyID, &s.NodeGroupID, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get site: %w", err)
	}
	return &s, nil
}

func (r *PGRepository) Update(ctx context.Context, s *Site) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE sites SET name=$2, domain=$3, upstream_host=$4, upstream_port=$5,
			protocol=$6, protection_mode=$7, policy_id=$8, node_group_id=$9,
			status=$10, updated_at=$11
		WHERE id=$1`,
		s.ID, s.Name, s.Domain, s.UpstreamHost, s.UpstreamPort, s.Protocol,
		s.ProtectionMode, s.PolicyID, s.NodeGroupID, s.Status, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update site: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PGRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM sites WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete site: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PGRepository) List(ctx context.Context, offset, limit int) ([]Site, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sites`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sites: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, domain, upstream_host, upstream_port, protocol,
			protection_mode, policy_id, node_group_id, status, created_at, updated_at
		FROM sites ORDER BY created_at DESC OFFSET $1 LIMIT $2`, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list sites: %w", err)
	}
	defer rows.Close()
	var list []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.UpstreamHost, &s.UpstreamPort, &s.Protocol,
			&s.ProtectionMode, &s.PolicyID, &s.NodeGroupID, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, rows.Err()
}
