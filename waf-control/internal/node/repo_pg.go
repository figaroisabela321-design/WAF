package node

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

func (r *PGRepository) Create(ctx context.Context, n *Node) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO nodes (id,name,ip,region,node_group,coraza_version,crs_version,
			config_version,status,last_heartbeat,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		n.ID, n.Name, n.IP, n.Region, n.NodeGroup, n.CorazaVersion, n.CRSVersion,
		n.ConfigVersion, n.Status, n.LastHeartbeat, n.CreatedAt, n.UpdatedAt)
	return err
}

func (r *PGRepository) GetByID(ctx context.Context, id uuid.UUID) (*Node, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id,name,ip,region,node_group,coraza_version,crs_version,
			config_version,status,last_heartbeat,created_at,updated_at
		FROM nodes WHERE id=$1`, id)
	var n Node
	err := row.Scan(&n.ID, &n.Name, &n.IP, &n.Region, &n.NodeGroup, &n.CorazaVersion, &n.CRSVersion,
		&n.ConfigVersion, &n.Status, &n.LastHeartbeat, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}
	return &n, nil
}

func (r *PGRepository) Update(ctx context.Context, n *Node) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE nodes SET name=$2,ip=$3,region=$4,node_group=$5,coraza_version=$6,
			crs_version=$7,config_version=$8,status=$9,last_heartbeat=$10,updated_at=$11
		WHERE id=$1`,
		n.ID, n.Name, n.IP, n.Region, n.NodeGroup, n.CorazaVersion, n.CRSVersion,
		n.ConfigVersion, n.Status, n.LastHeartbeat, n.UpdatedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PGRepository) List(ctx context.Context, offset, limit int) ([]Node, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id,name,ip,region,node_group,coraza_version,crs_version,
			config_version,status,last_heartbeat,created_at,updated_at
		FROM nodes ORDER BY created_at DESC OFFSET $1 LIMIT $2`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Name, &n.IP, &n.Region, &n.NodeGroup, &n.CorazaVersion, &n.CRSVersion,
			&n.ConfigVersion, &n.Status, &n.LastHeartbeat, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, n)
	}
	return list, total, rows.Err()
}
