package audit

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

func (r *PGRepository) Create(ctx context.Context, log *Log) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (id,actor_id,actor_name,method,path,resource,resource_id,
			status_code,ip,user_agent,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		log.ID, log.ActorID, log.ActorName, log.Method, log.Path, log.Resource,
		log.ResourceID, log.StatusCode, log.IP, log.UserAgent, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	return nil
}

func (r *PGRepository) GetByID(ctx context.Context, id uuid.UUID) (*Log, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id,actor_id,actor_name,method,path,resource,resource_id,
			status_code,ip,user_agent,created_at FROM audit_logs WHERE id=$1`, id)
	var l Log
	err := row.Scan(&l.ID, &l.ActorID, &l.ActorName, &l.Method, &l.Path, &l.Resource,
		&l.ResourceID, &l.StatusCode, &l.IP, &l.UserAgent, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &l, err
}

func (r *PGRepository) List(ctx context.Context, offset, limit int) ([]Log, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id,actor_id,actor_name,method,path,resource,resource_id,
			status_code,ip,user_agent,created_at
		FROM audit_logs ORDER BY created_at DESC OFFSET $1 LIMIT $2`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Log
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.ActorID, &l.ActorName, &l.Method, &l.Path, &l.Resource,
			&l.ResourceID, &l.StatusCode, &l.IP, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, l)
	}
	return list, total, rows.Err()
}
