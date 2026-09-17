package site

import (
	"context"

	"github.com/google/uuid"
)

// Repository persists sites.
type Repository interface {
	Create(ctx context.Context, s *Site) error
	GetByID(ctx context.Context, id uuid.UUID) (*Site, error)
	Update(ctx context.Context, s *Site) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]Site, int, error)
}
