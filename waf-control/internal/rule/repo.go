package rule

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, r *Rule) error
	GetByID(ctx context.Context, id uuid.UUID) (*Rule, error)
	Update(ctx context.Context, r *Rule) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]Rule, int, error)
}
