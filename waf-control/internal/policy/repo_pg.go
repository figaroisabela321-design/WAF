package policy

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGRepository struct{ pool *pgxpool.Pool }

func NewPGRepository(pool *pgxpool.Pool) *PGRepository { return &PGRepository{pool: pool} }

const cols = `id,name,mode,protection_level,paranoia_level,inbound_threshold,outbound_threshold,
	sql_injection,xss,rce,lfi,scanner,protocol_attack,created_at,updated_at`

func scanPolicy(row pgx.Row) (*Policy, error) {
	var p Policy
	err := row.Scan(&p.ID, &p.Name, &p.Mode, &p.ProtectionLevel, &p.ParanoiaLevel,
		&p.InboundThreshold, &p.OutboundThreshold, &p.SQLInjection, &p.XSS, &p.RCE,
		&p.LFI, &p.Scanner, &p.ProtocolAttack, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PGRepository) Create(ctx context.Context, p *Policy) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO policies (`+cols+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		p.ID, p.Name, p.Mode, p.ProtectionLevel, p.ParanoiaLevel, p.InboundThreshold,
		p.OutboundThreshold, p.SQLInjection, p.XSS, p.RCE, p.LFI, p.Scanner,
		p.ProtocolAttack, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PGRepository) GetByID(ctx context.Context, id uuid.UUID) (*Policy, error) {
	return scanPolicy(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM policies WHERE id=$1`, id))
}

func (r *PGRepository) Update(ctx context.Context, p *Policy) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE policies SET name=$2,mode=$3,protection_level=$4,paranoia_level=$5,
			inbound_threshold=$6,outbound_threshold=$7,sql_injection=$8,xss=$9,rce=$10,
			lfi=$11,scanner=$12,protocol_attack=$13,updated_at=$14 WHERE id=$1`,
		p.ID, p.Name, p.Mode, p.ProtectionLevel, p.ParanoiaLevel, p.InboundThreshold,
		p.OutboundThreshold, p.SQLInjection, p.XSS, p.RCE, p.LFI, p.Scanner,
		p.ProtocolAttack, p.UpdatedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PGRepository) List(ctx context.Context, offset, limit int) ([]Policy, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM policies`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM policies ORDER BY created_at DESC OFFSET $1 LIMIT $2`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Name, &p.Mode, &p.ProtectionLevel, &p.ParanoiaLevel,
			&p.InboundThreshold, &p.OutboundThreshold, &p.SQLInjection, &p.XSS, &p.RCE,
			&p.LFI, &p.Scanner, &p.ProtocolAttack, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

// silence unused import if any
