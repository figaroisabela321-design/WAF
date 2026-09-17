package policy

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, p *Policy) error
	GetByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	Update(ctx context.Context, p *Policy) error
	List(ctx context.Context, offset, limit int) ([]Policy, int, error)
}
