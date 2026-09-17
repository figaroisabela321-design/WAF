package audit

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, log *Log) error
	List(ctx context.Context, offset, limit int) ([]Log, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Log, error)
}
